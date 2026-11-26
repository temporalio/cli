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

	"github.com/stretchr/testify/assert"
	"github.com/temporalio/cli/internal/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/workflow"
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

func (s *SharedServerSuite) TestSchedule_Update_SearchAttributes() {
	res := s.Execute("operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "ScheduleUpdateKeyword", "--type", "Keyword")
	s.NoError(res.Err)

	initial := map[string]string{
		"CustomKeywordField":    `"before"`,
		"ScheduleUpdateKeyword": `"remove-me"`,
	}
	for _, tc := range []struct {
		name string
		args []string
		want map[string]string
	}{
		{name: "preserve when omitted", want: initial},
		{
			name: "replace entire collection",
			args: []string{"--schedule-search-attribute", `CustomKeywordField="after"`},
			want: map[string]string{"CustomKeywordField": `"after"`},
		},
		{
			name: "replace multiple attributes",
			args: []string{"--schedule-search-attribute", `CustomKeywordField="after"`,
				"--schedule-search-attribute", `ScheduleUpdateKeyword="keep-me"`},
			want: map[string]string{"CustomKeywordField": `"after"`, "ScheduleUpdateKeyword": `"keep-me"`},
		},
		{
			name: "null removes a key",
			args: []string{"--schedule-search-attribute", `CustomKeywordField=null`,
				"--schedule-search-attribute", `ScheduleUpdateKeyword="keep-me"`},
			want: map[string]string{"ScheduleUpdateKeyword": `"keep-me"`},
		},
		{
			name: "only null clears collection",
			args: []string{"--schedule-search-attribute", `CustomKeywordField=null`},
			want: map[string]string{},
		},
		{
			name: "remove all",
			args: []string{"--remove-schedule-search-attributes"},
			want: map[string]string{},
		},
		{
			name: "false does not remove",
			args: []string{"--remove-schedule-search-attributes=false"},
			want: initial,
		},
	} {
		schedID, workflowID, res := s.createSchedule("--interval", "1h", "--paused",
			"--schedule-search-attribute", `CustomKeywordField="before"`,
			"--schedule-search-attribute", `ScheduleUpdateKeyword="remove-me"`,
			"--schedule-memo", `label="unchanged"`,
			"--search-attribute", `CustomKeywordField="workflow-value"`)
		s.NoError(res.Err)

		s.T().Run(tc.name, func(t *testing.T) {
			h := NewCommandHarness(t)
			res := h.Execute(append([]string{"schedule", "update",
				"--address", s.Address(),
				"--schedule-id", schedID,
				"--workflow-id", workflowID,
				"--task-queue", s.Worker().Options.TaskQueue,
				"--type", "DevWorkflow", "--interval", "1h", "--paused",
				"--search-attribute", `CustomKeywordField="workflow-value"`,
			}, tc.args...)...)
			h.NoError(res.Err)

			h.EventuallyWithT(func(t *assert.CollectT) {
				desc, err := s.Client.WorkflowService().DescribeSchedule(h.Context, &workflowservice.DescribeScheduleRequest{
					Namespace: s.Namespace(), ScheduleId: schedID,
				})
				if !assert.NoError(t, err) {
					return
				}
				got := make(map[string]string)
				for key, payload := range desc.GetSearchAttributes().GetIndexedFields() {
					got[key] = string(payload.GetData())
				}
				assert.Equal(t, tc.want, got)
				workflowFields := desc.GetSchedule().GetAction().GetStartWorkflow().GetSearchAttributes().GetIndexedFields()
				assert.Equal(t, `"workflow-value"`, string(workflowFields["CustomKeywordField"].GetData()))
				assert.Equal(t, `"unchanged"`, string(desc.GetMemo().GetFields()["label"].GetData()))
				assert.True(t, desc.GetSchedule().GetState().GetPaused())
			}, 10*time.Second, 100*time.Millisecond)

			h.EventuallyWithT(func(t *assert.CollectT) {
				res := h.Execute("schedule", "list", "--address", s.Address(),
					"--query", `CustomKeywordField = "after"`, "--output", "json")
				if !assert.NoError(t, res.Err) {
					return
				}
				var entries []struct{ ScheduleId string }
				if !assert.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &entries)) {
					return
				}
				found := false
				for _, entry := range entries {
					found = found || entry.ScheduleId == schedID
				}
				assert.Equal(t, tc.want["CustomKeywordField"] == `"after"`, found)
			}, 10*time.Second, 100*time.Millisecond)
		})
	}
}

func (s *SharedServerSuite) TestSchedule_Update_SearchAttributeTypes() {
	for _, attr := range []struct{ name, typ string }{
		{"ScheduleUpdateInt", "Int"},
		{"ScheduleUpdateDouble", "Double"},
		{"ScheduleUpdateBool", "Bool"},
		{"ScheduleUpdateText", "Text"},
		{"ScheduleUpdateDatetime", "Datetime"},
		{"ScheduleUpdateKeywordList", "KeywordList"},
	} {
		res := s.Execute("operator", "search-attribute", "create",
			"--address", s.Address(), "--name", attr.name, "--type", attr.typ)
		s.NoError(res.Err)
	}
	schedID, workflowID, res := s.createSchedule("--interval", "1h", "--paused")
	s.NoError(res.Err)
	res = s.Execute("schedule", "update",
		"--address", s.Address(), "--schedule-id", schedID,
		"--workflow-id", workflowID, "--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow", "--interval", "1h", "--paused",
		"--schedule-search-attribute", `CustomKeywordField="keyword"`,
		"--schedule-search-attribute", `ScheduleUpdateInt=42`,
		"--schedule-search-attribute", `ScheduleUpdateDouble=1.25`,
		"--schedule-search-attribute", `ScheduleUpdateBool=true`,
		"--schedule-search-attribute", `ScheduleUpdateText="some text"`,
		"--schedule-search-attribute", `ScheduleUpdateDatetime="2026-01-01T00:00:00Z"`,
		"--schedule-search-attribute", `ScheduleUpdateKeywordList=["one","two"]`)
	s.NoError(res.Err)
	s.EventuallyWithT(func(t *assert.CollectT) {
		desc, err := s.Client.WorkflowService().DescribeSchedule(s.Context, &workflowservice.DescribeScheduleRequest{
			Namespace: s.Namespace(), ScheduleId: schedID,
		})
		if !assert.NoError(t, err) {
			return
		}
		got := make(map[string]string)
		for key, payload := range desc.GetSearchAttributes().GetIndexedFields() {
			got[key] = string(payload.GetData())
		}
		assert.Equal(t, map[string]string{
			"CustomKeywordField":        `"keyword"`,
			"ScheduleUpdateInt":         `42`,
			"ScheduleUpdateDouble":      `1.25`,
			"ScheduleUpdateBool":        `true`,
			"ScheduleUpdateText":        `"some text"`,
			"ScheduleUpdateDatetime":    `"2026-01-01T00:00:00Z"`,
			"ScheduleUpdateKeywordList": `["one","two"]`,
		}, got)
	}, 10*time.Second, 100*time.Millisecond)
}

func TestScheduleUpdate_InvalidSearchAttributes(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		err  string
	}{
		{"conflicting flags", []string{"--remove-schedule-search-attributes", "--schedule-search-attribute", `Key="value"`}, "cannot combine"},
		{"missing equals", []string{"--schedule-search-attribute", "Key"}, "missing expected '='"},
		{"invalid JSON", []string{"--schedule-search-attribute", "Key=invalid"}, "invalid JSON value"},
		{"trailing JSON", []string{"--schedule-search-attribute", `Key="value" true`}, "unexpected trailing data"},
		{"object", []string{"--schedule-search-attribute", `Key={"nested":true}`}, "unsupported search attribute type"},
		{"non-string list", []string{"--schedule-search-attribute", `Key=[1]`}, "array element 0"},
		{"number overflow", []string{"--schedule-search-attribute", `Key=1e999`}, "invalid number"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewCommandHarness(t)
			// Invalid input must fail before attempting to connect to a server.
			res := h.Execute(append([]string{"schedule", "update",
				"--address", "127.0.0.1:1", "--schedule-id", "schedule",
				"--task-queue", "queue", "--type", "Workflow",
			}, tc.args...)...)
			h.ErrorContains(res.Err, tc.err)
			h.Empty(res.Stdout.String())
		})
	}
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
