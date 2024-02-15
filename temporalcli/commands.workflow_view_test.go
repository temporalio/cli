package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

// TODO(cretz): To test:
// * Workflow list
// * Workflow describe with just auto-reset points

func (s *SharedServerSuite) TestWorkflow_Describe_ActivityFailing() {
	// Set activity to just continually error
	s.Worker.OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional error")
	})

	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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

func (s *SharedServerSuite) TestWorkflow_Show_Follow() {
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
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
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
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
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
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
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
		DevWorkflow,
		"ignored",
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
