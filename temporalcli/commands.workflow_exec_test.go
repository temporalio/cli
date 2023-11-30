package temporalcli_test

import "testing"

// TODO(cretz): To test:
// * Start workflow JSON and text
// * Execute workflow JSON
// * Execute workflow text ensure history shown including following runs

func TestWorkflowStart_Simple(t *testing.T) {
	h := StartWorkerCommandHarness(t, WorkerCommandHarnessOptions{})
	defer h.Close()

	// Execute text first
	h.Worker.Options.DevWorkflowOutput = map[string]string{"foo": "bar"}
	res := h.Execute(
		"workflow", "start",
		"--address", h.Server.Address(),
		"--task-queue", h.Worker.Options.TaskQueue,
		"--type", "DevWorkflow",
	)
	h.NoError(res.Err)
	// TODO(cretz): more...
}
