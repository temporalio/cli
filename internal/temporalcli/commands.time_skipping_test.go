package temporalcli_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/temporalcli"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
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

// Run the CLI in a child process so its default Fail handler can call os.Exit.
// The regular command harness replaces that handler and only captures an error.
func TestTimeSkippingCLIProcess(t *testing.T) {
	if os.Getenv("TEMPORAL_CLI_TIME_SKIPPING_TEST_PROCESS") != "1" {
		return
	}
	separator := slices.Index(os.Args, "--")
	if separator < 0 {
		t.Fatal("missing CLI arguments")
	}
	options := temporalcli.CommandOptions{Args: os.Args[separator+1:]}
	options.DeprecatedEnvConfig.DisableEnvConfig = true
	options.DeprecatedEnvConfig.EnvConfigName = "default"
	temporalcli.Execute(context.Background(), options)
	os.Exit(0)
}

func (s *SharedServerSuite) TestTimeSkipping_PollExitStatus() {
	s.useTimeSkippingServer()
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return nil, workflow.Sleep(ctx, 10*time.Hour)
	})
	for _, noun := range []string{"workflow", "schedule"} {
		for _, outcome := range []string{"completed", "failed", "timeout"} {
			s.T().Run(noun+" "+outcome, func(t *testing.T) {
				var id string
				if noun == "workflow" {
					taskQueue := s.Worker().Options.TaskQueue
					if outcome != "completed" {
						// Keep fast-forward pending without an active worker.
						taskQueue = uuid.NewString()
					}
					run, err := s.Client.ExecuteWorkflow(s.Context,
						client.StartWorkflowOptions{TaskQueue: taskQueue}, DevWorkflow, nil)
					require.NoError(t, err)
					id = run.GetID()
					defer func() { _ = s.Client.CancelWorkflow(s.Context, id, "") }()
				} else {
					args := []string{"--interval", "10d"}
					if outcome != "completed" {
						args = append(args, "--paused")
					}
					var res *CommandResult
					id, _, res = s.createSchedule(args...)
					require.NoError(t, res.Err)
				}
				idFlag := "--" + noun + "-id"
				res := s.Execute(noun, "time-skipping", "enable", "--address", s.Address(),
					idFlag, id, "--fast-forward-duration", "5h", "--fast-forward-id", "ExitStatusTest")
				require.NoError(t, res.Err)
				if outcome == "failed" {
					res = s.Execute(noun, "time-skipping", "disable", "--address", s.Address(), idFlag, id)
					require.NoError(t, res.Err)
				}
				timeout := "15s"
				if outcome == "timeout" {
					timeout = "200ms"
				}
				for _, format := range []string{"text", "json", "jsonl"} {
					ctx, cancel := context.WithTimeout(s.Context, 20*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTimeSkippingCLIProcess$", "--",
						noun, "time-skipping", "poll", "--address", s.Address(), "--namespace", s.Namespace(),
						idFlag, id, "--fast-forward-id", "ExitStatusTest", "--command-timeout", timeout)
					if format != "text" {
						cmd.Args = append(cmd.Args, "--output", format)
					}
					cmd.Env = append(os.Environ(), "TEMPORAL_CLI_TIME_SKIPPING_TEST_PROCESS=1")
					var stdout, stderr bytes.Buffer
					cmd.Stdout, cmd.Stderr = &stdout, &stderr
					err := cmd.Run()
					if outcome == "completed" {
						require.NoError(t, err, "stderr: %s", stderr.String())
						if format == "text" {
							require.Equal(t, "completed", strings.TrimSpace(stdout.String()))
						} else {
							require.JSONEq(t, `{"completed":true}`, stdout.String())
							if format == "jsonl" {
								require.Len(t, strings.Split(strings.TrimSpace(stdout.String()), "\n"), 1)
							}
						}
						require.Empty(t, stderr.String())
						continue
					}
					var exitErr *exec.ExitError
					require.ErrorAs(t, err, &exitErr)
					require.Equal(t, 1, exitErr.ExitCode(), "stderr: %s", stderr.String())
					require.Empty(t, stdout.String())
					if outcome == "failed" {
						require.Regexp(t, `Error: fast-forward failed: \S`, stderr.String())
					} else {
						require.Equal(t, "Error: program interrupted", strings.TrimSpace(stderr.String()))
					}
				}
			})
		}
	}
}
