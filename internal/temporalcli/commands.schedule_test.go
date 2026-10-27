package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *SharedServerSuite) createSchedule(args ...string) (schedId, schedWfId string, res *CommandResult) {
	schedId = fmt.Sprintf("sched-%x", rand.Uint32())
	schedWfId = fmt.Sprintf("my-wf-id-%x", rand.Uint32())
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return nil, workflow.Sleep(ctx, 10*time.Second)
	})
	s.T().Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		options := temporalcli.CommandOptions{
			IOStreams: temporalcli.IOStreams{
				Stdout: io.Discard,
				Stderr: io.Discard,
			},
			Args: []string{
				"schedule", "delete",
				"--address", s.Address(),
				"-s", schedId,
			},
			Fail: func(error) {},
		}
		temporalcli.Execute(ctx, options)
	})
	res = s.Execute(append([]string{
		"schedule", "create",
		"--address", s.Address(),
		"-s", schedId,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", schedWfId,
	}, args...)...,
	)
	return
}

func (s *SharedServerSuite) TestSchedule_Create() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingEnable() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
	s.NoError(res.Err)
	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-duration", "5h",
		"--fast-forward-id", "YourFastForwardId",
		"--disable-propagation",
		"--max-skip-per-session", "17",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Time skipping: enabled")
	s.Contains(res.Stdout.String(), "Fast-forward duration: 5h 0m 0s")
	s.Contains(res.Stdout.String(), "Fast-forward ID: YourFastForwardId")
	s.Contains(res.Stdout.String(), "Maximum skips per session: 17")
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"--schedule-id", schedID,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), `"DisablePropagation":true`)
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingEnableRequiresDuration() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
	require.NoError(s.T(), res.Err)

	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
	)
	s.EqualError(res.Err, `required flag(s) "fast-forward-duration" not set`)
	s.Contains(res.Stderr.String(), `Error: required flag(s) "fast-forward-duration" not set`)
	s.NotContains(res.Stdout.String(), "Schedule updated:")
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingEnableGeneratesFastForwardID() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
	require.NoError(s.T(), res.Err)

	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-duration", "5h",
	)
	require.NoError(s.T(), res.Err)
	match := regexp.MustCompile(`(?m)^Fast-forward ID: (\S+)$`).FindStringSubmatch(res.Stdout.String())
	require.Len(s.T(), match, 2)
	fastForwardID := match[1]
	_, err := uuid.Parse(fastForwardID)
	require.NoError(s.T(), err)
	s.Contains(res.Stdout.String(), "temporal schedule time-skipping poll --namespace default --schedule-id "+schedID+" --fast-forward-id "+fastForwardID)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"--schedule-id", schedID,
	)
	require.NoError(s.T(), res.Err)
	s.Contains(res.Stdout.String(), `"Id":"`+fastForwardID+`"`)

	// The generated ID is usable in the suggested poll command.
	res = s.Execute(
		"schedule", "time-skipping", "poll",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-id", fastForwardID,
		"--timeout", "200ms",
	)
	s.NoError(res.Err)
	s.Equal("poll timed out", strings.TrimSpace(res.Stdout.String()))
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingPollTimeout() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
	s.NoError(res.Err)
	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-duration", "5h",
		"--fast-forward-id", "YourFastForwardId",
	)
	s.NoError(res.Err)

	// A timeout leaves the fast-forward available to poll again.
	for range 2 {
		res = s.Execute(
			"schedule", "time-skipping", "poll",
			"--address", s.Address(),
			"--schedule-id", schedID,
			"--fast-forward-id", "YourFastForwardId",
			"--timeout", "200ms",
		)
		s.NoError(res.Err)
		s.Equal("poll timed out", strings.TrimSpace(res.Stdout.String()))
	}
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingMinimalLifecycle() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-duration", "5h",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Time skipping: enabled")
	match := regexp.MustCompile(`(?m)^Fast-forward ID: (\S+)$`).FindStringSubmatch(res.Stdout.String())
	s.Len(match, 2)
	fastForwardID := match[1]
	_, err := uuid.Parse(fastForwardID)
	s.NoError(err)

	res = s.Execute(
		"schedule", "time-skipping", "poll",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-id", fastForwardID,
		"--timeout", "10s",
	)
	s.NoError(res.Err)
	s.Equal("completed", strings.TrimSpace(res.Stdout.String()))

	res = s.Execute("schedule", "describe", "--address", s.Address(), "--schedule-id", schedID)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "TimeSkippingInfo")
	s.Contains(res.Stdout.String(), `"Id":"`+fastForwardID+`"`)
	s.Contains(res.Stdout.String(), `"HasCompleted":true`)
	s.Contains(res.Stdout.String(), `"DisablePropagation":false`)

	for _, output := range []string{"json", "jsonl"} {
		res = s.Execute("schedule", "describe", "--address", s.Address(), "--schedule-id", schedID, "--output", output)
		s.NoError(res.Err)
		var description workflowservice.DescribeScheduleResponse
		s.NoError(protojson.Unmarshal(res.Stdout.Bytes(), &description))
		s.Equal("DevWorkflow", description.GetSchedule().GetAction().GetStartWorkflow().GetWorkflowType().GetName())
		info := description.GetInfo().GetTimeSkippingInfo()
		s.NotNil(info)
		s.Equal(fastForwardID, info.GetFastForwardInfo().GetFastForwardId())
		s.True(info.GetFastForwardInfo().GetHasCompleted())
		s.False(info.GetEffectiveConfig().GetDisablePropagation())
		if output == "jsonl" {
			s.Len(strings.Split(strings.TrimSpace(res.Stdout.String()), "\n"), 1)
		}
	}

	res = s.Execute("schedule", "time-skipping", "disable", "--address", s.Address(), "--schedule-id", schedID)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Time skipping: disabled")
	res = s.Execute("schedule", "describe", "--address", s.Address(), "--schedule-id", schedID)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), `"EffectiveEnabled":false`)
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingPollFailed() {
	s.useTimeSkippingServer()
	for _, change := range []string{"not enabled", "disabled", "replaced"} {
		s.T().Run(change, func(t *testing.T) {
			schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
			require.NoError(t, res.Err)
			if change != "not enabled" {
				res = s.Execute(
					"schedule", "time-skipping", "enable",
					"--address", s.Address(),
					"--schedule-id", schedID,
					"--fast-forward-duration", "5h",
					"--fast-forward-id", "OriginalFastForwardId",
				)
				require.NoError(t, res.Err)
			}
			switch change {
			case "disabled":
				res = s.Execute(
					"schedule", "time-skipping", "disable",
					"--address", s.Address(),
					"--schedule-id", schedID,
				)
				require.NoError(t, res.Err)
			case "replaced":
				res = s.Execute(
					"schedule", "time-skipping", "enable",
					"--address", s.Address(),
					"--schedule-id", schedID,
					"--fast-forward-duration", "5h",
					"--fast-forward-id", "ReplacementFastForwardId",
				)
				require.NoError(t, res.Err)
			}
			res = s.Execute(
				"schedule", "time-skipping", "poll",
				"--address", s.Address(),
				"--schedule-id", schedID,
				"--fast-forward-id", "OriginalFastForwardId",
				"--timeout", "5s",
			)
			require.NoError(t, res.Err)
			require.Regexp(t, `^failed: \S`, strings.TrimSpace(res.Stdout.String()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingPollCanceled() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
	s.NoError(res.Err)
	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-duration", "5h",
		"--fast-forward-id", "YourFastForwardId",
	)
	s.NoError(res.Err)

	ctx, cancel := context.WithCancel(s.Context)
	cancel()
	res = s.ExecuteWithContext(ctx,
		"schedule", "time-skipping", "poll",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-id", "YourFastForwardId",
	)
	s.ErrorContains(res.Err, "canceled")
	s.Empty(res.Stdout.String())
}

func (s *SharedServerSuite) TestSchedule_Delete() {
	schedId, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// check exists
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "delete",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)

	// doesn't exist anymore
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestSchedule_Describe() {
	schedId, schedWfId, res := s.createSchedule("--interval", "2s")
	s.NoError(res.Err)

	// run once manually so we see a running workflow

	res = s.Execute(
		"schedule", "trigger",
		"--address", s.Address(),
		"-s", schedId,
	)

	// text

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", schedId,
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		s.ContainsOnSameLine(out, "ScheduleId", schedId)
		s.ContainsOnSameLine(out, "Spec", "2s")
		return AssertContainsOnSameLine(out, "RunningWorkflows", schedWfId+"-") == nil
	}, 10*time.Second, 100*time.Millisecond)

	// json

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
		"-o", "json",
	)
	s.NoError(res.Err)
	var j struct {
		Schedule struct {
			Action struct {
				StartWorkflow struct {
					Id string `json:"workflowId"`
				} `json:"startWorkflow"`
			} `json:"action"`
		} `json:"schedule"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	s.Equal(schedWfId, j.Schedule.Action.StartWorkflow.Id)
}

func (s *SharedServerSuite) TestSchedule_CreateDescribeCalendar() {
	schedId, _, res := s.createSchedule("--calendar", `{"hour":"2,4","dayOfWeek":"thu,fri"}`)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ScheduleId", schedId)
	s.ContainsOnSameLine(out, "Spec", "dayOfWeek")
}

func (s *SharedServerSuite) TestSchedule_CreateDescribe_SearchAttributes_Memo() {
	s.t.Skip("Skipped until issue #590- Error printing typed search attributes inside schedule actions - is fixed")

	schedId, _, res := s.createSchedule("--interval", "10d",
		"--schedule-search-attribute", `CustomKeywordField="schedule-string-val"`,
		"--search-attribute", `CustomKeywordField="workflow-string-val"`,
		"--schedule-memo", `schedMemo="data here"`,
		"--memo", `wfMemo="other data"`,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	// TODO: We have to disable shorthand payload encoding for now so these come out as base64.
	// After https://github.com/temporalio/api-go/pull/154, ensure these come out as nice strings.
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	s.ContainsOnSameLine(out, "SearchAttributes", "CustomKeywordField", b64(`"schedule-string-val"`))
	s.ContainsOnSameLine(out, "Memo", "schedMemo", `"data here"`) // somehow this one comes out as a string anyway
	s.ContainsOnSameLine(out, "Action", "CustomKeywordField", b64(`"workflow-string-val"`))
	s.ContainsOnSameLine(out, "Action", "wfMemo", b64(`"other data"`))
}

func (s *SharedServerSuite) TestSchedule_CreateDescribe_UserMetadata() {
	schedId, _, res := s.createSchedule("--interval", "10d",
		"--static-summary", "summ",
		"--static-details", "details",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Action", "Summary", "summ")
	s.ContainsOnSameLine(out, "Action", "Details", "details")
}

func (s *SharedServerSuite) TestSchedule_List() {
	res := s.Execute(
		"operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "TestSchedule_List",
		"--type", "keyword",
	)
	s.NoError(res.Err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res = s.Execute(
			"operator", "search-attribute", "list",
			"--address", s.Address(),
			"-o", "json",
		)
		assert.NoError(t, res.Err)
		var jsonOut operatorservice.ListSearchAttributesResponse
		assert.NoError(t, temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
		assert.Equal(t, enums.INDEXED_VALUE_TYPE_KEYWORD, jsonOut.CustomAttributes["TestSchedule_List"])
	}, 10*time.Second, time.Second)

	schedId, _, res := s.createSchedule(
		"--interval",
		"10d",
		"--schedule-search-attribute", `TestSchedule_List="here"`,
	)
	s.NoError(res.Err)

	// table really-long
	var out string
	s.EventuallyWithT(func(t *assert.CollectT) {
		res = s.Execute(
			"schedule", "list",
			"--address", s.Address(),
			"--really-long",
		)
		assert.NoError(t, res.Err)
		out = res.Stdout.String()
		assert.Contains(t, out, schedId)
	}, 10*time.Second, time.Second)
	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "0s" /*jitter*/, "false", "{}" /*memo*/)
	s.ContainsOnSameLine(out, "TestSchedule_List")

	// table

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out = res.Stdout.String()

	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "false")

	// table long

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--long",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "false")

	// json

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res.Err)
	var j []struct {
		ScheduleId string `json:"scheduleId"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	ok := false
	for _, entry := range j {
		ok = ok || entry.ScheduleId == schedId
	}
	s.True(ok, "schedule not found in json result")

	// jsonl

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"-o", "jsonl",
	)
	s.NoError(res.Err)
	lines := bytes.Split(res.Stdout.Bytes(), []byte("\n"))
	ok = false
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var j struct {
			ScheduleId string `json:"scheduleId"`
		}
		s.NoError(json.Unmarshal(line, &j))
		ok = ok || j.ScheduleId == schedId
	}
	s.True(ok, "schedule not found in jsonl result")

	// JSON query (match)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'here'",
		"-o", "json",
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	ok = false
	for _, entry := range j {
		ok = ok || entry.ScheduleId == schedId
	}
	s.True(ok, "schedule not found in json result")

	// query (match)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'here'",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "false")

	// JSON query (no matches)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'notHere'",
		"-o", "json",
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	ok = false
	for _, entry := range j {
		ok = ok || entry.ScheduleId == schedId
	}
	s.False(ok, "schedule found in json result, but should not be found")

	// query (no matches)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'notHere'",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.NotContainsf(out, schedId, "schedule found, but should not be found")

	// query (invalid query field)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "unknownField = 'notHere'",
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestSchedule_Toggle() {
	schedId, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// pause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", schedId,
		"--pause",
		"--reason", "testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "true")
	s.ContainsOnSameLine(out, "Notes", "testing")

	// unpause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", schedId,
		"--unpause",
		"--reason", "we're done testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "false")
	s.ContainsOnSameLine(out, "Notes", "done testing")
}

func (s *SharedServerSuite) TestSchedule_Trigger() {
	schedId, schedWfId, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "trigger",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"workflow", "list",
			"--address", s.Address(),
			"-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, schedId),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		return AssertContainsOnSameLine(out, schedWfId) == nil
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_Backfill() {
	schedId, schedWfId, res := s.createSchedule("--interval", "10d/5h")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "backfill",
		"--address", s.Address(),
		"-s", schedId,
		"--start-time", "2022-02-02T00:00:00Z",
		"--end-time", "2022-02-28T00:00:00Z",
		"--overlap-policy", "AllowAll",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"workflow", "list",
			"--address", s.Address(),
			"-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, schedId),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		re := regexp.MustCompile(regexp.QuoteMeta(schedWfId + "-2022-02"))
		return len(re.FindAllString(out, -1)) == 3
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_Update() {
	schedId, schedWfId, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "update",
		"--address", s.Address(),
		"-s", schedId,
		"--task-queue", "SomeOtherTq",
		"--type", "SomeOtherWf",
		"--workflow-id", schedWfId,
		"--interval", "1h",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", schedId,
			"-o", "json",
		)
		s.NoError(res.Err)
		var j struct {
			Schedule struct {
				Spec struct {
					Interval []struct {
						Interval string `json:"interval"`
					} `json:"interval"`
				} `json:"spec"`
				Action struct {
					StartWorkflow struct {
						WorkflowType struct {
							Name string `json:"name"`
						} `json:"workflowType"`
						TaskQueue struct {
							Name string `json:"name"`
						} `json:"taskQueue"`
					} `json:"startWorkflow"`
				} `json:"action"`
			} `json:"schedule"`
		}
		s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
		return j.Schedule.Action.StartWorkflow.WorkflowType.Name == "SomeOtherWf" &&
			j.Schedule.Action.StartWorkflow.TaskQueue.Name == "SomeOtherTq" &&
			j.Schedule.Spec.Interval[0].Interval == "3600s"
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingDisable() {
	s.useTimeSkippingServer()
	schedID, _, res := s.createSchedule("--interval", "10d", "--paused")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--fast-forward-duration", "6h",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Time skipping: enabled")
	s.Contains(res.Stdout.String(), "Fast-forward ID:")

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"--schedule-id", schedID,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "TimeSkippingInfo")
	s.Contains(res.Stdout.String(), `"EffectiveEnabled":true`)
	s.Contains(res.Stdout.String(), `"FastForwardInfo":`)

	res = s.Execute(
		"schedule", "time-skipping", "disable",
		"--address", s.Address(),
		"--schedule-id", schedID,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Time skipping: disabled")

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--output", "json",
	)
	s.NoError(res.Err)
	var description struct {
		Schedule struct {
			Action struct {
				StartWorkflow struct {
					WorkflowType struct {
						Name string `json:"name"`
					} `json:"workflowType"`
				} `json:"startWorkflow"`
			} `json:"action"`
		} `json:"schedule"`
		Info struct {
			TimeSkippingInfo struct {
				EffectiveConfig struct {
					Enabled bool `json:"enabled"`
				} `json:"effectiveConfig"`
			} `json:"timeSkippingInfo"`
		} `json:"info"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &description))
	s.Equal("DevWorkflow", description.Schedule.Action.StartWorkflow.WorkflowType.Name)
	s.False(description.Info.TimeSkippingInfo.EffectiveConfig.Enabled)
}

func (s *SharedServerSuite) TestSchedule_Memo_Update() {
	schedId, schedWfId, res := s.createSchedule("--memo", "bar=1")
	s.NoError(res.Err)
	res = s.Execute(
		"schedule", "update",
		"--address", s.Address(),
		"-s", schedId,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", schedWfId,
		"--memo", "bar=2",
	)
	s.NoError(res.Err)
	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", schedId,
			"-o", "json",
		)
		s.NoError(res.Err)

		var j struct {
			Schedule struct {
				Action struct {
					StartWorkflow struct {
						Memo struct {
							Fields struct {
								Bar struct {
									Data string `json:"data"`
								} `json:"bar"`
							} `json:"fields"`
						} `json:"memo"`
					} `json:"startWorkflow"`
				} `json:"action"`
			} `json:"schedule"`
		}
		s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))

		// Decoded 'Mg==' is 2
		return j.Schedule.Action.StartWorkflow.Memo.Fields.Bar.Data == "Mg=="
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_ListMatchingTimes() {
	// use a calendar spec with known hours so results are deterministic
	schedId, _, res := s.createSchedule("--calendar", `{"hour":"3,6,9"}`)
	s.NoError(res.Err)

	// query a full day - should match exactly 3 times
	res = s.Execute(
		"schedule", "list-matching-times",
		"--address", s.Address(),
		"-s", schedId,
		"--start-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-01-01T23:59:59Z",
	)
	s.NoError(res.Err)
	// text output should contain parseable RFC3339 timestamps
	for _, line := range strings.Split(res.Stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Time" {
			continue
		}
		_, err := time.Parse(time.RFC3339, line)
		s.NoError(err, "should parse text line as time: %q", line)
	}

	// json output
	res = s.Execute(
		"schedule", "list-matching-times",
		"--address", s.Address(),
		"-s", schedId,
		"--start-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-01-01T23:59:59Z",
		"-o", "json",
	)
	s.NoError(res.Err)
	var resp struct {
		StartTime []string `json:"startTime"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &resp))
	s.Equal(3, len(resp.StartTime))
}

func (s *SharedServerSuite) TestSchedule_ListMatchingTimes_NotFound() {
	res := s.Execute(
		"schedule", "list-matching-times",
		"--address", s.Address(),
		"-s", "nonexistent-schedule-id",
		"--start-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-01-01T23:59:59Z",
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestSchedule_TimeSkippingPropagatesWithSkipOverlap() {
	s.useTimeSkippingServer()

	// The first action is at the next future UTC hour; the fast-forward target
	// falls between its third and fourth actions.
	firstAction := time.Now().UTC().Truncate(time.Hour).Add(2 * time.Hour)
	scheduleID, _, res := s.createSchedule(
		"--interval", "1h",
		"--start-time", firstAction.Add(-time.Second).Format(time.RFC3339Nano),
		"--overlap-policy", "Skip",
	)
	s.NoError(res.Err)

	duration := time.Until(firstAction.Add(2*time.Hour + time.Minute))
	res = s.Execute(
		"schedule", "time-skipping", "enable",
		"--address", s.Address(),
		"--schedule-id", scheduleID,
		"--fast-forward-duration", duration.String(),
		"--fast-forward-id", "ThreeActionsWithPropagation",
		"--disable-propagation=false",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "time-skipping", "poll",
		"--address", s.Address(),
		"--schedule-id", scheduleID,
		"--fast-forward-id", "ThreeActionsWithPropagation",
		"--timeout", "30s",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "completed")

	var description *client.ScheduleDescription
	s.EventuallyWithT(func(t *assert.CollectT) {
		var err error
		description, err = s.Client.ScheduleClient().GetHandle(s.Context, scheduleID).Describe(s.Context)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(t, 3, description.Info.NumActions)
		assert.Len(t, description.Info.RecentActions, 3)
		assert.Equal(t, 0, description.Info.NumActionsSkippedOverlap)
	}, 10*time.Second, 100*time.Millisecond)
	s.NotNil(description.Schedule.Policy)
	s.Equal(enums.SCHEDULE_OVERLAP_POLICY_SKIP, description.Schedule.Policy.Overlap)
	s.NotNil(description.Info.TimeSkippingInfo)
	s.False(description.Info.TimeSkippingInfo.GetEffectiveConfig().GetDisablePropagation())

	workflowIDs := make(map[string]struct{}, 3)
	for i, action := range description.Info.RecentActions {
		scheduledTime := firstAction.Add(time.Duration(i) * time.Hour)
		s.WithinDuration(scheduledTime, action.ScheduleTime, time.Second)
		s.NotNil(action.StartWorkflowResult)
		workflowID := action.StartWorkflowResult.WorkflowID
		_, seen := workflowIDs[workflowID]
		s.False(seen, "schedule action reused Workflow ID %q", workflowID)
		workflowIDs[workflowID] = struct{}{}
		runID := action.StartWorkflowResult.FirstExecutionRunID
		workflowDescription, err := s.Client.DescribeWorkflowExecution(s.Context, workflowID, runID)
		s.NoError(err)
		s.NotNil(workflowDescription.GetWorkflowExecutionInfo().GetStartTime())
		s.WithinDuration(scheduledTime, workflowDescription.GetWorkflowExecutionInfo().GetStartTime().AsTime(), 5*time.Second)
		s.NotNil(workflowDescription.GetWorkflowExtendedInfo().GetTimeSkippingInfo())
		s.True(workflowDescription.GetWorkflowExtendedInfo().GetTimeSkippingInfo().GetEffectiveConfig().GetEnabled())
	}
}
