package temporalcli_test

import (
	"context"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

func (s *SharedServerSuite) TestActivity_Complete() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	aid := "dev-activity-id"
	identity := "MyIdentity"
	res := s.Execute(
		"activity", "complete",
		"--activity-id", aid,
		"--workflow-id", wid,
		"--result", "\"complete-activity-result\"",
		"--identity", identity,
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	var actual string
	s.NoError(run.Get(s.Context, &actual))
	s.Equal("complete-activity-result", actual)

	started, completed, failed := s.getActivityEvents(wid, aid)
	s.NotNil(started)
	s.Nil(failed)
	s.NotNil(completed)
	s.Equal("\"complete-activity-result\"", string(completed.Result.Payloads[0].GetData()))
	s.Equal(identity, completed.GetIdentity())
}

func (s *SharedServerSuite) TestActivity_Fail() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	aid := "dev-activity-id"
	detail := "{\"myKey\": \"myValue\"}"
	reason := "MyReason"
	identity := "MyIdentity"
	res := s.Execute(
		"activity", "fail",
		"--activity-id", aid,
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

	started, completed, failed := s.getActivityEvents(wid, aid)
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
	wid := run.GetID()
	aid := "dev-activity-id"
	res := s.Execute(
		"activity", "complete",
		"--activity-id", aid,
		"--workflow-id", wid,
		"--result", "{not json}",
		"--address", s.Address(),
	)
	s.ErrorContains(res.Err, "is not valid JSON")

	started, completed, failed := s.getActivityEvents(wid, aid)
	s.Nil(started)
	s.Nil(completed)
	s.Nil(failed)
}

func (s *SharedServerSuite) TestActivity_Fail_InvalidDetail() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	aid := "dev-activity-id"
	res := s.Execute(
		"activity", "fail",
		"--activity-id", aid,
		"--workflow-id", wid,
		"--detail", "{not json}",
		"--address", s.Address(),
	)
	s.ErrorContains(res.Err, "is not valid JSON")

	started, completed, failed := s.getActivityEvents(wid, aid)
	s.Nil(started)
	s.Nil(completed)
	s.Nil(failed)
}

func (s *SharedServerSuite) TestActivityOptionsUpdate_Accept() {
	run := s.waitActivityStarted()
	wid := run.GetID()
	aid := "dev-activity-id"
	identity := "MyIdentity"

	res := s.Execute(
		"activity", "update-options",
		"--activity-id", aid,
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
	wid := run.GetID()
	aid := "dev-activity-id"
	identity := "MyIdentity"

	res := s.Execute(
		"activity", "update-options",
		"--activity-id", aid,
		"--workflow-id", wid,
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
	//if this test fails, check the default values of activity options
	s.ContainsOnSameLine(out, "StartToCloseTimeout", "10s")
	s.ContainsOnSameLine(out, "HeartbeatTimeout", "0s")
	s.ContainsOnSameLine(out, "MaximumInterval", "1m40s")
	s.ContainsOnSameLine(out, "BackoffCoefficient", "2")
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
