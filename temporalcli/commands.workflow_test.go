package temporalcli_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestWorkflow_Signal_SingleWorkflowSuccess() {
	// Make workflow wait for signal and then return it
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		var ret any
		workflow.GetSignalChannel(ctx, "my-signal").Receive(ctx, &ret)
		return ret, nil
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
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
				TaskQueue:        s.Worker.Options.TaskQueue,
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
		"--name", "my-signal",
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

func (s *SharedServerSuite) TestWorkflow_Terminate_SingleWorkflowSuccess_WithoutReason() {
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
		ctx.Done().Receive(ctx, nil)
		return nil, ctx.Err()
	})

	// Start the workflow
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
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
	res := s.testTerminateBatchWorkflow(false)
	s.Contains(res.Stdout.String(), "approximately 5 workflow(s)")
	s.Contains(res.Stdout.String(), "Started batch")
}

func (s *SharedServerSuite) TestWorkflow_Terminate_BatchWorkflowSuccessJSON() {
	res := s.testTerminateBatchWorkflow(true)
	var jsonRes map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonRes))
	s.NotEmpty(jsonRes["batchJobId"])
}

func (s *SharedServerSuite) testTerminateBatchWorkflow(json bool) *CommandResult {
	s.Worker.OnDevWorkflow(func(ctx workflow.Context, a any) (any, error) {
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
				TaskQueue:        s.Worker.Options.TaskQueue,
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
	return res
}

func (s *SharedServerSuite) TestWorkflow_Update() {
	updateName := "test-update"

	s.Worker.OnDevWorkflow(func(ctx workflow.Context, val any) (any, error) {
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
	run, err := s.Client.ExecuteWorkflow(
		s.Context,
		client.StartWorkflowOptions{TaskQueue: s.Worker.Options.TaskQueue},
		DevWorkflow,
		input,
	)
	s.NoError(err)

	// Stop the workflow when the test is complete
	defer func() {
		err := s.Client.SignalWorkflow(s.Context, run.GetID(), run.GetRunID(), "updates-done", nil)
		s.NoError(err)
	}()

	// successful update, should show the result
	res := s.Execute("workflow", "update", "--address", s.Address(), "-w", run.GetID(), "--name", updateName, "-i", strconv.Itoa(input))
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), strconv.Itoa(input))

	// successful update passing first-execution-run-id
	res = s.Execute("workflow", "update", "--address", s.Address(), "-w", run.GetID(), "--name", updateName, "-i", strconv.Itoa(input), "--first-execution-run-id", run.GetRunID())
	s.NoError(res.Err)

	// update rejected, when name is not available
	res = s.Execute("workflow", "update", "--address", s.Address(), "-w", run.GetID(), "-i", strconv.Itoa(input))
	s.ErrorContains(res.Err, "required flag(s) \"name\" not set")

	// update rejected, wrong workflowID
	res = s.Execute("workflow", "update", "--address", s.Address(), "-w", "nonexistent-wf-id", "--name", updateName, "-i", strconv.Itoa(input))
	s.ErrorContains(res.Err, "unable to update workflow")

	// update rejected, wrong update name
	res = s.Execute("workflow", "update", "--address", s.Address(), "-w", run.GetID(), "--name", "nonexistent-update-name", "-i", strconv.Itoa(input))
	s.ErrorContains(res.Err, "unable to update workflow")
}
