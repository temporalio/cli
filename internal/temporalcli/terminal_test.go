package temporalcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleTerminalErrorWritesOneRedactedReportAndRetainsBothErrors(t *testing.T) {
	commandErr := errors.New("request rejected for token exact-secret")
	writeErr := errors.New("stderr closed")
	writer := &countingErrorWriter{err: writeErr}

	result := handleTerminalError(commandErr, terminalOptions{
		Stderr:       writer,
		KnownSecrets: []string{"exact-secret"},
	})

	assert.Equal(t, 1, result.ExitStatus)
	assert.ErrorIs(t, result.CommandErr, commandErr)
	assert.ErrorIs(t, result.PresentationErr, writeErr)
	assert.Equal(t, 1, writer.writes)
	assert.Contains(t, string(writer.last), "Error: request rejected for token [REDACTED]")
	assert.NotContains(t, string(writer.last), "exact-secret")
}

func TestRenderErrorTextIsTotalAndUsesExplicitColorPolicy(t *testing.T) {
	report := errorReport{
		Summary: "connection refused",
		Checks: []errorCheck{{
			Outcome: checkFailed,
			Message: "TCP connection refused",
		}},
	}

	plain := string(renderErrorText(report, renderOptions{Color: false}))
	colored := string(renderErrorText(report, renderOptions{Color: true}))

	assert.Equal(t, "Error: connection refused\n\n  Connecting\n    ✗ TCP connection refused\n", plain)
	assert.NotContains(t, plain, "\x1b[")
	assert.Contains(t, colored, "\x1b[")
	assert.Equal(t, []byte("Error: unknown error\n"), renderErrorText(errorReport{}, renderOptions{}))
}

func TestConnectErrorCarriesSemanticFactsWithoutRenderedStateOrRawArguments(t *testing.T) {
	cause := errors.New("dial failed")
	diagnosis := &connectDiagnosis{
		Address: "127.0.0.1:7233",
		Cause:   causeTCPRefused,
		Stages:  []diagStage{{Status: diagFail, Label: "TCP connection refused"}},
	}
	err := newConnectError(diagnosis, connectMeta{Address: diagnosis.Address}, cause)
	diagnosis.Stages[0].Label = "mutated after construction"

	assert.Equal(t, "failed connecting to Temporal server at 127.0.0.1:7233: connection refused", err.Error())
	assert.ErrorIs(t, err, cause)
	report := normalizeError(err)
	assert.Equal(t, "TCP connection refused", report.Checks[0].Message, "connect error must copy diagnosis stages")
	require.NotNil(t, report.Action)
	require.Len(t, report.Action.Invocations, 1)
	assert.Equal(t, []string{"temporal", "server", "start-dev"}, report.Action.Invocations[0].Command)
	assert.NotContains(t, err.Error(), "\n")
}

func TestNormalizeErrorOwnsActivityNotFoundPresentationAndPreservesUnknownActivityErrors(t *testing.T) {
	notFound := &activityNotFoundError{
		activityID: "activity-id",
		cause:      errors.New("server detail"),
	}
	report := normalizeError(fmt.Errorf("poll activity: %w", notFound))

	assert.Equal(t, "standalone Activity not found", report.Summary)
	assert.Empty(t, report.Context)
	assert.Empty(t, report.Checks)
	assert.Nil(t, report.Action)

	unknown := errors.New("activity result unavailable")
	assert.Equal(t, unknown.Error(), normalizeError(unknown).Summary)
}

func TestRenderInvocationUsesExplicitShellQuotingAndEscapesControls(t *testing.T) {
	invocation := displayInvocation{Command: []string{"temporal", "config", "set"}, Args: []string{"--value", "space and 'quote'", "--profile", "-prod"}}
	posix, ok := renderInvocation(invocation, displayShellPOSIX)
	require.True(t, ok)
	assert.Contains(t, posix, `'space and '"'"'quote'"'"''`)
	assert.Contains(t, posix, "--profile -prod")
	powerShell, ok := renderInvocation(invocation, displayShellPowerShell)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(powerShell, "& 'temporal' "))
	assert.Contains(t, powerShell, `'space and ''quote'''`)
	assert.Contains(t, powerShell, `'--profile' '-prod'`)
	escaped, ok := renderInvocation(displayInvocation{Command: []string{"temporal"}, Args: []string{"unsafe\nvalue"}}, displayShellPOSIX)
	require.True(t, ok)
	assert.Equal(t, `temporal 'unsafe\nvalue'`, escaped)
}

func TestRedactReportDeepCopiesEscapesControlsAndUsesLongestSecretFirst(t *testing.T) {
	report := errorReport{
		Summary:      "token-long\x1b[31m",
		Context:      []safeField{{Label: "Target\nName", Value: "token-long"}},
		CheckHeading: "Check\rHeading",
		Checks:       []errorCheck{{Outcome: checkFailed, Message: "token\ncheck"}},
		Action: &displayAction{
			Label:       "use\ttoken-long",
			Invocations: []displayInvocation{{Command: []string{"temporal"}, Args: []string{"token-long"}}},
		},
	}
	secrets := []string{"token", "token-long"}
	got := redactReport(report, secrets)

	assert.Equal(t, `[REDACTED]\u{1b}[31m`, got.Summary)
	assert.Equal(t, `Target\nName`, got.Context[0].Label)
	assert.Equal(t, `Check\rHeading`, got.CheckHeading)
	assert.Equal(t, `[REDACTED]\ncheck`, got.Checks[0].Message)
	assert.Equal(t, `use\t[REDACTED]`, got.Action.Label)
	assert.Equal(t, "token-long", report.Context[0].Value, "source context must not be mutated")
	assert.Equal(t, "token-long", report.Action.Invocations[0].Args[0], "nested source slices must not be mutated")
	assert.Equal(t, "[REDACTED]", got.Action.Invocations[0].Args[0])
	assert.Equal(t, []string{"token", "token-long"}, secrets, "sorting must not mutate the caller's secret list")
}

func TestRenderInvocationPowerShellDoublesSingleQuotesWithinOneQuotedToken(t *testing.T) {
	got, ok := renderInvocation(displayInvocation{Command: []string{"temporal"}, Args: []string{"it's one token"}}, displayShellPowerShell)
	require.True(t, ok)
	assert.Equal(t, "& 'temporal' 'it''s one token'", got)
}

func TestFinishCommandPreservesOriginalErrorWhilePresentingCancellationCause(t *testing.T) {
	original := errors.New("original command failure")
	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(errors.New("command timed out after 2s"))
	var stderr bytes.Buffer
	cctx := &CommandContext{Context: ctx, Options: CommandOptions{IOStreams: IOStreams{Stderr: &stderr}, EnvLookup: testEnvLookup{}}}
	result := finishCommand(cctx, original, "")
	assert.ErrorIs(t, result.CommandErr, original)
	assert.Contains(t, stderr.String(), "command timed out after 2s")
	assert.NotContains(t, stderr.String(), original.Error())
}

func TestGeneratedLeavesReturnErrorsThroughRunE(t *testing.T) {
	source, err := os.ReadFile("commands.gen.go")
	require.NoError(t, err)
	runECalls := bytes.Count(source, []byte("s.Command.RunE = func"))
	returnCalls := bytes.Count(source, []byte("return s.run(cctx, args)"))
	markerCalls := bytes.Count(source, []byte("cctx.commandRunStarted = true"))

	assert.Greater(t, runECalls, 100, "expected to inspect every generated command leaf")
	assert.Equal(t, runECalls, returnCalls, "every generated RunE must return its run error")
	assert.Equal(t, runECalls, markerCalls, "every generated RunE must mark command execution")
	assert.NotContains(t, string(source), "Options.Fail")
}

func TestExecuteRestoresColorAcrossFailureSources(t *testing.T) {
	old := color.NoColor
	t.Cleanup(func() { color.NoColor = old })

	for name, args := range map[string][]string{
		"argument": {"workflow", "describe", "--color", "always"},
		"pre-run":  {"workflow", "list", "--time-format", "invalid", "--color", "always"},
		"runtime":  {"config", "get", "--disable-config-file", "--disable-config-env", "--color", "always"},
	} {
		t.Run(name, func(t *testing.T) {
			color.NoColor = true
			var stdout, stderr bytes.Buffer
			result := Execute(t.Context(), CommandOptions{
				Args:                args,
				IOStreams:           IOStreams{Stdout: &stdout, Stderr: &stderr},
				DeprecatedEnvConfig: DeprecatedEnvConfig{DisableEnvConfig: true, EnvConfigName: "default"},
			})
			assert.Error(t, result.CommandErr)
			assert.True(t, color.NoColor, "global color policy must be restored")
		})
	}
}

func TestEarlyJSONPolicyDisablesTerminalColor(t *testing.T) {
	cctx := &CommandContext{Options: CommandOptions{Args: []string{"workflow", "describe", "--output=jsonl", "--color=always"}}}
	assert.False(t, cctx.terminalColorEnabled())
}

func TestEarlyJSONFailuresRenderUsageWithoutANSI(t *testing.T) {
	old := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = old })

	for name, args := range map[string][]string{
		"json parse failure":     {"workflow", "describe", "--output", "json", "--color", "always", "--not-a-flag"},
		"jsonl required failure": {"workflow", "describe", "--output=jsonl", "--color=always"},
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			result := Execute(t.Context(), CommandOptions{
				Args:                args,
				IOStreams:           IOStreams{Stdout: &stdout, Stderr: &stderr},
				DeprecatedEnvConfig: DeprecatedEnvConfig{DisableEnvConfig: true, EnvConfigName: "default"},
			})

			assert.Error(t, result.CommandErr)
			assert.Equal(t, 1, result.ExitStatus)
			assert.Empty(t, stdout.String())
			assert.Contains(t, stderr.String(), "Usage:")
			assert.NotContains(t, stderr.String(), "\x1b[")
			assert.Equal(t, 1, bytes.Count(stderr.Bytes(), []byte("Error:")))
		})
	}
}

func TestExecuteRuntimeErrorPreservesIdentityWithoutUsageAndRetainsPresentationError(t *testing.T) {
	configPath := t.TempDir()
	writeErr := errors.New("stderr unavailable")
	stderr := &countingErrorWriter{err: writeErr}
	var stdout bytes.Buffer

	result := Execute(t.Context(), CommandOptions{
		Args: []string{
			"config", "get",
			"--config-file", configPath,
			"--disable-config-env",
		},
		IOStreams:           IOStreams{Stdout: &stdout, Stderr: stderr},
		DeprecatedEnvConfig: DeprecatedEnvConfig{DisableEnvConfig: true, EnvConfigName: "default"},
	})

	require.Error(t, result.CommandErr)
	var pathErr *os.PathError
	require.ErrorAs(t, result.CommandErr, &pathErr)
	assert.Equal(t, configPath, pathErr.Path)
	assert.ErrorIs(t, result.CommandErr, pathErr)
	assert.ErrorIs(t, result.PresentationErr, writeErr)
	assert.Equal(t, 1, result.ExitStatus)
	assert.Equal(t, 1, stderr.writes)
	assert.NotContains(t, string(stderr.last), "Usage:")
	assert.Empty(t, stdout.String())
}

func TestUnknownFlagWithFollowingValueIsNotMisclassifiedAsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	result := Execute(t.Context(), CommandOptions{
		Args:                []string{"--definitely-invalid", "value"},
		IOStreams:           IOStreams{Stdout: &stdout, Stderr: &stderr},
		DeprecatedEnvConfig: DeprecatedEnvConfig{DisableEnvConfig: true},
	})
	require.Error(t, result.CommandErr)
	assert.Contains(t, result.CommandErr.Error(), "unknown flag: --definitely-invalid")
	assert.NotContains(t, result.CommandErr.Error(), "unknown command")
}

func TestKnownSecretsExtractsAvailableScalarSliceHeaderAndEnvironmentValues(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("api-key", "", "")
	cmd.Flags().StringArray("grpc-meta", nil, "")
	require.NoError(t, cmd.Flags().Parse([]string{"--api-key", "flag-secret", "--grpc-meta", "Authorization=header-secret"}))
	cctx := &CommandContext{
		CurrentCommand: cmd,
		Options: CommandOptions{EnvLookup: testEnvLookup{
			"TEMPORAL_CODEC_AUTH": "env-secret",
		}},
	}

	secrets := cctx.knownSecrets()
	assert.Contains(t, secrets, "flag-secret")
	assert.Contains(t, secrets, "Authorization=header-secret")
	assert.Contains(t, secrets, "header-secret")
	assert.Contains(t, secrets, "env-secret")
}

func TestKnownSecretsIgnoresDisabledConfigEnvironment(t *testing.T) {
	cctx := &CommandContext{Options: CommandOptions{
		Args:      []string{"--disable-config-env"},
		EnvLookup: testEnvLookup{"TEMPORAL_API_KEY": "disabled-env-secret"},
	}}
	assert.NotContains(t, cctx.knownSecrets(), "disabled-env-secret")
}

type testEnvLookup map[string]string

func (e testEnvLookup) LookupEnv(name string) (string, bool) {
	value, ok := e[name]
	return value, ok
}

func (e testEnvLookup) Environ() []string { return nil }

type countingErrorWriter struct {
	err    error
	writes int
	last   []byte
}

func (w *countingErrorWriter) Write(p []byte) (int, error) {
	w.writes++
	w.last = append([]byte(nil), p...)
	return 0, w.err
}

var _ io.Writer = (*countingErrorWriter)(nil)
