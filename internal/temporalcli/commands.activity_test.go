package temporalcli_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
)

const (
	activityId   string = "dev-activity-id"
	activityType string = "DevActivity"
	identity     string = "MyIdentity"
)

func (s *SharedServerSuite) TestActivity_Complete() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	res := s.Execute(
		"activity", "complete",
		"--activity-id", activityId,
		"--workflow-id", wid,
		"--result", "\"complete-activity-result\"",
		"--identity", identity,
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	var actual string
	s.NoError(run.Get(s.Context, &actual))
	s.Equal("complete-activity-result", actual)

	started, completed, failed := s.getActivityEvents(wid, activityId)
	s.NotNil(started)
	s.Nil(failed)
	s.NotNil(completed)
	s.Equal("\"complete-activity-result\"", string(completed.Result.Payloads[0].GetData()))
	s.Equal(identity, completed.GetIdentity())
}

func (s *SharedServerSuite) TestActivity_Fail() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	detail := "{\"myKey\": \"myValue\"}"
	reason := "MyReason"
	identity := "MyIdentity"
	res := s.Execute(
		"activity", "fail",
		"--activity-id", activityId,
		"--workflow-id", wid,
		"--run-id", run.GetRunID(),
		"--detail", detail,
		"--reason", reason,
		"--identity", identity,
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	err := run.Get(s.Context, nil)
	s.NotNil(err)

	started, completed, failed := s.getActivityEvents(wid, activityId)
	s.NotNil(started)
	s.Nil(completed)
	s.NotNil(failed)
	s.Equal(
		detail,
		string(failed.GetFailure().GetApplicationFailureInfo().GetDetails().Payloads[0].GetData()),
	)
	s.Equal(reason, failed.GetFailure().Message)
	s.Equal(identity, failed.GetIdentity())
}

func (s *SharedServerSuite) TestActivity_Complete_InvalidResult() {
	run := s.waitActivityStarted()
	res := s.Execute(
		"activity", "complete",
		"--activity-id", activityId,
		"--workflow-id", run.GetID(),
		"--result", "{not json}",
		"--address", s.Address(),
	)
	s.ErrorContains(res.Err, "is not valid JSON")

	started, completed, failed := s.getActivityEvents(run.GetID(), activityId)
	s.Nil(started)
	s.Nil(completed)
	s.Nil(failed)
}

func (s *SharedServerSuite) TestActivity_Fail_InvalidDetail() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	res := s.Execute(
		"activity", "fail",
		"--activity-id", activityId,
		"--workflow-id", wid,
		"--detail", "{not json}",
		"--address", s.Address(),
	)
	s.ErrorContains(res.Err, "is not valid JSON")

	started, completed, failed := s.getActivityEvents(wid, activityId)
	s.Nil(started)
	s.Nil(completed)
	s.Nil(failed)
}

func (s *SharedServerSuite) TestActivityOptionsUpdate_Accept() {
	run := s.waitActivityStarted()
	wid := run.GetID()

	res := s.Execute(
		"activity", "update-options",
		"--activity-id", activityId,
		"--workflow-id", wid,
		"--run-id", run.GetRunID(),
		"--identity", identity,
		"--task-queue", "new-task-queue",
		"--schedule-to-close-timeout", "60s",
		"--schedule-to-start-timeout", "5s",
		"--start-to-close-timeout", "10s",
		"--heartbeat-timeout", "20s",
		"--retry-initial-interval", "5s",
		"--retry-maximum-interval", "60s",
		"--retry-backoff-coefficient", "2",
		"--retry-maximum-attempts", "5",
		"--address", s.Address(),
	)

	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ScheduleToCloseTimeout", "1m0s")
	s.ContainsOnSameLine(out, "ScheduleToStartTimeout", "5s")
	s.ContainsOnSameLine(out, "StartToCloseTimeout", "10s")
	s.ContainsOnSameLine(out, "HeartbeatTimeout", "10s")
	s.ContainsOnSameLine(out, "InitialInterval", "5s")
	s.ContainsOnSameLine(out, "MaximumInterval", "1m0s")
	s.ContainsOnSameLine(out, "BackoffCoefficient", "2")
	s.ContainsOnSameLine(out, "MaximumAttempts", "5")
}

func (s *SharedServerSuite) TestActivityOptionsUpdate_Partial() {
	run := s.waitActivityStarted()

	res := s.Execute(
		"activity", "update-options",
		"--activity-id", activityId,
		"--workflow-id", run.GetID(),
		"--run-id", run.GetRunID(),
		"--identity", identity,
		"--task-queue", "new-task-queue",
		"--schedule-to-close-timeout", "41s",
		"--schedule-to-start-timeout", "11s",
		"--retry-initial-interval", "4s",
		"--retry-maximum-attempts", "10",
		"--address", s.Address(),
	)

	s.NoError(res.Err)
	out := res.Stdout.String()

	// updated
	s.ContainsOnSameLine(out, "ScheduleToCloseTimeout", "41s")
	s.ContainsOnSameLine(out, "ScheduleToStartTimeout", "11s")
	s.ContainsOnSameLine(out, "StartToCloseTimeout", "10s")
	s.ContainsOnSameLine(out, "InitialInterval", "4s")
	s.ContainsOnSameLine(out, "MaximumAttempts", "10")

	// old value
	// note - this is a snapshot of current values
	// if this test fails, check the default values of activity options
	s.ContainsOnSameLine(out, "StartToCloseTimeout", "10s")
	s.ContainsOnSameLine(out, "HeartbeatTimeout", "0s")
	s.ContainsOnSameLine(out, "MaximumInterval", "1m40s")
	s.ContainsOnSameLine(out, "BackoffCoefficient", "2")
}

func sendActivityCommand(command string, run client.WorkflowRun, s *SharedServerSuite, extraArgs ...string) *CommandResult {
	args := []string{
		"activity", command,
		"--workflow-id", run.GetID(),
		"--run-id", run.GetRunID(),
		"--identity", identity,
		"--address", s.Address(),
	}

	args = append(args, extraArgs...)

	res := s.Execute(args...)
	return res
}

func (s *SharedServerSuite) TestActivityPauseUnpause() {
	run := s.waitActivityStarted()

	res := sendActivityCommand("pause", run, s, "--activity-id", activityId)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		if resp.GetPendingActivities() == nil {
			return false
		}
		return len(resp.PendingActivities) > 0 && resp.PendingActivities[0].Paused
	}, 5*time.Second, 100*time.Millisecond)

	res = sendActivityCommand("unpause", run, s, "--activity-id", activityId, "--reset-attempts")
	s.NoError(res.Err)

	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		if resp.GetPendingActivities() == nil {
			return false
		}
		return len(resp.PendingActivities) > 0 && !resp.PendingActivities[0].Paused
	}, 5*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestActivityPauseUnpauseByType() {
	run := s.waitActivityStarted()
	res := sendActivityCommand("pause", run, s, "--activity-type", activityType)
	s.NoError(res.Err)

	res = sendActivityCommand("unpause", run, s, "--activity-type", activityType, "--reset-attempts")
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestActivityCommandFailed_NoActivityTpeOrId() {
	run := s.waitActivityStarted()

	commands := []string{"pause", "unpause", "reset"}
	for _, command := range commands {
		// should fail because both activity-id and activity-type are not provided
		res := sendActivityCommand(command, run, s)
		s.Error(res.Err)
	}
}

func (s *SharedServerSuite) TestActivityCommandFailed_BothActivityTpeOrId() {
	run := s.waitActivityStarted()

	commands := []string{"pause", "unpause", "reset"}
	for _, command := range commands {
		res := sendActivityCommand(command, run, s, "--activity-id", activityId, "--activity-type", activityType)
		s.Error(res.Err)
	}
}

func (s *SharedServerSuite) TestActivityReset() {
	run := s.waitActivityStarted()

	res := sendActivityCommand("reset", run, s, "--activity-id", activityId)
	s.NoError(res.Err)
	// make sure we receive a server response
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ServerResponse", "true")

	// reset should fail because activity is not found
	res = sendActivityCommand("reset", run, s, "--activity-id", "fake_id")
	s.Error(res.Err)
	// make sure we receive a NotFound error from the server`
	var notFound *serviceerror.NotFound
	s.ErrorAs(res.Err, &notFound)
}

// Test helpers

func (s *SharedServerSuite) waitActivityStarted() client.WorkflowRun {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		time.Sleep(0xFFFF * time.Hour)
		return nil, nil
	})
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
		return len(resp.PendingActivities) > 0
	}, 5*time.Second, 100*time.Millisecond)
	return run
}

func waitWorkflowStarted(s *SharedServerSuite) client.WorkflowRun {
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
		return len(resp.PendingActivities) > 0
	}, 5*time.Second, 100*time.Millisecond)
	return run
}

func (s *SharedServerSuite) getActivityEvents(workflowID, activityID string) (
	started *history.ActivityTaskStartedEventAttributes,
	completed *history.ActivityTaskCompletedEventAttributes,
	failed *history.ActivityTaskFailedEventAttributes,
) {
	iter := s.Client.GetWorkflowHistory(s.Context, workflowID, "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		s.NoError(err)
		if attrs := event.GetActivityTaskStartedEventAttributes(); attrs != nil {
			started = attrs
		} else if attrs := event.GetActivityTaskCompletedEventAttributes(); attrs != nil {
			completed = attrs
			s.Equal("json/plain", string(completed.Result.Payloads[0].Metadata["encoding"]))
		} else if attrs := event.GetActivityTaskFailedEventAttributes(); attrs != nil {
			failed = attrs
		}
	}
	return started, completed, failed
}

func checkActivitiesRunning(s *SharedServerSuite, run client.WorkflowRun) {
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		return len(resp.GetPendingActivities()) > 0
	}, 5*time.Second, 200*time.Millisecond)
}

func checkActivitiesPaused(s *SharedServerSuite, run client.WorkflowRun) {
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		if resp.GetPendingActivities() == nil {
			return false
		}
		return len(resp.GetPendingActivities()) > 0 && resp.GetPendingActivities()[0].Paused
	}, 5*time.Second, 200*time.Millisecond)
}

func (s *SharedServerSuite) TestUnpauseActivity_BatchSuccess() {
	var failActivity atomic.Bool
	failActivity.Store(true)
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		if failActivity.Load() {
			return nil, fmt.Errorf("update workflow received non-float input")
		}
		return nil, nil
	})

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		// override the activity options to allow activity to constantly fail
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ActivityID:          activityId,
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 0,
			},
		})
		var res any
		err := workflow.ExecuteActivity(ctx, DevActivity).Get(ctx, &res)
		return res, err
	})

	run1 := waitWorkflowStarted(s)
	run2 := waitWorkflowStarted(s)

	// Wait for all to appear in list
	query := fmt.Sprintf("WorkflowId = '%s' OR WorkflowId = '%s'", run1.GetID(), run2.GetID())
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: query,
		})
		s.NoError(err)
		return len(resp.Executions) == 2
	}, 3*time.Second, 100*time.Millisecond)

	// Pause the activities
	res := sendActivityCommand("pause", run1, s, "--activity-id", activityId)
	s.NoError(res.Err)
	res = sendActivityCommand("pause", run2, s, "--activity-id", activityId)
	s.NoError(res.Err)

	// wait for activities to be paused
	checkActivitiesPaused(s, run1)
	checkActivitiesPaused(s, run2)

	var lastRequestLock sync.Mutex
	var startBatchRequest *workflowservice.StartBatchOperationRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string, req, reply any,
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) error {
			lastRequestLock.Lock()
			if r, ok := req.(*workflowservice.StartBatchOperationRequest); ok {
				startBatchRequest = r
			}
			lastRequestLock.Unlock()
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)

	// Send batch activity unpause
	cmdRes := s.Execute("activity", "unpause",
		"--rps", "1",
		"--address", s.Address(),
		"--query", query,
		"--reason", "unpause-test",
		"--yes", "--match-all",
	)
	s.NoError(cmdRes.Err)
	s.NotEmpty(startBatchRequest.JobId)

	// check activities are running
	checkActivitiesRunning(s, run1)
	checkActivitiesRunning(s, run2)

	// unblock the activities to let them finish
	failActivity.Store(false)
}

func (s *SharedServerSuite) TestResetActivity_BatchSuccess() {
	var failActivity atomic.Bool
	failActivity.Store(true)
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		if failActivity.Load() {
			return nil, fmt.Errorf("update workflow received non-float input")
		}
		return nil, nil
	})

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		// override the activity options to allow activity to constantly fail
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ActivityID:          activityId,
			StartToCloseTimeout: 1 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 0,
			},
		})
		var res any
		err := workflow.ExecuteActivity(ctx, DevActivity).Get(ctx, &res)
		return res, err
	})

	run1 := waitWorkflowStarted(s)
	run2 := waitWorkflowStarted(s)

	// Wait for all to appear in list
	query := fmt.Sprintf("WorkflowId = '%s' OR WorkflowId = '%s'", run1.GetID(), run2.GetID())
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: query,
		})
		s.NoError(err)
		return len(resp.Executions) == 2
	}, 3*time.Second, 100*time.Millisecond)

	// Pause the activities
	res := sendActivityCommand("pause", run1, s, "--activity-id", activityId)
	s.NoError(res.Err)
	res = sendActivityCommand("pause", run2, s, "--activity-id", activityId)
	s.NoError(res.Err)

	// wait for activities to be paused
	checkActivitiesPaused(s, run1)
	checkActivitiesPaused(s, run2)

	var lastRequestLock sync.Mutex
	var startBatchRequest *workflowservice.StartBatchOperationRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string, req, reply any,
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) error {
			lastRequestLock.Lock()
			if r, ok := req.(*workflowservice.StartBatchOperationRequest); ok {
				startBatchRequest = r
			}
			lastRequestLock.Unlock()
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)

	// Send reset activity unpause
	cmdRes := s.Execute("activity", "reset",
		"--rps", "1",
		"--address", s.Address(),
		"--query", query,
		"--reason", "unpause-test",
		"--yes", "--match-all",
	)
	s.NoError(cmdRes.Err)
	s.NotEmpty(startBatchRequest.JobId)

	// check activities are running
	checkActivitiesRunning(s, run1)
	checkActivitiesRunning(s, run2)

	// unblock the activities to let them finish
	failActivity.Store(false)
}

func (s *SharedServerSuite) TestActivityExecute_RetriesOnEmptyPollResponse() {
	// Activity sleeps longer than the server's activity.longPollTimeout (2s),
	// forcing at least one empty poll response before the result arrives.
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		time.Sleep(3 * time.Second)
		return "standalone-result", nil
	})

	res := s.Execute(
		"activity", "execute",
		"--activity-id", "poll-retry-test",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "standalone-result")
}

func TestHelp_ActivitySubcommands(t *testing.T) {
	h := NewCommandHarness(t)

	res := h.Execute("help", "activity")
	assert.NoError(t, res.Err)
	out := res.Stdout.String()
	for _, sub := range []string{"cancel", "complete", "count", "describe", "execute", "fail", "list", "result", "start", "terminate"} {
		assert.Contains(t, out, sub, "missing subcommand %q in activity help", sub)
	}
}

func TestHelp_ActivityStartFlags(t *testing.T) {
	h := NewCommandHarness(t)

	res := h.Execute("activity", "start", "--help")
	assert.NoError(t, res.Err)
	out := res.Stdout.String()
	for _, flag := range []string{"--activity-id", "--type", "--task-queue", "--schedule-to-close-timeout", "--start-to-close-timeout", "--input"} {
		assert.Contains(t, out, flag, "missing flag %q in activity start help", flag)
	}
}

func TestHelp_ActivityCompleteFlags(t *testing.T) {
	h := NewCommandHarness(t)

	res := h.Execute("activity", "complete", "--help")
	assert.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "--activity-id")
	assert.Contains(t, out, "--workflow-id")
	assert.Contains(t, out, "--result")
}

func TestHelp_ActivityFailFlags(t *testing.T) {
	h := NewCommandHarness(t)

	res := h.Execute("activity", "fail", "--help")
	assert.NoError(t, res.Err)
	out := res.Stdout.String()
	assert.Contains(t, out, "--activity-id")
	assert.Contains(t, out, "--workflow-id")
	assert.Contains(t, out, "--detail")
	assert.Contains(t, out, "--reason")
}
