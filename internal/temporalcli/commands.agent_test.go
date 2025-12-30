package temporalcli_test

import (
	"context"
	"encoding/json"
	"strings"
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

func (s *SharedServerSuite) TestAgent_Failures_WithFollowChildren() {
	childWfType := "child-wf-failures-" + uuid.NewString()

	// Register child workflow that fails
	s.Worker().Worker.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input any) (any, error) {
			return "", temporal.NewApplicationError("deep child error", "DeepChildError")
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
	_ = run.Get(s.Context, &result)

	// Wait for workflow to be visible
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "WorkflowId = '" + run.GetID() + "' AND ExecutionStatus = 'Failed'",
		})
		s.NoError(err)
		return len(resp.Executions) == 1
	}, 5*time.Second, 100*time.Millisecond)

	// Run failures command with --follow-children
	res := s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--follow-children",
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult))

	// Find our specific failure
	var ourFailure *agent.FailureReport
	for i := range failuresResult.Failures {
		if failuresResult.Failures[i].RootWorkflow.WorkflowID == run.GetID() {
			ourFailure = &failuresResult.Failures[i]
			break
		}
	}

	s.NotNil(ourFailure, "should find our workflow in failures")
	s.Equal(1, ourFailure.Depth, "depth should be 1 (parent -> child)")
	s.Len(ourFailure.Chain, 2, "chain should have 2 elements")
	s.NotNil(ourFailure.LeafFailure, "should have leaf_failure populated")
	s.NotEqual(ourFailure.RootWorkflow.WorkflowID, ourFailure.LeafFailure.WorkflowID, "leaf should be different from root")
	s.Contains(ourFailure.RootCause, "deep child error", "root cause should mention the child error")
}

func (s *SharedServerSuite) TestAgent_Failures_ErrorContains() {
	uniqueError1 := "unique-error-" + uuid.NewString()
	uniqueError2 := "different-error-" + uuid.NewString()

	// Create workflows that fail with different error messages
	var errorToReturn string
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return "", temporal.NewApplicationError(errorToReturn, "TestError")
	})

	// Start workflow with first error
	errorToReturn = uniqueError1
	run1, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test1",
	)
	s.NoError(err)
	var result string
	_ = run1.Get(s.Context, &result)

	// Start workflow with second error
	errorToReturn = uniqueError2
	run2, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test2",
	)
	s.NoError(err)
	_ = run2.Get(s.Context, &result)

	// Wait for workflows to be visible
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "ExecutionStatus = 'Failed'",
		})
		s.NoError(err)
		return len(resp.Executions) >= 2
	}, 5*time.Second, 100*time.Millisecond)

	// Run failures command with --error-contains filtering for first error
	res := s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--error-contains", uniqueError1,
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult))

	// Should find exactly 1 failure with uniqueError1
	s.Len(failuresResult.Failures, 1, "should find exactly 1 failure with the specific error")
	s.Contains(failuresResult.Failures[0].RootCause, uniqueError1, "root cause should contain the filter string")

	// Run with case-insensitive matching
	res = s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--error-contains", strings.ToUpper(uniqueError1[:10]),
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult2 agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult2))

	// Should still find the failure (case-insensitive)
	s.Len(failuresResult2.Failures, 1, "should find failure with case-insensitive match")

	// Run with non-matching filter
	res = s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--error-contains", "nonexistent-error-string-xyz",
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult3 agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult3))

	// Should find no failures
	s.Len(failuresResult3.Failures, 0, "should find no failures with non-matching filter")
}

func (s *SharedServerSuite) TestAgent_Failures_MultipleStatuses() {
	// Create a workflow that can fail or be canceled
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return "", temporal.NewApplicationError("multi-status-test", "TestError")
	})

	// Start and wait for workflow to fail
	run1, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"test1",
	)
	s.NoError(err)
	var result string
	_ = run1.Get(s.Context, &result)

	// Wait for workflow to be visible
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "WorkflowId = '" + run1.GetID() + "' AND ExecutionStatus = 'Failed'",
		})
		s.NoError(err)
		return len(resp.Executions) == 1
	}, 5*time.Second, 100*time.Millisecond)

	// Test comma-separated statuses
	res := s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--status", "Failed,TimedOut",
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult))

	// Should find at least our 1 failure
	s.GreaterOrEqual(len(failuresResult.Failures), 1, "should find failures with comma-separated statuses")

	// Verify the query contains both statuses
	s.Contains(failuresResult.Query, "Failed", "query should contain Failed status")
	s.Contains(failuresResult.Query, "TimedOut", "query should contain TimedOut status")

	// Test multiple --status flags
	res = s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--status", "Failed",
		"--status", "TimedOut",
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult2 agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult2))

	// Should find the same failures
	s.GreaterOrEqual(len(failuresResult2.Failures), 1, "should find failures with multiple --status flags")
	s.Contains(failuresResult2.Query, "Failed", "query should contain Failed status")
	s.Contains(failuresResult2.Query, "TimedOut", "query should contain TimedOut status")

	// Test single status to verify filtering works
	res = s.Execute(
		"agent", "failures",
		"--address", s.Address(),
		"--since", "1h",
		"--status", "Canceled",
		"-o", "json",
	)
	s.NoError(res.Err)

	var failuresResult3 agent.FailuresResult
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &failuresResult3))

	// Should find no failures (we only have Failed workflows, not Canceled)
	s.Len(failuresResult3.Failures, 0, "should find no failures when filtering by Canceled only")
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
