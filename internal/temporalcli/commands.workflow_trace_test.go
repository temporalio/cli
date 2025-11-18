package temporalcli_test

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestWorkflow_Trace_Summary() {
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
		"workflow", "trace",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()

	s.ContainsOnSameLine(out, "WorkflowId", run.GetID())
	s.ContainsOnSameLine(out, "RunId", run.GetRunID())
	s.ContainsOnSameLine(out, "Type", "DevWorkflow")
	s.ContainsOnSameLine(out, "Namespace", s.Namespace())
	s.ContainsOnSameLine(out, "TaskQueue", s.Worker().Options.TaskQueue)
}

func (s *SharedServerSuite) TestWorkflow_Trace_Complete() {
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
		"workflow", "trace",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()

	s.ContainsOnSameLine(out, "WorkflowId", run.GetID())
	s.Contains(out, "╪ ✓ DevWorkflow")
	s.Contains(out, fmt.Sprintf("wfId: %s, runId: %s", run.GetID(), run.GetRunID()))
	s.Contains(out, " │   ┼ ✓ DevActivity")
}

func (s *SharedServerSuite) TestWorkflow_Trace_Failure() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional error")
	})
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Second,
		})
		var res any
		err := workflow.ExecuteActivity(ctx, DevActivity, input).Get(ctx, &res)
		return res, err
	})

	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue, WorkflowExecutionTimeout: time.Second},
		DevWorkflow,
		map[string]string{"foo": "bar"},
	)
	s.NoError(err)
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		return resp.WorkflowExecutionInfo.Status == enums.WORKFLOW_EXECUTION_STATUS_FAILED
	}, 5*time.Second, 100*time.Millisecond)

	// Text
	res := s.Execute(
		"workflow", "trace",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()

	s.Contains(out, "╪ ! DevWorkflow")
	s.Contains(out, fmt.Sprintf("wfId: %s, runId: %s", run.GetID(), run.GetRunID()))
	s.Contains(out, " │   Failure: activity error")
	s.Contains(out, " │   ┼ ! DevActivity")
	s.Contains(out, " │   │   Failure: intentional error")
}

func (s *SharedServerSuite) TestWorkflow_Trace_Fold() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional error")
	})
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Second,
		})
		var res any
		err := workflow.ExecuteActivity(ctx, DevActivity, input).Get(ctx, &res)
		return res, err
	})

	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue, WorkflowExecutionTimeout: time.Second},
		DevWorkflow,
		map[string]string{"foo": "bar"},
	)
	s.NoError(err)
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		return resp.WorkflowExecutionInfo.Status == enums.WORKFLOW_EXECUTION_STATUS_FAILED
	}, 5*time.Second, 100*time.Millisecond)

	// Text
	res := s.Execute(
		"workflow", "trace",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()

	s.Contains(out, "╪ ! DevWorkflow")
	s.Contains(out, fmt.Sprintf("wfId: %s, runId: %s", run.GetID(), run.GetRunID()))
	s.Contains(out, " │   Failure: activity error")
	s.Contains(out, " │   ┼ ! DevActivity")
	s.Contains(out, " │   │   Failure: intentional error")
}
