package temporalcli

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	expectedCommand := "temporal server start-dev"
	if runtime.GOOS == "windows" {
		expectedCommand = "& 'temporal' 'server' 'start-dev'"
	}
	assert.Contains(t, stderr.String(), expectedCommand)
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

	powerShell, ok := renderInvocation(invocation, displayShellPowerShell)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(powerShell, "& 'temporal' "))
	assert.Contains(t, powerShell, `'space and ''quote'''`)
	assert.Contains(t, powerShell, `'--profile' '-prod'`)

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
