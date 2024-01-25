package temporalcli_test

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func (s *SharedServerSuite) TestWorkflow_Signal_SimpleSuccess() {
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
