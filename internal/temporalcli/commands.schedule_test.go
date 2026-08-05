package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/temporalio/cli/internal/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/schedule/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type failAfterWriter struct {
	buf       bytes.Buffer
	remaining int
	err       error
}

type failListFinalizationWriter struct {
	buf bytes.Buffer
	err error
}

type shortWriteBuffer struct {
	buf bytes.Buffer
}

type failOnJSONItemWriter struct {
	buf        bytes.Buffer
	itemWrites int
	failItem   int
	err        error
}

type failItemAndFinalizationWriter struct {
	itemErr           error
	finalizationErr   error
	finalizationCalls int
}

func (w *failItemAndFinalizationWriter) Write(p []byte) (int, error) {
	switch {
	case bytes.Equal(p, []byte("\n]\n")):
		w.finalizationCalls++
		if w.finalizationErr == nil {
			return len(p), nil
		}
		return 0, w.finalizationErr
	case len(p) > 0 && p[0] == '{':
		return 0, w.itemErr
	default:
		return len(p), nil
	}
}

func (w *failOnJSONItemWriter) Write(p []byte) (int, error) {
	if len(p) > 0 && p[0] == '{' {
		w.itemWrites++
		if w.itemWrites == w.failItem {
			return 0, w.err
		}
	}
	return w.buf.Write(p)
}

func (w *shortWriteBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return w.buf.Write(p[:len(p)-1])
}

func (w *failListFinalizationWriter) Write(p []byte) (int, error) {
	if bytes.Equal(p, []byte("\n]\n")) {
		return 0, w.err
	}
	return w.buf.Write(p)
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, w.err
	}
	if len(p) > w.remaining {
		p = p[:w.remaining]
	}
	n, _ := w.buf.Write(p)
	w.remaining -= n
	if w.remaining == 0 {
		return n, w.err
	}
	return n, nil
}

func (s *SharedServerSuite) executeWithStdout(stdout io.Writer, args ...string) (error, string) {
	options := s.CommandHarness.Options
	var stderr bytes.Buffer
	options.Stdin = &s.Stdin
	options.Stdout = stdout
	options.Stderr = &stderr
	options.Args = args
	options.DeprecatedEnvConfig.DisableEnvConfig = true
	options.DeprecatedEnvConfig.EnvConfigName = "default"
	var commandErr error
	options.Fail = func(err error) {
		commandErr = err
		fmt.Fprintf(&stderr, "Error: %v\n", err)
	}
	temporalcli.Execute(s.Context, options)
	return commandErr, stderr.String()
}

func (s *SharedServerSuite) createSchedule(args ...string) (schedId, schedWfId string, res *CommandResult) {
	schedId = fmt.Sprintf("sched-%x", rand.Uint32())
	schedWfId = fmt.Sprintf("my-wf-id-%x", rand.Uint32())
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return nil, workflow.Sleep(ctx, 10*time.Second)
	})
	s.T().Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		options := temporalcli.CommandOptions{
			IOStreams: temporalcli.IOStreams{
				Stdout: io.Discard,
				Stderr: io.Discard,
			},
			Args: []string{
				"schedule", "delete",
				"--address", s.Address(),
				"-s", schedId,
			},
			Fail: func(error) {},
		}
		temporalcli.Execute(ctx, options)
	})
	res = s.Execute(append([]string{
		"schedule", "create",
		"--address", s.Address(),
		"-s", schedId,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", schedWfId,
	}, args...)...,
	)
	return
}

func (s *SharedServerSuite) updateSchedule(schedID, schedWorkflowID string, args ...string) *CommandResult {
	return s.Execute(append([]string{
		"schedule", "update",
		"--address", s.Address(),
		"--schedule-id", schedID,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", schedWorkflowID,
	}, args...)...)
}

func (s *SharedServerSuite) TestSchedule_Create() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestSchedule_CreateRejectsHeadersBeforeMutation() {
	var createRequests atomic.Int32
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := req.(*workflowservice.CreateScheduleRequest); ok {
			createRequests.Add(1)
		}
		return handler(ctx, req)
	})

	_, _, res := s.createSchedule("--interval", "10d", "--headers", "example=123")
	s.Error(res.Err)
	s.ErrorContains(res.Err, "headers are not supported for schedule actions")
	s.Equal(int32(0), createRequests.Load())
}

func (s *SharedServerSuite) TestSchedule_CreateAppliesPriorityAndFairness() {
	var createRequest *workflowservice.CreateScheduleRequest
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if request, ok := req.(*workflowservice.CreateScheduleRequest); ok {
			createRequest = request
		}
		return handler(ctx, req)
	})

	_, _, res := s.createSchedule(
		"--interval", "10d",
		"--priority-key", "2",
		"--fairness-key", "tenant-a",
		"--fairness-weight", "2.5",
	)
	s.NoError(res.Err)
	if createRequest == nil {
		s.Fail("CreateSchedule request was not captured")
		return
	}

	priority := createRequest.GetSchedule().GetAction().GetStartWorkflow().GetPriority()
	if priority == nil {
		s.Fail("CreateSchedule request did not include priority")
		return
	}
	s.Equal(int32(2), priority.GetPriorityKey())
	s.Equal("tenant-a", priority.GetFairnessKey())
	s.Equal(float32(2.5), priority.GetFairnessWeight())
}

func (s *SharedServerSuite) TestSchedule_CreateRejectsInvalidPriorityAndFairnessBeforeMutation() {
	var createRequests atomic.Int32
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := req.(*workflowservice.CreateScheduleRequest); ok {
			createRequests.Add(1)
		}
		return handler(ctx, req)
	})

	for _, testCase := range []struct {
		name string
		args []string
	}{
		{name: "negative priority", args: []string{"--priority-key", "-1"}},
		{name: "priority above maximum", args: []string{"--priority-key", "6"}},
		{name: "fairness key longer than 64 bytes", args: []string{"--fairness-key", strings.Repeat("a", 65)}},
		{name: "negative fairness weight", args: []string{"--fairness-weight", "-1"}},
		{name: "fairness weight below minimum", args: []string{"--fairness-weight", "0.0009"}},
		{name: "fairness weight above maximum", args: []string{"--fairness-weight", "1000.1"}},
	} {
		s.T().Run(testCase.name, func(t *testing.T) {
			args := append([]string{"--interval", "10d"}, testCase.args...)
			_, _, res := s.createSchedule(args...)
			if res.Err == nil {
				t.Fatal("schedule create returned nil error")
			}
		})
	}

	s.Equal(int32(0), createRequests.Load())
}

func (s *SharedServerSuite) TestSchedule_CreateAcceptsEmptyFairnessKeyAndZeroWeight() {
	var createRequest *workflowservice.CreateScheduleRequest
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if request, ok := req.(*workflowservice.CreateScheduleRequest); ok {
			createRequest = request
		}
		return handler(ctx, req)
	})

	_, _, res := s.createSchedule(
		"--interval", "10d",
		"--fairness-key=",
		"--fairness-weight", "0",
	)
	s.NoError(res.Err)
	if createRequest == nil {
		s.Fail("CreateSchedule request was not captured")
		return
	}

	s.Nil(createRequest.GetSchedule().GetAction().GetStartWorkflow().GetPriority())
}

func (s *SharedServerSuite) TestSchedule_Delete() {
	schedId, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// check exists
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "delete",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)

	// doesn't exist anymore
	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestSchedule_Describe() {
	schedId, schedWfId, res := s.createSchedule("--interval", "2s")
	s.NoError(res.Err)

	// run once manually so we see a running workflow

	res = s.Execute(
		"schedule", "trigger",
		"--address", s.Address(),
		"-s", schedId,
	)

	// text

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", schedId,
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		s.ContainsOnSameLine(out, "ScheduleId", schedId)
		s.ContainsOnSameLine(out, "Spec", "2s")
		return AssertContainsOnSameLine(out, "RunningWorkflows", schedWfId+"-") == nil
	}, 10*time.Second, 100*time.Millisecond)

	// json

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
		"-o", "json",
	)
	s.NoError(res.Err)
	var j struct {
		Schedule struct {
			Action struct {
				StartWorkflow struct {
					Id string `json:"workflowId"`
				} `json:"startWorkflow"`
			} `json:"action"`
		} `json:"schedule"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	s.Equal(schedWfId, j.Schedule.Action.StartWorkflow.Id)
}

func (s *SharedServerSuite) TestSchedule_DescribeReturnsWriteFailure() {
	schedID, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	for _, testCase := range []struct {
		name       string
		outputArgs []string
	}{
		{name: "text"},
		{name: "JSON", outputArgs: []string{"--output", "json"}},
		{name: "JSONL", outputArgs: []string{"--output", "jsonl"}},
	} {
		s.T().Run(testCase.name, func(t *testing.T) {
			wantErr := errors.New("stdout write failed")
			stdout := &failAfterWriter{remaining: 8, err: wantErr}
			args := []string{
				"schedule", "describe",
				"--address", s.Address(),
				"--schedule-id", schedID,
			}
			err, stderr := s.executeWithStdout(stdout, append(args, testCase.outputArgs...)...)

			assert.ErrorIs(t, err, wantErr)
			assert.Contains(t, stderr, wantErr.Error())
			assert.NotEmpty(t, stdout.buf.String())
		})
	}
}

func (s *SharedServerSuite) TestSchedule_DescribeStructuredReturnsSerializationFailure() {
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
			if _, ok := req.(*workflowservice.DescribeScheduleRequest); ok {
				resp := reply.(*workflowservice.DescribeScheduleResponse)
				resp.Schedule = &schedule.Schedule{
					State: &schedule.ScheduleState{Notes: string([]byte{0xff})},
				}
				return nil
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)

	for _, output := range []string{"json", "jsonl"} {
		s.T().Run(output, func(t *testing.T) {
			var stdout bytes.Buffer
			err, stderr := s.executeWithStdout(
				&stdout,
				"schedule", "describe",
				"--address", s.Address(),
				"--schedule-id", "serialization-failure",
				"--output", output,
			)

			assert.Error(t, err)
			assert.Contains(t, stderr, "invalid UTF-8")
		})
	}
}

func (s *SharedServerSuite) TestSchedule_DescribeTextReturnsSerializationFailure() {
	schedID, _, res := s.createSchedule("--calendar", `{"minute":"0","comment":"valid"}`)
	s.NoError(res.Err)

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
			if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
				return err
			}
			if _, ok := req.(*workflowservice.DescribeScheduleRequest); !ok {
				return nil
			}
			resp := reply.(*workflowservice.DescribeScheduleResponse)
			calendars := resp.GetSchedule().GetSpec().GetStructuredCalendar()
			if len(calendars) == 0 {
				return errors.New("valid Describe response has no structured calendar")
			}
			calendars[0].Comment = string([]byte{0xff})
			return nil
		}),
	)

	var stdout bytes.Buffer
	err, stderr := s.executeWithStdout(
		&stdout,
		"schedule", "describe",
		"--address", s.Address(),
		"--schedule-id", schedID,
	)

	s.Error(err)
	s.Contains(stderr, "invalid UTF-8")
	s.Empty(stdout.String())
}

func (s *SharedServerSuite) TestSchedule_CreateDescribeCalendar() {
	schedId, _, res := s.createSchedule("--calendar", `{"hour":"2,4","dayOfWeek":"thu,fri"}`)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "ScheduleId", schedId)
	s.ContainsOnSameLine(out, "Spec", "dayOfWeek")
}

func (s *SharedServerSuite) TestSchedule_CreateDescribe_SearchAttributes_Memo() {
	s.t.Skip("Skipped until issue #590- Error printing typed search attributes inside schedule actions - is fixed")

	schedId, _, res := s.createSchedule("--interval", "10d",
		"--schedule-search-attribute", `CustomKeywordField="schedule-string-val"`,
		"--search-attribute", `CustomKeywordField="workflow-string-val"`,
		"--schedule-memo", `schedMemo="data here"`,
		"--memo", `wfMemo="other data"`,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	// TODO: We have to disable shorthand payload encoding for now so these come out as base64.
	// After https://github.com/temporalio/api-go/pull/154, ensure these come out as nice strings.
	b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	s.ContainsOnSameLine(out, "SearchAttributes", "CustomKeywordField", b64(`"schedule-string-val"`))
	s.ContainsOnSameLine(out, "Memo", "schedMemo", `"data here"`) // somehow this one comes out as a string anyway
	s.ContainsOnSameLine(out, "Action", "CustomKeywordField", b64(`"workflow-string-val"`))
	s.ContainsOnSameLine(out, "Action", "wfMemo", b64(`"other data"`))
}

func (s *SharedServerSuite) TestSchedule_CreateDescribe_UserMetadata() {
	schedId, _, res := s.createSchedule("--interval", "10d",
		"--static-summary", "summ",
		"--static-details", "details",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Action", "Summary", "summ")
	s.ContainsOnSameLine(out, "Action", "Details", "details")
}

func (s *SharedServerSuite) TestSchedule_List() {
	res := s.Execute(
		"operator", "search-attribute", "create",
		"--address", s.Address(),
		"--name", "TestSchedule_List",
		"--type", "keyword",
	)
	s.NoError(res.Err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res = s.Execute(
			"operator", "search-attribute", "list",
			"--address", s.Address(),
			"-o", "json",
		)
		assert.NoError(t, res.Err)
		var jsonOut operatorservice.ListSearchAttributesResponse
		assert.NoError(t, temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
		assert.Equal(t, enums.INDEXED_VALUE_TYPE_KEYWORD, jsonOut.CustomAttributes["TestSchedule_List"])
	}, 10*time.Second, time.Second)

	schedId, _, res := s.createSchedule(
		"--interval",
		"10d",
		"--schedule-search-attribute", `TestSchedule_List="here"`,
	)
	s.NoError(res.Err)

	// table really-long
	var out string
	s.EventuallyWithT(func(t *assert.CollectT) {
		res = s.Execute(
			"schedule", "list",
			"--address", s.Address(),
			"--really-long",
		)
		assert.NoError(t, res.Err)
		out = res.Stdout.String()
		assert.Contains(t, out, schedId)
	}, 10*time.Second, time.Second)
	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "0s" /*jitter*/, "false", "{}" /*memo*/)
	s.ContainsOnSameLine(out, "TestSchedule_List")

	// table

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
	)
	s.NoError(res.Err)
	out = res.Stdout.String()

	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "false")

	// table long

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--long",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "false")

	// json

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"-o", "json",
	)
	s.NoError(res.Err)
	var j []struct {
		ScheduleId string `json:"scheduleId"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	ok := false
	for _, entry := range j {
		ok = ok || entry.ScheduleId == schedId
	}
	s.True(ok, "schedule not found in json result")

	// jsonl

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"-o", "jsonl",
	)
	s.NoError(res.Err)
	lines := bytes.Split(res.Stdout.Bytes(), []byte("\n"))
	ok = false
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var j struct {
			ScheduleId string `json:"scheduleId"`
		}
		s.NoError(json.Unmarshal(line, &j))
		ok = ok || j.ScheduleId == schedId
	}
	s.True(ok, "schedule not found in jsonl result")

	// JSON query (match)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'here'",
		"-o", "json",
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	ok = false
	for _, entry := range j {
		ok = ok || entry.ScheduleId == schedId
	}
	s.True(ok, "schedule not found in json result")

	// query (match)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'here'",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, schedId, "DevWorkflow", "false")

	// JSON query (no matches)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'notHere'",
		"-o", "json",
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
	ok = false
	for _, entry := range j {
		ok = ok || entry.ScheduleId == schedId
	}
	s.False(ok, "schedule found in json result, but should not be found")

	// query (no matches)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "TestSchedule_List = 'notHere'",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.NotContainsf(out, schedId, "schedule found, but should not be found")

	// query (invalid query field)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--query", "unknownField = 'notHere'",
	)
	s.Error(res.Err)
}

func (s *SharedServerSuite) TestSchedule_ListReturnsOpeningFailure() {
	wantErr := errors.New("stdout opening failed")
	stdout := &failAfterWriter{err: wantErr}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
		"--output", "json",
	)

	assert.ErrorIs(s.T(), err, wantErr)
	assert.Contains(s.T(), stderr, wantErr.Error())
}

func (s *SharedServerSuite) TestSchedule_ListReturnsFirstItemWriteFailure() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
	wantErr := errors.New("stdout first item failed")
	stdout := &failAfterWriter{err: wantErr}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
		"--output", "jsonl",
	)

	assert.ErrorIs(s.T(), err, wantErr)
	assert.Contains(s.T(), stderr, wantErr.Error())
}

func (s *SharedServerSuite) TestSchedule_ListFinalizesAfterItemFailureAndPreservesBothErrors() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
	itemErr := errors.New("stdout item failed")
	finalizationErr := errors.New("stdout finalization also failed")
	stdout := &failItemAndFinalizationWriter{
		itemErr:         itemErr,
		finalizationErr: finalizationErr,
	}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
		"--output", "json",
	)

	assert.ErrorIs(s.T(), err, itemErr)
	assert.ErrorIs(s.T(), err, finalizationErr)
	assert.Contains(s.T(), stderr, itemErr.Error())
	assert.Contains(s.T(), stderr, finalizationErr.Error())
	assert.Equal(s.T(), 1, stdout.finalizationCalls)
}

func (s *SharedServerSuite) TestSchedule_ListFinalizesAfterRPCFailureAndPreservesPrimaryError() {
	rpcErr := errors.New("list RPC failed")
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
			if _, ok := req.(*workflowservice.ListSchedulesRequest); ok {
				return rpcErr
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	stdout := &failItemAndFinalizationWriter{}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
		"--output", "json",
	)

	assert.ErrorContains(s.T(), err, rpcErr.Error())
	assert.Contains(s.T(), stderr, rpcErr.Error())
	assert.Equal(s.T(), 1, stdout.finalizationCalls)
}

func (s *SharedServerSuite) TestSchedule_ListTextFinalizesAfterIteratorFailure() {
	const iteratorErrMessage = "list iterator failed"
	iteratorErr := status.Error(codes.InvalidArgument, iteratorErrMessage)
	options := s.CommandHarness.Options
	options.AdditionalClientGRPCDialOptions = append(
		options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string,
			req any,
			reply any,
			cc *grpc.ClientConn,
			invoker grpc.UnaryInvoker,
			opts ...grpc.CallOption,
		) error {
			if _, ok := req.(*workflowservice.ListSchedulesRequest); ok {
				return iteratorErr
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	var stdout, stderr bytes.Buffer
	options.Stdin = &s.Stdin
	options.Stdout = &stdout
	options.Stderr = &stderr
	options.Args = []string{
		"schedule", "list",
		"--address", s.Address(),
	}
	options.DeprecatedEnvConfig.DisableEnvConfig = true
	options.DeprecatedEnvConfig.EnvConfigName = "default"
	var commandErr error
	options.Fail = func(err error) {
		commandErr = err
	}
	cctx, cancel, err := temporalcli.NewCommandContext(s.Context, options)
	s.NoError(err)
	defer cancel()
	cmd := temporalcli.NewTemporalCommand(cctx)
	cmd.Command.SetArgs(cctx.Options.Args)
	cmd.Command.SetOut(cctx.Options.Stdout)
	cmd.Command.SetErr(cctx.Options.Stderr)

	err = cmd.Command.ExecuteContext(cctx)

	s.NoError(err)
	s.ErrorContains(commandErr, iteratorErrMessage)
	var restartErr error
	s.NotPanics(func() {
		restartErr = cctx.Printer.StartListErr()
	})
	s.NoError(restartErr)
	s.NoError(cctx.Printer.EndListErr())
}

func (s *SharedServerSuite) TestSchedule_ListReturnsMiddleItemWriteFailure() {
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
			if _, ok := req.(*workflowservice.ListSchedulesRequest); !ok {
				return invoker(ctx, method, req, reply, cc, opts...)
			}
			reply.(*workflowservice.ListSchedulesResponse).Schedules = []*schedule.ScheduleListEntry{
				{ScheduleId: "first-item"},
				{ScheduleId: "middle-item"},
			}
			return nil
		}),
	)
	wantErr := errors.New("stdout middle item failed")
	stdout := &failOnJSONItemWriter{failItem: 2, err: wantErr}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
		"--output", "jsonl",
	)

	assert.ErrorIs(s.T(), err, wantErr)
	assert.Contains(s.T(), stderr, wantErr.Error())
	assert.Equal(s.T(), 2, stdout.itemWrites)
	assert.NotEmpty(s.T(), stdout.buf.String())
}

func (s *SharedServerSuite) TestSchedule_ListReturnsEveryItemSerializationFailure() {
	var schedules []*schedule.ScheduleListEntry
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
			if _, ok := req.(*workflowservice.ListSchedulesRequest); !ok {
				return invoker(ctx, method, req, reply, cc, opts...)
			}
			reply.(*workflowservice.ListSchedulesResponse).Schedules = schedules
			return nil
		}),
	)

	for _, testCase := range []struct {
		name      string
		schedules []*schedule.ScheduleListEntry
		wantPrior bool
	}{
		{
			name: "first item",
			schedules: []*schedule.ScheduleListEntry{
				{ScheduleId: string([]byte{0xff})},
			},
		},
		{
			name: "middle item",
			schedules: []*schedule.ScheduleListEntry{
				{ScheduleId: "valid-prior-item"},
				{ScheduleId: string([]byte{0xff})},
				{ScheduleId: "unreached-item"},
			},
			wantPrior: true,
		},
	} {
		s.T().Run(testCase.name, func(t *testing.T) {
			schedules = testCase.schedules
			var stdout bytes.Buffer

			err, stderr := s.executeWithStdout(
				&stdout,
				"schedule", "list",
				"--address", s.Address(),
				"--output", "jsonl",
			)

			assert.Error(t, err)
			assert.Contains(t, stderr, "invalid UTF-8")
			assert.Equal(t, testCase.wantPrior, strings.Contains(stdout.String(), "valid-prior-item"))
			assert.NotContains(t, stdout.String(), "unreached-item")
		})
	}
}

func (s *SharedServerSuite) TestSchedule_ListTextReturnsItemShortWrite() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
	stdout := &shortWriteBuffer{}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
	)

	assert.ErrorIs(s.T(), err, io.ErrShortWrite)
	assert.Contains(s.T(), stderr, io.ErrShortWrite.Error())
}

func (s *SharedServerSuite) TestSchedule_ListReturnsFinalizationFailure() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
	wantErr := errors.New("stdout finalization failed")
	stdout := &failListFinalizationWriter{err: wantErr}

	err, stderr := s.executeWithStdout(
		stdout,
		"schedule", "list",
		"--address", s.Address(),
		"--output", "json",
	)

	assert.ErrorIs(s.T(), err, wantErr)
	assert.Contains(s.T(), stderr, wantErr.Error())
	assert.NotEmpty(s.T(), stdout.buf.String())
}

func (s *SharedServerSuite) TestSchedule_ListNoneOutputRemainsEmpty() {
	_, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "list",
		"--address", s.Address(),
		"--output", "none",
	)

	s.NoError(res.Err)
	s.Empty(res.Stdout.String())
}

func (s *SharedServerSuite) TestSchedule_Toggle() {
	schedId, _, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	// pause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", schedId,
		"--pause",
		"--reason", "testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "true")
	s.ContainsOnSameLine(out, "Notes", "testing")

	// unpause

	res = s.Execute(
		"schedule", "toggle",
		"--address", s.Address(),
		"-s", schedId,
		"--unpause",
		"--reason", "we're done testing",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "describe",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "Paused", "false")
	s.ContainsOnSameLine(out, "Notes", "done testing")
}

func (s *SharedServerSuite) TestSchedule_Trigger() {
	schedId, schedWfId, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "trigger",
		"--address", s.Address(),
		"-s", schedId,
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"workflow", "list",
			"--address", s.Address(),
			"-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, schedId),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		return AssertContainsOnSameLine(out, schedWfId) == nil
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_Backfill() {
	schedId, schedWfId, res := s.createSchedule("--interval", "10d/5h")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "backfill",
		"--address", s.Address(),
		"-s", schedId,
		"--start-time", "2022-02-02T00:00:00Z",
		"--end-time", "2022-02-28T00:00:00Z",
		"--overlap-policy", "AllowAll",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"workflow", "list",
			"--address", s.Address(),
			"-q", fmt.Sprintf(`TemporalScheduledById = "%s"`, schedId),
		)
		s.NoError(res.Err)
		out := res.Stdout.String()
		re := regexp.MustCompile(regexp.QuoteMeta(schedWfId + "-2022-02"))
		return len(re.FindAllString(out, -1)) == 3
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_Update() {
	schedId, schedWfId, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.Execute(
		"schedule", "update",
		"--address", s.Address(),
		"-s", schedId,
		"--task-queue", "SomeOtherTq",
		"--type", "SomeOtherWf",
		"--workflow-id", schedWfId,
		"--interval", "1h",
	)
	s.NoError(res.Err)

	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", schedId,
			"-o", "json",
		)
		s.NoError(res.Err)
		var j struct {
			Schedule struct {
				Spec struct {
					Interval []struct {
						Interval string `json:"interval"`
					} `json:"interval"`
				} `json:"spec"`
				Action struct {
					StartWorkflow struct {
						WorkflowType struct {
							Name string `json:"name"`
						} `json:"workflowType"`
						TaskQueue struct {
							Name string `json:"name"`
						} `json:"taskQueue"`
					} `json:"startWorkflow"`
				} `json:"action"`
			} `json:"schedule"`
		}
		s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))
		return j.Schedule.Action.StartWorkflow.WorkflowType.Name == "SomeOtherWf" &&
			j.Schedule.Action.StartWorkflow.TaskQueue.Name == "SomeOtherTq" &&
			j.Schedule.Spec.Interval[0].Interval == "3600s"
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_UpdateAppliesPriorityAndFairness() {
	var updateRequest *workflowservice.UpdateScheduleRequest
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if request, ok := req.(*workflowservice.UpdateScheduleRequest); ok {
			updateRequest = request
		}
		return handler(ctx, req)
	})

	schedID, schedWorkflowID, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)

	res = s.updateSchedule(
		schedID, schedWorkflowID,
		"--interval", "1h",
		"--priority-key", "2",
		"--fairness-key", "tenant-a",
		"--fairness-weight", "2.5",
	)
	s.NoError(res.Err)
	if updateRequest == nil {
		s.Fail("UpdateSchedule request was not captured")
		return
	}

	priority := updateRequest.GetSchedule().GetAction().GetStartWorkflow().GetPriority()
	if priority == nil {
		s.Fail("UpdateSchedule request did not include priority")
		return
	}
	s.Equal(int32(2), priority.GetPriorityKey())
	s.Equal("tenant-a", priority.GetFairnessKey())
	s.Equal(float32(2.5), priority.GetFairnessWeight())
}

func (s *SharedServerSuite) TestSchedule_UpdateResetsOmittedFairness() {
	var updateRequest *workflowservice.UpdateScheduleRequest
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if request, ok := req.(*workflowservice.UpdateScheduleRequest); ok {
			updateRequest = request
		}
		return handler(ctx, req)
	})

	schedID, schedWorkflowID, res := s.createSchedule(
		"--interval", "10d",
		"--priority-key", "2",
		"--fairness-key", "tenant-a",
		"--fairness-weight", "2.5",
	)
	s.NoError(res.Err)

	res = s.updateSchedule(
		schedID, schedWorkflowID,
		"--interval", "1h",
	)
	s.NoError(res.Err)
	if updateRequest == nil {
		s.Fail("UpdateSchedule request was not captured")
		return
	}

	priority := updateRequest.GetSchedule().GetAction().GetStartWorkflow().GetPriority()
	s.Equal(int32(0), priority.GetPriorityKey())
	s.Equal("", priority.GetFairnessKey())
	s.Equal(float32(0), priority.GetFairnessWeight())
}

func (s *SharedServerSuite) TestSchedule_UpdateRejectsHeadersBeforeMutation() {
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

	schedID, schedWorkflowID, res := s.createSchedule("--interval", "10d")
	s.NoError(res.Err)
	scheduleRequests.Store(0)

	res = s.updateSchedule(
		schedID, schedWorkflowID,
		"--interval", "1h",
		"--headers", "example=123",
	)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "headers are not supported for schedule actions")
	s.Equal(int32(0), scheduleRequests.Load())
}

func (s *SharedServerSuite) TestSchedule_UpdateHelpExplainsFullReplacement() {
	res := s.Execute("schedule", "update", "--help")
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "full replacement")
	s.Contains(res.Stdout.String(), "Any options not provided will be reset to their default\nvalues")
	s.Contains(res.Stdout.String(), "temporal schedule describe")
}

func (s *SharedServerSuite) TestSchedule_PatchHelpRegistersNotes() {
	res := s.Execute("schedule", "patch", "--help")
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "--notes")
	s.Contains(res.Stdout.String(), "--unset-notes")
	s.NotContains(res.Stdout.String(), "--headers")

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

func (s *SharedServerSuite) TestSchedule_PatchRejectsInvalidRequestsBeforeMutation() {
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
			name:          "no operation",
			args:          []string{"schedule", "patch", "--schedule-id", "schedule-id"},
			errorContains: "one of --notes or --unset-notes is required",
		},
		{
			name:          "set and unset notes",
			args:          []string{"schedule", "patch", "--schedule-id", "schedule-id", "--notes", "note", "--unset-notes"},
			errorContains: "--notes and --unset-notes are mutually exclusive",
		},
		{
			name:          "missing schedule ID",
			args:          []string{"schedule", "patch", "--notes", "note"},
			errorContains: "required flag(s) \"schedule-id\" not set",
		},
		{
			name:          "empty schedule ID",
			args:          []string{"schedule", "patch", "--address", s.Address(), "--schedule-id=", "--notes", "note"},
			errorContains: "schedule ID is required",
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

func (s *SharedServerSuite) TestSchedule_PatchSetsNotesWithSingleRawUpdate() {
	const (
		namespace  = "patch-notes-namespace"
		scheduleID = "patch-notes-schedule"
		identity   = "patch-notes-identity"
	)
	conflictToken := []byte("patch-notes-conflict-token")
	describedSchedule := &schedule.Schedule{
		Spec:  &schedule.ScheduleSpec{CronString: []string{"0 12 * * *"}},
		State: &schedule.ScheduleState{Notes: "existing notes"},
	}
	describedScheduleSnapshot := proto.Clone(describedSchedule).(*schedule.Schedule)

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
			assert.True(t, proto.Equal(describedScheduleSnapshot, describedSchedule))
		})
	}
}

func (s *SharedServerSuite) TestSchedule_PatchUnsetsNotesWithSingleRawUpdate() {
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
	describedScheduleSnapshot := proto.Clone(describedSchedule).(*schedule.Schedule)

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
	s.True(proto.Equal(describedScheduleSnapshot, describedSchedule))
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

func (s *SharedServerSuite) TestSchedule_UpdateDoesNotPrompt() {
	var updateRequests atomic.Int32
	s.DevServer.SetServerInterceptor(func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := req.(*workflowservice.UpdateScheduleRequest); ok {
			updateRequests.Add(1)
		}
		return handler(ctx, req)
	})

	for _, output := range []string{"text", "json"} {
		s.T().Run(output, func(t *testing.T) {
			schedID, schedWorkflowID, res := s.createSchedule("--interval", "10d")
			if res.Err != nil {
				t.Fatalf("schedule create returned an unexpected error: %v", res.Err)
			}
			updateRequests.Store(0)

			const sentinel = "stdin must remain unread"
			s.Stdin.Reset()
			_, _ = s.Stdin.WriteString(sentinel)

			res = s.updateSchedule(
				schedID, schedWorkflowID,
				"--output", output,
				"--interval", "1h",
			)
			if res.Err != nil {
				t.Fatalf("schedule update returned an unexpected error: %v", res.Err)
			}
			assert.Equal(t, int32(1), updateRequests.Load())
			assert.Equal(t, sentinel, s.Stdin.String())
		})
	}
}

func (s *SharedServerSuite) TestSchedule_Memo_Update() {
	schedId, schedWfId, res := s.createSchedule("--memo", "bar=1")
	s.NoError(res.Err)
	res = s.Execute(
		"schedule", "update",
		"--address", s.Address(),
		"-s", schedId,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", schedWfId,
		"--memo", "bar=2",
	)
	s.NoError(res.Err)
	s.Eventually(func() bool {
		res = s.Execute(
			"schedule", "describe",
			"--address", s.Address(),
			"-s", schedId,
			"-o", "json",
		)
		s.NoError(res.Err)

		var j struct {
			Schedule struct {
				Action struct {
					StartWorkflow struct {
						Memo struct {
							Fields struct {
								Bar struct {
									Data string `json:"data"`
								} `json:"bar"`
							} `json:"fields"`
						} `json:"memo"`
					} `json:"startWorkflow"`
				} `json:"action"`
			} `json:"schedule"`
		}
		s.NoError(json.Unmarshal(res.Stdout.Bytes(), &j))

		// Decoded 'Mg==' is 2
		return j.Schedule.Action.StartWorkflow.Memo.Fields.Bar.Data == "Mg=="
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestSchedule_ListMatchingTimes() {
	// use a calendar spec with known hours so results are deterministic
	schedId, _, res := s.createSchedule("--calendar", `{"hour":"3,6,9"}`)
	s.NoError(res.Err)

	// query a full day - should match exactly 3 times
	res = s.Execute(
		"schedule", "list-matching-times",
		"--address", s.Address(),
		"-s", schedId,
		"--start-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-01-01T23:59:59Z",
	)
	s.NoError(res.Err)
	// text output should contain parseable RFC3339 timestamps
	for _, line := range strings.Split(res.Stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Time" {
			continue
		}
		_, err := time.Parse(time.RFC3339, line)
		s.NoError(err, "should parse text line as time: %q", line)
	}

	// json output
	res = s.Execute(
		"schedule", "list-matching-times",
		"--address", s.Address(),
		"-s", schedId,
		"--start-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-01-01T23:59:59Z",
		"-o", "json",
	)
	s.NoError(res.Err)
	var resp struct {
		StartTime []string `json:"startTime"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &resp))
	s.Equal(3, len(resp.StartTime))
}

func (s *SharedServerSuite) TestSchedule_ListMatchingTimes_NotFound() {
	res := s.Execute(
		"schedule", "list-matching-times",
		"--address", s.Address(),
		"-s", "nonexistent-schedule-id",
		"--start-time", "2025-01-01T00:00:00Z",
		"--end-time", "2025-01-01T23:59:59Z",
	)
	s.Error(res.Err)
}
