package temporalcli_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_ValidatesArguments_MissingRequiredFlag() {
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", "test-workflow",
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset",
	)
	require.Error(s.T(), res.Err)
	require.Contains(s.T(), res.Err.Error(), "required flag")
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_ValidatesArguments_PinnedWithoutVersion() {
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", "test-workflow",
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset",
		"--versioning-override-behavior", "pinned",
	)
	require.Error(s.T(), res.Err)
	require.Contains(s.T(), res.Err.Error(), "missing version with 'pinned' behavior")
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_ValidatesArguments_AutoUpgradeWithVersion() {
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", "test-workflow",
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset",
		"--versioning-override-behavior", "auto_upgrade",
		"--versioning-override-pinned-version", "some-version",
	)
	require.Error(s.T(), res.Err)
	require.Contains(s.T(), res.Err.Error(), "cannot set pinned version with")
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_Single_AutoUpgradeBehavior() {
	var wfExecutions int
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		wfExecutions++
		return "result", nil
	})

	// Start the workflow
	searchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
		},
		DevWorkflow,
		"test-input",
	)
	s.NoError(err)
	var result any
	s.NoError(run.Get(s.Context, &result))
	s.Equal(1, wfExecutions)

	// Reset with auto upgrade versioning behavior
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset-with-auto-upgrade",
		"--versioning-override-behavior", "auto_upgrade",
	)
	require.NoError(s.T(), res.Err)

	// Wait for reset to complete
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == 2 && resp.Executions[0].Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	}, 3*time.Second, 100*time.Millisecond)

	s.Equal(2, wfExecutions, "Should have re-executed the workflow")
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_Single_PinnedBehavior() {
	var wfExecutions int
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		wfExecutions++
		return "result", nil
	})

	// Start the workflow
	searchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
		},
		DevWorkflow,
		"test-input",
	)
	s.NoError(err)
	var result any
	s.NoError(run.Get(s.Context, &result))
	s.Equal(1, wfExecutions)

	// Reset with pinned versioning behavior and properly formatted version
	pinnedVersion := "test-deployment.v1.0"
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset-with-pinned-version",
		"--versioning-override-behavior", "pinned",
		"--versioning-override-pinned-version", pinnedVersion,
	)
	require.NoError(s.T(), res.Err)

	// Wait for reset to complete
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		if len(resp.Executions) < 2 { // there should be two executions.
			return false
		}
		resetRunID := resp.Executions[0].Execution.RunId // the first result is the reset execution.
		descResult, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), resetRunID)
		s.NoError(err)
		s.NotNil(descResult)

		info := descResult.GetWorkflowExecutionInfo()
		pinnedVersionOverride := info.VersioningInfo.VersioningOverride.GetPinned().GetVersion()
		pinnedVersionOverrideString := pinnedVersionOverride.GetDeploymentName() + "." + pinnedVersionOverride.GetBuildId()
		return pinnedVersionOverrideString == pinnedVersion // the second execution should have the pinned version override.
	}, 5*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestWorkflow_ResetBatchWithWorkflowUpdateOptions_AutoUpgradeBehavior() {
	var wfExecutions int
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		wfExecutions++
		return "result", nil
	})

	// Start the workflow
	searchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
		},
		DevWorkflow,
		"test-input",
	)
	s.NoError(err)
	var result any
	s.NoError(run.Get(s.Context, &result))
	s.Equal(1, wfExecutions)

	// Reset batch with auto_upgrade versioning behavior
	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"--query", fmt.Sprintf("CustomKeywordField = '%s'", searchAttr),
		"-t", "FirstWorkflowTask",
		"--reason", "test-batch-reset-with-update-options",
		"--versioning-override-behavior", "auto_upgrade",
	)
	require.NoError(s.T(), res.Err)

	// Wait for batch reset to complete
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == 2 && resp.Executions[0].Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	}, 3*time.Second, 100*time.Millisecond)

	s.Equal(2, wfExecutions, "Should have re-executed the workflow from batch reset")
}

func (s *SharedServerSuite) TestWorkflow_ResetBatchWithWorkflowUpdateOptions_PinnedBehavior() {
	var wfExecutions int
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		wfExecutions++
		return "result", nil
	})

	// Start the workflow
	searchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
		},
		DevWorkflow,
		"test-input",
	)
	s.NoError(err)
	var result any
	s.NoError(run.Get(s.Context, &result))
	s.Equal(1, wfExecutions)

	// Reset batch with pinned versioning behavior and properly formatted version
	pinnedVersion := "batch-deployment.v1.0"
	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"--query", fmt.Sprintf("CustomKeywordField = '%s'", searchAttr),
		"-t", "FirstWorkflowTask",
		"--reason", "test-batch-reset-with-pinned-version",
		"--versioning-override-behavior", "pinned",
		"--versioning-override-pinned-version", pinnedVersion,
	)
	require.NoError(s.T(), res.Err)

	// Wait for batch reset to complete
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		if len(resp.Executions) < 2 { // there should be two executions.
			return false
		}
		resetRunID := resp.Executions[0].Execution.RunId // the first result is the reset execution.
		descResult, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), resetRunID)
		s.NoError(err)
		s.NotNil(descResult)

		info := descResult.GetWorkflowExecutionInfo()
		pinnedVersionOverride := info.VersioningInfo.VersioningOverride.GetPinned().GetVersion()
		pinnedVersionOverrideString := pinnedVersionOverride.GetDeploymentName() + "." + pinnedVersionOverride.GetBuildId()
		return pinnedVersionOverrideString == pinnedVersion // the second execution should have the pinned version override.
	}, 5*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_InheritsParentFlags() {
	// Test that the subcommand inherits parent flags correctly
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", "test-workflow",
		"-r", "test-run-id",
		"-e", "10",
		"--reason", "test-reset-with-inherited-flags",
		"--versioning-override-behavior", "auto_upgrade",
		"--reapply-exclude", "Signal",
	)

	// The command should fail because the workflow doesn't exist, but the error
	// should be about the missing workflow, not about invalid flags
	require.Error(s.T(), res.Err)
	require.NotContains(s.T(), res.Err.Error(), "required flag")
	require.NotContains(s.T(), res.Err.Error(), "invalid argument")
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_WithLastWorkflowTask() {
	var wfExecutions, activityExecutions int
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		activityExecutions++
		return nil, nil
	})
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		workflow.ExecuteActivity(ctx, DevActivity, 1).Get(ctx, nil)
		wfExecutions++
		return nil, nil
	})

	// Start the workflow
	searchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
		},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)
	var junk any
	s.NoError(run.Get(s.Context, &junk))
	s.Equal(1, wfExecutions)

	// Reset to the last workflow task with update options
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-t", "LastWorkflowTask",
		"--reason", "test-reset-last-workflow-task-with-options",
		"--versioning-override-behavior", "auto_upgrade",
	)
	require.NoError(s.T(), res.Err)

	// Wait for reset to complete
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == 2 && resp.Executions[0].Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	}, 3*time.Second, 100*time.Millisecond)

	s.Equal(2, wfExecutions, "Should re-executed the workflow")
	s.Equal(1, activityExecutions, "Should not have re-executed the activity")
}

func (s *SharedServerSuite) TestWorkflow_ResetWithWorkflowUpdateOptions_WithEventID() {
	// Test that the new subcommand works with event ID reset type
	var activityCount int
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		activityCount++
		return a, nil
	})

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		var res any
		if err := workflow.ExecuteActivity(ctx, DevActivity, 1).Get(ctx, &res); err != nil {
			return res, err
		}
		err := workflow.ExecuteActivity(ctx, DevActivity, 2).Get(ctx, &res)
		return res, err
	})

	// Start the workflow
	searchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
		},
		DevWorkflow,
		"ignored",
	)
	require.NoError(s.T(), err)
	var ignored any
	s.NoError(run.Get(s.Context, &ignored))
	s.Equal(2, activityCount)

	// Reset with event ID and update options
	res := s.Execute(
		"workflow", "reset", "with-workflow-update-options",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-e", "3", // Use a known early event ID
		"--reason", "test-reset-event-id-with-options",
		"--versioning-override-behavior", "auto_upgrade",
	)
	require.NoError(s.T(), res.Err)

	// Wait for reset to complete
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == 2 && resp.Executions[0].Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	}, 5*time.Second, 100*time.Millisecond)

	s.Greater(activityCount, 2, "Should have re-executed activities after reset")
}
