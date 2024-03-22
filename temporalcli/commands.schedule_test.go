package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) schedId() string   { return s.testVar("sched") }
func (s *SharedServerSuite) schedWfId() string { return s.testVar("my-wf-id") }

func (s *SharedServerSuite) createSchedule(args ...string) *CommandResult {
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return nil, workflow.Sleep(ctx, 10*time.Second)
	})
	s.T().Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		options := temporalcli.CommandOptions{
			Stdout: io.Discard,
			Stderr: io.Discard,
			Args: []string{
				"schedule", "delete",
				"--address", s.Address(),
				"-s", s.schedId(),
			},
			Fail: func(error) {},
		}
		temporalcli.Execute(ctx, options)
	})
	return s.Execute(append([]string{
		"schedule", "create",
		"--address", s.Address(),
		"-s", s.schedId(),
		"--task-queue", s.Worker.Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", s.schedWfId(),
	}, args...)...,
	)
}

func (s *SharedServerSuite) TestSchedule_Create() {
	res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestSchedule_Delete() {
	res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// check exists
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "delete",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.NoError(res.Err)

	// doesn't exist anymore
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestSchedule_Describe() {
	res := s.createSchedule("--interval", "2s")
	s.NoError(res.Err)

	// run once manually so we see a running workflow

	res = s.Execute(
		"schedule", "trigger",
		"--address", s.Address(),
		"-s", s.schedId(),
	)

	// text

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", s.schedId(),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		s.ContainsOnSameLine(out, "ScheduleId", s.schedId())
		s.ContainsOnSameLine(out, "Spec", "2s")
		return AssertContainsOnSameLine(out, "RunningWorkflows", s.schedWfId()+"-") == nil
	}, 10*time.Second, 100*time.Millisecond)

	// json

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", s.schedId(),
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
	s.Equal(s.schedWfId(), j.Schedule.Action.StartWorkflow.Id)
}

func (s *SharedServerSuite) TestSchedule_CreateDescribeCalendar() {
	res := s.createSchedule("--calendar", `{"hour":"2,4","dayOfWeek":"thu,fri"}`)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ScheduleId", s.schedId())
	s.ContainsOnSameLine(out, "Spec", "dayOfWeek")
}

func (s *SharedServerSuite) TestSchedule_List() {
	res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// table

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "list",
			"--address", s.Address(),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		return AssertContainsOnSameLine(out, s.schedId(), "DevWorkflow", "false") == nil
	}, 10*time.Second, time.Second)

	// table long

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--long",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, s.schedId(), "DevWorkflow", "false")

	// table really-long

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--really-long",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, s.schedId(), "DevWorkflow", "0s" /*jitter*/, "false", "nil" /*memo*/)

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
		ok = ok || entry.ScheduleId == s.schedId()
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
		ok = ok || j.ScheduleId == s.schedId()
	}
	s.True(ok, "schedule not found in jsonl result")
}

func (s *SharedServerSuite) TestSchedule_Toggle() {
	res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// pause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", s.schedId(),
		"--pause",
		"--reason", "testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "true")
	s.ContainsOnSameLine(out, "Notes", "testing")

	// unpause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", s.schedId(),
		"--unpause",
		"--reason", "we're done testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "false")
	s.ContainsOnSameLine(out, "Notes", "done testing")
}

func (s *SharedServerSuite) TestSchedule_Trigger() {
	res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "trigger",
		"--address", s.Address(),
		"-s", s.schedId(),
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"workflow", "list",
			"--address", s.Address(),
			"-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, s.schedId()),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		return AssertContainsOnSameLine(out, s.schedWfId()) == nil
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_Backfill() {
	res := s.createSchedule("--interval", "10d/5h")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "backfill",
		"--address", s.Address(),
		"-s", s.schedId(),
		"--start-time", "2022-02-02T00:00:00Z",
		"--end-time", "2022-02-28T00:00:00Z",
		"--overlap-policy", "AllowAll",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"workflow", "list",
			"--address", s.Address(),
			"-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, s.schedId()),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		re := regexp.MustCompile(regexp.QuoteMeta(s.schedWfId() + "-2022-02"))
		return len(re.FindAllString(out, -1)) == 3
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_Update() {
	res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "update",
		"--address", s.Address(),
		"-s", s.schedId(),
		"--task-queue", "SomeOtherTq",
		"--type", "SomeOtherWf",
		"--workflow-id", s.schedWfId(),
		"--interval", "1h",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", s.schedId(),
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
