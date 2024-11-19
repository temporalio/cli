package temporalcli_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.temporal.io/server/common/rpc"
	"go.temporal.io/server/common/searchattribute"
	"go.temporal.io/server/common/worker_versioning"
)

func (s *SharedServerSuite) awaitNextWorkflow(searchAttr string) {
	var lastExecs []*workflowpb.WorkflowExecutionInfo
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		lastExecs = resp.Executions
		return len(resp.Executions) == 2 && resp.Executions[0].Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	}, 3*time.Second, 100*time.Millisecond, "Reset execution failed to complete", lastExecs)
}

func (s *SharedServerSuite) TestWorkflow_Reset_ToFirstWorkflowTask() {
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

	// Reset to the first workflow task
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset-FirstWorkflowTask",
	)
	require.NoError(s.T(), res.Err)
	s.awaitNextWorkflow(searchAttr)
	s.Equal(2, wfExecutions, "Should have re-executed the workflow from the beginning")
	s.Greater(activityExecutions, 1, "Should have re-executed the workflow from the beginning")
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ToFirstWorkflowTask() {
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

	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"--query", fmt.Sprintf("CustomKeywordField = '%s'", searchAttr),
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset-FirstWorkflowTask",
	)
	require.NoError(s.T(), res.Err)
	s.awaitNextWorkflow(searchAttr)
	s.Equal(2, wfExecutions, "Should have re-executed the workflow from the beginning")
	s.Greater(activityExecutions, 1, "Should have re-executed the workflow from the beginning")
}

func (s *SharedServerSuite) TestWorkflow_Reset_ToLastWorkflowTask() {
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

	// Reset to the final workflow task
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-t", "LastWorkflowTask",
		"--reason", "test-reset-LastWorkflowTask",
	)
	require.NoError(s.T(), res.Err)
	s.awaitNextWorkflow(searchAttr)
	s.Equal(2, wfExecutions, "Should re-executed the workflow")
	s.Equal(1, activityExecutions, "Should not have re-executed the activity")
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ToLastWorkflowTask() {
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

	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"--query", fmt.Sprintf("CustomKeywordField = '%s'", searchAttr),
		"-t", "LastWorkflowTask",
		"--reason", "test-reset-LastWorkflowTask",
	)
	require.NoError(s.T(), res.Err)
	s.awaitNextWorkflow(searchAttr)
	s.Equal(2, wfExecutions, "Should re-executed the workflow")
	s.Equal(1, activityExecutions, "Should not have re-executed the activity")
}

func (s *SharedServerSuite) TestWorkflow_Reset_ToEventID() {
	// We execute two activities and will resume just before the second one. We use the same activity for both
	// but a unique input so we can check which fake activity is executed
	var oneExecutions, twoExecutions int
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		n, ok := a.(float64)
		if !ok {
			return nil, fmt.Errorf("expected float64, not %T (%v)", a, a)
		}
		switch n {
		case 1:
			oneExecutions++
		case 2:
			twoExecutions++
		default:
			return 0, fmt.Errorf("activity expected input 1 or 2 got %v", n)
		}
		return n, nil
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
	s.Equal(1, oneExecutions)
	s.Equal(1, twoExecutions)

	// We want to reset to the last WFTCompleted event before the second activity, so we
	// need to search the history for it. I could just pick the event ID, but I don't want
	// this test to break if new event types are added in the future
	wfsvc := s.Client.WorkflowService()
	ctx := context.Background()
	req := workflowservice.GetWorkflowExecutionHistoryReverseRequest{
		Namespace: s.Namespace(),
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: run.GetID(),
			RunId:      run.GetRunID(),
		},
		MaximumPageSize: 250,
		NextPageToken:   nil,
	}
	beforeSecondActivity := int64(-1)
	var takeNextWorkflowTaskCompleted bool
	for beforeSecondActivity == -1 {
		resp, err := wfsvc.GetWorkflowExecutionHistoryReverse(ctx, &req)
		s.NoError(err)
		for _, e := range resp.GetHistory().GetEvents() {
			s.T().Logf("Event: %d %s", e.GetEventId(), e.GetEventType())
			if e.GetEventType() == enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED && beforeSecondActivity == -1 {
				takeNextWorkflowTaskCompleted = true
			} else if e.GetEventType() == enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED && takeNextWorkflowTaskCompleted {
				beforeSecondActivity = e.EventId
				break
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}

	// Reset to the before the second activity execution
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-e", fmt.Sprintf("%d", beforeSecondActivity),
		"--reason", "test-reset-event-id",
	)
	require.NoError(s.T(), res.Err)

	s.awaitNextWorkflow(searchAttr)
	s.Equal(1, oneExecutions, "Should not have re-executed the first activity")
	s.Equal(2, twoExecutions, "Should have re-executed the second activity")
}

func (s *SharedServerSuite) TestBatchResetByBuildId() {
	sut := newSystemUnderTest(s)

	sut.startWorkerFor(originalWorkflow, workflowOptions{name: "wf", version: "v1"})
	sut.executeWorkflow("wf")
	sut.waitWorkflowBlockedAfterFirstActivity()
	sut.stopWorkerFor("v1")

	sut.startWorkerFor(extendedWorkflowWithBuggyActivity, workflowOptions{name: "wf", version: "v2"})
	sut.allowWorkflowToContinue()
	sut.waitBadActivityExecuted()
	sut.stopWorkerFor("v2")

	sut.startWorkerFor(extendedWorkflowWithNonDeterministicFix, workflowOptions{name: "wf", version: "v3"})
	sut.waitBlockOnNonDeterministicError()

	query := fmt.Sprintf("WorkflowId = \"%s\"", sut.run.GetID())
	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"--reason", "test-reset-event-id",
		"--type", "BuildId",
		"--build-id", sut.buildPrefix+"v2",
		"--query", query,
	)
	require.NoError(s.T(), res.Err)

	sut.assertWorkflowComplete()
	sut.assertOnlySecondActivityRetried()
	sut.stopWorkerFor("v3")
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_OnlyMatchingQuery() {
	var resetWfExecutions, resetActivityExecutions int
	var nonResetWfExecutions, nonResetActivityExecutions int
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		isReset, ok := a.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, not %T (%v)", a, a)
		}
		if isReset {
			resetActivityExecutions++
		} else {
			nonResetActivityExecutions++
		}
		return nil, nil
	})
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		workflow.ExecuteActivity(ctx, DevActivity, a).Get(ctx, nil)
		isReset, ok := a.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, not %T (%v)", a, a)
		}
		if isReset {
			resetWfExecutions++
		} else {
			nonResetWfExecutions++
		}
		return nil, nil
	})

	resetSearchAttr := "keyword-" + uuid.NewString()
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": resetSearchAttr},
		},
		DevWorkflow,
		true,
	)
	s.NoError(err)
	var junk any
	s.NoError(run.Get(s.Context, &junk))
	s.Equal(1, resetWfExecutions)

	nonResetSearchAttr := "keyword-" + uuid.NewString()
	nonResetRun, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:        s.Worker().Options.TaskQueue,
			SearchAttributes: map[string]any{"CustomKeywordField": nonResetSearchAttr},
		},
		DevWorkflow,
		false,
	)
	s.NoError(err)
	s.NoError(nonResetRun.Get(s.Context, &junk))
	s.Equal(1, nonResetWfExecutions)

	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"--query", fmt.Sprintf("CustomKeywordField = '%s'", resetSearchAttr),
		"-t", "FirstWorkflowTask",
		"--reason", "test-reset-FirstWorkflowTask",
	)
	require.NoError(s.T(), res.Err)
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + resetSearchAttr + "'" + " OR " + "CustomKeywordField = '" + nonResetSearchAttr + "'",
		})
		s.NoError(err)
		if len(resp.Executions) != 3 {
			return false
		}
		for _, exec := range resp.Executions {
			if exec.Status != enums.WORKFLOW_EXECUTION_STATUS_COMPLETED {
				return false
			}
		}
		return true
	}, 3*time.Second, 100*time.Millisecond)
	s.Equal(2, resetWfExecutions, "Should have re-executed the workflow from the beginning")
	s.Equal(2, resetActivityExecutions, "Should have re-executed the workflow from the beginning")
	s.Equal(1, nonResetWfExecutions, "Should not have re-executed the non-matching workflow")
	s.Equal(1, nonResetActivityExecutions, "Should not have re-executed the non-matching workflow")
}

type WorkflowResetTest struct {
	s                      *SharedServerSuite
	reapplyType            string
	reapplyExclude         []string
	expectUpdatesReapplied bool
	expectSignalsReapplied bool
}

func (s *SharedServerSuite) TestWorkflow_Reset_DefaultReappliesAll() {
	t := WorkflowResetTest{
		s:                      s,
		expectUpdatesReapplied: true,
		expectSignalsReapplied: true,
	}
	t.runSingleReset()
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ReappliesAll() {
	t := WorkflowResetTest{
		s:                      s,
		expectUpdatesReapplied: true,
		expectSignalsReapplied: true,
	}
	t.runBatchReset()
}

func (s *SharedServerSuite) TestWorkflow_Reset_ExcludeUpdate() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyExclude:         []string{"Update"},
		expectUpdatesReapplied: false,
		expectSignalsReapplied: true,
	}
	t.runSingleReset()
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ExcludeUpdate() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyExclude:         []string{"Update"},
		expectUpdatesReapplied: false,
		expectSignalsReapplied: true,
	}
	t.runBatchReset()
}

func (s *SharedServerSuite) TestWorkflow_Reset_ExcludeSignal() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyExclude:         []string{"Signal"},
		expectUpdatesReapplied: true,
		expectSignalsReapplied: false,
	}
	t.runSingleReset()
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ExcludeSignal() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyExclude:         []string{"Signal"},
		expectUpdatesReapplied: true,
		expectSignalsReapplied: false,
	}
	t.runBatchReset()
}

func (s *SharedServerSuite) TestWorkflow_Reset_ExcludeSignalAndUpdate() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyExclude:         []string{"Signal", "Update"},
		expectUpdatesReapplied: false,
		expectSignalsReapplied: false,
	}
	t.runSingleReset()
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ExcludeSignalAndUpdate() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyExclude:         []string{"Signal", "Update"},
		expectUpdatesReapplied: false,
		expectSignalsReapplied: false,
	}
	t.runBatchReset()
}

func (s *SharedServerSuite) TestWorkflow_Reset_ReapplySignalOnly() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyType:            "Signal",
		expectUpdatesReapplied: false,
		expectSignalsReapplied: true,
	}
	t.runSingleReset()
}

func (s *SharedServerSuite) TestWorkflow_ResetBatch_ReapplySignalOnly() {
	t := WorkflowResetTest{
		s:                      s,
		reapplyType:            "Signal",
		expectUpdatesReapplied: false,
		expectSignalsReapplied: true,
	}
	t.runBatchReset()
}

func (t *WorkflowResetTest) runSingleReset() {
	t.run(false)
}

func (t *WorkflowResetTest) runBatchReset() {
	t.run(true)
}

func (t *WorkflowResetTest) run(resetBatch bool) {
	s := t.s
	var wfExecutions, updateHandlerExecutions, signalHandlerExecutions int
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {

		// Handle signals
		sigChan := workflow.GetSignalChannel(ctx, "mySignal")
		workflow.Go(ctx, func(ctx workflow.Context) {
			for {
				sigChan.Receive(ctx, nil)
				signalHandlerExecutions++
			}
		})
		// Handle updates
		workflow.SetUpdateHandler(ctx, "myUpdate", func(ctx workflow.Context) error {
			updateHandlerExecutions++
			return nil
		})
		err := workflow.Sleep(ctx, 100*time.Millisecond)
		if err != nil {
			return nil, err
		}
		wfExecutions++
		return nil, nil
	})

	searchAttr := "keyword-" + uuid.NewString()
	run := t.startWorkflowAndSendTwoSignalsAndTwoUpdates(searchAttr)
	s.Equal(2, updateHandlerExecutions)
	s.Equal(2, signalHandlerExecutions)
	s.Equal(1, wfExecutions)

	if resetBatch {
		t.resetBatchWorkflow(searchAttr)
	} else {
		t.resetWorkflow(run.GetID())
	}
	s.awaitNextWorkflow(searchAttr)

	if t.expectUpdatesReapplied {
		s.Equal(4, updateHandlerExecutions)
	} else {
		s.Equal(2, updateHandlerExecutions)
	}
	if t.expectSignalsReapplied {
		s.Equal(4, signalHandlerExecutions)
	} else {
		s.Equal(2, signalHandlerExecutions)
	}
	s.Equal(2, wfExecutions)
}

func (t *WorkflowResetTest) startWorkflowAndSendTwoSignalsAndTwoUpdates(searchAttr string) client.WorkflowRun {
	s := t.s
	// Start the workflow
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

	// Wait for the workflow to start before sending signals/updates.
	// This has to be done, as batch reset with type `FirstWorkflowTask`, will reset to first workflow task completed, so the first signal
	// sent before the workflow starts, will be reapplied, as the reset point is later in the history.
	// The same would happen with single reset to eventId 4.
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) > 0
	}, 3*time.Second, 100*time.Millisecond, "Workflow failed to start")

	// before sending signals, we wait for the workflow to execute the activity
	for i := 1; i <= 2; i++ {
		s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), run.GetRunID(), "mySignal", fmt.Sprintf("%d", i)))
		updateHandle, err := s.Client.UpdateWorkflow(s.Context, client.UpdateWorkflowOptions{
			WorkflowID:   run.GetID(),
			RunID:        run.GetRunID(),
			UpdateName:   "myUpdate",
			WaitForStage: client.WorkflowUpdateStageAccepted,
		})
		s.NoError(err)
		s.NoError(updateHandle.Get(s.Context, nil))
	}
	s.NoError(run.Get(s.Context, nil))
	return run
}

func (t *WorkflowResetTest) resetWorkflow(workflowID string) {
	s := t.s
	// Reset to the beginning
	args := []string{
		"workflow", "reset",
		"--address", s.Address(),
		"-w", workflowID,
		"--event-id", "3",
		"--reason", "test-workflow-reset",
	}
	if len(t.reapplyExclude) > 0 && t.reapplyType != "" {
		panic("--reapply-type cannot be used with --reapply-exclude")
	}
	if t.reapplyType != "" {
		args = append(args, "--reapply-type", t.reapplyType)
	}
	if len(t.reapplyExclude) > 0 {
		for _, exclude := range t.reapplyExclude {
			args = append(args, "--reapply-exclude", exclude)
		}
	}
	res := s.Execute(args...)
	require.NoError(s.T(), res.Err)
}

func (t *WorkflowResetTest) resetBatchWorkflow(searchAttr string) {
	s := t.s
	args := []string{
		"workflow", "reset",
		"--address", s.Address(),
		"--query", fmt.Sprintf("CustomKeywordField = '%s'", searchAttr),
		"--type", "FirstWorkflowTask",
		"--reason", "test-workflow-reset",
	}
	if len(t.reapplyExclude) > 0 && t.reapplyType != "" {
		panic("--reapply-type cannot be used with --reapply-exclude")
	}
	if t.reapplyType != "" {
		args = append(args, "--reapply-type", t.reapplyType)
	}
	if len(t.reapplyExclude) > 0 {
		for _, exclude := range t.reapplyExclude {
			args = append(args, "--reapply-exclude", exclude)
		}
	}
	s.CommandHarness.Stdin.WriteString("y\n")
	res := s.Execute(args...)
	require.NoError(s.T(), res.Err)
}

func (s *SharedServerSuite) TestWorkflow_Reset_DoesNotAllowBothReapplyOptions() {
	res := s.Execute(
		"workflow", "reset",
		"--address", s.Address(),
		"-w", "whatever",
		"--event-id", "3",
		"--reason", "test-reset-FirstWorkflowTask",
		"--reapply-exclude", "Signal",
		"--reapply-type", "Signal",
	)
	require.Error(s.T(), res.Err)
	s.Contains(res.Err.Error(), "--reapply-type cannot be used with --reapply-exclude")
}

const (
	badActivity = iota
	firstActivity
	secondActivity
	thirdActivity
)

type assertions interface {
	NoError(error, ...interface{})
	Error(error, ...interface{})
	True(bool, ...interface{})
	Eventually(func() bool, time.Duration, time.Duration, ...interface{})
}

type batchResetTestData struct {
	t           *testing.T
	buildPrefix string
	assert      assertions
	client      client.Client
	tq          string
	counters    []*atomic.Int32
	run         client.WorkflowRun
	ctx         context.Context
	ctxCancelF  context.CancelFunc
	workers     map[string]worker.Worker
	namespace   string
}

func newSystemUnderTest(suite *SharedServerSuite) *batchResetTestData {
	sut := batchResetTestData{
		t:           suite.T(),
		counters:    []*atomic.Int32{&atomic.Int32{}, &atomic.Int32{}, &atomic.Int32{}, &atomic.Int32{}},
		assert:      suite,
		client:      suite.Client,
		namespace:   suite.Namespace(),
		buildPrefix: uuid.NewString()[:6] + "-",
		tq:          suite.T().Name(),
		workers:     make(map[string]worker.Worker),
	}

	suite.T().Cleanup(func() { sut.stopAllWorkers() })

	sut.ctx, sut.ctxCancelF = context.WithTimeout(context.Background(), 30*time.Second)
	return &sut
}

func (sut *batchResetTestData) internalVersionFor(v string) string {
	return sut.buildPrefix + v
}

func (sut *batchResetTestData) firstActivity() error {
	sut.counters[firstActivity].Add(1)
	return nil
}

func (sut *batchResetTestData) secondActivity() error {
	sut.counters[secondActivity].Add(1)
	return nil
}

func (sut *batchResetTestData) thirdActivity() error {
	sut.counters[thirdActivity].Add(1)
	return nil
}

func (sut *batchResetTestData) badActivity() error {
	sut.counters[badActivity].Add(1)
	return nil
}

type workflowOptions struct {
	name    string
	version string
}

func (sut *batchResetTestData) startWorkerFor(wf VersionedTestWorkflow, options workflowOptions) {
	sut.t.Helper()
	sut.workers[options.version] = worker.New(sut.client, sut.tq, worker.Options{BuildID: sut.internalVersionFor(options.version)})
	sut.workers[options.version].RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{Name: options.name})
	sut.workers[options.version].RegisterActivityWithOptions(sut.firstActivity, activity.RegisterOptions{Name: "firstActivity"})
	sut.workers[options.version].RegisterActivityWithOptions(sut.secondActivity, activity.RegisterOptions{Name: "secondActivity"})
	sut.workers[options.version].RegisterActivityWithOptions(sut.thirdActivity, activity.RegisterOptions{Name: "thirdActivity"})
	sut.workers[options.version].RegisterActivityWithOptions(sut.badActivity, activity.RegisterOptions{Name: "badActivity"})
	err := sut.workers[options.version].Start()
	sut.assert.NoError(err, "Could not start worker for %v, workflowVersion %s", wf, options.version)
}

func (sut *batchResetTestData) stopWorkerFor(version string) {
	if w, ok := sut.workers[version]; ok {
		w.Stop()
		delete(sut.workers, version)
	}
}

func (sut *batchResetTestData) stopAllWorkers() {
	for version := range sut.workers {
		sut.stopWorkerFor(version)
	}
}

func (sut *batchResetTestData) executeWorkflow(workflowName string) {
	sut.t.Helper()
	run, err := sut.client.ExecuteWorkflow(sut.ctx, client.StartWorkflowOptions{TaskQueue: sut.tq}, workflowName)
	sut.assert.NoError(err, "Failed to execute workflow %s", workflowName)
	sut.run = run
}

type VersionedTestWorkflow = func(workflow.Context, any) (string, error)

func originalWorkflow(ctx workflow.Context, _ any) (string, error) {
	ao := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{ScheduleToCloseTimeout: 5 * time.Second})
	if err := workflow.ExecuteActivity(ao, "firstActivity").Get(ctx, nil); err != nil {
		return "original workflow, first activity", err
	}

	ch := workflow.GetSignalChannel(ctx, "wait")
	ch.Receive(ctx, nil)

	if err := workflow.ExecuteActivity(ao, "secondActivity").Get(ctx, nil); err != nil {
		return "first workflow, second activity", fmt.Errorf("firstWorkflow:secondActivity:%v", err)
	}
	return "done 1!", nil
}

func extendedWorkflowWithBuggyActivity(ctx workflow.Context, arg any) (string, error) {
	if result, err := originalWorkflow(ctx, arg); err != nil {
		return "buggy workflow: " + result, fmt.Errorf("buggyWorkflow:firtstActivity:%v", err)
	}

	// (we run activity in a loop so that's visible in history, not just failing workflow tasks,
	// otherwise we wouldn't need a reset to "fix" it, just a new build would be enough.)
	ao := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{ScheduleToCloseTimeout: 5 * time.Second})
	for {
		if err := workflow.ExecuteActivity(ao, "badActivity").Get(ctx, nil); err != nil {
			return "buggy workflow, bad activity", fmt.Errorf("buggyWorkflow:badActivity:%v", err)
		}
		workflow.Sleep(ctx, time.Second)
	}
}

func extendedWorkflowWithNonDeterministicFix(ctx workflow.Context, arg any) (string, error) {
	if result, err := originalWorkflow(ctx, arg); err != nil {
		return "fixed workflow: " + result, fmt.Errorf("fixedWorkflow:firtstActivity:%v", err)
	}
	// introduce non-determinism by replacing badActivity with Sleep. (Replacing one activity with another does not
	// result in non-determinism)
	workflow.Sleep(ctx, time.Second)

	ao := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{ScheduleToCloseTimeout: 5 * time.Second})
	if err := workflow.ExecuteActivity(ao, "thirdActivity").Get(ctx, nil); err != nil {
		return "fixed workflow, third activity", fmt.Errorf("fixedWorkflow:thirdActivity:%v", err)
	}

	return "done 3!", nil
}

func (sut *batchResetTestData) waitWorkflowBlockedAfterFirstActivity() {
	sut.t.Helper()
	sut.assert.Eventually(
		func() bool {
			workflowHistory, err := sut.getWorkflowHistory()
			return err == nil && len(workflowHistory) >= 10
		},
		5*time.Second,
		100*time.Millisecond,
	)
}

func (sut *batchResetTestData) allowWorkflowToContinue() {
	sut.t.Helper()
	err := sut.client.SignalWorkflow(sut.ctx, sut.run.GetID(), sut.run.GetRunID(), "wait", nil)
	sut.assert.NoError(err, "failed to signal workflow to allow it to continue")
}

func (sut *batchResetTestData) waitBadActivityExecuted() {
	sut.t.Helper()
	sut.assert.Eventually(func() bool { return sut.counters[badActivity].Load() >= 3 }, 10*time.Second, 200*time.Millisecond)
}

func (sut *batchResetTestData) waitBlockOnNonDeterministicError() {
	sut.t.Helper()
	// but v3 is not quite compatible, the workflow should be blocked on non-determinism errors for now.
	waitCtx, cancel := context.WithTimeout(sut.ctx, 2*time.Second)
	defer cancel()
	sut.assert.Error(sut.run.Get(waitCtx, nil))

	// wait for it to appear in visibility
	query := fmt.Sprintf(`%s = "%s" and %s = "%s"`,
		searchattribute.ExecutionStatus, "Running",
		searchattribute.BuildIds, worker_versioning.UnversionedBuildIdSearchAttribute(sut.internalVersionFor("v2")))
	sut.assert.Eventually(func() bool {
		resp, err := sut.client.ListWorkflow(sut.ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: sut.namespace,
			Query:     query,
		})
		return err == nil && len(resp.Executions) == 1
	}, 10*time.Second, 500*time.Millisecond)

}

func (sut *batchResetTestData) assertWorkflowComplete() {
	sut.t.Helper()
	// need to loop since runid will be resolved early and we need to re-resolve to pick up
	// the new run instead of the terminated one
	sut.assert.Eventually(func() bool {
		var out string
		if sut.client.GetWorkflow(sut.ctx, sut.run.GetID(), "").Get(sut.ctx, &out) == nil {
			if out == "done 3!" {
				return true
			}
		}
		return false
	}, 10*time.Second, 200*time.Millisecond)
}

func (sut *batchResetTestData) assertOnlySecondActivityRetried() {
	sut.t.Helper()
	firstActivityAttempts := sut.counters[firstActivity].Load()
	secondActivityAttempts := sut.counters[secondActivity].Load()
	thirdActivityAttempts := sut.counters[thirdActivity].Load()
	sut.assert.True(
		int32(1) == firstActivityAttempts && int32(2) == secondActivityAttempts && int32(1) == thirdActivityAttempts,
		"expected only second activity restarted, got first attempts: %d, second attempts: %d, third attempts: %d",
		firstActivityAttempts,
		secondActivityAttempts,
		thirdActivityAttempts,
	)
}

func (sut *batchResetTestData) getWorkflowHistory() ([]*history.HistoryEvent, error) {
	ctx, _ := rpc.NewContextWithTimeoutAndVersionHeaders(90 * time.Second)
	iter := sut.client.GetWorkflowHistory(ctx, sut.run.GetID(), "", false, 1)

	events := make([]*history.HistoryEvent, 0, 20)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return events, err
		}
		events = append(events, event)
	}

	return events, nil
}
