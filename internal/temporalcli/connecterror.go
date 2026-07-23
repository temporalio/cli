package temporalcli

import (
	"fmt"
	"net"
)

// connectError is returned by dialClient when connecting to the server fails.
// It retains semantic observations for terminal normalization; Error stays a
// concise, uncolored compatibility string and Unwrap preserves the cause.
type connectError struct {
	diagnosis connectDiagnosis
	meta      connectMeta
	cause     error
}

func (e *connectError) Error() string {
	if e.meta.Address == "" {
		return "failed preparing Temporal server connection: " + connectSummary(&e.diagnosis, e.cause)
	}
	return fmt.Sprintf("failed connecting to Temporal server at %s: %s", e.meta.Address, connectSummary(&e.diagnosis, e.cause))
}

func (e *connectError) Unwrap() error { return e.cause }

// connectMeta contains the effective non-secret connection settings returned
// by the same ClientOptionsBuilder.Build call used for the failed dial.
type connectMeta struct {
	Address       string
	Namespace     string
	TLSConfigured bool
}

func newConnectError(d *connectDiagnosis, meta connectMeta, origErr error) *connectError {
	if meta.Address == "" {
		meta.Address = d.Address
	}
	diagnosis := *d
	diagnosis.Stages = append([]diagStage(nil), d.Stages...)
	return &connectError{diagnosis: diagnosis, meta: meta, cause: origErr}
}

// connectSummary is the one-line cause appended to the first error line. For
// unclassified failures and timeouts it keeps the original error text so
// scripts matching on strings like "context deadline exceeded" keep working.
func connectSummary(d *connectDiagnosis, origErr error) string {
	switch d.Cause {
	case causeDNS:
		return "could not resolve host"
	case causeTCPRefused:
		return "connection refused"
	case causeTCPTimeout:
		return "connection timed out"
	case causeServerPlaintext:
		return "TLS is enabled but the server did not respond with TLS; check tls configuration"
	case causeServerSpeaksTLS:
		return "the server requires TLS but the CLI is connecting without it"
	case causeClientCertRequired:
		return "TLS handshake failed: server requires client certificate (mTLS)"
	case causeCAVerify:
		return "cannot verify server TLS certificate"
	case causeHostnameMismatch:
		return "server TLS certificate does not match host"
	case causeCertFileUnreadable:
		return fmt.Sprintf("cannot read file %q", d.Detail)
	case causeUnauthenticated:
		return "authentication failed"
	case causePermissionDenied:
		return "permission denied"
	default:
		return shortErr(origErr)
	}
}

// suggestAction maps only directly observed failures to a typed next step.
// It never replays argv or guesses credential kind or configuration provenance.
func suggestAction(d *connectDiagnosis, meta connectMeta) *displayAction {
	host, port, _ := net.SplitHostPort(meta.Address)
	switch d.Cause {
	case causeCertFileUnreadable:
		return &displayAction{Label: fmt.Sprintf("Cannot read %q — check that the path exists and is readable.", d.Detail)}
	case causeClientCertRequired:
		return &displayAction{Label: "The server requires client certificates (mTLS). Configure both --tls-cert-path and --tls-key-path."}
	case causeUnauthenticated, causePermissionDenied:
		// client.Options keeps credentials opaque, so no credential-specific
		// action is authoritative here.
		return nil
	case causeServerPlaintext:
		return &displayAction{Label: fmt.Sprintf("The server at %s does not appear to use TLS. Remove --tls and related TLS flags, or check the address.", meta.Address)}
	case causeServerSpeaksTLS:
		return &displayAction{Label: "The server requires TLS. Retry with --tls."}
	case causeDNS:
		return &displayAction{Label: fmt.Sprintf("Could not resolve %q — check the server address.", host)}
	case causeTCPRefused:
		if isLoopbackHost(host) && port == "7233" {
			return &displayAction{
				Label:       fmt.Sprintf("No Temporal server is running at %s. Start a local dev server:", meta.Address),
				Invocations: []displayInvocation{{Command: []string{"temporal", "server", "start-dev"}}},
			}
		}
		return &displayAction{Label: fmt.Sprintf("Nothing is listening at %s — verify the address and that the server is running.", meta.Address)}
	case causeCAVerify:
		return &displayAction{Label: "The server certificate is not trusted. Configure its CA certificate with --tls-ca-path."}
	case causeHostnameMismatch:
		return &displayAction{Label: fmt.Sprintf("The server certificate is not valid for %q. Set --tls-server-name to a certificate name.", host)}
	case causeTCPTimeout:
		return &displayAction{Label: fmt.Sprintf("The TCP check timed out. Verify the address and network path to %s.", meta.Address)}
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
