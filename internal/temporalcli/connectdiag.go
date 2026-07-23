package temporalcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
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
	causeCAVerify
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

// connectDiagnosis is the result of probing a failed connection.
type connectDiagnosis struct {
	Address string
	Stages  []diagStage
	Cause   connectCause
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
func diagnoseConnection(ctx context.Context, address string, tlsCfg *tls.Config, origErr error) *connectDiagnosis {
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
	ctx, cancel := context.WithTimeout(ctx, connectDiagnosisBudget)
	defer cancel()

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		d.inconclusive("Connection checks unavailable: address is not in host:port form")
		d.Cause, d.Detail = classifyGRPCError(origErr)
		return d
	}

	// Stage: DNS (skipped for IP literals)
	if net.ParseIP(host) == nil {
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
		d.probeTLS(ctx, conn, host, tlsCfg)
		if ctx.Err() != nil {
			return d
		}
		if d.Cause != causeUnknown {
			return d
		}
	} else if cause, detail := probeServerSpeaksTLS(ctx, conn, host); cause != causeUnknown {
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
	if cfg.ServerName == "" && !cfg.InsecureSkipVerify {
		cfg.ServerName = host
	}
	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		if d.interrupted(ctx, "TLS handshake") {
			return
		}
		d.classifyTLSError(err)
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
			d.classifyTLSError(err)
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

func (d *connectDiagnosis) classifyTLSError(err error) {
	var recordErr tls.RecordHeaderError
	var certErr *tls.CertificateVerificationError
	var unknownAuthErr x509.UnknownAuthorityError
	var hostnameErr x509.HostnameError
	switch {
	case errors.As(err, &recordErr):
		d.fail("TLS handshake failed: server did not respond with TLS (it may be a plaintext gRPC endpoint)")
		d.Cause = causeServerPlaintext
	case errors.As(err, &hostnameErr):
		d.fail("TLS handshake failed: server certificate is not valid for this host: " + shortErr(err))
		d.Cause = causeHostnameMismatch
	case errors.As(err, &unknownAuthErr), errors.As(err, &certErr):
		d.fail("TLS handshake failed: cannot verify server certificate: " + shortErr(err))
		d.Cause = causeCAVerify
	case classifyTLSAlert(err) == causeClientCertRequired:
		d.fail("TLS handshake failed: server requires mTLS, no valid client certificate was provided")
		d.Cause = causeClientCertRequired
	default:
		d.fail("TLS handshake failed: " + shortErr(err))
		d.Cause = causeUnknown
		d.Detail = err.Error()
	}
}

// classifyTLSAlert detects remote TLS alerts that explicitly indicate the
// server required or rejected a client certificate. Go does not export alert
// types for remote errors, so this matches the alert descriptions crypto/tls
// emits: alert 116 "certificate required" (TLS 1.3) and alert 42 "bad
// certificate". Generic alerts such as alert 40 "handshake failure" are not
// sufficient evidence of mTLS. Matching is scoped to TLS probe errors only;
// if these strings drift in a future Go release, unit tests pin them and the
// diagnosis degrades to showing the raw error without a suggestion.
func classifyTLSAlert(err error) connectCause {
	msg := err.Error()
	if strings.Contains(msg, "certificate required") ||
		strings.Contains(msg, "bad certificate") {
		return causeClientCertRequired
	}
	return causeUnknown
}

// probeServerSpeaksTLS opportunistically offers a TLS handshake to a server
// the CLI is configured to reach in plaintext. If the server negotiates TLS
// (or rejects us at the certificate step, which still means it spoke TLS), the
// mismatch is the likely root cause.
func probeServerSpeaksTLS(ctx context.Context, conn net.Conn, host string) (connectCause, string) {
	tlsConn := tls.Client(conn, &tls.Config{InsecureSkipVerify: true, ServerName: host})
	err := tlsConn.HandshakeContext(ctx)
	if err == nil {
		return causeServerSpeaksTLS, ""
	}
	if classifyTLSAlert(err) == causeClientCertRequired {
		return causeServerSpeaksTLS, " (and it requires client certificates)"
	}
	return causeUnknown, ""
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
