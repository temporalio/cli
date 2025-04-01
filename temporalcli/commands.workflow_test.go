package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
)

func (s *SharedServerSuite) TestWorkflow_Signal_SingleWorkflowSuccess() {
	// Make workflow wait for signal and then return it
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		var ret any
		workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, &ret)
		return ret, nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	// Send signal
	res := s.Execute(
		"workflow", "signal",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--name", "my-signal",
		"-i", `{"foo": "bar"}`,
	)
	s.NoError(res.Err)

	// Confirm workflow result was as expected
	var actual any
	s.NoError(run.Get(s.Context, &actual))
	s.Equal(map[string]any{"foo": "bar"}, actual)
}

func (s *SharedServerSuite) TestWorkflow_Signal_BatchWorkflowSuccess() {
	res := s.testSignalBatchWorkflow(false)
	s.Contains(res.Stdout.String(), "approximately 5 workflow(s)")
	s.Contains(res.Stdout.String(), "Started batch")
}

func (s *SharedServerSuite) TestWorkflow_Signal_BatchWorkflowSuccessJSON() {
	res := s.testSignalBatchWorkflow(true)
	var jsonRes map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonRes))
	s.NotEmpty(jsonRes["batchJobId"])
}

func (s *SharedServerSuite) testSignalBatchWorkflow(json bool) *CommandResult {
	// Make workflow wait for signal and then return it
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		var ret any
		workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, &ret)
		return ret, nil
	})

	// Start 5 workflows
	runs := make([]client.WorkflowRun, 5)
	searchAttr := "keyword-" + uuid.NewString()
	for i := range runs {
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
		runs[i] = run
	}

	// Wait for all to appear in list
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == len(runs)
	}, 3*time.Second, 100*time.Millisecond)

	// Send batch signal with a "y" for non-json or "--yes" for json
	args := []string{
		"workflow", "signal",
		"--address", s.Address(),
		"--query", "CustomKeywordField = '" + searchAttr + "'",
		// Use --type here to make sure the alias works
		"--type", "my-signal",
		"-i", `{"key": "val"}`,
	}
	if json {
		args = append(args, "--yes", "-o", "json")
	} else {
		s.CommandHarness.Stdin.WriteString("y\n")
	}
	res := s.Execute(args...)
	s.NoError(res.Err)

	// Confirm that all workflows complete with the signal value
	for _, run := range runs {
		var ret map[string]string
		s.NoError(run.Get(s.Context, &ret))
		s.Equal(map[string]string{"key": "val"}, ret)
	}
	return res
}

func (s *SharedServerSuite) TestWorkflow_Delete_BatchWorkflowSuccess() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start some workflows
	prefix := "delete-test-"
	ids := []string{prefix + "1", prefix + "2", prefix + "3"}
	for _, id := range ids {
		_, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{ID: id, TaskQueue: "delete-test"},
			DevWorkflow,
			"ignored",
		)
		s.NoError(err)
	}

	// Confirm workflows exist in visibility
	s.Eventually(func() bool {
		wfs, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: s.Namespace(),
			Query:     "TaskQueue = 'delete-test'",
		})
		s.NoError(err)

		if len(wfs.GetExecutions()) == 3 {
			return true
		}

		return false
	}, 5*time.Second, 100*time.Millisecond, "timed out awaiting for workflows to exist in visibility")

	// Send delete
	res := s.Execute(
		"workflow", "delete",
		"--address", s.Address(),
		"--query", "TaskQueue = 'delete-test' AND (WorkflowId = 'delete-test-1' OR WorkflowId = 'delete-test-2')",
		"--reason", "test",
		"-y",
	)
	s.NoError(res.Err)

	// Confirm workflows were deleted
	s.Eventually(func() bool {
		wfs, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace: s.Namespace(),
			Query:     "TaskQueue = 'delete-test'",
		})
		s.NoError(err)

		if len(wfs.GetExecutions()) == 1 && wfs.GetExecutions()[0].GetExecution().GetWorkflowId() == "delete-test-3" {
			return true
		}

		return false
	}, 8*time.Second, 100*time.Millisecond, "timed out awaiting for workflows termination")
}

func (s *SharedServerSuite) TestWorkflow_Terminate_SingleWorkflowSuccess_WithoutReason() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	// Send terminate
	res := s.Execute(
		"workflow", "terminate",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)

	// Confirm workflow was terminated
	s.Contains(run.Get(s.Context, nil).Error(), "terminated")
	// Ensure the termination reason was recorded
	iter := s.Client.GetWorkflowHistory(s.Context, run.GetID(), run.GetRunID(), false, enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
	var foundReason bool
	for iter.HasNext() {
		event, err := iter.Next()
		s.NoError(err)
		if term := event.GetWorkflowExecutionTerminatedEventAttributes(); term != nil {
			foundReason = true
			// We're not going to check the value here so we don't pin ourselves to our particular default, but there _should_ be a default reason
			s.NotEmpty(term.Reason)
		}
	}
	s.True(foundReason)
}

func (s *SharedServerSuite) TestWorkflow_Terminate_SingleWorkflowSuccess_WithReason() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	// Send terminate
	res := s.Execute(
		"workflow", "terminate",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reason", "terminate-test",
	)
	s.NoError(res.Err)

	// Confirm workflow was terminated
	s.Contains(run.Get(s.Context, nil).Error(), "terminated")

	// Ensure the termination reason was recorded
	iter := s.Client.GetWorkflowHistory(s.Context, run.GetID(), run.GetRunID(), false, enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
	var foundReason bool
	for iter.HasNext() {
		event, err := iter.Next()
		s.NoError(err)
		if term := event.GetWorkflowExecutionTerminatedEventAttributes(); term != nil {
			foundReason = true
			s.Equal("terminate-test", term.Reason)
		}
	}
	s.True(foundReason)
}

func (s *SharedServerSuite) TestWorkflow_Terminate_BatchWorkflowSuccess() {
	_, res := s.testTerminateBatchWorkflow(5, 0, false)
	s.Contains(res.Stdout.String(), "approximately 5 workflow(s)")
	s.Contains(res.Stdout.String(), "Started batch")
}

func (s *SharedServerSuite) TestWorkflow_Terminate_BatchWorkflow_Ratelimit_MissingQuery() {
	res := s.Execute(
		"workflow", "terminate",
		"--address", s.Address(),
		"--rps", "5",
	)
	s.EqualError(res.Err, "must set either workflow ID or query")

	res = s.Execute(
		"workflow", "terminate",
		"-w", "some-workflow-id",
		"--address", s.Address(),
		"--rps", "5",
	)
	s.EqualError(res.Err, "cannot set rps when workflow ID is set")
}

func (s *SharedServerSuite) TestWorkflow_Terminate_BatchWorkflowSuccess_Ratelimit() {
	var rps float32 = 1
	req, _ := s.testTerminateBatchWorkflow(2, rps, false)
	s.Equal(rps, req.MaxOperationsPerSecond)
}

func (s *SharedServerSuite) TestWorkflow_Terminate_BatchWorkflowSuccess_JSON() {
	_, res := s.testTerminateBatchWorkflow(5, 0, true)
	var jsonRes map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonRes))
	s.NotEmpty(jsonRes["batchJobId"])
}

func (s *SharedServerSuite) testTerminateBatchWorkflow(
	total int,
	rps float32,
	json bool,
) (*workflowservice.StartBatchOperationRequest, *CommandResult) {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

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

	// Start workflows
	runs := make([]client.WorkflowRun, total)
	searchAttr := "keyword-" + uuid.NewString()
	for i := range runs {
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
		runs[i] = run
	}

	// Wait for all to appear in list
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == len(runs)
	}, 3*time.Second, 100*time.Millisecond)

	// Send batch terminate with a "y" for non-json or "--yes" for json
	args := []string{
		"workflow", "terminate",
		"--rps", fmt.Sprint(rps),
		"--address", s.Address(),
		"--query", "CustomKeywordField = '" + searchAttr + "'",
		"--reason", "terminate-test",
	}
	if json {
		args = append(args, "--yes", "-o", "json")
	} else {
		s.CommandHarness.Stdin.WriteString("y\n")
	}
	res := s.Execute(args...)
	s.NoError(res.Err)

	// Wait for batch job to complete.
	listRes := s.Execute("batch", "list", "--address", s.Address())
	s.NoError(listRes.Err)
	s.Contains(listRes.Stdout.String(), "Completed")

	// Confirm that all workflows are terminated
	for _, run := range runs {
		s.Contains(run.Get(s.Context, nil).Error(), "terminated")
		// Ensure the termination reason was recorded
		iter := s.Client.GetWorkflowHistory(s.Context, run.GetID(), run.GetRunID(), false, enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
		var foundReason bool
		for iter.HasNext() {
			event, err := iter.Next()
			s.NoError(err)
			if term := event.GetWorkflowExecutionTerminatedEventAttributes(); term != nil {
				foundReason = true
				s.Equal("terminate-test", term.Reason)
			}
		}
		s.True(foundReason)
	}
	return startBatchRequest, res
}

func (s *SharedServerSuite) TestWorkflow_Cancel_SingleWorkflowSuccess() {
	// Make workflow wait for cancel and then return the context's error
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	// Send cancel
	res := s.Execute(
		"workflow", "cancel",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)

	// Confirm workflow was cancelled
	s.Error(workflow.ErrCanceled, run.Get(s.Context, nil))
}

func (s *SharedServerSuite) TestWorkflow_Batch_Update_Options_Versioning_Override() {
	buildId1 := uuid.NewString()
	buildId2 := uuid.NewString()
	deploymentName := uuid.NewString()
	version1 := deploymentName + "." + buildId1
	version2 := deploymentName + "." + buildId2
	// Workflow that waits to be canceled.
	waitingWorkflow := func(ctx workflow.Context) error {
		ctx.Done().Receive(ctx, nil)
		return ctx.Err()
	}
	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version1,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
		Workflows: []any{waitingWorkflow},
	})
	defer w.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--version", version1,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--version", version1,
		"--yes",
	)
	s.NoError(res.Err)

	// Start workflows
	numWorkflows := 5
	runs := make([]client.WorkflowRun, numWorkflows)
	searchAttr := "keyword-" + uuid.NewString()
	for i := range runs {
		run, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{
				TaskQueue:        w.Options.TaskQueue,
				SearchAttributes: map[string]any{"CustomKeywordField": searchAttr},
			},
			waitingWorkflow,
		)
		s.NoError(err)
		runs[i] = run
	}

	s.EventuallyWithT(func(t *assert.CollectT) {
		for _, run := range runs {
			res = s.Execute(
				"workflow", "describe",
				"--address", s.Address(),
				"-w", run.GetID(),
			)
			assert.NoError(t, res.Err)
			assert.Contains(t, res.Stdout.String(), version1)
			assert.Contains(t, res.Stdout.String(), "Pinned")
		}
	}, 30*time.Second, 100*time.Millisecond)

	// Wait for all to appear in list
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == len(runs)
	}, 3*time.Second, 100*time.Millisecond)

	s.CommandHarness.Stdin.WriteString("y\n")
	res = s.Execute(
		"workflow", "update-options",
		"--address", s.Address(),
		"--query", "CustomKeywordField = '"+searchAttr+"'",
		"--versioning-override-behavior", "pinned",
		"--versioning-override-pinned-version", version2,
	)
	s.NoError(res.Err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		for _, run := range runs {
			// json
			res = s.Execute(
				"workflow", "describe",
				"--address", s.Address(),
				"-w", run.GetID(),
				"--output", "json",
			)
			require.NoError(t, res.Err)

			var jsonResp workflowservice.DescribeWorkflowExecutionResponse
			require.NoError(t, temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonResp, true))

			versioningInfo := jsonResp.GetWorkflowExecutionInfo().GetVersioningInfo()
			require.NotNil(t, versioningInfo)
			require.NotNil(t, versioningInfo.VersioningOverride)
			require.Equal(t, version2, versioningInfo.VersioningOverride.PinnedVersion)
		}
	}, 30*time.Second, 100*time.Millisecond)
}

func (s *SharedServerSuite) TestWorkflow_Update_Options_Versioning_Override() {
	buildId1 := uuid.NewString()
	buildId2 := uuid.NewString()
	deploymentName := uuid.NewString()
	version1 := deploymentName + "." + buildId1
	version2 := deploymentName + "." + buildId2

	// Workflow that waits to be canceled.
	waitingWorkflow := func(ctx workflow.Context) error {
		ctx.Done().Receive(ctx, nil)
		return ctx.Err()
	}
	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			DeploymentOptions: worker.DeploymentOptions{
				UseVersioning:             true,
				Version:                   version1,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
		Workflows: []any{waitingWorkflow},
	})
	defer w.Stop()

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "list",
			"--address", s.Address(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), deploymentName)
	}, 30*time.Second, 100*time.Millisecond)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "deployment", "describe-version",
			"--address", s.Address(),
			"--version", version1,
		)
		assert.NoError(t, res.Err)
	}, 30*time.Second, 100*time.Millisecond)

	res := s.Execute(
		"worker", "deployment", "set-current-version",
		"--address", s.Address(),
		"--version", version1,
		"--yes",
	)
	s.NoError(res.Err)

	// Start the workflow and wait until the operation is started.
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: w.Options.TaskQueue},
		waitingWorkflow,
	)
	s.NoError(err)

	s.EventuallyWithT(func(t *assert.CollectT) {
		res = s.Execute(
			"workflow", "describe",
			"--address", s.Address(),
			"-w", run.GetID(),
		)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), version1)
		assert.Contains(t, res.Stdout.String(), "Pinned")
	}, 30*time.Second, 100*time.Millisecond)

	res = s.Execute(
		"workflow", "update-options",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--versioning-override-behavior", "pinned",
		"--versioning-override-pinned-version", version2,
	)
	s.NoError(res.Err)

	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "OverrideBehavior", "Pinned")
	s.ContainsOnSameLine(res.Stdout.String(), "OverridePinnedVersion", version2)

	// remove override
	res = s.Execute(
		"workflow", "update-options",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--versioning-override-behavior", "unspecified",
	)
	s.NoError(res.Err)

	// json
	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--output", "json",
	)
	s.NoError(res.Err)

	var jsonResp workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonResp, true))
	versioningInfo := jsonResp.WorkflowExecutionInfo.VersioningInfo
	s.Nil(versioningInfo.VersioningOverride)
}

func (s *SharedServerSuite) TestWorkflow_Update_Execute() {
	workflowUpdateTest{
		s:        s,
		useStart: false,
	}.testWorkflowUpdateHelper()
}

func (s *SharedServerSuite) TestWorkflow_Update_Start() {
	workflowUpdateTest{
		s:        s,
		useStart: true,
	}.testWorkflowUpdateHelper()
}

type workflowUpdateTest struct {
	s        *SharedServerSuite
	useStart bool
}

func (t workflowUpdateTest) testWorkflowUpdateHelper() {
	updateName := "test-update"

	t.s.Worker().OnDevWorkflow(func(ctx workflow.Context, val any) (any, error) {
		// setup a simple workflow which receives non-negative floats in updates and adds them to a running counter
		counter, ok := val.(float64)
		if !ok {
			return nil, fmt.Errorf("update workflow received non-float input")
		}
		err := workflow.SetUpdateHandlerWithOptions(
			ctx,
			updateName,
			func(ctx workflow.Context, i float64) (float64, error) {
				tmp := counter
				counter += i
				workflow.GetLogger(ctx).Info("counter updated", "added", i, "new-value", counter)
				return tmp, nil
			},
			workflow.UpdateHandlerOptions{
				Validator: func(ctx workflow.Context, i float64) error {
					if i < 0 {
						return fmt.Errorf("add value must be non-negative (%v)", i)
					}
					return nil
				}},
		)
		if err != nil {
			return 0, err
		}

		// wait on a signal to indicate the test is complete
		if ok := workflow.GetSignalChannel(ctx, "updates-done").Receive(ctx, nil); !ok {
			return 0, fmt.Errorf("signal channel was closed")
		}
		return counter, nil
	})

	// Start the workflow
	input := rand.Intn(100)
	run, err := t.s.Client.ExecuteWorkflow(
		t.s.Context,
		client.StartWorkflowOptions{TaskQueue: t.s.Worker().Options.TaskQueue},
		DevWorkflow,
		input,
	)
	t.s.NoError(err)

	// Stop the workflow when the test is complete
	defer func() {
		err := t.s.Client.SignalWorkflow(t.s.Context, run.GetID(), run.GetRunID(), "updates-done", nil)
		t.s.NoError(err)
	}()

	// successful update, output should contain the result
	res := t.execute("workflow", "update", "execute", "--address", t.s.Address(), "-w", run.GetID(), "--name", updateName, "-i", strconv.Itoa(input))
	t.s.NoError(res.Err)
	t.s.ContainsOnSameLine(res.Stdout.String(), "Result", strconv.Itoa(1*input))

	// successful update passing first-execution-run-id
	// Use --type here to make sure the alias works
	res = t.execute("workflow", "update", "execute", "--address", t.s.Address(), "-w", run.GetID(), "--type", updateName, "-i", strconv.Itoa(input), "--first-execution-run-id", run.GetRunID())
	t.s.NoError(res.Err)
	t.s.ContainsOnSameLine(res.Stdout.String(), "Result", strconv.Itoa(2*input))

	// successful update passing update-id
	res = t.execute("workflow", "update", "execute", "--address", t.s.Address(), "--update-id", strconv.Itoa(input), "-w", run.GetID(), "--name", updateName, "-i", strconv.Itoa(input))
	t.s.NoError(res.Err)
	t.s.ContainsOnSameLine(res.Stdout.String(), "UpdateID", strconv.Itoa(input))
	t.s.ContainsOnSameLine(res.Stdout.String(), "Result", strconv.Itoa(3*input))

	// successful update without input
	res = t.execute("workflow", "update", "execute", "--address", t.s.Address(), "--update-id", strconv.Itoa(input), "-w", run.GetID(), "--name", updateName)
	t.s.NoError(res.Err)
	t.s.ContainsOnSameLine(res.Stdout.String(), "UpdateID", strconv.Itoa(input))
	t.s.ContainsOnSameLine(res.Stdout.String(), "Result", strconv.Itoa(3*input))

	if t.useStart {
		// update rejected, name not supplied
		res = t.s.Execute("workflow", "update", "start", "--wait-for-stage", "accepted", "--address", t.s.Address(), "-w", run.GetID(), "-i", strconv.Itoa(input))
		t.s.ErrorContains(res.Err, "required flag(s) \"name\" not set")

		// update rejected, wrong workflowID
		res = t.s.Execute("workflow", "update", "start", "--wait-for-stage", "accepted", "--address", t.s.Address(), "-w", "nonexistent-wf-id", "--name", updateName, "-i", strconv.Itoa(input))
		t.s.ErrorContains(res.Err, "unable to update workflow")

		// update rejected, wrong update name
		res = t.s.Execute("workflow", "update", "start", "--wait-for-stage", "accepted", "--address", t.s.Address(), "-w", run.GetID(), "--name", "nonexistent-update-name", "-i", strconv.Itoa(input))
		t.s.ErrorContains(res.Err, "unable to update workflow")
	} else {
		// update rejected, name not supplied
		res = t.s.Execute("workflow", "update", "execute", "--address", t.s.Address(), "-w", run.GetID(), "-i", strconv.Itoa(input))
		t.s.ErrorContains(res.Err, "required flag(s) \"name\" not set")

		// update rejected, wrong workflowID
		res = t.s.Execute("workflow", "update", "execute", "--address", t.s.Address(), "-w", "nonexistent-wf-id", "--name", updateName, "-i", strconv.Itoa(input))
		t.s.ErrorContains(res.Err, "unable to update workflow")

		// update rejected, wrong update name
		res = t.s.Execute("workflow", "update", "execute", "--address", t.s.Address(), "-w", run.GetID(), "--name", "nonexistent-update-name", "-i", strconv.Itoa(input))
		t.s.ErrorContains(res.Err, "unable to update workflow")
	}
}

func (t workflowUpdateTest) execute(args ...string) *CommandResult {
	if !(len(args) >= 3 && args[0] == "workflow" && args[1] == "update" && (args[2] == "execute" || args[2] == "start")) {
		panic("invalid args passed to execute in `workflow update` test")
	}
	if t.useStart {
		// Test `update start` by confirming that we can start the update and
		// then use `update execute` to wait for it to complete.
		startArgs := append([]string{"workflow", "update", "start", "--wait-for-stage", "accepted"}, args[3:]...)
		res := t.s.Execute(startArgs...)
		t.s.NoError(res.Err)
		workflowID, updateID, name := t.extractWorkflowIDUpdateIDAndUpdateName(args, res.Stdout.String())
		return t.s.Execute("workflow", "update", "execute", "--address", t.s.Address(), "-w", workflowID, "--update-id", updateID, "--name", name)
	} else {
		return t.s.Execute(args...)
	}
}

func (t workflowUpdateTest) extractWorkflowIDUpdateIDAndUpdateName(args []string, stdout string) (string, string, string) {
	match := regexp.MustCompile(`UpdateID\s+(\S+)`).FindStringSubmatch(stdout)
	t.s.Equal(2, len(match), "stdout did not contain update ID in expected format")
	updateID := match[1]

	var workflowID string
	var name string
	for i, arg := range args {
		if arg == "-w" && i+1 < len(args) {
			workflowID = args[i+1]
		} else if arg == "--name" || arg == "--type" && i+1 < len(args) {
			name = args[i+1]
		}
	}
	if workflowID == "" {
		panic("No workflow ID found in args")
	}
	if name == "" {
		panic("No update name found in args")
	}

	return workflowID, updateID, name
}

func (s *SharedServerSuite) TestWorkflow_Update_Result() {
	updateName := "test-update"

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, val any) (any, error) {
		err := workflow.SetUpdateHandlerWithOptions(
			ctx,
			updateName,
			func(ctx workflow.Context, val string) (string, error) {
				if val == "fail" {
					return "", fmt.Errorf("failing")
				}
				return "hi " + val, nil
			},
			workflow.UpdateHandlerOptions{
				Validator: func(ctx workflow.Context, val string) error {
					if val == "reject" {
						return fmt.Errorf("rejecting")
					}
					return nil
				}},
		)
		if err != nil {
			return nil, err
		}
		if ok := workflow.GetSignalChannel(ctx, "updates-done").Receive(ctx, nil); !ok {
			return nil, fmt.Errorf("signal channel was closed")
		}
		return "done", nil
	})

	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	passingUpdateId := "passing"
	res := s.Execute("workflow", "update", "execute", "--address", s.Address(), "-w", run.GetID(),
		"--name", updateName, "-i", `"pass"`, "--update-id", passingUpdateId)
	s.NoError(res.Err)
	res = s.Execute("workflow", "update", "result", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", passingUpdateId)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Result", "hi pass")
	// JSON
	res = s.Execute("workflow", "update", "result", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", passingUpdateId, "-o", "json")
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), `"result": "hi pass"`)

	rejectedUpdateId := "rejected"
	res = s.Execute("workflow", "update", "execute", "--address", s.Address(), "-w", run.GetID(),
		"--name", updateName, "-i", `"reject"`, "--update-id", rejectedUpdateId)
	s.Error(res.Err)
	// Can't be found when rejected
	res = s.Execute("workflow", "update", "result", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", rejectedUpdateId)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "not found")

	failUpdateId := "fail"
	res = s.Execute("workflow", "update", "execute", "--address", s.Address(), "-w", run.GetID(),
		"--name", updateName, "-i", `"fail"`, "--update-id", failUpdateId)
	s.Error(res.Err)
	res = s.Execute("workflow", "update", "result", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", failUpdateId)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "update is failed")
	s.ContainsOnSameLine(res.Stdout.String(), "Failure", "failing")
	// JSON
	res = s.Execute("workflow", "update", "result", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", failUpdateId, "-o", "json")
	s.Error(res.Err)
	s.ErrorContains(res.Err, "update is failed")
	s.Contains(res.Stdout.String(), `"message": "failing"`)
}

func (s *SharedServerSuite) TestWorkflow_Update_Describe() {
	updateName := "test-update"

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, val any) (any, error) {
		updateProceed := workflow.GetSignalChannel(ctx, "update-proceed")
		err := workflow.SetUpdateHandlerWithOptions(
			ctx,
			updateName,
			func(ctx workflow.Context, val string) (string, error) {
				if val == "fail" {
					return "", fmt.Errorf("failing")
				}
				updateProceed.Receive(ctx, nil)
				return "hi " + val, nil
			},
			workflow.UpdateHandlerOptions{
				Validator: func(ctx workflow.Context, val string) error {
					if val == "reject" {
						return fmt.Errorf("rejecting")
					}
					return nil
				}},
		)
		if err != nil {
			return nil, err
		}
		if ok := workflow.GetSignalChannel(ctx, "updates-done").Receive(ctx, nil); !ok {
			return nil, fmt.Errorf("signal channel was closed")
		}
		return "done", nil
	})

	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	passingUpdateId := "passing"
	res := s.Execute("workflow", "update", "start", "--wait-for-stage", "accepted",
		"--address", s.Address(), "-w", run.GetID(), "--name", updateName,
		"-i", `"pass"`, "--update-id", passingUpdateId)
	s.NoError(res.Err)
	res = s.Execute("workflow", "update", "describe", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", passingUpdateId)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Stage", "Accepted")
	// Unblock it
	err = s.Client.SignalWorkflow(s.Context, run.GetID(), run.GetRunID(), "update-proceed", nil)
	s.NoError(err)
	s.Eventually(func() bool {
		res = s.Execute("workflow", "update", "describe", "--address", s.Address(), "-w", run.GetID(),
			"--update-id", passingUpdateId)
		s.NoError(res.Err)
		return AssertContainsOnSameLine(res.Stdout.String(), "Stage", "Completed") == nil
	}, 3*time.Second, 100*time.Millisecond)
	s.ContainsOnSameLine(res.Stdout.String(), "Result", "hi pass")
	// JSON
	res = s.Execute("workflow", "update", "describe", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", passingUpdateId, "-o", "json")
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), `"result": "hi pass"`)
	s.Contains(res.Stdout.String(), `"stage": "Completed"`)

	rejectedUpdateId := "rejected"
	res = s.Execute("workflow", "update", "execute", "--address", s.Address(), "-w", run.GetID(),
		"--name", updateName, "-i", `"reject"`, "--update-id", rejectedUpdateId)
	s.Error(res.Err)
	// Can't be found when rejected
	res = s.Execute("workflow", "update", "describe", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", rejectedUpdateId)
	s.Error(res.Err)
	s.ErrorContains(res.Err, "not found")

	failUpdateId := "fail"
	res = s.Execute("workflow", "update", "execute", "--address", s.Address(), "-w", run.GetID(),
		"--name", updateName, "-i", `"fail"`, "--update-id", failUpdateId)
	s.Error(res.Err)
	res = s.Execute("workflow", "update", "describe", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", failUpdateId)
	// Should *not* return nonzero here
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "Failure", "failing")
	// JSON
	res = s.Execute("workflow", "update", "describe", "--address", s.Address(), "-w", run.GetID(),
		"--update-id", failUpdateId, "-o", "json")
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), `"message": "failing"`)
}

func (s *SharedServerSuite) TestWorkflow_Cancel_BatchWorkflowSuccess() {
	res := s.testCancelBatchWorkflow(false)
	s.Contains(res.Stdout.String(), "approximately 5 workflow(s)")
	s.Contains(res.Stdout.String(), "Started batch")
}

func (s *SharedServerSuite) TestWorkflow_Cancel_BatchWorkflowSuccessJSON() {
	res := s.testCancelBatchWorkflow(true)
	var jsonRes map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonRes))
	s.NotEmpty(jsonRes["batchJobId"])
}

func (s *SharedServerSuite) testCancelBatchWorkflow(json bool) *CommandResult {
	// Make workflow wait for cancel and then return the context's error
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start 5 workflows
	runs := make([]client.WorkflowRun, 5)
	searchAttr := "keyword-" + uuid.NewString()
	for i := range runs {
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
		runs[i] = run
	}

	// Wait for all to appear in list
	s.Eventually(func() bool {
		resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
			Query: "CustomKeywordField = '" + searchAttr + "'",
		})
		s.NoError(err)
		return len(resp.Executions) == len(runs)
	}, 3*time.Second, 100*time.Millisecond)

	// Send batch cancel with a "y" for non-json or "--yes" for json
	args := []string{
		"workflow", "cancel",
		"--address", s.Address(),
		"--query", "CustomKeywordField = '" + searchAttr + "'",
		"--reason", "cancellation-test",
	}
	if json {
		args = append(args, "--yes", "-o", "json")
	} else {
		s.CommandHarness.Stdin.WriteString("y\n")
	}
	res := s.Execute(args...)
	s.NoError(res.Err)

	// Confirm that all workflows fail with ErrCanceled
	for _, run := range runs {
		s.Error(workflow.ErrCanceled, run.Get(s.Context, nil))
	}
	return res
}

func (s *SharedServerSuite) TestWorkflow_Query_SingleWorkflowSuccess() {
	s.testQueryWorkflow(false)
}

func (s *SharedServerSuite) TestWorkflow_Query_SingleWorkflowSuccessJSON() {
	s.testQueryWorkflow(true)
}

func (s *SharedServerSuite) testQueryWorkflow(json bool) {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		err := workflow.SetQueryHandler(ctx, "my-query", func(arg string) (any, error) {
			retme := struct {
				Echo  string `json:"input"`
				Other string `json:"other"`
			}{}
			retme.Echo = arg
			retme.Other = "yoyo"
			return retme, nil
		})
		if err != nil {
			return nil, err
		}
		return "done", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	args := []string{
		"workflow", "query",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--name", "my-query",
		"-i", `"hi"`,
	}
	if json {
		args = append(args, "-o", "json")
	}
	// Do the query
	res := s.Execute(args...)
	s.NoError(res.Err)
	if json {
		s.Contains(res.Stdout.String(), `"queryResult"`)
		s.Contains(res.Stdout.String(), `"input": "hi"`)
		s.Contains(res.Stdout.String(), `"other": "yoyo"`)
	} else {
		s.Contains(res.Stdout.String(), `{"input":"hi","other":"yoyo"}`)
	}

	s.NoError(run.Get(s.Context, nil))

	// Ensure query is rejected when using not open rejection condition
	args = []string{
		"workflow", "query",
		"--address", s.Address(),
		"-w", run.GetID(),
		// Use --type here to make sure the alias works
		"--type", "my-query",
		"-i", `"hi"`,
		"--reject-condition", "not_open",
	}
	if json {
		args = append(args, "-o", "json")
	}
	res = s.Execute(args...)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "query was rejected, workflow has status: Completed")
}

func (s *SharedServerSuite) TestWorkflow_Stack_SingleWorkflowSuccess() {
	s.testStackWorkflow(false)
}

func (s *SharedServerSuite) TestWorkflow_Stack_SingleWorkflowSuccessJSON() {
	s.testStackWorkflow(true)
}

func (s *SharedServerSuite) testStackWorkflow(json bool) {
	// Make workflow wait for signal and then return it
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		done := false
		workflow.Go(ctx, func(ctx workflow.Context) {
			_ = workflow.Await(ctx, func() bool {
				return done
			})
		})
		workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
		done = true
		return nil, nil

	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	args := []string{
		"workflow", "stack",
		"--address", s.Address(),
		"-w", run.GetID(),
	}
	if json {
		args = append(args, "-o", "json")
	}
	// Do the query
	res := s.Execute(args...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "coroutine root")
	s.Contains(res.Stdout.String(), "coroutine 2")

	// Unblock the workflow with a signal
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))

	s.NoError(run.Get(s.Context, nil))

	// Ensure query is rejected when using not open rejection condition
	args = []string{
		"workflow", "stack",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reject-condition", "not_open",
	}
	if json {
		args = append(args, "-o", "json")
	}
	res = s.Execute(args...)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "query was rejected, workflow has status: Completed")
}

func (s *SharedServerSuite) TestWorkflow_MetadataJSON() {
	s.testWorkflowMetadata(true)
}

func (s *SharedServerSuite) TestWorkflow_Metadata() {
	s.testWorkflowMetadata(false)
}

func (s *SharedServerSuite) testWorkflowMetadata(json bool) {
	// Make workflow wait for signal and then return it
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		done := false
		workflow.Go(ctx, func(ctx workflow.Context) {
			_ = workflow.Await(ctx, func() bool {
				return done
			})
		})
		workflow.SetQueryHandlerWithOptions(ctx, "my-query", func(arg string) (string, error) {
			return "hi", nil
		}, workflow.QueryHandlerOptions{Description: "q-desc"})
		workflow.SetUpdateHandlerWithOptions(ctx, "my-update",
			func(ctx workflow.Context, arg string) (string, error) { return "hi", nil },
			workflow.UpdateHandlerOptions{Description: "upd-desc"})
		workflow.SetCurrentDetails(ctx, "current-deets")
		workflow.GetSignalChannelWithOptions(ctx, "my-signal",
			workflow.SignalChannelOptions{Description: "sig-desc"}).Receive(ctx, nil)
		done = true
		return nil, nil

	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{
			TaskQueue:     s.Worker().Options.TaskQueue,
			StaticSummary: "summie",
			StaticDetails: "deets",
		},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	args := []string{
		"workflow", "metadata",
		"--address", s.Address(),
		"-w", run.GetID(),
	}
	if json {
		args = append(args, "-o", "json")
	}

	res := s.Execute(args...)
	s.NoError(res.Err)
	if !json {
		s.Contains(res.Stdout.String(), "Query Definitions:")
		s.ContainsOnSameLine(res.Stdout.String(), "my-query", "q-desc")
		s.Contains(res.Stdout.String(), "Signal Definitions:")
		s.ContainsOnSameLine(res.Stdout.String(), "my-signal", "sig-desc")
		s.Contains(res.Stdout.String(), "Update Definitions:")
		s.ContainsOnSameLine(res.Stdout.String(), "my-update", "upd-desc")
		s.Contains(res.Stdout.String(), "Current Details:")
		s.Contains(res.Stdout.String(), "current-deets")
	} else {
		s.Contains(res.Stdout.String(), "queryDefinitions")
		s.ContainsOnSameLine(res.Stdout.String(), "name", "my-query")
		s.Contains(res.Stdout.String(), "signalDefinitions")
		s.ContainsOnSameLine(res.Stdout.String(), "name", "my-signal")
		s.Contains(res.Stdout.String(), "updateDefinitions")
		s.ContainsOnSameLine(res.Stdout.String(), "name", "my-update")
		s.ContainsOnSameLine(res.Stdout.String(), "currentDetails", "current-deets")
	}

	// Unblock the workflow with a signal
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(run.Get(s.Context, nil))

	// Ensure query is rejected when using not open rejection condition
	args = []string{
		"workflow", "metadata",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reject-condition", "not_open",
	}
	if json {
		args = append(args, "-o", "json")
	}
	res = s.Execute(args...)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "query was rejected, workflow has status: Completed")
}
