package temporalcli_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/api/common/v1"

	"github.com/google/uuid"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/assert"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestWorkflow_Describe_ActivityFailing() {
	// Set activity to just continually error
	s.Worker().OnDevActivity(func(ctx context.Context, a any) (any, error) {
		return nil, fmt.Errorf("intentional error")
	})

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		})
		var res any
		err := workflow.ExecuteActivity(ctx, DevActivity, input).Get(ctx, &res)
		return res, err
	})

	// Start the workflow and wait until it has at least reached activity failure
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
		return len(resp.PendingActivities) > 0 && resp.PendingActivities[0].LastFailure != nil
	}, 5*time.Second, 100*time.Millisecond)

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", run.GetID())
	s.Contains(out, "Pending Activities: 1")
	s.ContainsOnSameLine(out, "LastFailure", "intentional error")
	s.Contains(out, "Pending Child Workflows: 0")

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	var jsonOut workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
	s.Equal("intentional error", jsonOut.PendingActivities[0].LastFailure.Message)
}

func (s *SharedServerSuite) TestWorkflow_Describe_Completed() {
	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		map[string]string{"foo": "bar"},
	)
	s.NoError(err)
	s.NoError(run.Get(s.Context, nil))

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.ContainsOnSameLine(out, "Result", `{"foo":"bar"}`)

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.NotNil(jsonOut["closeEvent"])
	s.Equal(map[string]any{"foo": "bar"}, jsonOut["result"])
}

func (s *SharedServerSuite) TestWorkflow_Describe_Versioned() {
	buildIdTaskQueue := uuid.NewString()

	buildId := "id1"
	res := s.Execute(
		"task-queue", "versioning", "insert-assignment-rule",
		"--build-id", buildId,
		"-y",
		"--address", s.Address(),
		"--task-queue", buildIdTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)

	var run client.WorkflowRun
	var err error
	s.Eventually(
		func() bool {
			run, err = s.Client.ExecuteWorkflow(
				s.Context,
				client.StartWorkflowOptions{TaskQueue: buildIdTaskQueue},
				DevWorkflow,
				map[string]string{"foo": "bar"},
			)
			s.NoError(err)

			// Text
			res = s.Execute(
				"workflow", "describe",
				"--address", s.Address(),
				"-w", run.GetID(),
			)
			s.NoError(res.Err)
			out := res.Stdout.String()
			return AssertContainsOnSameLine(out, "AssignedBuildId", buildId) == nil
		}, 10*time.Second, 100*time.Millisecond)

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	var jsonOut map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	execInfo, ok := jsonOut["workflowExecutionInfo"].(map[string]any)
	s.True(ok)
	s.Equal(buildId, execInfo["assignedBuildId"])
}

func (s *SharedServerSuite) TestWorkflow_Describe_NotDecodable() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return temporalcli.RawValue{
			Payload: &common.Payload{
				Metadata: map[string][]byte{"encoding": []byte("some-encoding")},
				Data:     []byte("some-data"),
			},
		}, nil
	})
	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		nil, // input is irrelevant
	)
	s.NoError(err)
	s.NoError(run.Get(s.Context, nil))

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Status", "COMPLETED")
	s.ContainsOnSameLine(out, "ResultEncoding", "some-encoding")

	// TODO: Enable once updated to api-go >= 1.33
	// JSON
	//res = s.Execute(
	//	"workflow", "describe",
	//	"-o", "json",
	//	"--address", s.Address(),
	//	"-w", run.GetID(),
	//)
	//s.NoError(res.Err)
	//var jsonOut map[string]any
	//s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	//s.NotNil(jsonOut["closeEvent"])
	//s.Equal("some-encoding", jsonOut["resultEncoding"])
}

func (s *SharedServerSuite) TestWorkflow_Describe_ResetPoints() {
	// Start the workflow and wait until it has at least reached activity failure
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		map[string]string{"foo": "bar"},
	)
	s.NoError(err)
	s.NoError(run.Get(s.Context, nil))

	// Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reset-points",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.NotContains(out, "Status")
	s.NotContains(out, "Result")
	s.Contains(out, "Auto Reset Points")

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--reset-points",
	)
	s.NoError(res.Err)
	var jsonOut []map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.NotNil(jsonOut[0])
	s.NotNil(jsonOut[0]["EventId"])
}

func (s *SharedServerSuite) TestWorkflow_Show_Follow() {
	s.testWorkflowShowFollow(true)
	s.testWorkflowShowFollow(false)
}

func (s *SharedServerSuite) testWorkflowShowFollow(detailed bool) {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	outputCh := make(chan *CommandResult)
	// Follow the workflow
	go func() {
		args := []string{"workflow", "show",
			"--address", s.Address(),
			"-w", run.GetID(),
			"--follow"}
		if detailed {
			args = append(args, "--detailed")
		}
		res := s.Execute(args...)
		outputCh <- res
		close(outputCh)
	}()

	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))

	// Ensure following completes
	res := <-outputCh
	s.NoError(res.Err)
	output := res.Stdout.String()
	// Confirm result present
	s.ContainsOnSameLine(output, "Result", `"hi!"`)
	s.NoError(run.Get(s.Context, nil))

	// Detailed uses sections, non-detailed uses table
	if detailed {
		s.Contains(output, "input[0]: ignored")
		s.Contains(output, "signalName: my-signal")
	} else {
		s.Contains(output, "WorkflowExecutionSignaled")
	}
}

func (s *SharedServerSuite) TestWorkflow_Show_NoFollow() {
	s.testWorkflowShowNoFollow(true)
	s.testWorkflowShowNoFollow(false)
}
func (s *SharedServerSuite) testWorkflowShowNoFollow(detailed bool) {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)

	args := []string{"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID()}
	if detailed {
		args = append(args, "--detailed")
	}
	res := s.Execute(args...)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.NotContains(out, "my-signal")
	s.NotContains(out, "Results:")

	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(run.Get(s.Context, nil))

	res = s.Execute(args...)
	s.NoError(res.Err)
	out = res.Stdout.String()
	if detailed {
		s.Contains(out, "my-signal")
	}
	s.ContainsOnSameLine(out, "Result", `"hi!"`)
}

func (s *SharedServerSuite) TestWorkflow_Show_JSON() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, nil)
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"workflow-param",
	)
	s.NoError(err)

	res := s.Execute(
		"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Contains(out, `"events": [`)
	s.Contains(out, `"eventType": "EVENT_TYPE_WORKFLOW_EXECUTION_STARTED"`)
	// Make sure payloads are still encoded non-shorthand
	s.Contains(out, base64.StdEncoding.EncodeToString([]byte(`"workflow-param"`)))

	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(run.Get(s.Context, nil))

	res = s.Execute(
		"workflow", "show",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, `"events": [`)
	s.Contains(out, `"signalName": "my-signal"`)
	s.NotContains(out, "Results:")
}

func (s *SharedServerSuite) TestWorkflow_List() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		return a, nil
	})

	// Start the workflow
	for i := 0; i < 3; i++ {
		run, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
			DevWorkflow,
			strconv.Itoa(i),
		)
		s.NoError(err)
		s.NoError(run.Get(s.Context, nil))
	}

	res := s.Execute(
		"workflow", "list",
		"--address", s.Address(),
		"--query", fmt.Sprintf(`TaskQueue="%s"`, s.Worker().Options.TaskQueue),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Completed", "DevWorkflow")

	// JSON
	res = s.Execute(
		"workflow", "list",
		"--address", s.Address(),
		"--query", fmt.Sprintf(`TaskQueue="%s"`, s.Worker().Options.TaskQueue),
		"-o", "json",
	)
	s.NoError(res.Err)
	// Output is currently a series of JSON objects
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "name", "DevWorkflow")
	s.ContainsOnSameLine(out, "status", "WORKFLOW_EXECUTION_STATUS_COMPLETED")
}

func (s *SharedServerSuite) TestWorkflow_Count() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, shouldComplete any) (any, error) {
		// Only complete if shouldComplete is a true bool
		shouldCompleteBool, _ := shouldComplete.(bool)
		return nil, workflow.Await(ctx, func() bool { return shouldCompleteBool })
	})

	// Create 3 that complete and 2 that don't
	for i := 0; i < 5; i++ {
		_, err := s.Client.ExecuteWorkflow(
			s.Context,
			client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
			DevWorkflow,
			i < 3,
		)
		s.NoError(err)
	}

	// List and confirm they are all there in expected statuses
	s.Eventually(
		func() bool {
			resp, err := s.Client.ListWorkflow(s.Context, &workflowservice.ListWorkflowExecutionsRequest{
				Query: "TaskQueue = '" + s.Worker().Options.TaskQueue + "'",
			})
			s.NoError(err)
			var completed, running int
			for _, exec := range resp.Executions {
				if exec.Status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED {
					completed++
				} else if exec.Status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
					running++
				}
			}
			return completed == 3 && running == 2
		},
		10*time.Second,
		100*time.Millisecond,
	)

	// Simple count w/out grouping
	res := s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"'",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Equal("Total: 5", strings.TrimSpace(out))

	// Grouped
	res = s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"' GROUP BY ExecutionStatus",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, "Total: 5")
	s.Contains(out, "Group total: 2, values: Running")
	s.Contains(out, "Group total: 3, values: Completed")

	// Simple count w/out grouping JSON
	res = s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"'",
		"-o", "json",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	// Proto JSON makes this count a string
	s.Contains(out, `"count": "5"`)

	// Grouped JSON
	res = s.Execute(
		"workflow", "count",
		"--address", s.Address(),
		"--query", "TaskQueue = '"+s.Worker().Options.TaskQueue+"' GROUP BY ExecutionStatus",
		"-o", "jsonl",
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.Contains(out, `"count":"5"`)
	s.Contains(out, `{"groupValues":["Running"],"count":"2"}`)
	s.Contains(out, `{"groupValues":["Completed"],"count":"3"}`)
}

func (s *SharedServerSuite) TestWorkflow_Describe_Deployment() {
	buildId := uuid.NewString()
	seriesName := uuid.NewString()
	// Workflow that waits to be canceled.
	waitingWorkflow := func(ctx workflow.Context) error {
		ctx.Done().Receive(ctx, nil)
		return ctx.Err()
	}
	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Worker: worker.Options{
			BuildID:                 buildId,
			UseBuildIDForVersioning: true,
			DeploymentOptions: worker.DeploymentOptions{
				DeploymentSeriesName:      seriesName,
				DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
			},
		},
		Workflows: []any{waitingWorkflow},
	})
	defer w.Stop()

	res := s.Execute(
		"worker", "deployment", "set-current",
		"--address", s.Address(),
		"--series-name", seriesName,
		"--build-id", buildId,
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
		assert.Contains(t, res.Stdout.String(), buildId)
		assert.Contains(t, res.Stdout.String(), "Pinned")
	}, 30*time.Second, 100*time.Millisecond)

	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "Behavior", "Pinned")
	s.ContainsOnSameLine(out, "DeploymentBuildID", buildId)
	s.ContainsOnSameLine(out, "DeploymentSeriesName", seriesName)
	s.ContainsOnSameLine(out, "OverrideBehavior", "Unspecified")

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
	s.Equal("Pinned", versioningInfo.Behavior.String())
	s.Equal(buildId, versioningInfo.Deployment.BuildId)
	s.Equal(seriesName, versioningInfo.Deployment.SeriesName)
	s.Nil(versioningInfo.VersioningOverride)
	s.Nil(versioningInfo.DeploymentTransition)
}

func (s *SharedServerSuite) TestWorkflow_Describe_NexusOperationAndCallback() {
	handlerWorkflowID := uuid.NewString()
	endpointName := validEndpointName(s.T())

	// Workflow that waits to be canceled.
	handlerWorkflow := func(ctx workflow.Context, input nexus.NoValue) (nexus.NoValue, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	}

	// Expose the workflow above as an operation.
	op := temporalnexus.NewWorkflowRunOperation("test-op", handlerWorkflow, func(ctx context.Context, _ nexus.NoValue, opts nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{ID: handlerWorkflowID}, nil
	})
	service := nexus.NewService("test-service")
	s.NoError(service.Register(op))

	// Call the operation from this workflow.
	callerWorkflow := func(ctx workflow.Context) error {
		client := workflow.NewNexusClient(endpointName, service.Name)
		opCtx, cancel := workflow.WithCancel(ctx)
		fut := client.ExecuteOperation(opCtx, op, nil, workflow.NexusOperationOptions{})
		var exec workflow.NexusOperationExecution
		if err := fut.GetNexusOperationExecution().Get(ctx, &exec); err != nil {
			return err
		}
		// Cancel the operation on this signal.
		ch := workflow.GetSignalChannel(ctx, "cancel-op")
		ch.Receive(ctx, nil)
		cancel()
		// Wait for the operation to be canceled.
		return fut.Get(ctx, nil)
	}

	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Workflows:     []any{handlerWorkflow, callerWorkflow},
		NexusServices: []*nexus.Service{service},
	})
	defer w.Stop()

	// Create an endpoint for this test.
	_, err := s.Client.OperatorService().CreateNexusEndpoint(s.Context, &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexuspb.EndpointSpec{
			Name: endpointName,
			Target: &nexuspb.EndpointTarget{
				Variant: &nexuspb.EndpointTarget_Worker_{
					Worker: &nexuspb.EndpointTarget_Worker{
						Namespace: s.Namespace(),
						TaskQueue: w.Options.TaskQueue,
					},
				},
			},
		},
	})
	s.NoError(err)

	// Start the workflow and wait until the operation is started.
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: w.Options.TaskQueue},
		callerWorkflow,
	)
	s.NoError(err)

	// Wait for the operation to be started.
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		return len(resp.PendingNexusOperations) > 0 && resp.PendingNexusOperations[0].State == enums.PENDING_NEXUS_OPERATION_STATE_STARTED
	}, 30*time.Second, 100*time.Millisecond)

	// Operations - Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", run.GetID())
	s.Contains(out, "Pending Nexus Operations: 1")
	s.ContainsOnSameLine(out, "Endpoint", endpointName)
	s.ContainsOnSameLine(out, "Service", "test-service")
	s.ContainsOnSameLine(out, "Operation", "test-op")

	// Operations - JSON
	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--output", "json",
	)
	s.NoError(res.Err)
	var callerDesc workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &callerDesc, true))
	s.Equal(endpointName, callerDesc.PendingNexusOperations[0].Endpoint)
	s.Equal("test-service", callerDesc.PendingNexusOperations[0].Service)
	s.Equal("test-op", callerDesc.PendingNexusOperations[0].Operation)

	// Cancel the operation and check the callback on the handler workflow.
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), run.GetRunID(), "cancel-op", nil))
	s.ErrorAs(run.Get(s.Context, nil), new(*temporal.CanceledError))

	// Callbacks - Text
	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", handlerWorkflowID,
	)
	s.NoError(res.Err)
	out = res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", handlerWorkflowID)
	s.Contains(out, "Callbacks: 1")
	s.ContainsOnSameLine(out, "URL", "http://"+s.DevServer.Options.FrontendIP)
	s.ContainsOnSameLine(out, "Trigger", "WorkflowClosed")
	s.ContainsOnSameLine(out, "State", "Succeeded")

	// Callbacks - JSON
	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", handlerWorkflowID,
		"--output", "json",
	)
	s.NoError(res.Err)
	var handlerDesc workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &handlerDesc, true))
	s.Equal(enums.CALLBACK_STATE_SUCCEEDED, handlerDesc.Callbacks[0].State)
}

func (s *SharedServerSuite) TestWorkflow_Describe_NexusOperationBlocked() {
	endpointName := validEndpointName(s.T())

	// Call an unreachable operation from this workflow.
	callerWorkflow := func(ctx workflow.Context) error {
		client := workflow.NewNexusClient(endpointName, "test-service")
		fut := client.ExecuteOperation(ctx, "test-op", nil, workflow.NexusOperationOptions{})
		// Destination is unreachable, the future will never complete.
		return fut.GetNexusOperationExecution().Get(ctx, nil)
	}

	w := s.DevServer.StartDevWorker(s.Suite.T(), DevWorkerOptions{
		Workflows: []any{callerWorkflow},
	})
	defer w.Stop()

	// Create an endpoint for this test.
	_, err := s.Client.OperatorService().CreateNexusEndpoint(
		s.Context,
		&operatorservice.CreateNexusEndpointRequest{
			Spec: &nexuspb.EndpointSpec{
				Name: endpointName,
				Target: &nexuspb.EndpointTarget{
					Variant: &nexuspb.EndpointTarget_External_{
						External: &nexuspb.EndpointTarget_External{
							Url: "http://localhost:12345", // unreachable destination
						},
					},
				},
			},
		},
	)
	s.NoError(err)

	// Start the workflow and wait until the operation is started.
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: w.Options.TaskQueue},
		callerWorkflow,
	)
	s.NoError(err)

	// Start another workflow so it will trigger the circuit breaker faster.
	dummyRun, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: w.Options.TaskQueue},
		callerWorkflow,
	)
	s.NoError(err)

	// Wait for the operation to be blocked
	s.Eventually(func() bool {
		resp, err := s.Client.DescribeWorkflowExecution(s.Context, run.GetID(), run.GetRunID())
		s.NoError(err)
		return len(resp.PendingNexusOperations) > 0 &&
			resp.PendingNexusOperations[0].State == enums.PENDING_NEXUS_OPERATION_STATE_BLOCKED
	}, 30*time.Second, 100*time.Millisecond)

	// Operations - Text
	res := s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "WorkflowId", run.GetID())
	s.Contains(out, "Pending Nexus Operations: 1")
	s.ContainsOnSameLine(out, "Endpoint", endpointName)
	s.ContainsOnSameLine(out, "Service", "test-service")
	s.ContainsOnSameLine(out, "Operation", "test-op")
	s.ContainsOnSameLine(out, "BlockedReason", "The circuit breaker is open.")

	// Operations - JSON
	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", run.GetID(),
		"--output", "json",
	)
	s.NoError(res.Err)
	var callerDesc workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &callerDesc, true))
	s.Equal(endpointName, callerDesc.PendingNexusOperations[0].Endpoint)
	s.Equal("test-service", callerDesc.PendingNexusOperations[0].Service)
	s.Equal("test-op", callerDesc.PendingNexusOperations[0].Operation)
	s.Equal(enums.PENDING_NEXUS_OPERATION_STATE_BLOCKED, callerDesc.PendingNexusOperations[0].State)
	s.Equal("The circuit breaker is open.", callerDesc.PendingNexusOperations[0].BlockedReason)

	s.NoError(s.Client.TerminateWorkflow(s.Context, run.GetID(), run.GetRunID(), ""))
	s.NoError(s.Client.TerminateWorkflow(s.Context, dummyRun.GetID(), dummyRun.GetRunID(), ""))
}

func (s *SharedServerSuite) Test_WorkflowResult() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		sigs := 0
		for {
			var val string
			workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, &val)
			if val == "fail" {
				return nil, fmt.Errorf("failed on purpose")
			}
			sigs += 1
			if sigs == 2 {
				break
			}
		}
		return "hi!", nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)
	// Send signals to complete
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", nil))

	args := []string{"workflow", "result",
		"--address", s.Address(),
		"-w", run.GetID()}
	res := s.Execute(args...)
	s.NoError(res.Err)
	output := res.Stdout.String()
	// Confirm result present
	s.ContainsOnSameLine(output, "Status", "COMPLETED")
	s.ContainsOnSameLine(output, "Result", `"hi!"`)

	s.NoError(run.Get(s.Context, nil))

	// JSON
	args = []string{"workflow", "result",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json"}
	res = s.Execute(args...)
	s.NoError(res.Err)
	output = res.Stdout.String()
	// Confirm result present
	s.Contains(output, `"status": "COMPLETED"`)
	s.Contains(output, `"result": "hi!"`)
	s.Contains(output, "workflowExecutionCompletedEventAttributes")

	// Failed version
	run, err = s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker().Options.TaskQueue},
		DevWorkflow,
		"ignored",
	)
	s.NoError(err)
	s.NoError(s.Client.SignalWorkflow(s.Context, run.GetID(), "", "my-signal", "fail"))

	args = []string{"workflow", "result",
		"--address", s.Address(),
		"-w", run.GetID()}
	res = s.Execute(args...)
	s.Error(res.Err)
	output = res.Stdout.String()
	// Confirm result present
	s.ContainsOnSameLine(output, "Status", "FAILED")
	s.ContainsOnSameLine(output, "Message", "failed on purpose")

	s.Error(run.Get(s.Context, nil))

	// JSON
	args = []string{"workflow", "result",
		"--address", s.Address(),
		"-w", run.GetID(),
		"-o", "json"}
	res = s.Execute(args...)
	s.Error(res.Err)
	output = res.Stdout.String()
	// Confirm result present
	s.Contains(output, `"status": "FAILED"`)
	s.Contains(output, `"message": "failed on purpose"`)
	s.Contains(output, "workflowExecutionFailedEventAttributes")
}

func (s *SharedServerSuite) TestWorkflow_Describe_WorkflowMetadata() {
	workflowId := uuid.NewString()

	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return map[string]string{"foo": "bar"}, nil
	})

	res := s.Execute(
		"workflow", "start",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", workflowId,
		"--static-summary", "summie",
		"--static-details", "deets",
	)
	s.NoError(res.Err)

	// Text
	res = s.Execute(
		"workflow", "describe",
		"--address", s.Address(),
		"-w", workflowId,
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.ContainsOnSameLine(out, "StaticSummary", "summie")
	s.ContainsOnSameLine(out, "StaticDetails", "deets")

	// JSON
	res = s.Execute(
		"workflow", "describe",
		"-o", "json",
		"--address", s.Address(),
		"-w", workflowId,
	)
	s.NoError(res.Err)
	var jsonOut workflowservice.DescribeWorkflowExecutionResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &jsonOut, true))
	s.NotNil(jsonOut.ExecutionConfig.UserMetadata.Summary)
	s.NotNil(jsonOut.ExecutionConfig.UserMetadata.Details)
}
