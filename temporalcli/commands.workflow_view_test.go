package temporalcli_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestWorkflow_Describe_ActivityFailing() {
	// Set activity to just continually error
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional error")
	})

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		})
		var res any
		err := workflow.ExecuteActivity(ctx, DevActivity, input).Get(ctx, &res)
		return res, err
	})

	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		return len(resp.PendingActivities) > 0 && resp.PendingActivities[0].LastFailure != nil
	}, 5*time.Second, 100*time.Millisecond)

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", run.GetID())
	s.Contains(out, "Pending Activities: 1")
	s.ContainsOnSameLine(out, "LastFailure", "intentional error")
	s.Contains(out, "Pending Child Workflows: 0")

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	var jsonOut workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
	s.Equal("intentional error", jsonOut.PendingActivities[0].LastFailure.Message)
}

func (s *SharedServerSuite) TestWorkflow_Describe_Completed() {
	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		map[string]string{"foo": "bar"},
	)
	s.NoError(err)
	s.NoError(run.Get(s.Context, nil))

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.ContainsOnSameLine(out, "Result", `{"foo":"bar"}`)

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.NotNil(jsonOut["closeEvent"])
	s.Equal(map[string]any{"foo": "bar"}, jsonOut["result"])
}

func (s *SharedServerSuite) TestWorkflow_Describe_Versioned() {
	buildIdTaskQueue := uuid.NewString()

	buildId := "id1"
	res := s.Execute(
		"task-queue", "versioning", "insert-assignment-rule",
		"--build-id", buildId,
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	var run client.WorkflowRun
	var err error
	s.Eventually(
		func() bool {
			run, err = s.Client.ExecuteWorkflow(
				s.Context,
				client.StartWorkflowOptions{TaskQueue: buildIdTaskQueue},
				DevWorkflow,
				map[string]string{"foo": "bar"},
			)
			s.NoError(err)

			// Text
			res = s.Execute(
				"workflow", "describe",
				"--address", s.Address(),
				"-w", run.GetID(),
			)
			s.NoError(res.Err)
			out := res.Stdout.String()
			return AssertContainsOnSameLine(out, "AssignedBuildId", buildId) == nil
		}, 10 * time.Second, 100 * time.Millisecond)

		// JSON
		res = s.Execute(
			"workflow", "describe",
			"-o", "json",
			"--address", s.Address(),
			"-w", run.GetID(),
		)
		s.NoError(res.Err)
		var jsonOut map[string]any
		s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
		execInfo, ok := jsonOut["workflowExecutionInfo"].(map[string]any)
		s.True(ok)
		s.Equal(buildId, execInfo["assignedBuildId"])
}

func (s *SharedServerSuite) TestWorkflow_Describe_ResetPoints() {
	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		map[string]string{"foo": "bar"},
	)
	s.NoError(err)
	s.NoError(run.Get(s.Context, nil))

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reset-points",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.NotContains(out, "Status")
	s.NotContains(out, "Result")
	s.Contains(out, "Auto Reset Points")

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reset-points",
	)
	s.NoError(res.Err)
	var jsonOut []map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.NotNil(jsonOut[0])
	s.NotNil(jsonOut[0]["EventId"])
}

func (s *SharedServerSuite) TestWorkflow_Show_Follow() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	doneFollowingCh := make(chan struct{})
	// Follow the workflow
	go func() {
		res := s.Execute(
			"workflow", "show",
			"--address", s.Address(),
			"-w", run.GetID(),
			"--follow",
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		s.Contains(out, "my-signal")
		s.Contains(out, "Result  \"hi!\"")
		close(doneFollowingCh)
	}()

	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))

	// Ensure following completes
	<-doneFollowingCh
	s.NoError(run.Get(s.Context, nil))
}

func (s *SharedServerSuite) TestWorkflow_Show_NoFollow() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	res := s.Execute(
		"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.NotContains(out, "my-signal")
	s.NotContains(out, "Results:")

	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(run.Get(s.Context, nil))

	res = s.Execute(
		"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, "my-signal")
	s.Contains(out, "Result  \"hi!\"")
}

func (s *SharedServerSuite) TestWorkflow_Show_JSON() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"workflow-param",
	)
	s.NoError(err)

	res := s.Execute(
		"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Contains(out, `"events": [`)
	s.Contains(out, `"eventType": "EVENT_TYPE_WORKFLOW_EXECUTION_STARTED"`)
	// Make sure payloads are still encoded non-shorthand
	s.Contains(out, base64.StdEncoding.EncodeToString([]byte(`"workflow-param"`)))

	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(run.Get(s.Context, nil))

	res = s.Execute(
		"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, `"events": [`)
	s.Contains(out, `"signalName": "my-signal"`)
	s.NotContains(out, "Results:")
}

func (s *SharedServerSuite) TestWorkflow_List() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		return a, nil
	})

	// Start the workflow
	for i := 0; i < 3; i++ {
		run, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
			DevWorkflow,
			strconv.Itoa(i),
		)
		s.NoError(err)
		s.NoError(run.Get(s.Context, nil))
	}

	res := s.Execute(
		"workflow", "list",
		"--address", s.Address(),
		"--query", fmt.Sprintf(`TaskQueue="%s"`, s.Worker().Options.TaskQueue),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Completed", "DevWorkflow")

	// JSON
	res = s.Execute(
		"workflow", "list",
		"--address", s.Address(),
		"--query", fmt.Sprintf(`TaskQueue="%s"`, s.Worker().Options.TaskQueue),
		"-o", "json",
	)
	s.NoError(res.Err)
	// Output is currently a series of JSON objects
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "name", "DevWorkflow")
	s.ContainsOnSameLine(out, "status", "WORKFLOW_EXECUTION_STATUS_COMPLETED")
}

func (s *SharedServerSuite) TestWorkflow_Count() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, shouldComplete any) (any, error) {
		// Only complete if shouldComplete is a true bool
		shouldCompleteBool, _ := shouldComplete.(bool)
		return nil, workflow.Await(ctx, func() bool { return shouldCompleteBool })
	})

	// Create 3 that complete and 2 that don't
	for i := 0; i < 5; i++ {
		_, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
			DevWorkflow,
			i < 3,
		)
		s.NoError(err)
	}

	// List and confirm they are all there in expected statuses
	s.Eventually(
		func() bool {
			resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
				Query: "TaskQueue = '" + s.Worker().Options.TaskQueue + "'",
			})
			s.NoError(err)
			var completed, running int
			for _, exec := range resp.Executions {
				if exec.Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED {
					completed++
				} else if exec.Status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
					running++
				}
			}
			return completed == 3 && running == 2
		},
		10*time.Second,
		100*time.Millisecond,
	)

	// Simple count w/out grouping
	res := s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"'",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Equal("Total: 5", strings.TrimSpace(out))

	// Grouped
	res = s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"' GROUP BY ExecutionStatus",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, "Total: 5")
	s.Contains(out, "Group total: 2, values: Running")
	s.Contains(out, "Group total: 3, values: Completed")

	// Simple count w/out grouping JSON
	res = s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"'",
		"-o", "json",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	// Proto JSON makes this count a string
	s.Contains(out, `"count": "5"`)

	// Grouped JSON
	res = s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"' GROUP BY ExecutionStatus",
		"-o", "jsonl",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, `"count":"5"`)
	s.Contains(out, `{"groupValues":["Running"],"count":"2"}`)
	s.Contains(out, `{"groupValues":["Completed"],"count":"3"}`)
}
