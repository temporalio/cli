package temporalcli

import (
	"fmt"
	"net"
	"strings"

	"github.com/fatih/color"
)

// connectError is returned by dialClient when connecting to the server fails.
// Error() renders the full multi-line diagnosis (summary, probe stages, and a
// suggested fix); Unwrap() preserves the original dial error for
// errors.Is/As.
type connectError struct {
	rendered string
	cause    error
}

func (e *connectError) Error() string { return e.rendered }
func (e *connectError) Unwrap() error { return e.cause }

// connectMeta carries the request context the suggestion engine needs to
// propose a concrete fix (e.g. reconstructing the exact command the user
// typed with the missing flags appended).
type connectMeta struct {
	// Args are the CLI args as typed (without the binary name).
	Args          []string
	Address       string
	AddressSource string // "flag", "profile", or "default"
	ProfileName   string
	HasAPIKey     bool
	TLSConfigured bool
}

// newConnectError renders a connection failure into a connectError. It must
// be called while command execution is active so that color state (JSON mode,
// --color) is applied correctly; the rendering is captured eagerly.
func newConnectError(d *connectDiagnosis, meta connectMeta, origErr error) *connectError {
	var b strings.Builder
	b.WriteString("failed connecting to Temporal server at ")
	b.WriteString(meta.Address)
	b.WriteString(": ")
	b.WriteString(connectSummary(d, origErr))
	if len(d.Stages) > 0 {
		b.WriteString("\n\n  Connecting to ")
		b.WriteString(meta.Address)
		for _, stage := range d.Stages {
			if stage.Status == diagOK {
				b.WriteString("\n    " + color.GreenString("✓") + " " + stage.Label)
			} else {
				b.WriteString("\n    " + color.RedString("✗") + " " + stage.Label)
			}
		}
	}
	if suggestion := suggestFix(d, meta); suggestion != "" {
		b.WriteString("\n\n")
		b.WriteString(indentLines(suggestion, "  "))
	}
	return &connectError{rendered: b.String(), cause: origErr}
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
		return "TLS is enabled but the server did not respond with TLS"
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
		if d.Detail != "" {
			return "authentication failed: " + d.Detail
		}
		return "authentication failed"
	default:
		return shortErr(origErr)
	}
}

// suggestFix maps a classified failure to one concrete next step. First match
// wins; an empty return means no suggestion (the stages still render).
func suggestFix(d *connectDiagnosis, meta connectMeta) string {
	host, port, _ := net.SplitHostPort(meta.Address)
	switch d.Cause {
	case causeCertFileUnreadable:
		return fmt.Sprintf("Cannot read %q — check that the path exists and is readable.", d.Detail)
	case causeClientCertRequired:
		var b strings.Builder
		if isCloudHost(host) {
			b.WriteString("This looks like a Temporal Cloud endpoint secured with mTLS. Provide client certificates:")
		} else {
			b.WriteString("The server requires client certificates (mTLS). Provide them:")
		}
		b.WriteString("\n\n")
		b.WriteString(indentLines(reconstructCommand(meta.Args, "--tls-cert-path YourCert.pem", "--tls-key-path YourKey.pem"), "  "))
		b.WriteString("\n\nOr configure once:\n\n")
		b.WriteString("  temporal config set --prop tls.client_cert_path YourCert.pem\n")
		b.WriteString("  temporal config set --prop tls.client_key_path YourKey.pem")
		if strings.HasSuffix(host, ".api.temporal.io") {
			b.WriteString("\n\nIf your namespace uses API-key auth instead, pass --api-key YourApiKey.")
		}
		return b.String()
	case causeAuth:
		if meta.HasAPIKey {
			return "The server rejected the provided API key. Verify the key is valid and that the address is your namespace's gRPC endpoint (for Temporal Cloud API keys, a regional endpoint like us-west-2.aws.api.temporal.io:7233)."
		}
		return "The server rejected the request as unauthenticated. If it requires an API key, pass --api-key; if it requires mTLS, pass --tls-cert-path and --tls-key-path."
	case causeServerPlaintext:
		return fmt.Sprintf("The server at %s does not appear to use TLS. Remove --tls and certificate flags, or double-check the address and port.", meta.Address)
	case causeServerSpeaksTLS:
		var b strings.Builder
		b.WriteString("The server requires TLS. Add --tls:")
		b.WriteString("\n\n")
		b.WriteString(indentLines(reconstructCommand(meta.Args, "--tls"), "  "))
		if isCloudHost(host) {
			b.WriteString("\n\nTemporal Cloud endpoints also need client credentials: --tls-cert-path/--tls-key-path or --api-key.")
		}
		return b.String()
	case causeDNS:
		s := fmt.Sprintf("Could not resolve %q — check the server address.", host)
		if meta.AddressSource == "profile" {
			profile := meta.ProfileName
			if profile == "" {
				profile = "default"
			}
			s += fmt.Sprintf("\nThe address comes from config profile %q — inspect it with:\n\n  temporal config get --prop address", profile)
		}
		return s
	case causeTCPRefused:
		if isLoopbackHost(host) && port == "7233" {
			return fmt.Sprintf("No Temporal server is running at %s. Start a local dev server with:\n\n  temporal server start-dev", meta.Address)
		}
		return fmt.Sprintf("Nothing is listening at %s — verify the address and that the server is running.", meta.Address)
	case causeCAVerify:
		return "The server's TLS certificate is not trusted by your system roots. If the server uses a private CA, pass --tls-ca-path YourServerCA.pem."
	case causeHostnameMismatch:
		return fmt.Sprintf("The server's TLS certificate is not valid for %q. If you connect via an IP or alternate name, set --tls-server-name to the name in the certificate.", host)
	case causeTCPTimeout, causeTimeout:
		return fmt.Sprintf("The connection stalled — a firewall or proxy may be blocking traffic to %s. Verify the address, port, and network path. Bound the wait with --client-connect-timeout.", meta.Address)
	}
	return ""
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

// reconstructCommand re-renders the command the user typed, one option per
// line (matching the style used in help text), with extraFlags appended.
func reconstructCommand(args []string, extraFlags ...string) string {
	// Split leading command words from options.
	var words, opts []string
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			opts = args[i:]
			break
		}
		words = append(words, arg)
	}
	var lines []string
	lines = append(lines, "temporal "+strings.Join(words, " "))
	for i := 0; i < len(opts); i++ {
		line := opts[i]
		// Attach the option's value, if the next arg isn't another option.
		if !strings.Contains(line, "=") && i+1 < len(opts) && !strings.HasPrefix(opts[i+1], "-") {
			i++
			line += " " + quoteArg(opts[i])
		}
		lines = append(lines, "    "+line)
	}
	for _, flag := range extraFlags {
		lines = append(lines, "    "+flag)
	}
	return strings.Join(lines, " \\\n")
}

func quoteArg(arg string) string {
	if strings.ContainsAny(arg, " \t\"'") {
		return fmt.Sprintf("%q", arg)
	}
	return arg
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
