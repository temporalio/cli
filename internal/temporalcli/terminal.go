package temporalcli

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"unicode"
)

type checkOutcome int

const (
	checkSucceeded checkOutcome = iota
	checkFailed
	checkInconclusive
	checkSkipped
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

type connectionReport struct {
	Summary      string
	Context      []safeField
	CheckHeading string
	Checks       []errorCheck
	Action       *displayAction
}

type displayShell int

const (
	displayShellPOSIX displayShell = iota
	displayShellPowerShell
)

// writeConnectionError renders err only when it contains a connectError. It
// returns whether the error was handled, leaving all generic failures on the
// CLI's longstanding fallback path.
func writeConnectionError(stderr io.Writer, err error, useColor bool) bool {
	var connectionErr *connectError
	if !errors.As(err, &connectionErr) {
		return false
	}

	shell := displayShellPOSIX
	if runtime.GOOS == "windows" {
		shell = displayShellPowerShell
	}
	_, _ = stderr.Write(renderConnectionReport(connectionErrorReport(connectionErr), useColor, shell))
	return true
}

func connectionErrorReport(err *connectError) connectionReport {
	report := connectionReport{
		Summary: err.Error(),
		Action:  suggestAction(&err.diagnosis, err.meta),
	}
	if err.meta.Namespace != "" {
		report.Context = append(report.Context, safeField{Label: "Namespace", Value: err.meta.Namespace})
	}
	if err.meta.Address != "" {
		report.Context = append(report.Context, safeField{Label: "TLS", Value: fmt.Sprintf("%t", err.meta.TLSConfigured)})
	}
	if len(err.diagnosis.Stages) > 0 {
		report.CheckHeading = "Connection checks"
		if err.meta.Address != "" {
			report.CheckHeading += " for " + err.meta.Address
		}
	}
	for _, stage := range err.diagnosis.Stages {
		outcome := checkInconclusive
		switch stage.Status {
		case diagOK:
			outcome = checkSucceeded
		case diagFail:
			outcome = checkFailed
		case diagSkipped:
			outcome = checkSkipped
		}
		report.Checks = append(report.Checks, errorCheck{Outcome: outcome, Message: stage.Label})
	}
	return report
}

func renderConnectionReport(report connectionReport, useColor bool, shell displayShell) []byte {
	var b strings.Builder
	b.WriteString("Error: ")
	b.WriteString(escapeTerminalControls(report.Summary))
	b.WriteByte('\n')
	for _, field := range report.Context {
		fmt.Fprintf(
			&b,
			"  %s: %s\n",
			escapeTerminalControls(field.Label),
			escapeTerminalControls(field.Value),
		)
	}
	if len(report.Checks) > 0 {
		heading := report.CheckHeading
		if heading == "" {
			heading = "Connecting"
		}
		fmt.Fprintf(&b, "\n  %s\n", escapeTerminalControls(heading))
		for _, check := range report.Checks {
			symbol := "✓"
			colorCode := "32"
			switch check.Outcome {
			case checkFailed:
				symbol = "✗"
				colorCode = "31"
			case checkInconclusive:
				symbol = "?"
				colorCode = "33"
			case checkSkipped:
				symbol = "-"
				colorCode = "33"
			}
			if useColor {
				symbol = "\x1b[" + colorCode + "m" + symbol + "\x1b[0m"
			}
			fmt.Fprintf(&b, "    %s %s\n", symbol, escapeTerminalControls(check.Message))
		}
	}
	if report.Action != nil {
		b.WriteByte('\n')
		if report.Action.Label != "" {
			b.WriteString(indentLines(escapeTerminalControls(report.Action.Label), "  "))
			b.WriteByte('\n')
		}
		for _, invocation := range report.Action.Invocations {
			rendered, ok := renderInvocation(invocation, shell)
			if !ok {
				continue
			}
			b.WriteByte('\n')
			b.WriteString("    ")
			b.WriteString(rendered)
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
	quoted := false
	for i := range parts {
		parts[i] = escapeTerminalControls(parts[i])
		if isBareWord(parts[i]) {
			continue
		}
		quoted = true
		if shell == displayShellPowerShell {
			parts[i] = quotePowerShell(parts[i])
		} else {
			parts[i] = quotePOSIX(parts[i])
		}
	}
	rendered := strings.Join(parts, " ")
	// The "&" call operator and single quotes are PowerShell-only; cmd.exe
	// treats both literally. GOOS cannot tell the two apart, so an invocation
	// that needs no quoting is emitted bare and runs in either shell.
	if shell == displayShellPowerShell && quoted {
		rendered = "& " + rendered
	}
	return rendered, true
}

// isBareWord reports whether a value can be passed to a shell verbatim.
func isBareWord(value string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_@%+=:,./-", r))
	}) < 0
}

func quotePOSIX(value string) string {
	if isBareWord(value) {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func quotePowerShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func escapeTerminalControls(value string) string {
	var b strings.Builder
	for _, r := range value {
		if !unicode.IsControl(r) {
			b.WriteRune(r)
			continue
		}
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			fmt.Fprintf(&b, `\u{%x}`, r)
		}
	}
	return b.String()
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
