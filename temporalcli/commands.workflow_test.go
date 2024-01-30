package temporalcli_test

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
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
	s.ContainsOnSameLine(res.Stderr.String(), "Started batch", "jobId")
}

func (s *SharedServerSuite) TestWorkflow_Signal_BatchWorkflowSuccessJSON() {
	res := s.testSignalBatchWorkflow(true)
	s.ContainsOnSameLine(res.Stderr.String(), "Started batch", "jobId")
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
		"--reason", "for a test",
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
