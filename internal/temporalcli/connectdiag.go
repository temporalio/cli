package temporalcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"syscall"
	"time"

	"go.temporal.io/api/serviceerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// connectCause classifies why a connection to the server failed. It drives
// the suggested fix shown alongside the staged diagnosis.
type connectCause int

const (
	causeUnknown connectCause = iota
	causeDNS
	causeTCPRefused
	causeTCPTimeout
	// causeServerPlaintext means TLS was configured but the server responded
	// with a non-TLS handshake.
	causeServerPlaintext
	// causeServerSpeaksTLS means TLS was not configured but the server
	// completed a TLS handshake when offered one.
	causeServerSpeaksTLS
	causeClientCertRequired
	// causeClientCertRejected means a client certificate was presented and the
	// server rejected it, which is a different remedy from not having one.
	causeClientCertRejected
	causeCAVerify
	// causeCertInvalid means the server certificate chained to a trusted root
	// but is not usable: expired, not yet valid, or an incompatible usage.
	// Configuring a CA cannot fix these.
	causeCertInvalid
	causeHostnameMismatch
	causeCertFileUnreadable
	causeUnauthenticated
	causePermissionDenied
	causeTimeout
)

type diagStatus int

const (
	diagOK diagStatus = iota
	diagFail
	diagInconclusive
	diagSkipped
)

type diagStage struct {
	Status diagStatus
	Label  string
}

// connectTarget describes the effective connection settings behind a failed
// dial, so the diagnosis can follow the same transport path the client did.
type connectTarget struct {
	Address   string
	Authority string
	TLS       *tls.Config
	// CustomDialer reports whether the client was given gRPC dial options that
	// can replace the transport. The configured address is then not what was
	// dialed, so the probes are skipped rather than blaming it.
	CustomDialer bool
}

// connectDiagnosis is the result of probing a failed connection.
type connectDiagnosis struct {
	Address string
	Stages  []diagStage
	Cause   connectCause
	// ServerName is the name the TLS probe verified against, which is the
	// configured --tls-server-name when set, otherwise one derived from the
	// authority or the address. ServerNameConfigured distinguishes the two,
	// because the remedy differs: correct an override, or add one.
	ServerName           string
	ServerNameConfigured bool
	// Detail carries cause-specific info (e.g. an unreadable file path or the
	// raw TLS alert text) for use in stage labels and suggestions.
	Detail string
}

const (
	connectDiagnosisBudget    = 3 * time.Second
	connectDiagnosisReadProbe = 500 * time.Millisecond
)

// diagnoseConnection probes address in stages (DNS, TCP, TLS) to pinpoint why
// a dial failed, and classifies origErr for anything past the transport. It
// must only be called on an already-failed dial: it makes fresh network
// connections (including, when TLS is not configured, one anonymous TLS
// handshake to detect a TLS-only server, which may appear in server logs).
func diagnoseConnection(ctx context.Context, target connectTarget, origErr error) *connectDiagnosis {
	address, tlsCfg := target.Address, target.TLS
	d := &connectDiagnosis{Address: address, Cause: causeUnknown}
	origCause, origDetail := classifyGRPCError(origErr)
	switch origCause {
	case causeUnauthenticated:
		d.Cause, d.Detail = origCause, origDetail
		d.fail("gRPC request was unauthenticated")
		return d
	case causePermissionDenied:
		d.Cause, d.Detail = origCause, origDetail
		d.fail("gRPC request was denied")
		return d
	}
	// A custom dialer can replace the transport entirely — a Unix socket, a
	// tunnel, an in-memory pipe — so the configured address is not what was
	// dialed. Probing it would report DNS or TCP failures for an address the
	// client never used.
	if target.CustomDialer {
		d.skipped("Connection checks skipped: a custom gRPC dialer is configured")
		d.Cause, d.Detail = origCause, origDetail
		return d
	}

	ctx, cancel := context.WithTimeout(ctx, connectDiagnosisBudget)
	defer cancel()

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		d.inconclusive("Connection checks unavailable: address is not in host:port form")
		d.Cause, d.Detail = classifyGRPCError(origErr)
		return d
	}

	// Stage: DNS (skipped for IP literals)
	if !isIPLiteral(host) {
		addrs, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			if d.interrupted(ctx, "DNS") {
				return d
			}
			d.fail(fmt.Sprintf("DNS lookup for %q failed: %v", host, dnsErrShort(err)))
			d.Cause = causeDNS
			return d
		}
		if d.interrupted(ctx, "DNS") {
			return d
		}
		plural := "es"
		if len(addrs) == 1 {
			plural = ""
		}
		d.ok(fmt.Sprintf("DNS resolved (%d address%s)", len(addrs), plural))
	} else {
		d.skipped("DNS lookup skipped: address uses an IP literal")
	}

	// Stage: TCP
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		if d.interrupted(ctx, "TCP") {
			return d
		}
		if isConnRefused(err) {
			d.fail("TCP connection refused: nothing is listening at " + address)
			d.Cause = causeTCPRefused
		} else if isTimeout(err) {
			d.fail("TCP connection timed out")
			d.Cause = causeTCPTimeout
		} else {
			d.inconclusive(fmt.Sprintf("TCP check inconclusive: %v", err))
			d.Cause = causeUnknown
		}
		return d
	}
	defer conn.Close()
	d.ok("TCP connection established")

	if tlsCfg != nil {
		// Deliberately not the authority. The SDK passes WithAuthority only on
		// its insecure path, so with TLS configured gRPC derives the handshake
		// name from the TLS ServerName, or the endpoint when that is unset.
		// --client-authority never reaches this handshake.
		d.probeTLS(ctx, conn, host, tlsCfg)
		if ctx.Err() != nil {
			return d
		}
		if d.Cause != causeUnknown {
			return d
		}
	} else if cause, detail := probeServerSpeaksTLS(ctx, conn, effectiveServerName(host, target.Authority)); cause != causeUnknown {
		d.fail("server expects TLS, but the CLI is connecting without it" + detail)
		d.Cause = cause
		return d
	} else if d.interrupted(ctx, "TLS") {
		return d
	}

	// Stage: gRPC — no re-dial; classify the original error.
	d.Cause, d.Detail = origCause, origDetail
	switch d.Cause {
	case causeUnauthenticated:
		d.fail("gRPC request was unauthenticated")
	case causePermissionDenied:
		d.fail("gRPC request was denied")
	case causeTimeout:
		d.inconclusive("gRPC failure could not be diagnosed before its deadline")
	default:
		d.inconclusive("gRPC failure was not classified: " + shortErr(origErr))
	}
	return d
}

// probeTLS performs a TLS handshake over conn using the client's own config
// and classifies the failure, if any. On handshake success it briefly reads to
// catch post-handshake rejections: TLS 1.3 servers requiring client
// certificates complete the handshake first and only then send an alert.
func (d *connectDiagnosis) probeTLS(ctx context.Context, conn net.Conn, host string, tlsCfg *tls.Config) {
	cfg := tlsCfg.Clone()
	d.ServerNameConfigured = cfg.ServerName != ""
	// Always send SNI. grpc-go sets ServerName from the authority
	// unconditionally, and InsecureSkipVerify disables certificate
	// verification, not SNI. Omitting it can route an SNI-based endpoint to a
	// default virtual host and fail a handshake the real client completes.
	if cfg.ServerName == "" {
		cfg.ServerName = host
	}
	d.ServerName = cfg.ServerName
	// The real gRPC transport advertises h2. A server that requires ALPN
	// rejects a probe that offers nothing with "no application protocol",
	// which would be reported as a TLS failure that the client never had.
	if len(cfg.NextProtos) == 0 {
		cfg.NextProtos = []string{"h2"}
	}
	// Alert 42 is ambiguous without this: a TLS 1.2 server sends it both for a
	// rejected certificate and for none at all.
	clientCertConfigured := len(cfg.Certificates) > 0 || cfg.GetClientCertificate != nil

	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		if d.interrupted(ctx, "TLS handshake") {
			return
		}
		d.classifyTLSError(err, clientCertConfigured)
		return
	}
	// Handshake OK; a short read distinguishes an mTLS rejection alert from a
	// genuinely healthy connection (where the read just times out).
	stopCancel := armReadDeadline(ctx, tlsConn)
	defer stopCancel()
	buf := make([]byte, 1)
	_, err := tlsConn.Read(buf)
	if ctx.Err() != nil {
		d.interrupted(ctx, "TLS post-handshake")
		return
	}
	if err != nil && !isTimeout(err) {
		if cause := classifyTLSAlert(err); cause != causeUnknown {
			d.classifyTLSError(err, clientCertConfigured)
			return
		}
	}
	d.ok("TLS handshake succeeded")
}

func armReadDeadline(ctx context.Context, conn net.Conn) func() bool {
	readDeadline := time.Now().Add(connectDiagnosisReadProbe)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(readDeadline) {
		readDeadline = contextDeadline
	}
	_ = conn.SetReadDeadline(readDeadline)
	return context.AfterFunc(ctx, func() { _ = conn.SetReadDeadline(time.Now()) })
}

func (d *connectDiagnosis) classifyTLSError(err error, clientCertConfigured bool) {
	var recordErr tls.RecordHeaderError
	var certErr *tls.CertificateVerificationError
	var unknownAuthErr x509.UnknownAuthorityError
	var hostnameErr x509.HostnameError
	var invalidErr x509.CertificateInvalidError
	alert := classifyTLSAlert(err)
	// Only a client that presented a certificate can have had one rejected.
	// A TLS 1.2 server sends alert 42 for an empty certificate message too,
	// where alert 116 is TLS 1.3 only, so the alert alone cannot separate
	// absence from rejection.
	if alert == causeClientCertRejected && !clientCertConfigured {
		alert = causeClientCertRequired
	}
	switch {
	case errors.As(err, &recordErr):
		d.fail("TLS handshake failed: server did not respond with TLS (it may be a plaintext gRPC endpoint)")
		d.Cause = causeServerPlaintext
	case errors.As(err, &hostnameErr):
		d.fail("TLS handshake failed: server certificate is not valid for this host: " + shortErr(err))
		d.Cause = causeHostnameMismatch
	// Checked before the generic verification error below: Go wraps
	// x509.CertificateInvalidError (expired, not yet valid, incompatible
	// usage) in *tls.CertificateVerificationError, and those failures are not
	// fixed by configuring a CA.
	case errors.As(err, &invalidErr):
		d.fail("TLS handshake failed: server certificate is not usable: " + shortErr(err))
		d.Cause = causeCertInvalid
		d.Detail = certInvalidReason(invalidErr)
	case errors.As(err, &unknownAuthErr), errors.As(err, &certErr):
		d.fail("TLS handshake failed: cannot verify server certificate: " + shortErr(err))
		d.Cause = causeCAVerify
	case alert == causeClientCertRequired:
		d.fail("TLS handshake failed: server requires mTLS, no valid client certificate was provided")
		d.Cause = causeClientCertRequired
	case alert == causeClientCertRejected:
		d.fail("TLS handshake failed: server rejected the client certificate")
		d.Cause = causeClientCertRejected
	default:
		d.fail("TLS handshake failed: " + shortErr(err))
		d.Cause = causeUnknown
		d.Detail = err.Error()
	}
}

// certInvalidReason names why a trusted certificate is unusable, so the
// suggestion can be specific about what to fix.
func certInvalidReason(err x509.CertificateInvalidError) string {
	switch err.Reason {
	case x509.Expired:
		return "expired or not yet valid"
	case x509.IncompatibleUsage:
		return "incompatible key usage"
	case x509.NotAuthorizedToSign, x509.TooManyIntermediates, x509.CANotAuthorizedForThisName:
		return "invalid certificate chain"
	}
	return ""
}

// classifyTLSAlert detects remote TLS alerts that explicitly indicate the
// server required or rejected a client certificate. Go does not export alert
// types for remote errors, so this matches the alert descriptions crypto/tls
// emits: alert 116 "certificate required" (TLS 1.3) and alert 42 "bad
// certificate". The two are kept apart because they call for different
// remedies: the first means no certificate was sent, the second means one was
// sent and refused. Generic alerts such as alert 40 "handshake failure" are
// not sufficient evidence of mTLS. Matching is scoped to TLS probe errors
// only; if these strings drift in a future Go release, unit tests pin them and
// the diagnosis degrades to showing the raw error without a suggestion.
func classifyTLSAlert(err error) connectCause {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "certificate required"):
		return causeClientCertRequired
	case strings.Contains(msg, "bad certificate"):
		return causeClientCertRejected
	}
	return causeUnknown
}

// probeServerSpeaksTLS opportunistically offers a TLS handshake to a server
// the CLI is configured to reach in plaintext. If the server negotiates TLS
// (or rejects us at the certificate step, which still means it spoke TLS), the
// mismatch is the likely root cause.
func probeServerSpeaksTLS(ctx context.Context, conn net.Conn, serverName string) (connectCause, string) {
	tlsConn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
		// Match the real gRPC transport; see probeTLS.
		NextProtos: []string{"h2"},
	})
	err := tlsConn.HandshakeContext(ctx)
	if err == nil {
		return causeServerSpeaksTLS, ""
	}
	// Either certificate alert still proves the server spoke TLS. This probe
	// never sends a client certificate, so both mean the server wants one.
	if classifyTLSAlert(err) != causeUnknown {
		return causeServerSpeaksTLS, " (and it requires client certificates)"
	}
	return causeUnknown, ""
}

// isIPLiteral reports whether host is an IP address rather than a name to
// resolve. net.ParseIP rejects a zone, so a scoped IPv6 literal such as
// "fe80::1%eth0" would otherwise be sent to the resolver and reported as a
// DNS failure before its valid TCP endpoint was ever tried.
func isIPLiteral(host string) bool {
	_, err := netip.ParseAddr(host)
	return err == nil
}

// effectiveServerName returns the SNI name for the opportunistic probe against
// a server the CLI is configured to reach in plaintext. The SDK passes
// WithAuthority on that path, so --client-authority is the authority gRPC
// would use. It deliberately does not apply to the configured-TLS probe, where
// the SDK never forwards the authority.
func effectiveServerName(host, authority string) string {
	if authority == "" {
		return host
	}
	if h, _, err := net.SplitHostPort(authority); err == nil {
		return h
	}
	return authority
}

func classifyGRPCError(err error) (connectCause, string) {
	var deadlineErr *serviceerror.DeadlineExceeded
	if errors.As(err, &deadlineErr) || errors.Is(err, context.DeadlineExceeded) {
		return causeTimeout, ""
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return causeUnauthenticated, ""
		case codes.PermissionDenied:
			return causePermissionDenied, ""
		case codes.DeadlineExceeded:
			return causeTimeout, ""
		}
	}
	return causeUnknown, ""
}

func (d *connectDiagnosis) ok(label string) { d.Stages = append(d.Stages, diagStage{diagOK, label}) }
func (d *connectDiagnosis) fail(label string) {
	d.Stages = append(d.Stages, diagStage{diagFail, label})
}
func (d *connectDiagnosis) inconclusive(label string) {
	d.Stages = append(d.Stages, diagStage{diagInconclusive, label})
}
func (d *connectDiagnosis) skipped(label string) {
	d.Stages = append(d.Stages, diagStage{diagSkipped, label})
}

func (d *connectDiagnosis) interrupted(ctx context.Context, stage string) bool {
	if ctx.Err() == nil {
		return false
	}
	d.Cause = causeUnknown
	d.Detail = ""
	if errors.Is(ctx.Err(), context.Canceled) {
		d.skipped(stage + " check skipped: diagnosis canceled")
	} else {
		d.inconclusive(stage + " check inconclusive: diagnosis deadline reached")
	}
	return true
}

// isConnRefused reports whether err is a refused TCP connection. errors.Is
// against syscall.ECONNREFUSED matches on unix, but Windows dials fail with
// WSAECONNREFUSED ("No connection could be made because the target machine
// actively refused it"), which Go does not map to syscall.ECONNREFUSED.
func isConnRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED) ||
		strings.Contains(err.Error(), "refused")
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) ||
		(errors.As(err, &netErr) && netErr.Timeout())
}

// shortErr renders an error compactly for a stage line, trimming the SDK's
// "failed reaching server:" prefix that would be redundant inside a
// connection diagnosis.
func shortErr(err error) string {
	return strings.TrimPrefix(err.Error(), "failed reaching server: ")
}

func dnsErrShort(err error) string {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return "no such host"
	}
	return err.Error()
}
