package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) createSchedule() *CommandResult {
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
				"-s", "mysched",
			},
			Fail: func(error) {},
		}
		temporalcli.Execute(ctx, options)
	})
	return s.Execute(
		"schedule", "create",
		"--address", s.Address(),
		"-s", "mysched",
		"--interval", "2s",
		"--task-queue", s.Worker.Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id",
	)
}

func (s *SharedServerSuite) TestSchedule_Create() {
	res := s.createSchedule()
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestSchedule_Describe() {
	res := s.createSchedule()
	s.NoError(res.Err)

	time.Sleep(4 * time.Second) // run at least once

	// text

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", "mysched",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ScheduleId", "mysched")
	s.ContainsOnSameLine(out, "Spec", "2s")
	s.ContainsOnSameLine(out, "RunningWorkflows", "my-id-")

	// json

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", "mysched",
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
	s.Equal("my-id", j.Schedule.Action.StartWorkflow.Id)
}

func (s *SharedServerSuite) TestSchedule_List() {
	res := s.createSchedule()
	s.NoError(res.Err)

	// table

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "mysched", "DevWorkflow", "false")

	// table long

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--long",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "mysched", "DevWorkflow", "false")

	// table really-long

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--really-long",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "mysched", "DevWorkflow", "0s" /*jitter*/, "false", "nil" /*memo*/)

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
		ok = ok || entry.ScheduleId == "mysched"
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
		ok = ok || j.ScheduleId == "mysched"
	}
	s.True(ok, "schedule not found in jsonl result")
}

func (s *SharedServerSuite) TestSchedule_Toggle() {
	res := s.createSchedule()
	s.NoError(res.Err)

	// pause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", "mysched",
		"--pause",
		"--reason", "testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", "mysched",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "true")
	s.ContainsOnSameLine(out, "Notes", "testing")

	// unpause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", "mysched",
		"--unpause",
		"--reason", "we're done testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", "mysched",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "false")
	s.ContainsOnSameLine(out, "Notes", "done testing")
}
