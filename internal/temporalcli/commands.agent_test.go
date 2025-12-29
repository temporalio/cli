package temporalcli_test

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/internal/agent"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestAgent_Timeline_BasicWorkflow() {
	// Create a simple workflow that completes
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		inputStr, _ := input.(string)
		return "result-" + inputStr, nil
	})

	// Start and wait for workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test-input",
	)
	s.NoError(err)

	// Wait for completion
	var result string
	s.NoError(run.Get(s.Context, &result))
	s.Equal("result-test-input", result)

	// Run timeline command
	res := s.Execute(
		"agent", "timeline",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)

	// Parse JSON output
	var timeline agent.TimelineResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &timeline))

	// Verify basic structure
	s.Equal(run.GetID(), timeline.Workflow.WorkflowID)
	s.Equal("Completed", timeline.Status)
	s.NotNil(timeline.StartTime)
	s.NotNil(timeline.CloseTime)
	s.Greater(len(timeline.Events), 0)

	// Check for workflow started and completed events
	foundStarted := false
	foundCompleted := false
	for _, ev := range timeline.Events {
		if ev.Type == "WorkflowExecutionStarted" {
			foundStarted = true
			s.Equal("workflow", ev.Category)
		}
		if ev.Type == "WorkflowExecutionCompleted" {
			foundCompleted = true
			s.Equal("workflow", ev.Category)
		}
	}
	s.True(foundStarted, "should have workflow started event")
	s.True(foundCompleted, "should have workflow completed event")
}

func (s *SharedServerSuite) TestAgent_Timeline_WithActivity() {
	// Create a workflow with an activity
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		var result string
		err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
			}),
			DevActivity,
			input,
		).Get(ctx, &result)
		return result, err
	})
	s.Worker().OnDevActivity(func(ctx context.Context, input any) (any, error) {
		inputStr, _ := input.(string)
		return "activity-" + inputStr, nil
	})

	// Start and wait for workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test",
	)
	s.NoError(err)

	var result string
	s.NoError(run.Get(s.Context, &result))

	// Run timeline command
	res := s.Execute(
		"agent", "timeline",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)

	var timeline agent.TimelineResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &timeline))

	// Check for activity events
	foundScheduled := false
	foundCompleted := false
	for _, ev := range timeline.Events {
		if ev.Category == "activity" {
			if ev.Status == "scheduled" {
				foundScheduled = true
			}
			if ev.Status == "completed" {
				foundCompleted = true
			}
		}
	}
	s.True(foundScheduled, "should have activity scheduled event")
	s.True(foundCompleted, "should have activity completed event")
}

func (s *SharedServerSuite) TestAgent_Trace_SimpleWorkflow() {
	// Create a simple failing workflow
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return "", temporal.NewApplicationError("test error", "TestError")
	})

	// Start and wait for workflow to fail
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test",
	)
	s.NoError(err)

	var result string
	err = run.Get(s.Context, &result)
	s.Error(err)

	// Run trace command
	res := s.Execute(
		"agent", "trace",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)

	var trace agent.TraceResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &trace))

	// Verify structure
	s.Equal(1, len(trace.Chain))
	s.Equal(run.GetID(), trace.Chain[0].WorkflowID)
	s.Equal("Failed", trace.Chain[0].Status)
	s.True(trace.Chain[0].IsLeaf)

	// Verify root cause
	s.NotNil(trace.RootCause)
	s.Equal("WorkflowFailed", trace.RootCause.Type)
	s.Contains(trace.RootCause.Error, "test error")
}

func (s *SharedServerSuite) TestAgent_Trace_WithChildWorkflow() {
	childWfType := "child-wf-" + uuid.NewString()

	// Register child workflow that fails
	s.Worker().Worker.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			return "", temporal.NewApplicationError("child error", "ChildError")
		},
		workflow.RegisterOptions{Name: childWfType},
	)

	// Register parent workflow that calls child
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		var result string
		err := workflow.ExecuteChildWorkflow(
			workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				WorkflowExecutionTimeout: 30 * time.Second,
			}),
			childWfType,
			input,
		).Get(ctx, &result)
		return result, err
	})

	// Start parent workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test",
	)
	s.NoError(err)

	// Wait for failure
	var result string
	err = run.Get(s.Context, &result)
	s.Error(err)

	// Run trace command
	res := s.Execute(
		"agent", "trace",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)

	var trace agent.TraceResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &trace))

	// Should have 2 nodes in chain (parent and child)
	s.Equal(2, len(trace.Chain))
	s.Equal(run.GetID(), trace.Chain[0].WorkflowID)
	s.Equal(0, trace.Chain[0].Depth)
	s.False(trace.Chain[0].IsLeaf)

	s.Equal(1, trace.Chain[1].Depth)
	s.True(trace.Chain[1].IsLeaf)
	s.Equal("Failed", trace.Chain[1].Status)

	// Root cause should be from child
	s.NotNil(trace.RootCause)
	s.Contains(trace.RootCause.Error, "child error")
}

func (s *SharedServerSuite) TestAgent_Failures_FindsFailedWorkflows() {
	searchAttr := "keyword-" + uuid.NewString()

	// Create a failing workflow
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return "", temporal.NewApplicationError("failure test", "FailureTest")
	})

	// Start multiple failing workflows
	for i := 0; i < 3; i++ {
		run, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{
				TaskQueue:        s.Worker().Options.TaskQueue,
				SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
			},
			DevWorkflow,
			"test",
		)
		s.NoError(err)

		// Wait for failure
		var result string
		_ = run.Get(s.Context, &result)
	}

	// Wait for workflows to be visible
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "' AND ExecutionStatus = 'Failed'",
		})
		s.NoError(err)
		return len(resp.Executions) == 3
	}, 5*time.Second, 100*time.Millisecond)

	// Run failures command
	res := s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult))

	// Should find at least our 3 failures
	s.GreaterOrEqual(len(failuresResult.Failures), 3)

	// Verify structure of failures
	for _, f := range failuresResult.Failures {
		s.NotEmpty(f.RootWorkflow.WorkflowID)
		s.NotEmpty(f.Status)
	}
}

func (s *SharedServerSuite) TestAgent_Timeline_Compact() {
	// Create a simple completing workflow
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return "done", nil
	})

	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test",
	)
	s.NoError(err)

	var result string
	s.NoError(run.Get(s.Context, &result))

	// Run with compact flag
	res := s.Execute(
		"agent", "timeline",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--compact",
		"-o", "json",
	)
	s.NoError(res.Err)

	var timeline agent.TimelineResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &timeline))

	// In compact mode, should still have key events
	s.Greater(len(timeline.Events), 0)
}

