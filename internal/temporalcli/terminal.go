package temporalcli

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"unicode"
)

const defaultFailureExitStatus = 1

// Result describes the terminal outcome without exiting the process. The
// original command error remains separate from any failure writing stderr.
type Result struct {
	CommandErr      error
	PresentationErr error
	ExitStatus      int
}

type checkOutcome int

const (
	checkSucceeded checkOutcome = iota
	checkFailed
)

type safeField struct {
	Label string
	Value string
}

type errorCheck struct {
	Outcome checkOutcome
	Message string
}

type displayInvocation struct {
	Command []string
	Args    []string
}

type displayAction struct {
	Label       string
	Invocations []displayInvocation
}

type errorReport struct {
	Summary      string
	Context      []safeField
	CheckHeading string
	Checks       []errorCheck
	Action       *displayAction
	Usage        string
}

type renderOptions struct {
	Color bool
	Shell displayShell
}

type displayShell int

const (
	displayShellPOSIX displayShell = iota
	displayShellPowerShell
)

type terminalOptions struct {
	Stderr       io.Writer
	Color        bool
	KnownSecrets []string
	Usage        string
	DisplayErr   error
}

type commandErrorRecorder struct {
	err error
}

func (r *commandErrorRecorder) Record(err error) {
	if r.err == nil {
		r.err = err
	}
	panic(recordedCommandError{})
}

func (r *commandErrorRecorder) Err() error { return r.err }

type recordedCommandError struct{}

func runRecordedCommand(run func()) (panicked bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(recordedCommandError); !ok {
				panic(recovered)
			}
			panicked = true
		}
	}()
	run()
	return false
}

func handleTerminalError(commandErr error, options terminalOptions) Result {
	result := Result{CommandErr: commandErr, ExitStatus: defaultFailureExitStatus}
	displayErr := options.DisplayErr
	if displayErr == nil {
		displayErr = commandErr
	}
	report := normalizeError(displayErr)
	report.Usage = options.Usage
	report = redactReport(report, options.KnownSecrets)
	shell := displayShellPOSIX
	if runtime.GOOS == "windows" {
		shell = displayShellPowerShell
	}
	rendered := renderErrorText(report, renderOptions{Color: options.Color, Shell: shell})
	n, err := options.Stderr.Write(rendered)
	if err == nil && n != len(rendered) {
		err = io.ErrShortWrite
	}
	result.PresentationErr = err
	return result
}

func normalizeError(err error) errorReport {
	var connectionErr *connectError
	if errors.As(err, &connectionErr) {
		return connectionErr.report()
	}
	var activityErr *activityNotFoundError
	if errors.As(err, &activityErr) {
		return activityErr.report()
	}
	if err == nil || err.Error() == "" {
		return errorReport{Summary: "unknown error"}
	}
	return errorReport{Summary: err.Error()}
}

func redactReport(report errorReport, secrets []string) errorReport {
	// This is defense in depth for exact values already known to the runtime,
	// not a claim that arbitrary legacy error prose is generically sanitized.
	// Structured adapters remain responsible for admitting only safe fields.
	redact := func(value string) string {
		for _, secret := range secrets {
			if secret != "" {
				value = strings.ReplaceAll(value, secret, "[REDACTED]")
			}
		}
		return value
	}
	report.Summary = redact(report.Summary)
	report.Usage = redact(report.Usage)
	for i := range report.Context {
		report.Context[i].Value = redact(report.Context[i].Value)
	}
	for i := range report.Checks {
		report.Checks[i].Message = redact(report.Checks[i].Message)
	}
	if report.Action != nil {
		report.Action.Label = redact(report.Action.Label)
		for invocationIndex := range report.Action.Invocations {
			invocation := &report.Action.Invocations[invocationIndex]
			for i := range invocation.Command {
				invocation.Command[i] = redact(invocation.Command[i])
			}
			for i := range invocation.Args {
				invocation.Args[i] = redact(invocation.Args[i])
			}
		}
	}
	return report
}

// renderErrorText is total over errorReport and performs no I/O or global
// color lookup. Reports are redacted before reaching this function.
func renderErrorText(report errorReport, options renderOptions) []byte {
	if report.Summary == "" {
		report.Summary = "unknown error"
	}
	var b strings.Builder
	b.WriteString("Error: ")
	b.WriteString(report.Summary)
	b.WriteByte('\n')
	for _, field := range report.Context {
		fmt.Fprintf(&b, "  %s: %s\n", field.Label, field.Value)
	}
	if len(report.Checks) > 0 {
		heading := report.CheckHeading
		if heading == "" {
			heading = "Connecting"
		}
		fmt.Fprintf(&b, "\n  %s\n", heading)
		for _, check := range report.Checks {
			symbol := "✓"
			colorCode := "32"
			if check.Outcome == checkFailed {
				symbol = "✗"
				colorCode = "31"
			}
			if options.Color {
				symbol = "\x1b[" + colorCode + "m" + symbol + "\x1b[0m"
			}
			fmt.Fprintf(&b, "    %s %s\n", symbol, check.Message)
		}
	}
	if report.Action != nil {
		b.WriteByte('\n')
		if report.Action.Label != "" {
			b.WriteString(indentLines(report.Action.Label, "  "))
			b.WriteByte('\n')
		}
		for _, invocation := range report.Action.Invocations {
			rendered, ok := renderInvocation(invocation, options.Shell)
			if !ok {
				continue
			}
			b.WriteByte('\n')
			b.WriteString("    ")
			b.WriteString(rendered)
			b.WriteByte('\n')
		}
	}
	if report.Usage != "" {
		b.WriteByte('\n')
		b.WriteString(strings.TrimLeft(report.Usage, "\n"))
		if !strings.HasSuffix(report.Usage, "\n") {
			b.WriteByte('\n')
		}
	}
	return []byte(b.String())
}

func renderInvocation(invocation displayInvocation, shell displayShell) (string, bool) {
	parts := append([]string(nil), invocation.Command...)
	parts = append(parts, invocation.Args...)
	if len(parts) == 0 {
		return "", false
	}
	for i := range parts {
		if containsControl(parts[i]) {
			return "", false
		}
		if shell == displayShellPowerShell {
			parts[i] = quotePowerShell(parts[i])
		} else {
			parts[i] = quotePOSIX(parts[i])
		}
	}
	return strings.Join(parts, " "), true
}

func containsControl(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}

func quotePOSIX(value string) string {
	if value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_@%+=:,./-", r))
	}) < 0 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func quotePowerShell(value string) string {
	if value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_@%+=:,./-", r))
	}) < 0 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
