package temporalcli

import (
	"fmt"
	"net"
	"strings"
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
	return fmt.Sprintf("failed connecting to Temporal server at %s: %s", e.meta.Address, connectSummary(&e.diagnosis, e.cause))
}
func (e *connectError) Unwrap() error { return e.cause }

// connectMeta contains only allowlisted connection provenance. Raw argv and
// credential values must never enter this value.
type connectMeta struct {
	// CommandPath is retained only as non-sensitive provenance for callers that
	// still populate it. Suggestions never replay the current command.
	CommandPath   []string
	Address       string
	AddressSource string // "flag", "profile", or "default"
	ProfileName   string
	HasAPIKey     bool
	HasOAuth      bool
	TLSConfigured bool
}

func newConnectError(d *connectDiagnosis, meta connectMeta, origErr error) *connectError {
	if meta.Address == "" {
		meta.Address = d.Address
	}
	return &connectError{diagnosis: *d, meta: meta, cause: origErr}
}

func (e *connectError) report() errorReport {
	report := errorReport{
		Summary:      e.Error(),
		CheckHeading: "Connecting to " + e.meta.Address,
		Action:       suggestAction(&e.diagnosis, e.meta),
	}
	for _, stage := range e.diagnosis.Stages {
		outcome := checkFailed
		if stage.Status == diagOK {
			outcome = checkSucceeded
		}
		report.Checks = append(report.Checks, errorCheck{Outcome: outcome, Message: stage.Label})
	}
	return report
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
	case causeAuth:
		return "authentication failed"
	default:
		return shortErr(origErr)
	}
}

// suggestAction maps a classified failure to one typed next step. Invocation
// arguments are known literals plus safe command metadata, never raw argv.
func suggestAction(d *connectDiagnosis, meta connectMeta) *displayAction {
	host, port, _ := net.SplitHostPort(meta.Address)
	switch d.Cause {
	case causeCertFileUnreadable:
		return &displayAction{Label: fmt.Sprintf("Cannot read %q — check that the path exists and is readable.", d.Detail)}
	case causeClientCertRequired:
		message := "The server requires client certificates (mTLS). Configure both certificate paths:"
		if isCloudHost(host) {
			message = "This looks like a Temporal Cloud endpoint secured with mTLS. Provide client certificates:"
		}
		if strings.HasSuffix(host, ".api.temporal.io") {
			message += " If the namespace uses API-key auth instead, pass --api-key."
		}
		return configAction(message, meta,
			[]string{"--prop", "tls.client_cert_path", "--value", "YourCert.pem"},
			[]string{"--prop", "tls.client_key_path", "--value", "YourKey.pem"})
	case causeAuth:
		if meta.HasAPIKey && meta.HasOAuth {
			return &displayAction{Label: "The server rejected the configured API key or OAuth credentials. Verify the active credential and namespace gRPC endpoint."}
		}
		if meta.HasAPIKey {
			return &displayAction{Label: "The server rejected the configured API key. Verify it and the namespace gRPC endpoint."}
		}
		if meta.HasOAuth {
			return &displayAction{Label: "The server rejected the configured OAuth credentials. Verify them and the namespace gRPC endpoint."}
		}
		return &displayAction{Label: "The server rejected the request as unauthenticated. Configure an API key or mTLS credentials."}
	case causeServerPlaintext:
		return &displayAction{Label: fmt.Sprintf("The server at %s does not appear to use TLS. Remove TLS settings or check the address.", meta.Address)}
	case causeServerSpeaksTLS:
		message := "The server requires TLS. Add --tls:"
		if isCloudHost(host) {
			message += " Temporal Cloud also requires client credentials."
		}
		return configAction(message, meta, []string{"--prop", "tls", "--value", "true"})
	case causeDNS:
		s := fmt.Sprintf("Could not resolve %q — check the server address.", host)
		if meta.AddressSource == "profile" {
			profile := meta.ProfileName
			if profile == "" {
				profile = "default"
			}
			s += fmt.Sprintf(" The address comes from config profile %q.", profile)
			return &displayAction{Label: s, Invocations: []displayInvocation{{Command: []string{"temporal", "config", "get"}, Args: appendProfile([]string{"--prop", "address"}, profile)}}}
		}
		return &displayAction{Label: s}
	case causeTCPRefused:
		if isLoopbackHost(host) && port == "7233" {
			return &displayAction{Label: fmt.Sprintf("No Temporal server is running at %s. Start a local dev server:", meta.Address), Invocations: []displayInvocation{{Command: []string{"temporal", "server", "start-dev"}}}}
		}
		return &displayAction{Label: fmt.Sprintf("Nothing is listening at %s — verify the address and that the server is running.", meta.Address)}
	case causeCAVerify:
		return configAction("The server certificate is not trusted. Configure its private CA:", meta, []string{"--prop", "tls.server_ca_cert_path", "--value", "YourServerCA.pem"})
	case causeHostnameMismatch:
		return &displayAction{Label: fmt.Sprintf("The server certificate is not valid for %q. Set --tls-server-name to a certificate name.", host)}
	case causeTCPTimeout, causeTimeout:
		return &displayAction{Label: fmt.Sprintf("The connection stalled. Verify the address and network path to %s.", meta.Address)}
	}
	return nil
}

func configAction(label string, meta connectMeta, propertyArgs ...[]string) *displayAction {
	action := &displayAction{Label: label}
	for _, args := range propertyArgs {
		action.Invocations = append(action.Invocations, displayInvocation{
			Command: []string{"temporal", "config", "set"},
			Args:    appendProfile(args, meta.ProfileName),
		})
	}
	return action
}

func appendProfile(args []string, profile string) []string {
	result := append([]string(nil), args...)
	if profile != "" {
		result = append(result, "--profile", profile)
	}
	return result
}

func isCloudHost(host string) bool {
	return strings.HasSuffix(host, ".tmprl.cloud") || strings.HasSuffix(host, ".api.temporal.io")
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func indentLines(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}
