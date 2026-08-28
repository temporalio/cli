package temporalcli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
)

func TestWriteConnectionErrorHandlesWrappedConnectError(t *testing.T) {
	cause := errors.New("dial failed")
	err := fmt.Errorf("client setup: %w", newConnectError(&connectDiagnosis{
		Address: "127.0.0.1:7233",
		Cause:   causeTCPRefused,
		Stages:  []diagStage{{Status: diagFail, Label: "TCP connection refused"}},
	}, connectMeta{Address: "127.0.0.1:7233", Namespace: "default"}, cause))

	stderr := &countingWriter{}
	assert.True(t, writeConnectionError(stderr, err, false))
	assert.Equal(t, 1, stderr.writes)
	assert.Contains(t, stderr.String(), "Error: failed connecting to Temporal server at 127.0.0.1:7233: connection refused")
	assert.Contains(t, stderr.String(), "Namespace: default")
	assert.Contains(t, stderr.String(), "✗ TCP connection refused")
	// Rendered identically on every platform. GOOS cannot distinguish
	// PowerShell from cmd.exe, so a command needing no quoting is emitted
	// bare: no call operator and no single quotes, which cmd takes literally.
	assert.Contains(t, stderr.String(), "\n    temporal server start-dev\n")
	assert.NotContains(t, stderr.String(), "\x1b[")
}

func TestWriteConnectionErrorLeavesGenericErrorsUnhandled(t *testing.T) {
	var stderr bytes.Buffer
	assert.False(t, writeConnectionError(&stderr, errors.New("ordinary failure"), true))
	assert.Empty(t, stderr.String())
}

func TestRenderConnectionReportUsesExplicitColorPolicy(t *testing.T) {
	report := connectionReport{
		Summary: "connection refused",
		Checks: []errorCheck{{
			Outcome: checkFailed,
			Message: "TCP connection refused",
		}},
	}

	plain := string(renderConnectionReport(report, false, displayShellPOSIX))
	colored := string(renderConnectionReport(report, true, displayShellPOSIX))

	assert.Equal(t, "Error: connection refused\n\n  Connecting\n    ✗ TCP connection refused\n", plain)
	assert.NotContains(t, plain, "\x1b[")
	assert.Contains(t, colored, "\x1b[31m✗\x1b[0m")
}

// An invocation needing no quoting must be runnable in cmd.exe as well as
// PowerShell, since GOOS cannot tell the two apart. The "&" call operator and
// single quotes are PowerShell-only, so neither may appear.
func TestRenderInvocationStaysRunnableInBothWindowsShells(t *testing.T) {
	invocation := displayInvocation{Command: []string{"temporal", "server", "start-dev"}}

	powerShell, ok := renderInvocation(invocation, displayShellPowerShell)
	require.True(t, ok)
	assert.Equal(t, "temporal server start-dev", powerShell)
	assert.NotContains(t, powerShell, "&")
	assert.NotContains(t, powerShell, "'")

	posix, ok := renderInvocation(invocation, displayShellPOSIX)
	require.True(t, ok)
	assert.Equal(t, powerShell, posix, "a bare invocation renders identically in every shell")
}

// The report is written to Options.Stderr, but fatih/color derives its global
// NoColor from os.Stdout. With --color auto, a terminal stdout and a
// redirected stderr would otherwise put ANSI escapes into captured output.
func TestUseColorForStderrFollowsTheConfiguredWriter(t *testing.T) {
	newCtx := func(stderr io.Writer, policy string) *CommandContext {
		c := &CommandContext{Options: CommandOptions{IOStreams: IOStreams{Stderr: stderr}}}
		if policy != "" {
			root := &TemporalCommand{}
			root.Color = cliext.NewFlagStringEnum([]string{"always", "never", "auto"}, "auto")
			root.Color.Value = policy
			c.RootCommand = root
		}
		return c
	}

	// Force color.NoColor off so "auto" is decided by the writer alone.
	orig := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = orig })

	var buf bytes.Buffer
	assert.False(t, newCtx(&buf, "auto").useColorForStderr(),
		"a non-terminal stderr must not receive escape sequences")
	assert.True(t, newCtx(&buf, "always").useColorForStderr(),
		"an explicit policy still wins")
	assert.False(t, newCtx(&buf, "never").useColorForStderr())

	jsonCtx := newCtx(&buf, "always")
	jsonCtx.JSONOutput = true
	assert.False(t, jsonCtx.useColorForStderr(), "JSON output is never colored")

	// color.NoColor still vetoes: it carries NO_COLOR and TERM=dumb.
	color.NoColor = true
	assert.False(t, newCtx(&buf, "auto").useColorForStderr())
}

func TestConnectionReportEscapesTerminalControls(t *testing.T) {
	report := connectionReport{
		Summary:      "bad\x1b[31m",
		Context:      []safeField{{Label: "Target\nName", Value: "host\tname"}},
		CheckHeading: "Check\rHeading",
		Checks:       []errorCheck{{Outcome: checkFailed, Message: "failed\ncheck"}},
		Action:       &displayAction{Label: "retry\tnow"},
	}

	rendered := string(renderConnectionReport(report, false, displayShellPOSIX))
	assert.Contains(t, rendered, `bad\u{1b}[31m`)
	assert.Contains(t, rendered, `Target\nName: host\tname`)
	assert.Contains(t, rendered, `Check\rHeading`)
	assert.Contains(t, rendered, `failed\ncheck`)
	assert.Contains(t, rendered, `retry\tnow`)
}

func TestRenderInvocationUsesExplicitShellQuoting(t *testing.T) {
	invocation := displayInvocation{
		Command: []string{"temporal", "config", "set"},
		Args:    []string{"--value", "space and 'quote'", "--profile", "-prod"},
	}
	posix, ok := renderInvocation(invocation, displayShellPOSIX)
	require.True(t, ok)
	assert.Contains(t, posix, `'space and '"'"'quote'"'"''`)
	assert.Contains(t, posix, "--profile -prod")

	// An argument that needs quoting forces PowerShell syntax: single quotes
	// doubled, and the "&" call operator so the quoted command still runs.
	powerShell, ok := renderInvocation(invocation, displayShellPowerShell)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(powerShell, "& temporal "))
	assert.Contains(t, powerShell, `'space and ''quote'''`)
	// Bare words stay bare in both shells.
	assert.Contains(t, powerShell, "--profile -prod")

	escaped, ok := renderInvocation(
		displayInvocation{Command: []string{"temporal"}, Args: []string{"unsafe\nvalue"}},
		displayShellPOSIX,
	)
	require.True(t, ok)
	assert.Equal(t, `temporal 'unsafe\nvalue'`, escaped)
}

func TestConnectErrorCopiesDiagnosisBeforeRendering(t *testing.T) {
	cause := errors.New("dial failed")
	diagnosis := &connectDiagnosis{
		Address: "127.0.0.1:7233",
		Cause:   causeTCPRefused,
		Stages:  []diagStage{{Status: diagFail, Label: "TCP connection refused"}},
	}
	err := newConnectError(diagnosis, connectMeta{Address: diagnosis.Address}, cause)
	diagnosis.Stages[0].Label = "mutated after construction"

	report := connectionErrorReport(err)
	require.Len(t, report.Checks, 1)
	assert.Equal(t, "TCP connection refused", report.Checks[0].Message)
	assert.ErrorIs(t, err, cause)
}

type countingWriter struct {
	bytes.Buffer
	writes int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	return w.Buffer.Write(p)
}
