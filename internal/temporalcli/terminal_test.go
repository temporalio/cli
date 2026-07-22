package temporalcli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/cliext"
	"golang.org/x/oauth2"
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

	assert.Equal(t, "failed connecting to Temporal server at 127.0.0.1:7233: connection refused", err.Error())
	assert.ErrorIs(t, err, cause)
	report := normalizeError(err)
	require.NotNil(t, report.Action)
	require.Len(t, report.Action.Invocations, 1)
	assert.Equal(t, []string{"temporal", "server", "start-dev"}, report.Action.Invocations[0].Command)
	assert.NotContains(t, err.Error(), "\n")
}

func TestRecorderSeamCapturesFirstLeafFailureWithoutChangingItsIdentity(t *testing.T) {
	// Seam decision: generated Run callbacks already call Fail as their final
	// operation. A command-scoped recorder therefore captures every leaf error
	// without changing generated Run to RunE. RunE was rejected because Cobra
	// skips post-run cleanup when RunE returns an error.
	want := errors.New("leaf failure")
	var recorder commandErrorRecorder

	assert.True(t, runRecordedCommand(func() { recorder.Record(want) }))

	assert.ErrorIs(t, recorder.Err(), want)
}

func TestRecorderSeamStopsExecutionAfterFailureAndRunsCleanup(t *testing.T) {
	var recorder commandErrorRecorder
	workedAfterFailure := false
	cleanupRan := false
	func() {
		defer func() { cleanupRan = true }()
		assert.True(t, runRecordedCommand(func() {
			recorder.Record(errors.New("stop here"))
			workedAfterFailure = true
		}))
	}()
	assert.False(t, workedAfterFailure)
	assert.True(t, cleanupRan)
}

func TestRenderInvocationUsesExplicitShellQuotingAndRejectsControls(t *testing.T) {
	invocation := displayInvocation{Command: []string{"temporal", "config", "set"}, Args: []string{"--value", "space and 'quote'", "--profile", "-prod"}}
	posix, ok := renderInvocation(invocation, displayShellPOSIX)
	require.True(t, ok)
	assert.Contains(t, posix, `'space and '"'"'quote'"'"''`)
	assert.Contains(t, posix, "--profile -prod")
	powerShell, ok := renderInvocation(invocation, displayShellPowerShell)
	require.True(t, ok)
	assert.Contains(t, powerShell, `'space and ''quote'''`)
	_, ok = renderInvocation(displayInvocation{Command: []string{"temporal"}, Args: []string{"unsafe\nvalue"}}, displayShellPOSIX)
	assert.False(t, ok)
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

func TestRecorderSeamGeneratedLeavesStopAfterRecording(t *testing.T) {
	source, err := os.ReadFile("commands.gen.go")
	require.NoError(t, err)
	allCalls := bytes.Count(source, []byte("cctx.Options.Fail(err)"))
	terminalCalls := bytes.Count(source, []byte("cctx.Options.Fail(err)\n\t\t}"))

	assert.Greater(t, allCalls, 100, "expected to inspect every generated command leaf")
	assert.Equal(t, allCalls, terminalCalls, "every generated Fail call must be the final operation in its branch")
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
				DeprecatedEnvConfig: DeprecatedEnvConfig{DisableEnvConfig: true},
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

func TestCaptureEffectiveConnectionSecretsIncludesProfileValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`[profile.prod]
api_key = "profile-api-secret"
codec = { endpoint = "https://codec.example", auth = "profile-codec-secret" }
grpc_meta = { authorization = "profile-header-secret" }
tls = { client_cert_data = "profile-cert-secret", client_key_data = "profile-key-secret", server_ca_cert_data = "profile-ca-secret" }
`), 0o600))
	cctx := &CommandContext{
		Options: CommandOptions{EnvLookup: testEnvLookup{}},
		RootCommand: &TemporalCommand{CommonOptions: cliext.CommonOptions{
			ConfigFile: configPath, Profile: "prod", DisableConfigEnv: true,
		}},
	}
	captureEffectiveConnectionSecrets(cctx, &cliext.ClientOptions{})
	for _, secret := range []string{"profile-api-secret", "profile-codec-secret", "profile-header-secret", "profile-cert-secret", "profile-key-secret", "profile-ca-secret"} {
		assert.Contains(t, cctx.knownSecrets(), secret)
	}
	assert.True(t, cctx.hasAPIKey)
}

func TestCaptureEffectiveConnectionSecretsIncludesOAuthValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "temporal.toml")
	require.NoError(t, cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		ConfigFilePath: configPath,
		ProfileName:    "prod",
		OAuth: &cliext.OAuthConfig{
			ClientConfig: &oauth2.Config{ClientID: "client", ClientSecret: "oauth-client-secret"},
			Token:        &oauth2.Token{AccessToken: "oauth-access-secret", RefreshToken: "oauth-refresh-secret"},
		},
	}))
	cctx := &CommandContext{
		Options: CommandOptions{EnvLookup: testEnvLookup{}},
		RootCommand: &TemporalCommand{CommonOptions: cliext.CommonOptions{
			ConfigFile: configPath, Profile: "prod", DisableConfigEnv: true,
		}},
	}
	captureEffectiveConnectionSecrets(cctx, &cliext.ClientOptions{})
	assert.ElementsMatch(t, []string{"oauth-client-secret", "oauth-access-secret", "oauth-refresh-secret"}, cctx.knownSecrets())
	assert.True(t, cctx.hasOAuth)
	assert.False(t, cctx.hasAPIKey)
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
