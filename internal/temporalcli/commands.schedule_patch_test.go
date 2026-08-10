package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/temporalio/cli/internal/temporalcli"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/schedule/v1"
	sdkpb "go.temporal.io/api/sdk/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *SharedServerSuite) TestSchedule_PatchHelpRegistersPatchOptions() {
	scheduleHelp := s.Execute("schedule", "--help")
	s.NoError(scheduleHelp.Err)
	s.Contains(strings.Join(strings.Fields(scheduleHelp.Stdout.String()), " "), "patch Change selected Schedule fields")

	res := s.Execute("schedule", "patch", "--help")
	s.NoError(res.Err)
	normalizedHelp := strings.Join(strings.Fields(res.Stdout.String()), " ")
	s.Contains(normalizedHelp, "Change selected fields on an existing Schedule while preserving unspecified fields")
	s.Regexp("temporal schedule patch \\\\\n\\s+--schedule-id \\\"YourScheduleId\\\" \\\\\n\\s+--notes \\\"Runs every hour\\\" \\\\\n\\s+--interval \\\"1h\\\"", res.Stdout.String())
	s.Contains(normalizedHelp, "A successful command confirms that the patch was submitted, not that it has been applied on every backend")
	s.Contains(normalizedHelp, "Calendar JSON specification. May be passed multiple times.")
	s.Contains(normalizedHelp, "Cron expression. May be passed multiple times.")
	s.Contains(normalizedHelp, "Interval specification. May be passed multiple times.")
	s.Regexp(`--static-summary string Set the static Workflow summary for human consumption in UIs\. Uses Temporal Markdown formatting, should be a single line\. EXPERIMENTAL\.`, normalizedHelp)
	s.Regexp(`--static-details string Set the static Workflow details for human consumption in UIs\. Uses Temporal Markdown formatting, may be multiple lines\. EXPERIMENTAL\.`, normalizedHelp)
	s.Contains(res.Stdout.String(), "--overlap-policy")
	s.Contains(res.Stdout.String(), "--catchup-window")
	s.Contains(res.Stdout.String(), "--unset-catchup-window")
	s.Contains(res.Stdout.String(), "--pause-on-failure")
	s.Contains(res.Stdout.String(), "--paused")
	s.Contains(res.Stdout.String(), "--remaining-actions")
	s.Contains(res.Stdout.String(), "--notes")
	s.Contains(res.Stdout.String(), "--unset-notes")
	for _, option := range []string{
		"--calendar",
		"--cron",
		"--interval",
		"--cadence-clear-all",
		"--start-time",
		"--unset-start-time",
		"--end-time",
		"--unset-end-time",
		"--jitter",
		"--unset-jitter",
		"--time-zone",
		"--unset-time-zone",
		"--workflow-id",
		"--type",
		"--task-queue",
		"--execution-timeout",
		"--unset-execution-timeout",
		"--run-timeout",
		"--unset-run-timeout",
		"--task-timeout",
		"--unset-task-timeout",
		"--static-summary",
		"--unset-static-summary",
		"--static-details",
		"--unset-static-details",
	} {
		s.Contains(res.Stdout.String(), option)
	}
	s.Contains(res.Stdout.String(), "Aliased as \"--name\"")
	s.Contains(res.Stdout.String(), "Remove the explicit Workflow")
	s.Contains(res.Stdout.String(), "Restore the inherited Workflow Run")
	s.Contains(res.Stdout.String(), "Restore the 10-second default")
	s.Contains(res.Stdout.String(), "Remove the static Workflow summary")
	s.Contains(res.Stdout.String(), "Remove the static Workflow details")
	s.Contains(res.Stdout.String(), "calendar, cron, and interval sources")
	s.Contains(res.Stdout.String(), "Omitting cadence options preserves the existing cadence")
	s.Contains(res.Stdout.String(), "`--cron '0 12 * * *'` to replace cadence")
	s.Contains(normalizedHelp, "`--cadence-clear-all` to clear it")
	s.Contains(res.Stdout.String(), "pause explicitly with `--paused=true` when needed")
	s.Contains(res.Stdout.String(), "Clear all cadence sources only when")
	s.NotContains(res.Stdout.String(), "--headers")
	s.NotContains(res.Stdout.String(), "--memo")
	s.NotContains(res.Stdout.String(), "--search-attribute")
	s.NotContains(res.Stdout.String(), "--input")
	s.NotContains(res.Stdout.String(), "--priority-key")
	s.NotContains(res.Stdout.String(), "--fairness-key")
	s.NotContains(res.Stdout.String(), "--fairness-weight")
	s.NotContains(res.Stdout.String(), "--unset-workflow-id")
	s.NotContains(res.Stdout.String(), "--unset-type")
	s.NotContains(res.Stdout.String(), "--unset-task-queue")

	res = s.Execute("schedule", "update", "--help")
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "full replacement")
	s.Contains(res.Stdout.String(), "field-preserving")
	s.Contains(res.Stdout.String(), "temporal schedule patch")
}

func (s *SharedServerSuite) TestSchedule_PatchUpdatesStoredNotes() {
	const (
		initialNotes = "initial notes"
		patchedNotes = "patched notes"
	)
	scheduleID, _, res := s.createSchedule("--interval", "10d", "--notes", initialNotes)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--schedule-id", scheduleID,
		"--notes", patchedNotes,
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"--schedule-id", scheduleID,
			"--output", "json",
		)
		if res.Err != nil {
			return false
		}
		var description struct {
			Schedule struct {
				State struct {
					Notes string `json:"notes"`
				} `json:"state"`
			} `json:"schedule"`
		}
		if err := json.Unmarshal(res.Stdout.Bytes(), &description); err != nil {
			return false
		}
		return description.Schedule.State.Notes == patchedNotes
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_PatchRejectsInvalidArgumentsBeforeMutation() {
	var scheduleRequests atomic.Int32
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		switch req.(type) {
		case *workflowservice.DescribeScheduleRequest, *workflowservice.UpdateScheduleRequest:
			scheduleRequests.Add(1)
		}
		return handler(ctx, req)
	})

	for _, tc := range []struct {
		name          string
		args          []string
		errorContains string
	}{
		{
			name:          "missing schedule ID",
			args:          []string{"schedule", "patch", "--notes", "note"},
			errorContains: "required flag(s) \"schedule-id\" not set",
		},
		{
			name:          "headers",
			args:          []string{"schedule", "patch", "--schedule-id", "schedule-id", "--notes", "note", "--headers", "example=123"},
			errorContains: "unknown flag: --headers",
		},
	} {
		scheduleRequests.Store(0)
		res := s.Execute(tc.args...)
		assert.Error(s.T(), res.Err, tc.name)
		assert.ErrorContains(s.T(), res.Err, tc.errorContains, tc.name)
		assert.Equal(s.T(), int32(0), scheduleRequests.Load(), tc.name)
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSemanticValidationFailsBeforeDial() {
	dialErr := errors.New("unexpected gRPC dial")

	for _, tc := range []struct {
		name          string
		args          []string
		errorContains string
	}{
		{
			name:          "no operation",
			args:          []string{"--schedule-id", "schedule-id"},
			errorContains: "at least one patch operation is required",
		},
		{
			name:          "set and unset notes",
			args:          []string{"--schedule-id", "schedule-id", "--notes", "note", "--unset-notes"},
			errorContains: "--notes and --unset-notes are mutually exclusive",
		},
		{
			name:          "set and unset catchup window",
			args:          []string{"--schedule-id", "schedule-id", "--catchup-window", "10s", "--unset-catchup-window"},
			errorContains: "--catchup-window and --unset-catchup-window are mutually exclusive",
		},
		{
			name:          "catchup window below minimum",
			args:          []string{"--schedule-id", "schedule-id", "--catchup-window", "9s"},
			errorContains: "catchup window must be at least 10s",
		},
		{
			name:          "zero catchup window",
			args:          []string{"--schedule-id", "schedule-id", "--catchup-window", "0s"},
			errorContains: "catchup window must be at least 10s",
		},
		{
			name:          "negative remaining actions",
			args:          []string{"--schedule-id", "schedule-id", "--remaining-actions", "-1"},
			errorContains: "remaining actions must not be negative",
		},
		{name: "malformed calendar", args: []string{"--schedule-id", "schedule-id", "--calendar", "{"}, errorContains: "failed to parse json calendar spec"},
		{name: "malformed interval", args: []string{"--schedule-id", "schedule-id", "--interval", "nonsense"}, errorContains: "invalid interval"},
		{name: "interval below one second", args: []string{"--schedule-id", "schedule-id", "--interval", "500ms"}, errorContains: "interval must be at least 1s"},
		{name: "negative interval phase", args: []string{"--schedule-id", "schedule-id", "--interval", "1h/-1s"}, errorContains: "interval phase must not be negative"},
		{name: "interval phase too large", args: []string{"--schedule-id", "schedule-id", "--interval", "1h/1h"}, errorContains: "interval phase must be less than the interval"},
		{name: "negative jitter", args: []string{"--schedule-id", "schedule-id", "--jitter", "-1s"}, errorContains: "jitter must not be negative"},
		{name: "TZ cron time zone", args: []string{"--schedule-id", "schedule-id", "--cron", "TZ=UTC 0 12 * * *"}, errorContains: "use --time-zone"},
		{name: "CRON_TZ cron time zone", args: []string{"--schedule-id", "schedule-id", "--cron", "CRON_TZ=UTC 0 12 * * *"}, errorContains: "use --time-zone"},
		{name: "leading space TZ cron time zone", args: []string{"--schedule-id", "schedule-id", "--cron", " TZ=UTC 0 12 * * *"}, errorContains: "use --time-zone"},
		{name: "leading tab TZ cron time zone", args: []string{"--schedule-id", "schedule-id", "--cron", "\tTZ=UTC 0 12 * * *"}, errorContains: "use --time-zone"},
		{name: "leading space CRON_TZ cron time zone", args: []string{"--schedule-id", "schedule-id", "--cron", " CRON_TZ=UTC 0 12 * * *"}, errorContains: "use --time-zone"},
		{name: "leading tab CRON_TZ cron time zone", args: []string{"--schedule-id", "schedule-id", "--cron", "\tCRON_TZ=UTC 0 12 * * *"}, errorContains: "use --time-zone"},
		{name: "start timestamp before protobuf range", args: []string{"--schedule-id", "schedule-id", "--start-time", "0000-01-01T00:00:00Z"}, errorContains: "invalid start time"},
		{name: "start timestamp normalized before protobuf range", args: []string{"--schedule-id", "schedule-id", "--start-time", "0001-01-01T00:00:00+14:00"}, errorContains: "invalid start time"},
		{name: "end timestamp normalized after protobuf range", args: []string{"--schedule-id", "schedule-id", "--end-time", "9999-12-31T23:59:59-14:00"}, errorContains: "invalid end time"},
		{name: "empty time zone set", args: []string{"--schedule-id", "schedule-id", "--time-zone="}, errorContains: "use --unset-time-zone"},
		{name: "whitespace time zone set", args: []string{"--schedule-id", "schedule-id", "--time-zone", " \t "}, errorContains: "use --unset-time-zone"},
		{name: "clear and calendar", args: []string{"--schedule-id", "schedule-id", "--cadence-clear-all", "--calendar", `{"minute":"5"}`}, errorContains: "cannot be combined"},
		{name: "clear and cron", args: []string{"--schedule-id", "schedule-id", "--cadence-clear-all", "--cron", "0 12 * * *"}, errorContains: "cannot be combined"},
		{name: "clear and interval", args: []string{"--schedule-id", "schedule-id", "--cadence-clear-all", "--interval", "1h"}, errorContains: "cannot be combined"},
		{
			name:          "empty schedule ID",
			args:          []string{"--schedule-id=", "--notes", "note"},
			errorContains: "schedule ID is required",
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			var dialAttempts atomic.Int32
			options := s.CommandHarness.Options
			options.Args = append([]string{"schedule", "patch"}, tc.args...)
			options.AdditionalClientGRPCDialOptions = append(
				options.AdditionalClientGRPCDialOptions,
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					dialAttempts.Add(1)
					return nil, dialErr
				}),
			)
			var commandErr error
			options.Fail = func(err error) { commandErr = err }

			temporalcli.Execute(context.Background(), options)

			assert.ErrorContains(t, commandErr, tc.errorContains)
			assert.Equal(t, int32(0), dialAttempts.Load())
			assert.NotErrorIs(t, commandErr, dialErr)
		})
	}

}

func (s *SharedServerSuite) TestSchedule_PatchMalformedPolicyAndStateFlagsFailBeforeScheduleRPC() {
	var scheduleRequests atomic.Int32
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		switch req.(type) {
		case *workflowservice.DescribeScheduleRequest, *workflowservice.UpdateScheduleRequest:
			scheduleRequests.Add(1)
		}
		return handler(ctx, req)
	})

	for _, tc := range []struct {
		name          string
		args          []string
		errorContains string
	}{
		{name: "enum", args: []string{"--overlap-policy=not-a-policy"}, errorContains: "invalid argument"},
		{name: "duration", args: []string{"--catchup-window=not-a-duration"}, errorContains: "invalid duration"},
		{name: "integer", args: []string{"--remaining-actions=not-an-int"}, errorContains: "invalid argument"},
		{name: "boolean", args: []string{"--paused=not-a-bool"}, errorContains: "invalid argument"},
		{name: "timestamp", args: []string{"--start-time=not-a-timestamp"}, errorContains: "cannot parse"},
		{name: "jitter duration", args: []string{"--jitter=not-a-duration"}, errorContains: "invalid duration"},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			scheduleRequests.Store(0)
			var failures []error
			options := s.CommandHarness.Options
			options.Args = append([]string{"schedule", "patch", "--schedule-id", "schedule-id"}, tc.args...)
			options.Fail = func(err error) { failures = append(failures, err) }
			temporalcli.Execute(context.Background(), options)
			assert.NotEmpty(t, failures)
			assert.ErrorContains(t, failures[0], tc.errorContains)
			assert.Equal(t, int32(0), scheduleRequests.Load())
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchRejectsDescribeResponseWithoutSchedule() {
	var describeRequests atomic.Int32
	var updateRequests atomic.Int32
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				describeRequests.Add(1)
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = nil
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests.Add(1)
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	assert.NotPanics(s.T(), func() {
		res := s.Execute(
			"schedule", "patch",
			"--address", s.Address(),
			"--schedule-id", "missing-schedule",
			"--notes", "updated notes",
		)
		assert.ErrorContains(s.T(), res.Err, `DescribeSchedule response for Schedule "missing-schedule" did not contain a Schedule`)
	})
	assert.Equal(s.T(), int32(1), describeRequests.Load())
	assert.Equal(s.T(), int32(0), updateRequests.Load())
}

func (s *SharedServerSuite) TestSchedule_PatchSubmitsOneUpdateThatChangesOnlyNotes() {
	const (
		namespace  = "patch-notes-namespace"
		scheduleID = "patch-notes-schedule"
		identity   = "patch-notes-identity"
	)
	conflictToken := []byte("patch-notes-conflict-token")
	describedSchedule := &schedule.Schedule{
		Spec: &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
		Policies: &schedule.SchedulePolicies{
			OverlapPolicy:  enums.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL,
			CatchupWindow:  durationpb.New(20 * time.Second),
			PauseOnFailure: true,
		},
		State: &schedule.ScheduleState{
			Notes:            "existing notes",
			Paused:           true,
			LimitedActions:   true,
			RemainingActions: 4,
		},
	}
	var lock sync.Mutex
	var describeRequests []*workflowservice.DescribeScheduleRequest
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests = append(describeRequests, proto.Clone(request).(*workflowservice.DescribeScheduleRequest))
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictToken...)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				lock.Unlock()
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	requestIDs := map[string]struct{}{}
	for _, tc := range []struct {
		name  string
		notes string
	}{
		{name: "different value", notes: "updated notes"},
		{name: "explicit empty string", notes: ""},
		{name: "same value", notes: "existing notes"},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			lock.Lock()
			describeRequests = nil
			updateRequests = nil
			lock.Unlock()

			res := s.Execute(
				"schedule", "patch",
				"--address", s.Address(),
				"--namespace", namespace,
				"--identity", identity,
				"--schedule-id", scheduleID,
				"--notes", tc.notes,
			)
			assert.NoError(t, res.Err)

			lock.Lock()
			gotDescribeRequests := append([]*workflowservice.DescribeScheduleRequest(nil), describeRequests...)
			gotUpdateRequests := append([]*workflowservice.UpdateScheduleRequest(nil), updateRequests...)
			lock.Unlock()
			assert.Len(t, gotDescribeRequests, 1)
			assert.Len(t, gotUpdateRequests, 1)
			if len(gotDescribeRequests) != 1 || len(gotUpdateRequests) != 1 {
				return
			}

			describeRequest := gotDescribeRequests[0]
			assert.Equal(t, namespace, describeRequest.GetNamespace())
			assert.Equal(t, scheduleID, describeRequest.GetScheduleId())

			updateRequest := gotUpdateRequests[0]
			assert.Equal(t, namespace, updateRequest.GetNamespace())
			assert.Equal(t, scheduleID, updateRequest.GetScheduleId())
			assert.Equal(t, conflictToken, updateRequest.GetConflictToken())
			assert.Equal(t, identity, updateRequest.GetIdentity())
			assert.NotEmpty(t, updateRequest.GetRequestId())
			_, exists := requestIDs[updateRequest.GetRequestId()]
			assert.False(t, exists)
			requestIDs[updateRequest.GetRequestId()] = struct{}{}
			assert.Nil(t, updateRequest.GetMemo())
			assert.Nil(t, updateRequest.GetSearchAttributes())

			expectedSchedule := proto.Clone(describedSchedule).(*schedule.Schedule)
			expectedSchedule.State.Notes = tc.notes
			assert.True(t, proto.Equal(expectedSchedule, updateRequest.GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsWorkflowIdentityIndependently() {
	describedSchedule := &schedule.Schedule{
		Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
			StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
				WorkflowId:   "existing-workflow-id",
				WorkflowType: &commonpb.WorkflowType{Name: "ExistingWorkflow"},
				TaskQueue: &taskqueuepb.TaskQueue{
					Name:       "existing-task-queue",
					Kind:       enums.TASK_QUEUE_KIND_STICKY,
					NormalName: "existing-normal-name",
				},
				WorkflowExecutionTimeout: durationpb.New(time.Hour),
				WorkflowRunTimeout:       durationpb.New(30 * time.Minute),
				WorkflowTaskTimeout:      durationpb.New(5 * time.Second),
			},
		}},
		Spec:  &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
		State: &schedule.ScheduleState{Notes: "preserved notes"},
	}
	var describeCount int
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				describeCount++
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = []byte("identity-conflict-token")
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		args  []string
		apply func(*workflowpb.NewWorkflowExecutionInfo)
	}{
		{name: "workflow ID", args: []string{"--workflow-id", "updated-workflow-id"}, apply: func(action *workflowpb.NewWorkflowExecutionInfo) { action.WorkflowId = "updated-workflow-id" }},
		{name: "workflow type", args: []string{"--type", "UpdatedWorkflow"}, apply: func(action *workflowpb.NewWorkflowExecutionInfo) { action.WorkflowType.Name = "UpdatedWorkflow" }},
		{name: "workflow type name alias", args: []string{"--name", "AliasedWorkflow"}, apply: func(action *workflowpb.NewWorkflowExecutionInfo) { action.WorkflowType.Name = "AliasedWorkflow" }},
		{name: "task queue", args: []string{"--task-queue", "updated-task-queue"}, apply: func(action *workflowpb.NewWorkflowExecutionInfo) { action.TaskQueue.Name = "updated-task-queue" }},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describeCount = 0
			updateRequests = nil
			res := s.Execute(append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id"}, tc.args...)...)
			assert.NoError(t, res.Err)
			assert.Equal(t, 1, describeCount)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected.GetAction().GetStartWorkflow())
			assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
		})
	}

	describedSchedule = &schedule.Schedule{Spec: &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}}}
	describeCount = 0
	updateRequests = nil
	res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", "--workflow-id", "updated-workflow-id")
	s.EqualError(res.Err, "schedule action does not contain a StartWorkflow action")
	s.Equal(1, describeCount)
	s.Empty(updateRequests)
}

func (s *SharedServerSuite) TestSchedule_PatchRejectsExactlyEmptyIdentityAndPreservesWhitespace() {
	dialErr := errors.New("unexpected gRPC dial")
	for _, tc := range []struct {
		name          string
		args          []string
		errorContains string
	}{
		{name: "empty workflow ID", args: []string{"--workflow-id="}, errorContains: "workflow ID must not be empty"},
		{name: "empty workflow type", args: []string{"--type="}, errorContains: "workflow type must not be empty"},
		{name: "empty task queue", args: []string{"--task-queue="}, errorContains: "task queue must not be empty"},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			var dialAttempts atomic.Int32
			options := s.CommandHarness.Options
			options.Args = append([]string{"schedule", "patch", "--schedule-id", "schedule-id"}, tc.args...)
			options.AdditionalClientGRPCDialOptions = append(
				options.AdditionalClientGRPCDialOptions,
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					dialAttempts.Add(1)
					return nil, dialErr
				}),
			)
			var commandErr error
			options.Fail = func(err error) { commandErr = err }

			temporalcli.Execute(context.Background(), options)

			assert.ErrorContains(t, commandErr, tc.errorContains)
			assert.Equal(t, int32(0), dialAttempts.Load())
			assert.NotErrorIs(t, commandErr, dialErr)
		})
	}

	describedSchedule := &schedule.Schedule{Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
		StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
			WorkflowId:   "existing-workflow-id",
			WorkflowType: &commonpb.WorkflowType{Name: "ExistingWorkflow"},
			TaskQueue:    &taskqueuepb.TaskQueue{Name: "existing-task-queue"},
		},
	}}}
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		args  []string
		value func(*workflowpb.NewWorkflowExecutionInfo) string
	}{
		{name: "workflow ID whitespace", args: []string{"--workflow-id", " \t "}, value: func(action *workflowpb.NewWorkflowExecutionInfo) string { return action.GetWorkflowId() }},
		{name: "workflow type whitespace", args: []string{"--type", " \t "}, value: func(action *workflowpb.NewWorkflowExecutionInfo) string { return action.GetWorkflowType().GetName() }},
		{name: "task queue whitespace", args: []string{"--task-queue", " \t "}, value: func(action *workflowpb.NewWorkflowExecutionInfo) string { return action.GetTaskQueue().GetName() }},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequests = nil
			res := s.Execute(append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id"}, tc.args...)...)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) == 1 {
				assert.Equal(t, " \t ", tc.value(updateRequests[0].GetSchedule().GetAction().GetStartWorkflow()))
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsPositiveWorkflowTimeoutsIndependently() {
	describedSchedule := &schedule.Schedule{
		Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
			StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
				WorkflowId:               "workflow-id",
				WorkflowType:             &commonpb.WorkflowType{Name: "Workflow"},
				TaskQueue:                &taskqueuepb.TaskQueue{Name: "task-queue"},
				WorkflowExecutionTimeout: durationpb.New(time.Hour),
				WorkflowRunTimeout:       durationpb.New(30 * time.Minute),
				WorkflowTaskTimeout:      durationpb.New(5 * time.Second),
			},
		}},
		Spec:  &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
		State: &schedule.ScheduleState{Notes: "preserved notes"},
	}
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		flag  string
		value time.Duration
		apply func(*workflowpb.NewWorkflowExecutionInfo, *durationpb.Duration)
	}{
		{name: "execution timeout", flag: "--execution-timeout", value: 2 * time.Hour, apply: func(action *workflowpb.NewWorkflowExecutionInfo, value *durationpb.Duration) {
			action.WorkflowExecutionTimeout = value
		}},
		{name: "run timeout", flag: "--run-timeout", value: 45 * time.Minute, apply: func(action *workflowpb.NewWorkflowExecutionInfo, value *durationpb.Duration) {
			action.WorkflowRunTimeout = value
		}},
		{name: "task timeout", flag: "--task-timeout", value: 7 * time.Second, apply: func(action *workflowpb.NewWorkflowExecutionInfo, value *durationpb.Duration) {
			action.WorkflowTaskTimeout = value
		}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", tc.flag, tc.value.String())
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected.GetAction().GetStartWorkflow(), durationpb.New(tc.value))
			assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsExplicitZeroWorkflowTimeoutsAsPresent() {
	describedSchedule := &schedule.Schedule{Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
		StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
			WorkflowId:   "workflow-id",
			WorkflowType: &commonpb.WorkflowType{Name: "Workflow"},
			TaskQueue:    &taskqueuepb.TaskQueue{Name: "task-queue"},
		},
	}}}
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name string
		flag string
		get  func(*workflowpb.NewWorkflowExecutionInfo) *durationpb.Duration
	}{
		{name: "execution timeout", flag: "--execution-timeout", get: func(action *workflowpb.NewWorkflowExecutionInfo) *durationpb.Duration {
			return action.GetWorkflowExecutionTimeout()
		}},
		{name: "run timeout", flag: "--run-timeout", get: func(action *workflowpb.NewWorkflowExecutionInfo) *durationpb.Duration {
			return action.GetWorkflowRunTimeout()
		}},
		{name: "task timeout", flag: "--task-timeout", get: func(action *workflowpb.NewWorkflowExecutionInfo) *durationpb.Duration {
			return action.GetWorkflowTaskTimeout()
		}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", tc.flag, "0s")
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			got := tc.get(updateRequests[0].GetSchedule().GetAction().GetStartWorkflow())
			assert.NotNil(t, got)
			if got != nil {
				assert.Equal(t, time.Duration(0), got.AsDuration())
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchResetsWorkflowTimeoutsToDistinctDefaults() {
	describedSchedule := &schedule.Schedule{
		Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
			StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
				WorkflowId:               "workflow-id",
				WorkflowType:             &commonpb.WorkflowType{Name: "Workflow"},
				TaskQueue:                &taskqueuepb.TaskQueue{Name: "task-queue"},
				WorkflowExecutionTimeout: durationpb.New(time.Hour),
				WorkflowRunTimeout:       durationpb.New(30 * time.Minute),
				WorkflowTaskTimeout:      durationpb.New(5 * time.Second),
			},
		}},
		State: &schedule.ScheduleState{Notes: "preserved notes"},
	}
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		flag  string
		apply func(*workflowpb.NewWorkflowExecutionInfo)
	}{
		{name: "execution timeout", flag: "--unset-execution-timeout", apply: func(action *workflowpb.NewWorkflowExecutionInfo) { action.WorkflowExecutionTimeout = nil }},
		{name: "run timeout", flag: "--unset-run-timeout", apply: func(action *workflowpb.NewWorkflowExecutionInfo) { action.WorkflowRunTimeout = nil }},
		{name: "task timeout", flag: "--unset-task-timeout", apply: func(action *workflowpb.NewWorkflowExecutionInfo) {
			action.WorkflowTaskTimeout = durationpb.New(10 * time.Second)
		}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", tc.flag)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected.GetAction().GetStartWorkflow())
			assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchRejectsTimeoutConflictsAndNegativesBeforeDial() {
	dialErr := errors.New("unexpected gRPC dial")
	for _, tc := range []struct {
		name          string
		args          []string
		errorContains string
	}{
		{name: "execution timeout conflict", args: []string{"--execution-timeout", "1s", "--unset-execution-timeout"}, errorContains: "--execution-timeout and --unset-execution-timeout are mutually exclusive"},
		{name: "run timeout conflict", args: []string{"--run-timeout", "1s", "--unset-run-timeout"}, errorContains: "--run-timeout and --unset-run-timeout are mutually exclusive"},
		{name: "task timeout conflict", args: []string{"--task-timeout", "1s", "--unset-task-timeout"}, errorContains: "--task-timeout and --unset-task-timeout are mutually exclusive"},
		{name: "negative execution timeout", args: []string{"--execution-timeout", "-1s"}, errorContains: "execution timeout must not be negative"},
		{name: "negative run timeout", args: []string{"--run-timeout", "-1s"}, errorContains: "run timeout must not be negative"},
		{name: "negative task timeout", args: []string{"--task-timeout", "-1s"}, errorContains: "task timeout must not be negative"},
		{name: "static summary conflict", args: []string{"--static-summary", "summary", "--unset-static-summary"}, errorContains: "--static-summary and --unset-static-summary are mutually exclusive"},
		{name: "static details conflict", args: []string{"--static-details", "details", "--unset-static-details"}, errorContains: "--static-details and --unset-static-details are mutually exclusive"},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			var dialAttempts atomic.Int32
			options := s.CommandHarness.Options
			options.Args = append([]string{"schedule", "patch", "--schedule-id", "schedule-id"}, tc.args...)
			options.AdditionalClientGRPCDialOptions = append(
				options.AdditionalClientGRPCDialOptions,
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					dialAttempts.Add(1)
					return nil, dialErr
				}),
			)
			var commandErr error
			options.Fail = func(err error) { commandErr = err }

			temporalcli.Execute(context.Background(), options)

			assert.ErrorContains(t, commandErr, tc.errorContains)
			assert.Equal(t, int32(0), dialAttempts.Load())
			assert.NotErrorIs(t, commandErr, dialErr)
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsStaticMetadataIndependently() {
	existingSummary, err := temporalcli.DataConverterWithRawValue.ToPayload("existing summary")
	s.NoError(err)
	existingDetails, err := temporalcli.DataConverterWithRawValue.ToPayload("existing details")
	s.NoError(err)
	metadata := &sdkpb.UserMetadata{Summary: existingSummary, Details: existingDetails}
	metadata.ProtoReflect().SetUnknown([]byte{0xa0, 0x06, 0x01})
	describedSchedule := &schedule.Schedule{Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
		StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
			WorkflowId:   "workflow-id",
			WorkflowType: &commonpb.WorkflowType{Name: "Workflow"},
			TaskQueue:    &taskqueuepb.TaskQueue{Name: "task-queue"},
			UserMetadata: metadata,
		},
	}}}
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		flag  string
		value string
		apply func(*sdkpb.UserMetadata, *commonpb.Payload)
	}{
		{name: "summary", flag: "--static-summary", value: "updated summary", apply: func(metadata *sdkpb.UserMetadata, value *commonpb.Payload) { metadata.Summary = value }},
		{name: "details", flag: "--static-details", value: "updated details", apply: func(metadata *sdkpb.UserMetadata, value *commonpb.Payload) { metadata.Details = value }},
		{name: "explicit empty summary", flag: "--static-summary", value: "", apply: func(metadata *sdkpb.UserMetadata, value *commonpb.Payload) { metadata.Summary = value }},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", tc.flag, tc.value)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			expectedPayload, payloadErr := temporalcli.DataConverterWithRawValue.ToPayload(tc.value)
			assert.NoError(t, payloadErr)
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected.GetAction().GetStartWorkflow().GetUserMetadata(), expectedPayload)
			assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
			actualMetadata := updateRequests[0].GetSchedule().GetAction().GetStartWorkflow().GetUserMetadata()
			assert.True(t, bytes.Equal([]byte{0xa0, 0x06, 0x01}, actualMetadata.ProtoReflect().GetUnknown()))
			if tc.value == "" {
				assert.Equal(t, "json/plain", string(expectedPayload.GetMetadata()["encoding"]))
				assert.Equal(t, []byte(`""`), expectedPayload.GetData())
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchUnsetsStaticMetadataIndependently() {
	existingSummary, err := temporalcli.DataConverterWithRawValue.ToPayload("existing summary")
	s.NoError(err)
	existingDetails, err := temporalcli.DataConverterWithRawValue.ToPayload("existing details")
	s.NoError(err)
	var describedSchedule *schedule.Schedule
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				reply.(*workflowservice.DescribeScheduleResponse).Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name     string
		flag     string
		metadata *sdkpb.UserMetadata
		apply    func(*sdkpb.UserMetadata)
	}{
		{name: "summary", flag: "--unset-static-summary", metadata: &sdkpb.UserMetadata{Summary: existingSummary, Details: existingDetails}, apply: func(metadata *sdkpb.UserMetadata) { metadata.Summary = nil }},
		{name: "details", flag: "--unset-static-details", metadata: &sdkpb.UserMetadata{Summary: existingSummary, Details: existingDetails}, apply: func(metadata *sdkpb.UserMetadata) { metadata.Details = nil }},
		{name: "absent metadata", flag: "--unset-static-summary", metadata: nil, apply: func(*sdkpb.UserMetadata) {}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			if tc.metadata != nil {
				tc.metadata.ProtoReflect().SetUnknown([]byte{0xa0, 0x06, 0x01})
			}
			describedSchedule = &schedule.Schedule{Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{
				StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
					WorkflowId:   "workflow-id",
					WorkflowType: &commonpb.WorkflowType{Name: "Workflow"},
					TaskQueue:    &taskqueuepb.TaskQueue{Name: "task-queue"},
					UserMetadata: tc.metadata,
				},
			}}}
			updateRequests = nil

			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", tc.flag)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected.GetAction().GetStartWorkflow().GetUserMetadata())
			assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchReappliesStaticSummaryAfterConflictRefresh() {
	staleSummary, err := temporalcli.DataConverterWithRawValue.ToPayload("stale summary")
	s.NoError(err)
	staleDetails, err := temporalcli.DataConverterWithRawValue.ToPayload("stale details")
	s.NoError(err)
	refreshedSummary, err := temporalcli.DataConverterWithRawValue.ToPayload("refreshed summary")
	s.NoError(err)
	refreshedDetails, err := temporalcli.DataConverterWithRawValue.ToPayload("refreshed details")
	s.NoError(err)
	describedSchedules := []*schedule.Schedule{
		{Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
			WorkflowId:   "stale-workflow-id",
			WorkflowType: &commonpb.WorkflowType{Name: "StaleWorkflow"},
			TaskQueue:    &taskqueuepb.TaskQueue{Name: "stale-task-queue"},
			UserMetadata: &sdkpb.UserMetadata{Summary: staleSummary, Details: staleDetails},
		}}}},
		{Action: &schedule.ScheduleAction{Action: &schedule.ScheduleAction_StartWorkflow{StartWorkflow: &workflowpb.NewWorkflowExecutionInfo{
			WorkflowId:   "refreshed-workflow-id",
			WorkflowType: &commonpb.WorkflowType{Name: "RefreshedWorkflow"},
			TaskQueue:    &taskqueuepb.TaskQueue{Name: "refreshed-task-queue"},
			UserMetadata: &sdkpb.UserMetadata{Summary: refreshedSummary, Details: refreshedDetails},
		}}}},
	}
	conflictTokens := [][]byte{[]byte("stale-token"), []byte("refreshed-token")}
	var describeCount int
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedules[describeCount]).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictTokens[describeCount]...)
				describeCount++
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				if len(updateRequests) == 1 {
					return status.Error(codes.FailedPrecondition, "mismatched conflict token")
				}
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", "--static-summary", "requested summary")
	s.NoError(res.Err)
	s.Equal(2, describeCount)
	s.Len(updateRequests, 2)
	if len(updateRequests) != 2 {
		return
	}
	requestedSummary, err := temporalcli.DataConverterWithRawValue.ToPayload("requested summary")
	s.NoError(err)
	for index, request := range updateRequests {
		expected := proto.Clone(describedSchedules[index]).(*schedule.Schedule)
		expected.GetAction().GetStartWorkflow().UserMetadata.Summary = requestedSummary
		s.True(proto.Equal(expected, request.GetSchedule()))
		s.Equal(conflictTokens[index], request.GetConflictToken())
		s.NotEmpty(request.GetRequestId())
	}
	s.NotEqual(updateRequests[0].GetRequestId(), updateRequests[1].GetRequestId())
	s.Equal("refreshed-task-queue", updateRequests[1].GetSchedule().GetAction().GetStartWorkflow().GetTaskQueue().GetName())
	s.True(proto.Equal(refreshedDetails, updateRequests[1].GetSchedule().GetAction().GetStartWorkflow().GetUserMetadata().GetDetails()))
}

func (s *SharedServerSuite) TestSchedule_PatchSetsWorkflowFieldsAndRestoresOptionalDefaults() {
	scheduleID, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--schedule-id", scheduleID,
		"--workflow-id", "patched-workflow-id",
		"--type", "PatchedWorkflow",
		"--task-queue", "patched-task-queue",
		"--execution-timeout", "2h",
		"--run-timeout", "1h",
		"--task-timeout", "7s",
		"--static-summary", "patched summary",
		"--static-details", "patched details",
	)
	s.NoError(res.Err)

	expectedSummary, err := temporalcli.DataConverterWithRawValue.ToPayload("patched summary")
	s.NoError(err)
	expectedDetails, err := temporalcli.DataConverterWithRawValue.ToPayload("patched details")
	s.NoError(err)
	s.Eventually(func() bool {
		res = s.Execute("schedule", "describe", "--address", s.Address(), "--schedule-id", scheduleID, "--output", "json")
		if res.Err != nil {
			return false
		}
		var description workflowservice.DescribeScheduleResponse
		if err := temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &description, true); err != nil {
			return false
		}
		action := description.GetSchedule().GetAction().GetStartWorkflow()
		return action.GetWorkflowId() == "patched-workflow-id" &&
			action.GetWorkflowType().GetName() == "PatchedWorkflow" &&
			action.GetTaskQueue().GetName() == "patched-task-queue" &&
			action.GetWorkflowExecutionTimeout().AsDuration() == 2*time.Hour &&
			action.GetWorkflowRunTimeout().AsDuration() == time.Hour &&
			action.GetWorkflowTaskTimeout().AsDuration() == 7*time.Second &&
			proto.Equal(expectedSummary, action.GetUserMetadata().GetSummary()) &&
			proto.Equal(expectedDetails, action.GetUserMetadata().GetDetails())
	}, 10*time.Second, 100*time.Millisecond)

	res = s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--schedule-id", scheduleID,
		"--unset-execution-timeout",
		"--unset-run-timeout",
		"--unset-task-timeout",
		"--unset-static-summary",
		"--unset-static-details",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute("schedule", "describe", "--address", s.Address(), "--schedule-id", scheduleID, "--output", "json")
		if res.Err != nil {
			return false
		}
		var description workflowservice.DescribeScheduleResponse
		if err := temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &description, true); err != nil {
			return false
		}
		action := description.GetSchedule().GetAction().GetStartWorkflow()
		return action.GetWorkflowId() == "patched-workflow-id" &&
			action.GetWorkflowType().GetName() == "PatchedWorkflow" &&
			action.GetTaskQueue().GetName() == "patched-task-queue" &&
			action.GetWorkflowExecutionTimeout() == nil &&
			action.GetWorkflowRunTimeout() == nil &&
			action.GetWorkflowTaskTimeout().AsDuration() == 10*time.Second &&
			action.GetUserMetadata().GetSummary() == nil &&
			action.GetUserMetadata().GetDetails() == nil
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_PatchReplacesCadenceAggregate() {
	const scheduleID = "patch-cadence-schedule"
	describedSchedule := &schedule.Schedule{
		Spec: &schedule.ScheduleSpec{
			StructuredCalendar:        []*schedule.StructuredCalendarSpec{{}},
			Calendar:                  []*schedule.CalendarSpec{{Minute: "1"}},
			CronString:                []string{"0 12 * * *"},
			Interval:                  []*schedule.IntervalSpec{{Interval: durationpb.New(time.Hour)}},
			ExcludeCalendar:           []*schedule.CalendarSpec{{Minute: "2"}},
			ExcludeStructuredCalendar: []*schedule.StructuredCalendarSpec{{}},
			StartTime:                 timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			EndTime:                   timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
			Jitter:                    durationpb.New(time.Minute),
			TimezoneName:              "America/New_York",
		},
		Action:   &schedule.ScheduleAction{},
		Policies: &schedule.SchedulePolicies{PauseOnFailure: true},
		State:    &schedule.ScheduleState{Notes: "preserved", Paused: true},
	}
	var lock sync.Mutex
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = []byte("cadence-token")
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				lock.Unlock()
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name      string
		args      []string
		cron      []string
		intervals []*schedule.IntervalSpec
	}{
		{name: "calendar", args: []string{"--calendar", `{"minute":"5"}`}, cron: []string{"0 5 0 * * * *"}},
		{name: "cron", args: []string{"--cron", "0 5 * * *"}, cron: []string{"0 5 * * *"}},
		{name: "interval", args: []string{"--interval", "2h/30m"}, intervals: []*schedule.IntervalSpec{{Interval: durationpb.New(2 * time.Hour), Phase: durationpb.New(30 * time.Minute)}}},
		{name: "minimum interval with zero phase", args: []string{"--interval", "1s/0s"}, intervals: []*schedule.IntervalSpec{{Interval: durationpb.New(time.Second), Phase: durationpb.New(0)}}},
		{name: "phase just below interval", args: []string{"--interval", "1s/999ms"}, intervals: []*schedule.IntervalSpec{{Interval: durationpb.New(time.Second), Phase: durationpb.New(999 * time.Millisecond)}}},
		{name: "combined", args: []string{"--calendar", `{"minute":"5"}`, "--cron", "0 5 * * *", "--interval", "2h/30m"}, cron: []string{"0 5 0 * * * *", "0 5 * * *"}, intervals: []*schedule.IntervalSpec{{Interval: durationpb.New(2 * time.Hour), Phase: durationpb.New(30 * time.Minute)}}},
		{name: "same value", args: []string{"--cron", "0 12 * * *"}, cron: []string{"0 12 * * *"}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			lock.Lock()
			updateRequests = nil
			lock.Unlock()

			res := s.Execute(append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", scheduleID}, tc.args...)...)
			assert.NoError(t, res.Err)

			lock.Lock()
			gotUpdates := append([]*workflowservice.UpdateScheduleRequest(nil), updateRequests...)
			lock.Unlock()
			if !assert.Len(t, gotUpdates, 1) {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			expected.Spec.StructuredCalendar = nil
			expected.Spec.Calendar = nil
			expected.Spec.CronString = tc.cron
			expected.Spec.Interval = tc.intervals
			assert.True(t, proto.Equal(expected, gotUpdates[0].GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchOmitsCadenceWithoutChangingIt() {
	describedSchedule := &schedule.Schedule{
		Spec: &schedule.ScheduleSpec{
			StructuredCalendar:        []*schedule.StructuredCalendarSpec{{}},
			Calendar:                  []*schedule.CalendarSpec{{Minute: "1"}},
			CronString:                []string{"0 12 * * *"},
			Interval:                  []*schedule.IntervalSpec{{Interval: durationpb.New(time.Hour)}},
			ExcludeCalendar:           []*schedule.CalendarSpec{{Minute: "2"}},
			ExcludeStructuredCalendar: []*schedule.StructuredCalendarSpec{{}},
			TimezoneName:              "America/New_York",
		},
		State: &schedule.ScheduleState{Notes: "before"},
	}
	var updateRequest *workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequest = proto.Clone(request).(*workflowservice.UpdateScheduleRequest)
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "patch-omit-cadence", "--notes", "after")
	s.NoError(res.Err)
	if !assert.NotNil(s.T(), updateRequest) {
		return
	}
	expected := proto.Clone(describedSchedule).(*schedule.Schedule)
	expected.State.Notes = "after"
	s.True(proto.Equal(expected, updateRequest.GetSchedule()))
	describedSpec, err := proto.Marshal(describedSchedule.GetSpec())
	s.NoError(err)
	submittedSpec, err := proto.Marshal(updateRequest.GetSchedule().GetSpec())
	s.NoError(err)
	s.Equal(describedSpec, submittedSpec)
}

func (s *SharedServerSuite) TestSchedule_PatchClearsCadenceOnlyForPausedResult() {
	var describedSchedule *schedule.Schedule
	var updateRequest *workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequest = proto.Clone(request).(*workflowservice.UpdateScheduleRequest)
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name   string
		paused bool
		args   []string
	}{
		{name: "already paused", paused: true},
		{name: "same patch pauses", paused: false, args: []string{"--paused=true"}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describedSchedule = &schedule.Schedule{
				Spec: &schedule.ScheduleSpec{
					StructuredCalendar:        []*schedule.StructuredCalendarSpec{{}},
					Calendar:                  []*schedule.CalendarSpec{{Minute: "1"}},
					CronString:                []string{"0 12 * * *"},
					Interval:                  []*schedule.IntervalSpec{{Interval: durationpb.New(time.Hour)}},
					ExcludeCalendar:           []*schedule.CalendarSpec{{Minute: "2"}},
					ExcludeStructuredCalendar: []*schedule.StructuredCalendarSpec{{}},
					StartTime:                 timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
					EndTime:                   timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
					Jitter:                    durationpb.New(time.Minute),
					TimezoneName:              "America/New_York",
				},
				State: &schedule.ScheduleState{Paused: tc.paused},
			}
			updateRequest = nil

			args := append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "patch-clear-cadence", "--cadence-clear-all"}, tc.args...)
			res := s.Execute(args...)
			assert.NoError(t, res.Err)
			if !assert.NotNil(t, updateRequest) {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			expected.Spec.StructuredCalendar = nil
			expected.Spec.Calendar = nil
			expected.Spec.CronString = nil
			expected.Spec.Interval = nil
			if !tc.paused {
				expected.State.Paused = true
			}
			assert.True(t, proto.Equal(expected, updateRequest.GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchRejectsCadenceClearWhenResultIsUnpaused() {
	var updateRequests atomic.Int32
	var describedState *schedule.ScheduleState
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = &schedule.Schedule{State: describedState}
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests.Add(1)
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		state *schedule.ScheduleState
		args  []string
	}{
		{name: "schedule is active", state: &schedule.ScheduleState{Paused: false}, args: []string{"--cadence-clear-all"}},
		{name: "schedule state is absent", state: nil, args: []string{"--cadence-clear-all"}},
		{name: "patch explicitly unpauses", state: &schedule.ScheduleState{Paused: true}, args: []string{"--cadence-clear-all", "--paused=false"}},
	} {
		describedState = tc.state
		res := s.Execute(append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "patch-unpaused-clear"}, tc.args...)...)
		s.Error(res.Err, tc.name)
		s.ErrorContains(res.Err, "use --paused=true to pause explicitly", tc.name)
	}
	s.Equal(int32(0), updateRequests.Load())
}

func (s *SharedServerSuite) TestSchedule_PatchClearsPausedCadenceToManualOnly() {
	scheduleID, workflowID, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", scheduleID, "--paused=true", "--cadence-clear-all")
	s.NoError(res.Err)

	var description struct {
		Schedule struct {
			Spec struct {
				StructuredCalendar []json.RawMessage `json:"structuredCalendar"`
				Calendar           []json.RawMessage `json:"calendar"`
				CronString         []string          `json:"cronString"`
				Interval           []json.RawMessage `json:"interval"`
			} `json:"spec"`
			State struct {
				Paused bool `json:"paused"`
			} `json:"state"`
		} `json:"schedule"`
		Info struct {
			FutureActionTimes []json.RawMessage `json:"futureActionTimes"`
		} `json:"info"`
	}
	s.Eventually(func() bool {
		res = s.Execute("schedule", "describe", "--address", s.Address(), "--schedule-id", scheduleID, "--output", "json")
		description = struct {
			Schedule struct {
				Spec struct {
					StructuredCalendar []json.RawMessage `json:"structuredCalendar"`
					Calendar           []json.RawMessage `json:"calendar"`
					CronString         []string          `json:"cronString"`
					Interval           []json.RawMessage `json:"interval"`
				} `json:"spec"`
				State struct {
					Paused bool `json:"paused"`
				} `json:"state"`
			} `json:"schedule"`
			Info struct {
				FutureActionTimes []json.RawMessage `json:"futureActionTimes"`
			} `json:"info"`
		}{}
		if res.Err != nil || json.Unmarshal(res.Stdout.Bytes(), &description) != nil {
			return false
		}
		return description.Schedule.State.Paused && len(description.Schedule.Spec.StructuredCalendar) == 0 && len(description.Schedule.Spec.Calendar) == 0 && len(description.Schedule.Spec.CronString) == 0 && len(description.Schedule.Spec.Interval) == 0 && len(description.Info.FutureActionTimes) == 0
	}, 10*time.Second, 100*time.Millisecond)

	res = s.Execute("schedule", "trigger", "--address", s.Address(), "--schedule-id", scheduleID)
	s.NoError(res.Err)
	s.Eventually(func() bool {
		res = s.Execute("workflow", "list", "--address", s.Address(), "-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, scheduleID))
		return res.Err == nil && AssertContainsOnSameLine(res.Stdout.String(), workflowID) == nil
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_PatchRevalidatesCadenceClearAfterConflictRefresh() {
	var describes int
	var updates []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				describes++
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = &schedule.Schedule{
					Spec:  &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}, Interval: []*schedule.IntervalSpec{{Interval: durationpb.New(time.Hour)}}},
					State: &schedule.ScheduleState{Paused: describes == 1},
				}
				response.ConflictToken = []byte(fmt.Sprintf("token-%d", describes))
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updates = append(updates, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				if len(updates) == 1 {
					return status.Error(codes.FailedPrecondition, "mismatched conflict token")
				}
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name        string
		args        []string
		wantErr     string
		wantUpdates int
	}{
		{name: "refreshed unpaused fails", args: []string{"--cadence-clear-all"}, wantErr: "use --paused=true to pause explicitly", wantUpdates: 1},
		{name: "explicit pause reapplies", args: []string{"--cadence-clear-all", "--paused=true"}, wantUpdates: 2},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describes = 0
			updates = nil

			res := s.Execute(append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "patch-refresh-clear"}, tc.args...)...)
			if tc.wantErr != "" {
				assert.ErrorContains(t, res.Err, tc.wantErr)
			} else {
				assert.NoError(t, res.Err)
			}
			assert.Equal(t, 2, describes)
			assert.Len(t, updates, tc.wantUpdates)
			for _, update := range updates {
				assert.True(t, update.GetSchedule().GetState().GetPaused())
				assert.Empty(t, update.GetSchedule().GetSpec().GetStructuredCalendar())
				assert.Empty(t, update.GetSchedule().GetSpec().GetCalendar())
				assert.Empty(t, update.GetSchedule().GetSpec().GetCronString())
				assert.Empty(t, update.GetSchedule().GetSpec().GetInterval())
			}
			if tc.wantUpdates == 2 && len(updates) == 2 {
				assert.Equal(t, []byte("token-2"), updates[1].GetConflictToken())
				assert.NotEqual(t, updates[0].GetRequestId(), updates[1].GetRequestId())
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsTimingFieldsIndependently() {
	initialStart := timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	initialEnd := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	describedSchedule := &schedule.Schedule{
		Spec: &schedule.ScheduleSpec{
			CronString:      []string{"0 12 * * *"},
			Interval:        []*schedule.IntervalSpec{{Interval: durationpb.New(time.Hour)}},
			ExcludeCalendar: []*schedule.CalendarSpec{{Minute: "2"}},
			StartTime:       initialStart,
			EndTime:         initialEnd,
			Jitter:          durationpb.New(time.Minute),
			TimezoneName:    "America/New_York",
			TimezoneData:    []byte("tzif"),
		},
		Action:   &schedule.ScheduleAction{},
		Policies: &schedule.SchedulePolicies{PauseOnFailure: true},
		State:    &schedule.ScheduleState{Notes: "preserved", Paused: true},
	}
	var updateRequest *workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequest = proto.Clone(request).(*workflowservice.UpdateScheduleRequest)
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		args  []string
		apply func(*schedule.Schedule)
	}{
		{name: "start", args: []string{"--start-time", "2027-01-01T00:00:00Z"}, apply: func(s *schedule.Schedule) {
			s.Spec.StartTime = timestamppb.New(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
		}},
		{name: "end", args: []string{"--end-time", "2028-01-01T00:00:00Z"}, apply: func(s *schedule.Schedule) {
			s.Spec.EndTime = timestamppb.New(time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC))
		}},
		{name: "jitter", args: []string{"--jitter", "30s"}, apply: func(s *schedule.Schedule) { s.Spec.Jitter = durationpb.New(30 * time.Second) }},
		{name: "zero jitter", args: []string{"--jitter", "0s"}, apply: func(s *schedule.Schedule) { s.Spec.Jitter = durationpb.New(0) }},
		{name: "time zone", args: []string{"--time-zone", "Asia/Tokyo"}, apply: func(s *schedule.Schedule) { s.Spec.TimezoneName, s.Spec.TimezoneData = "Asia/Tokyo", nil }},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequest = nil
			res := s.Execute(append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "patch-timing"}, tc.args...)...)
			assert.NoError(t, res.Err)
			if !assert.NotNil(t, updateRequest) {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected)
			assert.True(t, proto.Equal(expected, updateRequest.GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchUnsetsTimingFieldsAndRejectsSetUnsetConflicts() {
	describedSchedule := &schedule.Schedule{Spec: &schedule.ScheduleSpec{
		CronString:      []string{"0 12 * * *"},
		Interval:        []*schedule.IntervalSpec{{Interval: durationpb.New(time.Hour)}},
		ExcludeCalendar: []*schedule.CalendarSpec{{Minute: "2"}},
		StartTime:       timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		EndTime:         timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		Jitter:          durationpb.New(time.Minute),
		TimezoneName:    "America/New_York",
		TimezoneData:    []byte("tzif"),
	}, Action: &schedule.ScheduleAction{}, Policies: &schedule.SchedulePolicies{PauseOnFailure: true}, State: &schedule.ScheduleState{Notes: "preserved", Paused: true}}
	var updateRequest *workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequest = proto.Clone(request).(*workflowservice.UpdateScheduleRequest)
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name  string
		arg   string
		apply func(*schedule.Schedule)
	}{
		{name: "start", arg: "--unset-start-time", apply: func(s *schedule.Schedule) { s.Spec.StartTime = nil }},
		{name: "end", arg: "--unset-end-time", apply: func(s *schedule.Schedule) { s.Spec.EndTime = nil }},
		{name: "jitter", arg: "--unset-jitter", apply: func(s *schedule.Schedule) { s.Spec.Jitter = nil }},
		{name: "time zone", arg: "--unset-time-zone", apply: func(s *schedule.Schedule) { s.Spec.TimezoneName, s.Spec.TimezoneData = "", nil }},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			updateRequest = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "patch-unset-timing", tc.arg)
			assert.NoError(t, res.Err)
			if !assert.NotNil(t, updateRequest) {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			tc.apply(expected)
			assert.True(t, proto.Equal(expected, updateRequest.GetSchedule()))
		})
	}
	for _, tc := range []struct {
		name          string
		args          []string
		errorContains string
	}{
		{name: "start time", args: []string{"--start-time", "2027-01-01T00:00:00Z", "--unset-start-time"}, errorContains: "--start-time and --unset-start-time are mutually exclusive"},
		{name: "end time", args: []string{"--end-time", "2027-01-01T00:00:00Z", "--unset-end-time"}, errorContains: "--end-time and --unset-end-time are mutually exclusive"},
		{name: "jitter", args: []string{"--jitter", "1s", "--unset-jitter"}, errorContains: "--jitter and --unset-jitter are mutually exclusive"},
		{name: "time zone", args: []string{"--time-zone", "UTC", "--unset-time-zone"}, errorContains: "--time-zone and --unset-time-zone are mutually exclusive"},
	} {
		var dials atomic.Int32
		options := s.CommandHarness.Options
		options.Args = append([]string{"schedule", "patch", "--schedule-id", "patch-unset-timing"}, tc.args...)
		options.AdditionalClientGRPCDialOptions = append(options.AdditionalClientGRPCDialOptions, grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			dials.Add(1)
			return nil, errors.New("unexpected dial")
		}))
		var commandErr error
		options.Fail = func(err error) { commandErr = err }
		temporalcli.Execute(context.Background(), options)
		assert.ErrorContains(s.T(), commandErr, tc.errorContains, tc.name)
		assert.Zero(s.T(), dials.Load(), tc.name)
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsOnlyOverlapPolicy() {
	var describedSchedule *schedule.Schedule
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = []byte("overlap-token")
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name     string
		overlap  string
		schedule *schedule.Schedule
	}{
		{
			name:    "preserves full schedule",
			overlap: "Skip",
			schedule: &schedule.Schedule{
				Spec: &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
				Policies: &schedule.SchedulePolicies{
					OverlapPolicy:  enums.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL,
					CatchupWindow:  durationpb.New(30 * time.Second),
					PauseOnFailure: true,
				},
				State: &schedule.ScheduleState{
					Notes:            "preserved notes",
					Paused:           true,
					LimitedActions:   true,
					RemainingActions: 4,
				},
			},
		},
		{
			name:     "creates missing policies",
			overlap:  "BufferAll",
			schedule: &schedule.Schedule{State: &schedule.ScheduleState{Notes: "preserved notes"}},
		},
		{
			name:    "keeps explicit same value",
			overlap: "BufferAll",
			schedule: &schedule.Schedule{Policies: &schedule.SchedulePolicies{
				OverlapPolicy: enums.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL,
			}},
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describedSchedule = tc.schedule
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", "--overlap-policy", tc.overlap)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) != 1 {
				return
			}
			expected := proto.Clone(describedSchedule).(*schedule.Schedule)
			if expected.Policies == nil {
				expected.Policies = &schedule.SchedulePolicies{}
			}
			expected.Policies.OverlapPolicy, _ = enums.ScheduleOverlapPolicyFromString(tc.overlap)
			assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsAndUnsetsCatchupWindow() {
	var describedSchedule *schedule.Schedule
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name          string
		args          []string
		schedule      *schedule.Schedule
		catchupWindow *durationpb.Duration
	}{
		{
			name: "set preserves full schedule",
			args: []string{"--catchup-window", "10s"},
			schedule: &schedule.Schedule{
				Spec: &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
				Policies: &schedule.SchedulePolicies{
					OverlapPolicy:  enums.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL,
					CatchupWindow:  durationpb.New(30 * time.Second),
					PauseOnFailure: true,
				},
				State: &schedule.ScheduleState{
					Notes:            "preserved notes",
					Paused:           true,
					LimitedActions:   true,
					RemainingActions: 4,
				},
			},
			catchupWindow: durationpb.New(10 * time.Second),
		},
		{
			name:          "set creates missing policies",
			args:          []string{"--catchup-window", "10s"},
			schedule:      &schedule.Schedule{State: &schedule.ScheduleState{Notes: "preserved notes"}},
			catchupWindow: durationpb.New(10 * time.Second),
		},
		{
			name:          "unset",
			args:          []string{"--unset-catchup-window"},
			schedule:      &schedule.Schedule{Policies: &schedule.SchedulePolicies{CatchupWindow: durationpb.New(30 * time.Second)}},
			catchupWindow: nil,
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describedSchedule = tc.schedule
			updateRequests = nil
			args := append([]string{"schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id"}, tc.args...)
			res := s.Execute(args...)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) == 1 {
				expected := proto.Clone(describedSchedule).(*schedule.Schedule)
				if expected.Policies == nil {
					expected.Policies = &schedule.SchedulePolicies{}
				}
				expected.Policies.CatchupWindow = tc.catchupWindow
				assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsPauseOnFailureWhenExplicit() {
	describedSchedule := &schedule.Schedule{State: &schedule.ScheduleState{Paused: true}}
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, pauseOnFailure := range []bool{true, false} {
		s.T().Run(strconv.FormatBool(pauseOnFailure), func(t *testing.T) {
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", "--pause-on-failure="+strconv.FormatBool(pauseOnFailure))
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) == 1 {
				assert.Equal(t, pauseOnFailure, updateRequests[0].GetSchedule().GetPolicies().GetPauseOnFailure())
				assert.True(t, updateRequests[0].GetSchedule().GetState().GetPaused())
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchChangesOnlyPausedState() {
	var describedSchedule *schedule.Schedule
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name     string
		paused   bool
		schedule *schedule.Schedule
	}{
		{name: "true with nil state", paused: true, schedule: &schedule.Schedule{}},
		{name: "false preserves notes", paused: false, schedule: &schedule.Schedule{State: &schedule.ScheduleState{Paused: true, Notes: "preserved notes"}}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describedSchedule = tc.schedule
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", "--paused="+strconv.FormatBool(tc.paused))
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) == 1 {
				expected := proto.Clone(tc.schedule).(*schedule.Schedule)
				if expected.State == nil {
					expected.State = &schedule.ScheduleState{}
				}
				expected.State.Paused = tc.paused
				assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchSetsRemainingActions() {
	var describedSchedule *schedule.Schedule
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name      string
		value     string
		schedule  *schedule.Schedule
		limited   bool
		remaining int64
	}{
		{
			name:  "positive preserves rich state",
			value: "3",
			schedule: &schedule.Schedule{
				Spec: &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
				Policies: &schedule.SchedulePolicies{
					OverlapPolicy:  enums.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL,
					CatchupWindow:  durationpb.New(20 * time.Second),
					PauseOnFailure: true,
				},
				State: &schedule.ScheduleState{
					Notes:  "preserved notes",
					Paused: true,
				},
			},
			limited:   true,
			remaining: 3,
		},
		{
			name:  "zero preserves rich state",
			value: "0",
			schedule: &schedule.Schedule{State: &schedule.ScheduleState{
				Notes:            "preserved notes",
				Paused:           true,
				LimitedActions:   true,
				RemainingActions: 9,
			}},
			limited:   false,
			remaining: 0,
		},
		{
			name:      "positive creates missing state",
			value:     "3",
			schedule:  &schedule.Schedule{},
			limited:   true,
			remaining: 3,
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			describedSchedule = tc.schedule
			updateRequests = nil
			res := s.Execute("schedule", "patch", "--address", s.Address(), "--schedule-id", "schedule-id", "--remaining-actions", tc.value)
			assert.NoError(t, res.Err)
			assert.Len(t, updateRequests, 1)
			if len(updateRequests) == 1 {
				expected := proto.Clone(describedSchedule).(*schedule.Schedule)
				if expected.State == nil {
					expected.State = &schedule.ScheduleState{}
				}
				expected.State.LimitedActions = tc.limited
				expected.State.RemainingActions = tc.remaining
				assert.True(t, proto.Equal(expected, updateRequests[0].GetSchedule()))
			}
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchRetriesExactConflictWithRefreshedState() {
	const (
		namespace  = "patch-conflict-namespace"
		scheduleID = "patch-conflict-schedule"
		identity   = "patch-conflict-identity"
	)
	conflictTokens := [][]byte{
		[]byte("stale-conflict-token"),
		[]byte("refreshed-conflict-token"),
	}
	describedSchedules := []*schedule.Schedule{
		{
			Spec: &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
			Policies: &schedule.SchedulePolicies{
				OverlapPolicy:  enums.SCHEDULE_OVERLAP_POLICY_SKIP,
				CatchupWindow:  durationpb.New(time.Minute),
				PauseOnFailure: true,
			},
			State: &schedule.ScheduleState{Notes: "stale notes"},
		},
		{
			Spec: &schedule.ScheduleSpec{
				CronString:   []string{"15 * * * *"},
				TimezoneName: "America/New_York",
			},
			Policies: &schedule.SchedulePolicies{
				OverlapPolicy:  enums.SCHEDULE_OVERLAP_POLICY_BUFFER_ALL,
				CatchupWindow:  durationpb.New(2 * time.Minute),
				PauseOnFailure: true,
			},
			State: &schedule.ScheduleState{
				Notes:            "concurrent notes",
				Paused:           true,
				LimitedActions:   true,
				RemainingActions: 7,
			},
		},
	}
	var lock sync.Mutex
	var describeRequests []*workflowservice.DescribeScheduleRequest
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests = append(describeRequests, proto.Clone(request).(*workflowservice.DescribeScheduleRequest))
				describeIndex := len(describeRequests) - 1
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedules[describeIndex]).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictTokens[describeIndex]...)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				updateAttempt := len(updateRequests)
				lock.Unlock()
				if updateAttempt == 1 {
					// Additional dial interceptors run inside the SDK error interceptor, so return
					// the wire status it converts to *serviceerror.FailedPrecondition for the command.
					return status.Error(codes.FailedPrecondition, "mismatched conflict token")
				}
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--namespace", namespace,
		"--identity", identity,
		"--schedule-id", scheduleID,
		"--notes", "requested notes",
		"--unset-catchup-window",
		"--pause-on-failure=false",
		"--paused=false",
		"--remaining-actions", "0",
	)
	s.NoError(res.Err)
	s.Equal("Schedule patch submitted\n", res.Stdout.String())
	s.Empty(res.Stderr.String())

	lock.Lock()
	gotDescribeRequests := append([]*workflowservice.DescribeScheduleRequest(nil), describeRequests...)
	gotUpdateRequests := append([]*workflowservice.UpdateScheduleRequest(nil), updateRequests...)
	lock.Unlock()
	s.Len(gotDescribeRequests, 2)
	s.Len(gotUpdateRequests, 2)
	if len(gotDescribeRequests) != 2 || len(gotUpdateRequests) != 2 {
		return
	}

	requestIDs := map[string]struct{}{}
	for i := range 2 {
		s.Equal(namespace, gotDescribeRequests[i].GetNamespace())
		s.Equal(scheduleID, gotDescribeRequests[i].GetScheduleId())

		updateRequest := gotUpdateRequests[i]
		s.Equal(namespace, updateRequest.GetNamespace())
		s.Equal(scheduleID, updateRequest.GetScheduleId())
		s.Equal(conflictTokens[i], updateRequest.GetConflictToken())
		s.Equal(identity, updateRequest.GetIdentity())
		s.NotEmpty(updateRequest.GetRequestId())
		_, exists := requestIDs[updateRequest.GetRequestId()]
		s.False(exists)
		requestIDs[updateRequest.GetRequestId()] = struct{}{}

		expectedSchedule := proto.Clone(describedSchedules[i]).(*schedule.Schedule)
		expectedSchedule.State.Notes = "requested notes"
		expectedSchedule.Policies.CatchupWindow = nil
		expectedSchedule.Policies.PauseOnFailure = false
		expectedSchedule.State.Paused = false
		expectedSchedule.State.LimitedActions = false
		expectedSchedule.State.RemainingActions = 0
		s.True(proto.Equal(expectedSchedule, updateRequest.GetSchedule()))
	}
}

func (s *SharedServerSuite) TestSchedule_PatchStopsAfterThreeExactConflicts() {
	const conflictMessage = "mismatched conflict token"
	conflictErrors := make([]error, 3)
	for i := range conflictErrors {
		conflictStatus, err := status.New(codes.FailedPrecondition, conflictMessage).WithDetails(
			wrapperspb.String(fmt.Sprintf("attempt-%d", i+1)),
		)
		if err != nil {
			s.T().Fatalf("failed to construct conflict status: %v", err)
		}
		conflictErrors[i] = conflictStatus.Err()
	}

	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				describeAttempt := describeRequests
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = &schedule.Schedule{State: &schedule.ScheduleState{Notes: "existing notes"}}
				response.ConflictToken = []byte(fmt.Sprintf("conflict-token-%d", describeAttempt))
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				updateAttempt := updateRequests
				lock.Unlock()
				return conflictErrors[updateAttempt-1]
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--schedule-id", "patch-repeated-conflict-schedule",
		"--notes", "requested notes",
	)
	s.Error(res.Err)
	s.NotContains(res.Stdout.String(), "Schedule patch submitted")
	s.NotContains(res.Stderr.String(), "Schedule patch may already have been submitted")

	conflictErr, ok := res.Err.(*serviceerror.FailedPrecondition)
	if !assert.True(s.T(), ok) {
		return
	}
	s.Equal(conflictMessage, conflictErr.Message)
	details := conflictErr.Status().Details()
	if !assert.Len(s.T(), details, 1) {
		return
	}
	marker, ok := details[0].(*wrapperspb.StringValue)
	if !assert.True(s.T(), ok) {
		return
	}
	s.Equal("attempt-3", marker.GetValue())

	lock.Lock()
	gotDescribeRequests := describeRequests
	gotUpdateRequests := updateRequests
	lock.Unlock()
	s.Equal(3, gotDescribeRequests)
	s.Equal(3, gotUpdateRequests)
}

func (s *SharedServerSuite) TestSchedule_PatchDoesNotRetryOtherTypedFailedPreconditions() {
	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	updateError := error(nil)
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = &schedule.Schedule{State: &schedule.ScheduleState{Notes: "existing notes"}}
				response.ConflictToken = []byte("conflict-token")
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				err := updateError
				lock.Unlock()
				return err
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name    string
		message string
	}{
		{name: "near-miss message", message: "mismatched conflict token "},
		{name: "unrelated failure", message: "schedule is paused"},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			lock.Lock()
			describeRequests = 0
			updateRequests = 0
			// The inner status is converted by the SDK to the concrete service error
			// whose message the command must compare exactly.
			updateError = status.Error(codes.FailedPrecondition, tc.message)
			lock.Unlock()

			res := s.Execute(
				"schedule", "patch",
				"--address", s.Address(),
				"--schedule-id", "patch-other-precondition-schedule",
				"--notes", "requested notes",
			)
			assert.Error(t, res.Err)
			assert.NotContains(t, res.Stdout.String(), "Schedule patch submitted")
			assert.NotContains(t, res.Stderr.String(), "Schedule patch may already have been submitted")
			conflictErr, ok := res.Err.(*serviceerror.FailedPrecondition)
			if assert.True(t, ok) {
				assert.Equal(t, tc.message, conflictErr.Message)
			}

			lock.Lock()
			gotDescribeRequests := describeRequests
			gotUpdateRequests := updateRequests
			lock.Unlock()
			assert.Equal(t, 1, gotDescribeRequests)
			assert.Equal(t, 1, gotUpdateRequests)
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchDoesNotRetryBroadOrAmbiguousErrors() {
	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	updateError := error(nil)
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		// WithUnaryInterceptor is prepended outside the SDK error interceptor. This
		// lets the broad status case reach the command without typed conversion.
		grpc.WithUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = &schedule.Schedule{State: &schedule.ScheduleState{Notes: "existing notes"}}
				response.ConflictToken = []byte("conflict-token")
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				err := updateError
				lock.Unlock()
				return err
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name string
		err  error
	}{
		{
			name: "broad status with exact conflict text",
			err:  status.Error(codes.FailedPrecondition, "mismatched conflict token"),
		},
		{name: "deadline", err: context.DeadlineExceeded},
		{name: "unavailable", err: status.Error(codes.Unavailable, "service unavailable")},
		{name: "transport", err: io.ErrUnexpectedEOF},
		{name: "ordinary", err: errors.New("ordinary update failure")},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			lock.Lock()
			describeRequests = 0
			updateRequests = 0
			updateError = tc.err
			lock.Unlock()

			res := s.Execute(
				"schedule", "patch",
				"--address", s.Address(),
				"--schedule-id", "patch-ambiguous-error-schedule",
				"--notes", "requested notes",
			)
			assert.ErrorIs(t, res.Err, tc.err)
			assert.NotContains(t, res.Stdout.String(), "Schedule patch submitted")
			assert.NotContains(t, res.Stderr.String(), "Schedule patch may already have been submitted")

			lock.Lock()
			gotDescribeRequests := describeRequests
			gotUpdateRequests := updateRequests
			lock.Unlock()
			assert.Equal(t, 1, gotDescribeRequests)
			assert.Equal(t, 1, gotUpdateRequests)
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchCommandDeadlineStopsDuringRefreshedDescribe() {
	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				describeAttempt := describeRequests
				lock.Unlock()
				if describeAttempt == 2 {
					<-ctx.Done()
					return status.FromContextError(ctx.Err()).Err()
				}
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = &schedule.Schedule{State: &schedule.ScheduleState{Notes: "existing notes"}}
				response.ConflictToken = []byte("stale-conflict-token")
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				lock.Unlock()
				return status.Error(codes.FailedPrecondition, "mismatched conflict token")
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--command-timeout", "500ms",
		"--schedule-id", "patch-deadline-schedule",
		"--notes", "requested notes",
	)
	s.Error(res.Err)
	deadlineErr, ok := res.Err.(*serviceerror.DeadlineExceeded)
	if assert.True(s.T(), ok) {
		s.Equal(context.DeadlineExceeded.Error(), deadlineErr.Message)
	}
	s.NotContains(res.Stdout.String(), "Schedule patch submitted")
	s.NotContains(res.Stderr.String(), "Schedule patch may already have been submitted")

	lock.Lock()
	gotDescribeRequests := describeRequests
	gotUpdateRequests := updateRequests
	lock.Unlock()
	s.Equal(2, gotDescribeRequests)
	s.Equal(1, gotUpdateRequests)
}

func (s *SharedServerSuite) TestSchedule_PatchSubmitsOneUpdateThatClearsOnlyNotes() {
	const (
		namespace  = "patch-unset-notes-namespace"
		scheduleID = "patch-unset-notes-schedule"
		identity   = "patch-unset-notes-identity"
	)
	conflictToken := []byte("patch-unset-notes-conflict-token")
	describedSchedule := &schedule.Schedule{
		Spec: &schedule.ScheduleSpec{
			CronString:   []string{"0 12 * * *"},
			TimezoneName: "America/New_York",
			TimezoneData: []byte{1, 2, 3},
		},
		Policies: &schedule.SchedulePolicies{
			PauseOnFailure:         true,
			KeepOriginalWorkflowId: true,
		},
		State: &schedule.ScheduleState{
			Notes:            "notes to clear",
			Paused:           true,
			LimitedActions:   true,
			RemainingActions: 3,
		},
	}
	var lock sync.Mutex
	var describeRequests []*workflowservice.DescribeScheduleRequest
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests = append(describeRequests, proto.Clone(request).(*workflowservice.DescribeScheduleRequest))
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictToken...)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				lock.Unlock()
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--namespace", namespace,
		"--identity", identity,
		"--schedule-id", scheduleID,
		"--unset-notes",
	)
	s.NoError(res.Err)

	lock.Lock()
	gotDescribeRequests := append([]*workflowservice.DescribeScheduleRequest(nil), describeRequests...)
	gotUpdateRequests := append([]*workflowservice.UpdateScheduleRequest(nil), updateRequests...)
	lock.Unlock()
	s.Len(gotDescribeRequests, 1)
	s.Len(gotUpdateRequests, 1)
	if len(gotDescribeRequests) != 1 || len(gotUpdateRequests) != 1 {
		return
	}

	s.Equal(namespace, gotDescribeRequests[0].GetNamespace())
	s.Equal(scheduleID, gotDescribeRequests[0].GetScheduleId())

	updateRequest := gotUpdateRequests[0]
	s.Equal(namespace, updateRequest.GetNamespace())
	s.Equal(scheduleID, updateRequest.GetScheduleId())
	s.Equal(conflictToken, updateRequest.GetConflictToken())
	s.Equal(identity, updateRequest.GetIdentity())
	s.NotEmpty(updateRequest.GetRequestId())

	expectedSchedule := proto.Clone(describedSchedule).(*schedule.Schedule)
	expectedSchedule.State.Notes = ""
	s.True(proto.Equal(expectedSchedule, updateRequest.GetSchedule()))
}

func (s *SharedServerSuite) TestSchedule_PatchDescribeErrorsStopBeforeUpdate() {
	const injectedDescribeError = "injected Describe error"

	var lock sync.Mutex
	var injectDescribeError bool
	var describeRequests []*workflowservice.DescribeScheduleRequest
	var updateRequests []*workflowservice.UpdateScheduleRequest
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch request := req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests = append(describeRequests, proto.Clone(request).(*workflowservice.DescribeScheduleRequest))
				inject := injectDescribeError
				lock.Unlock()
				if inject {
					return errors.New(injectedDescribeError)
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests = append(updateRequests, proto.Clone(request).(*workflowservice.UpdateScheduleRequest))
				lock.Unlock()
				return invoker(ctx, method, req, reply, cc, opts...)
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name          string
		scheduleID    string
		inject        bool
		errorContains string
	}{
		{
			name:          "injected Describe error",
			scheduleID:    "patch-injected-describe-error",
			inject:        true,
			errorContains: injectedDescribeError,
		},
		{
			name:       "nonexistent Schedule",
			scheduleID: "patch-nonexistent-schedule",
		},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			lock.Lock()
			injectDescribeError = tc.inject
			describeRequests = nil
			updateRequests = nil
			lock.Unlock()

			res := s.Execute(
				"schedule", "patch",
				"--address", s.Address(),
				"--schedule-id", tc.scheduleID,
				"--notes", "updated notes",
			)
			assert.Error(t, res.Err)
			if tc.errorContains != "" {
				assert.ErrorContains(t, res.Err, tc.errorContains)
			}

			lock.Lock()
			gotDescribeRequests := append([]*workflowservice.DescribeScheduleRequest(nil), describeRequests...)
			gotUpdateRequests := append([]*workflowservice.UpdateScheduleRequest(nil), updateRequests...)
			lock.Unlock()
			assert.Len(t, gotDescribeRequests, 1)
			assert.Len(t, gotUpdateRequests, 0)
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchUpdateFailureDoesNotClaimSubmission() {
	const (
		namespace  = "patch-update-error-namespace"
		scheduleID = "patch-update-error-schedule"
	)
	wantErr := errors.New("injected UpdateSchedule error")
	conflictToken := []byte("patch-update-error-conflict-token")
	describedSchedule := &schedule.Schedule{
		State: &schedule.ScheduleState{Notes: "existing notes"},
	}

	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictToken...)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				lock.Unlock()
				return wantErr
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	res := s.Execute(
		"schedule", "patch",
		"--address", s.Address(),
		"--namespace", namespace,
		"--schedule-id", scheduleID,
		"--notes", "updated notes",
	)
	assert.Error(s.T(), res.Err)
	assert.ErrorContains(s.T(), res.Err, wantErr.Error())
	assert.NotContains(s.T(), res.Stdout.String(), "Schedule patch submitted\n")
	assert.NotContains(s.T(), res.Stderr.String(), "Schedule patch may already have been submitted")

	lock.Lock()
	gotDescribeRequests := describeRequests
	gotUpdateRequests := updateRequests
	lock.Unlock()
	assert.Equal(s.T(), 1, gotDescribeRequests)
	assert.Equal(s.T(), 1, gotUpdateRequests)
}

func (s *SharedServerSuite) TestSchedule_PatchSuccessfulOutputModes() {
	const (
		namespace  = "patch-output-namespace"
		scheduleID = "patch-output-schedule"
	)
	conflictToken := []byte("patch-output-conflict-token")
	describedSchedule := &schedule.Schedule{
		State: &schedule.ScheduleState{Notes: "existing notes"},
	}

	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictToken...)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				lock.Unlock()
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, tc := range []struct {
		name       string
		outputArgs []string
		wantStdout string
	}{
		{name: "text", wantStdout: "Schedule patch submitted\n"},
		{name: "json", outputArgs: []string{"--output", "json"}},
		{name: "jsonl", outputArgs: []string{"--output", "jsonl"}},
		{name: "none", outputArgs: []string{"--output", "none"}},
	} {
		s.T().Run(tc.name, func(t *testing.T) {
			lock.Lock()
			describeRequests = 0
			updateRequests = 0
			lock.Unlock()

			args := append([]string{
				"schedule", "patch",
				"--address", s.Address(),
				"--namespace", namespace,
				"--schedule-id", scheduleID,
				"--notes", "updated notes",
			}, tc.outputArgs...)
			res := s.Execute(args...)
			assert.NoError(t, res.Err)
			assert.Equal(t, tc.wantStdout, res.Stdout.String())
			assert.Empty(t, res.Stderr.String())

			lock.Lock()
			gotDescribeRequests := describeRequests
			gotUpdateRequests := updateRequests
			lock.Unlock()
			assert.Equal(t, 1, gotDescribeRequests)
			assert.Equal(t, 1, gotUpdateRequests)
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchWarnsWhenAcknowledgementWriteFails() {
	const (
		namespace  = "patch-write-error-namespace"
		scheduleID = "patch-write-error-schedule"
	)
	conflictToken := []byte("patch-write-error-conflict-token")
	describedSchedule := &schedule.Schedule{
		State: &schedule.ScheduleState{Notes: "existing notes"},
	}

	var lock sync.Mutex
	describeRequests := 0
	updateRequests := 0
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			switch req.(type) {
			case *workflowservice.DescribeScheduleRequest:
				lock.Lock()
				describeRequests++
				lock.Unlock()
				response := reply.(*workflowservice.DescribeScheduleResponse)
				response.Schedule = proto.Clone(describedSchedule).(*schedule.Schedule)
				response.ConflictToken = append([]byte(nil), conflictToken...)
				return nil
			case *workflowservice.UpdateScheduleRequest:
				lock.Lock()
				updateRequests++
				lock.Unlock()
				return nil
			default:
				return invoker(ctx, method, req, reply, cc, opts...)
			}
		}),
	)

	for _, testCase := range []struct {
		name    string
		wantErr error
	}{
		{name: "ordinary write error", wantErr: errors.New("stdout write failed")},
		{name: "broken pipe", wantErr: syscall.EPIPE},
	} {
		s.T().Run(testCase.name, func(t *testing.T) {
			lock.Lock()
			describeRequests = 0
			updateRequests = 0
			lock.Unlock()

			stdout := &failAfterWriter{remaining: 1, err: testCase.wantErr}
			err, stderr := s.executeWithStdout(
				stdout,
				"schedule", "patch",
				"--address", s.Address(),
				"--namespace", namespace,
				"--schedule-id", scheduleID,
				"--notes", "updated notes",
			)
			assert.Equal(t, testCase.wantErr, err)
			assert.ErrorIs(t, err, testCase.wantErr)
			assert.NotContains(t, stdout.buf.String(), "Schedule patch submitted\n")
			assert.Equal(t, "Schedule patch may already have been submitted\nError: "+testCase.wantErr.Error()+"\n", stderr)

			lock.Lock()
			gotDescribeRequests := describeRequests
			gotUpdateRequests := updateRequests
			lock.Unlock()
			assert.Equal(t, 1, gotDescribeRequests)
			assert.Equal(t, 1, gotUpdateRequests)
		})
	}
}
