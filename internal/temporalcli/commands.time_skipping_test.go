package temporalcli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/temporalcli"
)

// Validate user input through the public CLI without contacting a server.
func TestTimeSkippingInvalidFlags(t *testing.T) {
	execute := func(t *testing.T, args ...string) *CommandResult {
		t.Helper()
		res := &CommandResult{}
		options := temporalcli.CommandOptions{
			IOStreams: temporalcli.IOStreams{Stdout: &res.Stdout, Stderr: &res.Stderr},
			Fail:      func(err error) { res.Err = err },
		}
		options.DeprecatedEnvConfig.DisableEnvConfig = true
		options.DeprecatedEnvConfig.EnvConfigName = "default"
		cctx, cancel, err := temporalcli.NewCommandContext(context.Background(), options)
		require.NoError(t, err)
		defer cancel()
		cmd := temporalcli.NewTemporalCommand(cctx)
		cmd.Command.SetArgs(args)
		cmd.Command.SetOut(&res.Stdout)
		cmd.Command.SetErr(&res.Stderr)
		if err := cmd.Command.ExecuteContext(cctx); err != nil {
			res.Err = err
		}
		return res
	}
	for _, noun := range []string{"workflow", "schedule"} {
		t.Run(noun, func(t *testing.T) {
			idFlag := "--" + noun + "-id"
			for _, verb := range []string{"enable", "disable", "poll"} {
				t.Run(verb+" missing execution ID", func(t *testing.T) {
					args := []string{noun, "time-skipping", verb}
					if verb == "enable" {
						args = append(args, "--fast-forward-duration", "5h")
					}
					if verb == "poll" {
						args = append(args, "--fast-forward-id", "YourFastForwardId")
					}
					res := execute(t, args...)
					require.EqualError(t, res.Err, `required flag(s) "`+noun+`-id" not set`)
				})
			}
			type invalidFlagsCase struct {
				name, verb string
				flags      []string
				want       string
			}
			cases := []invalidFlagsCase{
				{"missing poll ID", "poll", nil, `required flag(s) "fast-forward-id" not set`},
				{"zero duration", "enable", []string{"--fast-forward-duration", "0s"}, "--fast-forward-duration must be greater than zero"},
				{"negative duration", "enable", []string{"--fast-forward-duration", "-1h"}, "--fast-forward-duration must be greater than zero"},
				{"malformed duration", "enable", []string{"--fast-forward-duration", "invalid"}, "invalid argument"},
				{"zero skip limit", "enable", []string{"--fast-forward-duration", "5h", "--max-skip-per-session", "0"}, "--max-skip-per-session must be greater than zero"},
				{"negative skip limit", "enable", []string{"--fast-forward-duration", "5h", "--max-skip-per-session", "-1"}, "--max-skip-per-session must be greater than zero"},
				{"skip limit too large", "enable", []string{"--fast-forward-duration", "5h", "--max-skip-per-session", "10001"}, "--max-skip-per-session must be between 1 and 10000"},
				{"malformed skip limit", "enable", []string{"--fast-forward-duration", "5h", "--max-skip-per-session", "abc"}, "invalid argument"},
				{"malformed propagation", "enable", []string{"--fast-forward-duration", "5h", "--disable-propagation=abc"}, "invalid argument"},
				{"unsupported disable setting", "disable", []string{"--disable-propagation"}, "unknown flag: --disable-propagation"},
			}
			if noun == "workflow" {
				cases = append(cases, invalidFlagsCase{
					"ID without duration", "enable", []string{"--fast-forward-id", "YourFastForwardId"}, "--fast-forward-id requires --fast-forward-duration"})
			} else {
				cases = append(cases, invalidFlagsCase{
					"unsupported poll timeout", "poll", []string{"--fast-forward-id", "YourFastForwardId", "--timeout", "5m"}, "unknown flag: --timeout"})
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					args := append([]string{noun, "time-skipping", tc.verb, idFlag, "YourExecutionId"}, tc.flags...)
					res := execute(t, args...)
					require.ErrorContains(t, res.Err, tc.want)
					require.NotContains(t, res.Stdout.String(), "Time skipping: enabled")
				})
			}
		})
	}
}
