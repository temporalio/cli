package temporalcli_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/google/uuid"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *SharedServerSuite) TestWorkflow_Start_SimpleSuccess() {
	// Text
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return map[string]string{"foo": "bar"}, nil
	})
	res := s.Execute(
		"workflow", "start",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
	)
	s.NoError(res.Err)
	// Confirm text has key/vals as expected
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", "my-id1")
	s.Contains(out, "RunId")
	s.ContainsOnSameLine(out, "TaskQueue", s.Worker().Options.TaskQueue)
	s.ContainsOnSameLine(out, "Type", "DevWorkflow")
	s.ContainsOnSameLine(out, "Namespace", "default")

	// JSON
	res = s.Execute(
		"workflow", "start",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		// Use --name here to make sure the alias works
		"--name", "DevWorkflow",
		"--workflow-id", "my-id2",
	)
	s.NoError(res.Err)
	var jsonOut map[string]string
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("my-id2", jsonOut["workflowId"])
	s.NotEmpty(jsonOut["runId"])
	s.Equal(s.Worker().Options.TaskQueue, jsonOut["taskQueue"])
	s.Equal("DevWorkflow", jsonOut["type"])
	s.Equal("default", jsonOut["namespace"])
}

func (s *SharedServerSuite) TestWorkflow_Start_StartDelay() {
	// Capture request
	var lastRequest any
	var lastRequestLock sync.Mutex
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string, req, reply any,
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) error {
			lastRequestLock.Lock()
			lastRequest = req
			lastRequestLock.Unlock()
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)

	res := s.Execute(
		"workflow", "start",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
		"-i", `["val1", "val2"]`,
		"--start-delay", "1ms",
	)
	s.NoError(res.Err)
	s.Equal(
		1*time.Millisecond,
		lastRequest.(*workflowservice.StartWorkflowExecutionRequest).WorkflowStartDelay.AsDuration(),
	)
}

func (s *SharedServerSuite) TestWorkflow_Execute_SimpleSuccess() {
	// Text
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return map[string]string{"foo": "bar"}, nil
	})
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
		"-i", `["val1", "val2"]`,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	// Confirm running (most of this check is done on start test)
	s.ContainsOnSameLine(out, "WorkflowId", "my-id1")
	s.Equal([]any{"val1", "val2"}, s.Worker().DevWorkflowLastInput())
	// Confirm we have some events
	s.ContainsOnSameLine(out, "1", "WorkflowExecutionStarted")
	s.ContainsOnSameLine(out, "2", "WorkflowTaskScheduled")
	s.ContainsOnSameLine(out, "3", "WorkflowTaskStarted")
	// Confirm results
	s.Contains(out, "RunTime")
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.ContainsOnSameLine(out, "Result", `{"foo":"bar"}`)

	// JSON
	res = s.Execute(
		"workflow", "execute",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id2",
	)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("my-id2", jsonOut["workflowId"])
	s.Equal("COMPLETED", jsonOut["status"])
	s.NotNil(jsonOut["closeEvent"])
	s.Equal(map[string]any{"foo": "bar"}, jsonOut["result"])
}

func (s *SharedServerSuite) TestWorkflow_Execute_SimpleFailure() {
	// Text
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return nil, fmt.Errorf("intentional failure")
	})
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
	)
	s.ErrorContains(res.Err, "workflow failed")
	out := res.Stdout.String()
	// Confirm failure
	s.ContainsOnSameLine(out, "Status", "FAILED")
	s.Contains(out, "Failure")
	s.ContainsOnSameLine(out, "Message", "intentional failure")

	// JSON
	res = s.Execute(
		"workflow", "execute",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id2",
	)
	s.ErrorContains(res.Err, "workflow failed")
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("FAILED", jsonOut["status"])
	s.Equal("intentional failure",
		jsonPath(jsonOut, "closeEvent", "workflowExecutionFailedEventAttributes", "failure", "message"))
}

func (s *SharedServerSuite) TestWorkflow_Execute_NestedFailure() {
	// Text
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		err := workflow.ExecuteActivity(ctx, DevActivity).Get(ctx, nil)
		return nil, err
	})
	s.Worker().OnDevActivity(func(ctx context.Context, input any) (any, error) {
		return nil, fmt.Errorf("intentional activity failure")
	})
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
	)
	s.ErrorContains(res.Err, "workflow failed")
	out := res.Stdout.String()
	// Confirm failure
	s.ContainsOnSameLine(out, "Status", "FAILED")
	s.Contains(out, "Failure")
	s.ContainsOnSameLine(out, "Message", "intentional activity failure")

	// JSON
	res = s.Execute(
		"workflow", "execute",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id2",
	)
	s.ErrorContains(res.Err, "workflow failed")
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("FAILED", jsonOut["status"])
	s.Equal("activity error",
		jsonPath(jsonOut, "closeEvent", "workflowExecutionFailedEventAttributes", "failure", "message"))
	s.Equal("intentional activity failure",
		jsonPath(jsonOut, "closeEvent", "workflowExecutionFailedEventAttributes", "failure", "cause", "message"))
}

func (s *SharedServerSuite) TestWorkflow_Execute_Cancel() {
	// Very bad™️ channel tricks
	doCancelChan := make(chan struct{})
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		doCancelChan <- struct{}{}
		err := workflow.Await(ctx, func() bool {
			return false
		})
		return nil, err
	})

	// Text
	go func() {
		<-doCancelChan
		_ = s.Client.CancelWorkflow(s.Context, "my-id1", "")
	}()
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
	)
	s.ErrorContains(res.Err, "workflow failed")
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "CANCELED")

	// JSON
	go func() {
		<-doCancelChan
		_ = s.Client.CancelWorkflow(s.Context, "my-id2", "")
	}()
	res = s.Execute(
		"workflow", "execute",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id2",
	)
	s.ErrorContains(res.Err, "workflow failed")
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("CANCELED", jsonOut["status"])
}

func (s *SharedServerSuite) TestWorkflow_Execute_Timeout() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		err := workflow.Await(ctx, func() bool {
			return false
		})
		return nil, err
	})

	// Text
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--execution-timeout", "1ms",
		"--workflow-id", "my-id1",
	)
	s.ErrorContains(res.Err, "workflow failed")
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "TIMEOUT")

	// JSON
	res = s.Execute(
		"workflow", "execute",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--execution-timeout", "1ms",
		"--workflow-id", "my-id2",
	)
	s.ErrorContains(res.Err, "workflow failed")
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("TIMEOUT", jsonOut["status"])
}

func (s *SharedServerSuite) TestWorkflow_Execute_ContinueAsNew() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		if input.(float64) < 2 {
			return nil, workflow.NewContinueAsNewError(ctx, "DevWorkflow", input.(float64)+1)
		}
		return nil, nil
	})

	// Text
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"-i", "1",
		"--workflow-id", "my-id1",
	)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.Contains(out, "WorkflowExecutionContinuedAsNew")
}

func (s *SharedServerSuite) TestWorkflow_Execute_ProtoJSON_Input() {
	// Very meta, use a start workflow request proto as the input to the workflow.
	startWorkflowReq := &workflowservice.StartWorkflowExecutionRequest{
		// Just fill in a few different types of fields to make sure everything is [de]serialized
		WorkflowId: "enchi-cat",
		WorkflowRunTimeout: &durationpb.Duration{
			Seconds: 1,
			Nanos:   2,
		},
		Input: &common.Payloads{
			Payloads: []*common.Payload{
				{Data: []byte("meow")},
			},
		},
	}
	startWorkflowReqSerialized, err := protojson.Marshal(startWorkflowReq)
	s.NoError(err)

	s.Worker().Worker.RegisterWorkflowWithOptions(func(
		ctx workflow.Context,
		input *workflowservice.StartWorkflowExecutionRequest,
	) (*workflowservice.StartWorkflowExecutionRequest, error) {
		return input, nil
	}, workflow.RegisterOptions{Name: "ProtoJSONWorkflow"})

	// Text
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "ProtoJSONWorkflow",
		"--input-meta", "encoding=json/protobuf",
		"-i", string(startWorkflowReqSerialized),
		"--workflow-id", "my-id1",
	)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.Contains(out, "enchi")
}

func (s *SharedServerSuite) TestWorkflow_Failure_On_Start() {
	// Use too-long of an ID to force a failure on start
	veryLongID := string(bytes.Repeat([]byte("a"), 1024))
	for _, cmd := range []string{"start", "execute"} {
		res := s.Execute(
			"workflow", cmd,
			"--address", s.Address(),
			"--task-queue", s.Worker().Options.TaskQueue,
			"--type", "DevWorkflow",
			"--workflow-id", veryLongID,
		)
		s.ErrorContains(res.Err, "failed starting workflow")
	}
}

func (s *SharedServerSuite) TestWorkflow_Execute_ClientHeaders() {
	// Capture headers
	var lastHeadersClient metadata.MD
	var lastHeadersLock sync.Mutex
	// Capture from client
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(func(
			ctx context.Context,
			method string, req, reply any,
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) error {
			lastHeadersLock.Lock()
			lastHeadersClient, _ = metadata.FromOutgoingContext(ctx)
			lastHeadersLock.Unlock()
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)

	// Capture from server
	// TODO(cretz): Pending fix on server for gRPC interceptors
	// var lastHeadersServer metadata.MD
	// s.SetServerInterceptor(
	// 	func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// 		lastHeadersLock.Lock()
	// 		lastHeadersServer, _ = metadata.FromIncomingContext(ctx)
	// 		lastHeadersLock.Unlock()
	// 		return handler(ctx, req)
	// 	},
	// )

	// Exec workflow
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
		"-i", `["val1", "val2"]`,
	)
	s.NoError(res.Err)

	// Check that the client name is there
	s.Equal("temporal-cli", lastHeadersClient["client-name"][0])
}

func (s *SharedServerSuite) TestWorkflow_Execute_EnvVars() {
	s.CommandHarness.Options.LookupEnv = func(key string) (string, bool) {
		if key == "TEMPORAL_ADDRESS" {
			return s.Address(), true
		}
		return "", false
	}
	res := s.Execute(
		"workflow", "execute",
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
	)
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestWorkflow_Execute_EnvConfig() {
	// Temp file for env
	tmpFile, err := os.CreateTemp("", "")
	s.NoError(err)
	// s.CommandHarness.Options.EnvConfigFile = tmpFile.Name()
	defer os.Remove(tmpFile.Name())

	// Set config value for input (obviously `--input` is normally a poor choice
	// for an env file)
	res := s.Execute(
		"env", "set",
		"--env-file", tmpFile.Name(),
		"myenv.input", `"env-conf-input"`,
	)
	s.NoError(res.Err)

	// Command with its own input which overrides env
	res = s.Execute(
		"workflow", "execute",
		"--env", "myenv",
		"--env-file", tmpFile.Name(),
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
		"--input", `"cli-input"`,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Result", `"cli-input"`)

	// But if command does not have input, can use env's
	res = s.Execute(
		"workflow", "execute",
		"--env", "myenv",
		"--env-file", tmpFile.Name(),
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id2",
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Result", `"env-conf-input"`)

	// And we can specify `env` with a property
	s.CommandHarness.Options.LookupEnv = func(key string) (string, bool) {
		if key == "TEMPORAL_ENV" {
			return "myenv", true
		}
		return "", false
	}
	res = s.Execute(
		"workflow", "execute",
		"--env-file", tmpFile.Name(),
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id3",
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Result", `"env-conf-input"`)

}

func (s *SharedServerSuite) TestWorkflow_Execute_CodecEndpoint() {
	// Start HTTP server for our codec
	srv := httptest.NewServer(converter.NewPayloadCodecHTTPHandler(prefixingCodec{}))
	defer srv.Close()

	// Run a different worker than the suite on a different task queue that has
	// our codec
	prefixedDataConverter := converter.NewCodecDataConverter(converter.GetDefaultDataConverter(), prefixingCodec{})
	workerClient, err := client.NewClientFromExisting(s.Client, client.Options{DataConverter: prefixedDataConverter})
	s.NoError(err)
	defer workerClient.Close()
	taskQueue := uuid.NewString()
	worker := worker.New(workerClient, taskQueue, worker.Options{})
	worker.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, arg any) (any, error) { return arg, nil },
		workflow.RegisterOptions{Name: "test-workflow"},
	)
	s.NoError(worker.Start())
	defer worker.Stop()

	// Helper to confirm encoded
	assertWorkflowEncoded := func(workflowID string) {
		var foundStart, foundComplete bool
		iter := s.Client.GetWorkflowHistory(s.Context, workflowID, "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, err := iter.Next()
			s.NoError(err)
			if start := event.GetWorkflowExecutionStartedEventAttributes(); start != nil {
				foundStart = true
				s.Equal("binary/prefixed", string(start.Input.Payloads[0].Metadata["encoding"]))
			} else if complete := event.GetWorkflowExecutionCompletedEventAttributes(); complete != nil {
				foundComplete = true
				s.Equal("binary/prefixed", string(complete.Result.Payloads[0].Metadata["encoding"]))
			}
		}
		s.True(foundStart)
		s.True(foundComplete)
	}

	// Run a workflow with our codec endpoint
	res := s.Execute(
		"workflow", "execute",
		"--codec-endpoint", "http://"+srv.Listener.Addr().String(),
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--type", "test-workflow",
		"--workflow-id", "my-id1",
		"--input", `{"foo":"bar"}`,
	)
	s.NoError(res.Err)
	// Confirm result is proper, but when fetching history both input and result
	// are mangled
	s.ContainsOnSameLine(res.Stdout.String(), "Result", `{"foo":"bar"}`)
	assertWorkflowEncoded("my-id1")

	// Let's do the same with JSON and full details so we can check history is
	// actually decoded for the user
	res = s.Execute(
		"workflow", "execute",
		"-o", "json", "--detailed",
		"--codec-endpoint", "http://"+srv.Listener.Addr().String(),
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--type", "test-workflow",
		"--workflow-id", "my-id2",
		"--input", `{"foo":"bar"}`,
	)
	s.NoError(res.Err)
	assertWorkflowEncoded("my-id2")
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal("bar", jsonPath(jsonOut, "result", "foo"))
	input, err := base64.StdEncoding.DecodeString(jsonPath(jsonOut,
		"history", "events", "0", "workflowExecutionStartedEventAttributes", "input", "payloads", "0", "data").(string))
	s.NoError(err)
	s.Equal(`{"foo":"bar"}`, string(input))

	// Run without codec endpoint and confirm remains encoded for user
	res = s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--type", "test-workflow",
		"--workflow-id", "my-id3",
		"--input", `{"foo":"bar"}`,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Result",
		fmt.Sprintf("%q:%q", "encoding", base64.StdEncoding.EncodeToString([]byte("binary/prefixed"))))
}

type prefixingCodec struct{}

func (prefixingCodec) Encode(payloads []*common.Payload) ([]*common.Payload, error) {
	ret := make([]*common.Payload, len(payloads))
	for i, payload := range payloads {
		ret[i] = proto.Clone(payload).(*common.Payload)
		ret[i].Data = append([]byte("prefix-"), ret[i].Data...)
		ret[i].Metadata["old-encoding"] = ret[i].Metadata["encoding"]
		ret[i].Metadata["encoding"] = []byte("binary/prefixed")
	}
	return ret, nil
}

func (prefixingCodec) Decode(payloads []*common.Payload) ([]*common.Payload, error) {
	ret := make([]*common.Payload, len(payloads))
	for i, payload := range payloads {
		ret[i] = proto.Clone(payload).(*common.Payload)
		if bytes.HasPrefix(ret[i].Data, []byte("prefix-")) {
			ret[i].Data = bytes.TrimPrefix(ret[i].Data, []byte("prefix-"))
			ret[i].Metadata["encoding"] = ret[i].Metadata["old-encoding"]
			delete(ret[i].Metadata, "old-encoding")
		}
	}
	return ret, nil
}

func jsonPath(v any, path ...string) any {
	switch t := v.(type) {
	case map[string]any:
		v = t[path[0]]
	case []any:
		i, err := strconv.Atoi(path[0])
		if err != nil {
			panic(err)
		}
		v = t[i]
	default:
		panic(fmt.Sprintf("unknown type: %T", v))
	}
	if len(path) == 1 {
		return v
	}
	return jsonPath(v, path[1:]...)
}

func (s *SharedServerSuite) TestWorkflow_Execute_NullValue() {
	// Regression test: see https://github.com/temporalio/cli/pull/617
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return map[string]any{"foo": nil}, nil
	})
	res := s.Execute(
		"workflow", "execute",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.ContainsOnSameLine(out, "Result", `{"foo":null}`)
}

func (s *SharedServerSuite) TestWorkflow_SignalWithStart_WorkflowIDMandatory() {
	res := s.Execute(
		"workflow", "signal-with-start",
		"--type", "wfType",
		"--task-queue", "tq",
		"--signal-name", "sigName",
	)
	s.ErrorContains(res.Err, "--workflow-id flag must be provided")
}

func (s *SharedServerSuite) TestWorkflow_SignalWithStart_StartNewWorkflow() {
	s.testSignalWithStartHelper(false)
}

func (s *SharedServerSuite) TestWorkflow_SignalWithStart_SendSignalToExistingWorkflow() {
	s.testSignalWithStartHelper(true)
}

func (s *SharedServerSuite) testSignalWithStartHelper(useExistingWorkflow bool) {
	wfId := uuid.NewString()
	signalWfInput := `"workflow-input"`
	signalInput := `"signal-input"`
	expectedWfOutput := map[string]string{
		"workflow": "workflow-input",
		"signal":   "signal-input",
	}

	if useExistingWorkflow {
		run, err := s.Client.ExecuteWorkflow(s.Context, client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue}, DevWorkflow, "not-signal-with-start-input")
		s.NoError(err)
		// Re-assign wfId for the signal to be sent to an existing workflow.
		wfId = run.GetID()
		expectedWfOutput["workflow"] = "not-signal-with-start-input"
	}

	// Run workflow, block on signal.
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, wfInput any) (any, error) {
		wfState := make(map[string]string)
		wfState["workflow"] = wfInput.(string)
		var sigReceived string
		workflow.GetSignalChannel(ctx, "sigName").Receive(ctx, &sigReceived)
		wfState["signal"] = sigReceived
		return wfState, nil
	})

	// Send signal-with-start command.
	res := s.Execute(
		"workflow", "signal-with-start",
		"--address", s.Address(),
		"--workflow-id", wfId,
		"--type", "DevWorkflow",
		"--input", signalWfInput,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--signal-name", "sigName",
		"--signal-input", signalInput,
	)
	s.NoError(res.Err)

	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", wfId)
	s.Contains(out, "RunId")
	s.ContainsOnSameLine(out, "Type", "DevWorkflow")
	s.ContainsOnSameLine(out, "Namespace", "default")

	// Check that a new workflow was started with expected workflow ID.
	if !useExistingWorkflow {
		run := s.Client.GetWorkflow(s.Context, wfId, "")
		s.Equal(wfId, run.GetID())
	}

	// Wait for workflow to complete.
	wfReturn := make(map[string]string)
	err := s.Client.GetWorkflow(s.Context, wfId, "").Get(s.Context, &wfReturn)
	s.NoError(err)

	// Compare the extracted values with what the workflow returned
	s.Equal(expectedWfOutput["signal"], wfReturn["signal"])
	s.Equal(expectedWfOutput["workflow"], wfReturn["workflow"])
}

func (s *SharedServerSuite) TestWorkflow_StartUpdateWithStart_RuntimeOptionChecks() {
	res := s.Execute(
		"workflow", "start-update-with-start",
		"--type", "wfType",
		"--task-queue", "tq",
		"--update-name", "updateName",
		"--update-wait-for-stage", "accepted",
	)
	s.ErrorContains(res.Err, "--workflow-id flag must be provided")
	res = s.Execute(
		"workflow", "start-update-with-start",
		"--type", "wfType",
		"--task-queue", "tq",
		"--update-name", "updateName",
		"--update-wait-for-stage", "accepted",
		"--workflow-id", "wfId",
	)
	s.ErrorContains(res.Err, "--id-conflict-policy flag must be provided")
}

func (s *SharedServerSuite) TestWorkflow_ExecuteUpdateWithStart_RuntimeOptionChecks() {
	res := s.Execute(
		"workflow", "execute-update-with-start",
		"--type", "wfType",
		"--task-queue", "tq",
		"--update-name", "updateName",
	)
	s.ErrorContains(res.Err, "--workflow-id flag must be provided")
	res = s.Execute(
		"workflow", "execute-update-with-start",
		"--type", "wfType",
		"--task-queue", "tq",
		"--update-name", "updateName",
		"--workflow-id", "wfId",
	)
	s.ErrorContains(res.Err, "--id-conflict-policy flag must be provided")
}

type updateWithStartTest struct {
	updateWithStartSetup
	useStart             bool
	idConflictPolicy     string
	expectedError        string
	expectedUpdateResult string
	expectedWfOutput     map[string]string
}

type updateWithStartSetup struct {
	wfId                string
	updateName          string
	updateId            string
	useExistingWorkflow bool
}

func (s *SharedServerSuite) TestWorkflow_StartUpdateWithStart_StartsNewWorkflow() {
	updateWithStartSetup := s.updateWithStartTestSetup(false)
	s.testStartUpdateWithStartHelper(updateWithStartTest{
		updateWithStartSetup: updateWithStartSetup,
		useStart:             true,
		idConflictPolicy:     "Fail",
		expectedWfOutput:     map[string]string{"workflow": "workflow-input", "update": "update-input"},
	})
}

func (s *SharedServerSuite) TestWorkflow_StartUpdateWithStart_SendUpdateToExistingWorkflow() {
	updateWithStartSetup := s.updateWithStartTestSetup(true)
	s.testStartUpdateWithStartHelper(updateWithStartTest{
		updateWithStartSetup: updateWithStartSetup,
		useStart:             true,
		idConflictPolicy:     "Fail",
		expectedError:        "Workflow execution is already running",
	})
	s.testStartUpdateWithStartHelper(updateWithStartTest{
		updateWithStartSetup: updateWithStartSetup,
		useStart:             true,
		idConflictPolicy:     "UseExisting",
		expectedWfOutput:     map[string]string{"workflow": "not-update-with-start-workflow-input", "update": "update-input"},
	})
}

func (s *SharedServerSuite) TestWorkflow_ExecuteUpdateWithStart_StartsWorkflow() {
	updateWithStartSetup := s.updateWithStartTestSetup(false)
	s.testStartUpdateWithStartHelper(updateWithStartTest{
		updateWithStartSetup: updateWithStartSetup,
		useStart:             false,
		idConflictPolicy:     "Fail",
		expectedUpdateResult: "update-input",
		expectedWfOutput:     map[string]string{"workflow": "workflow-input", "update": "update-input"},
	})
}

func (s *SharedServerSuite) TestWorkflow_ExecuteUpdateWithStart_SendUpdateToExistingWorkflow() {
	updateWithStartSetup := s.updateWithStartTestSetup(true)
	s.testStartUpdateWithStartHelper(updateWithStartTest{
		updateWithStartSetup: updateWithStartSetup,
		useStart:             false,
		idConflictPolicy:     "Fail",
		expectedError:        "Workflow execution is already running",
		expectedWfOutput:     map[string]string{"workflow": "workflow-input", "update": "update-input"},
	})
	s.testStartUpdateWithStartHelper(updateWithStartTest{
		updateWithStartSetup: updateWithStartSetup,
		useStart:             false,
		idConflictPolicy:     "UseExisting",
		expectedWfOutput:     map[string]string{"workflow": "not-update-with-start-workflow-input", "update": "update-input"},
	})
}

func (s *SharedServerSuite) updateWithStartTestSetup(useExistingWorkflow bool) updateWithStartSetup {
	wfId := uuid.NewString()
	updateName := "test-update-name"
	updateId := uuid.NewString()
	if useExistingWorkflow {
		// Start a workflow with a specific workflow ID.
		run, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{
				TaskQueue: s.Worker().Options.TaskQueue,
			},
			DevWorkflow,
			"not-update-with-start-workflow-input",
		)
		s.NoError(err)
		// Re-assign wfId for the update to be sent to an existing workflow.
		wfId = run.GetID()
	}

	// Run workflow.
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, wfInput any) (any, error) {
		wfState := make(map[string]string)
		wfState["workflow"] = wfInput.(string)

		err := workflow.SetUpdateHandlerWithOptions(
			ctx,
			updateName,
			func(ctx workflow.Context, updateInput string) (string, error) {
				wfState["update"] = updateInput
				return updateInput, nil
			},
			workflow.UpdateHandlerOptions{},
		)
		if err != nil {
			return nil, err
		}
		// Block workflow completion on signal.
		workflow.GetSignalChannel(ctx, "complete").Receive(ctx, nil)
		return wfState, nil
	})
	return updateWithStartSetup{wfId, updateName, updateId, useExistingWorkflow}
}

func (s *SharedServerSuite) testStartUpdateWithStartHelper(opts updateWithStartTest) {
	cmdName := "execute-update-with-start"
	additionalArgs := []string{}

	if opts.useStart {
		cmdName = "start-update-with-start"
		additionalArgs = []string{"--update-wait-for-stage", "accepted"}
	}

	baseArgs := []string{
		"workflow", cmdName,
		"--address", s.Address(),
		"--workflow-id", opts.wfId,
		"--type", "DevWorkflow",
		"--input", `"workflow-input"`,
		"--task-queue", s.Worker().Options.TaskQueue,
		"--id-conflict-policy", opts.idConflictPolicy,
		"--update-name", opts.updateName,
		"--update-id", opts.updateId,
		"--update-input", `"update-input"`,
	}

	// Send start-update-with-start command.
	args := append(baseArgs, additionalArgs...)
	res := s.Execute(args...)

	// Check expected error.
	if opts.expectedError != "" {
		s.ErrorContains(res.Err, opts.expectedError)
		return
	}

	s.NoError(res.Err)

	// Confirm text output has key/vals as expected
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", opts.wfId)
	s.Contains(out, "RunId")
	s.ContainsOnSameLine(out, "Type", "DevWorkflow")
	s.ContainsOnSameLine(out, "Namespace", "default")
	s.ContainsOnSameLine(out, "UpdateName", opts.updateName)
	s.ContainsOnSameLine(out, "UpdateID", opts.updateId)

	// Check expected update result.
	if opts.expectedUpdateResult != "" {
		s.ContainsOnSameLine(out, "UpdateResult", opts.expectedUpdateResult)
	}

	// Check that new workflow was started with expected workflow ID.
	if !opts.useExistingWorkflow {
		run := s.Client.GetWorkflow(s.Context, opts.wfId, "")
		s.Equal(opts.wfId, run.GetID())
	}

	// Send signal to complete workflow.
	err := s.Client.SignalWorkflow(s.Context, opts.wfId, "", "complete", nil)
	s.NoError(err)

	// Wait for workflow to complete.
	wfReturn := make(map[string]string)
	err = s.Client.GetWorkflow(s.Context, opts.wfId, "").Get(s.Context, &wfReturn)
	s.NoError(err)

	// Expect workflow to have received update and given inputs from start-update-with-start.
	s.Equal(opts.expectedWfOutput["workflow"], wfReturn["workflow"])
	s.Equal(opts.expectedWfOutput["update"], wfReturn["update"])
}
