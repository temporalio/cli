package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

func (s *SharedServerSuite) TestStandaloneActivity_Complete() {
	activityStarted := make(chan struct{})
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		close(activityStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	started := s.startStandaloneActivity("sa-complete-test")
	<-activityStarted

	res := s.Execute(
		"activity", "complete",
		"--activity-id", "sa-complete-test",
		"--run-id", started["runId"].(string),
		"--result", `"completed-externally"`,
		"--address", s.Address(),
	)
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestStandaloneActivity_Fail() {
	activityStarted := make(chan struct{})
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		close(activityStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	started := s.startStandaloneActivity("sa-fail-test")
	<-activityStarted

	res := s.Execute(
		"activity", "fail",
		"--activity-id", "sa-fail-test",
		"--run-id", started["runId"].(string),
		"--reason", "external-failure",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
}

// startStandaloneActivity starts a standalone activity via the CLI and returns
// the parsed JSON response containing activityId and runId.
func (s *SharedServerSuite) startStandaloneActivity(activityID string, extraArgs ...string) map[string]any {
	args := []string{
		"activity", "start",
		"-o", "json",
		"--activity-id", activityID,
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--address", s.Address(),
	}
	args = append(args, extraArgs...)
	res := s.Execute(args...)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	return jsonOut
}

func (s *SharedServerSuite) TestStandaloneActivity_Start() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return "start-result", nil
	})

	res := s.Execute(
		"activity", "start",
		"--activity-id", "start-test",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ActivityId", "start-test")
	s.Contains(out, "RunId")
	s.ContainsOnSameLine(out, "Type", "DevActivity")
	s.ContainsOnSameLine(out, "Namespace", "default")
	s.Contains(out, "TaskQueue")

	// JSON
	res = s.Execute(
		"activity", "start",
		"-o", "json",
		"--activity-id", "start-test-json",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("start-test-json", jsonOut["activityId"])
	s.NotEmpty(jsonOut["runId"])
	s.Equal("DevActivity", jsonOut["type"])
	s.Equal("default", jsonOut["namespace"])
	s.NotEmpty(jsonOut["taskQueue"])
}

func (s *SharedServerSuite) TestStandaloneActivity_Execute_Success() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return map[string]string{"foo": "bar"}, nil
	})

	res := s.Execute(
		"activity", "execute",
		"--activity-id", "exec-test",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Result", `{"foo":"bar"}`)
}

func (s *SharedServerSuite) TestStandaloneActivity_Execute_Success_JSON() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return map[string]string{"foo": "bar"}, nil
	})

	res := s.Execute(
		"activity", "execute",
		"-o", "json",
		"--activity-id", "exec-json-test",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("COMPLETED", jsonOut["status"])
	s.Equal(map[string]any{"foo": "bar"}, jsonOut["result"])
}

func (s *SharedServerSuite) TestStandaloneActivity_Execute_Failure() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional failure")
	})

	res := s.Execute(
		"activity", "execute",
		"--activity-id", "exec-fail-test",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--retry-maximum-attempts", "1",
		"--address", s.Address(),
	)
	s.ErrorContains(res.Err, "activity failed")
	s.ErrorContains(res.Err, "intentional failure")
}

func (s *SharedServerSuite) TestStandaloneActivity_Execute_Failure_JSON() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional failure")
	})

	res := s.Execute(
		"activity", "execute",
		"-o", "json",
		"--activity-id", "exec-fail-json-test",
		"--type", "DevActivity",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--start-to-close-timeout", "30s",
		"--retry-maximum-attempts", "1",
		"--address", s.Address(),
	)
	s.Error(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("FAILED", jsonOut["status"])
	failureObj, ok := jsonOut["failure"].(map[string]any)
	s.True(ok, "failure should be a structured object, got: %T", jsonOut["failure"])
	s.NotEmpty(failureObj["message"])
	s.NotNil(failureObj["cause"])
}

func (s *SharedServerSuite) TestStandaloneActivity_Execute_RetriesOnEmptyPollResponse() {
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

func (s *SharedServerSuite) TestStandaloneActivity_Result() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return "result-value", nil
	})

	started := s.startStandaloneActivity("result-test")

	res := s.Execute(
		"activity", "result",
		"--activity-id", "result-test",
		"--run-id", started["runId"].(string),
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "result-value")
}

func (s *SharedServerSuite) TestStandaloneActivity_Describe() {
	activityStarted := make(chan struct{})
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		close(activityStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	started := s.startStandaloneActivity("describe-test")
	<-activityStarted

	res := s.Execute(
		"activity", "describe",
		"--activity-id", "describe-test",
		"--run-id", started["runId"].(string),
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ActivityId", "describe-test")
	s.Contains(out, "DevActivity")
}

func (s *SharedServerSuite) TestStandaloneActivity_List() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return "listed", nil
	})

	s.startStandaloneActivity("list-test-1")
	s.startStandaloneActivity("list-test-2")

	s.Eventually(func() bool {
		res := s.Execute(
			"activity", "list",
			"--address", s.Address(),
		)
		out := res.Stdout.String()
		return res.Err == nil && strings.Contains(out, "list-test-1") && strings.Contains(out, "list-test-2")
	}, 5*time.Second, 200*time.Millisecond)
}

func (s *SharedServerSuite) TestStandaloneActivity_Count() {
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return "counted", nil
	})

	s.startStandaloneActivity("count-test")

	s.Eventually(func() bool {
		res := s.Execute(
			"activity", "count",
			"--address", s.Address(),
		)
		return res.Err == nil && strings.Contains(res.Stdout.String(), "Total:")
	}, 5*time.Second, 200*time.Millisecond)
}

func (s *SharedServerSuite) TestStandaloneActivity_Cancel() {
	activityStarted := make(chan struct{})
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		close(activityStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	started := s.startStandaloneActivity("cancel-test")
	<-activityStarted

	res := s.Execute(
		"activity", "cancel",
		"--activity-id", "cancel-test",
		"--run-id", started["runId"].(string),
		"--reason", "test-cancel",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Cancellation requested")
}

func (s *SharedServerSuite) TestStandaloneActivity_Terminate() {
	activityStarted := make(chan struct{})
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		close(activityStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	started := s.startStandaloneActivity("terminate-test")
	<-activityStarted

	res := s.Execute(
		"activity", "terminate",
		"--activity-id", "terminate-test",
		"--run-id", started["runId"].(string),
		"--reason", "test-terminate",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Activity terminated")
}
