package temporalcli_test

import (
	"encoding/json"
	"time"

	"go.temporal.io/api/enums/v1"
)

func (s *SharedServerSuite) TestTaskQueue_Describe_Simple() {
	// Wait until the poller appears
	s.Eventually(func() bool {
		desc, err := s.Client.DescribeTaskQueue(s.Context, s.Worker.Options.TaskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
		s.NoError(err)
		for _, poller := range desc.Pollers {
			if poller.Identity == s.DevServer.Options.ClientOptions.Identity {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "Worker never appeared")

	// Text
	res := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", s.Worker.Options.TaskQueue,
	)
	s.NoError(res.Err)
	// For text, just making sure our client identity is present is good enough
	s.Contains(res.Stdout.String(), s.DevServer.Options.ClientOptions.Identity)

	// JSON
	res = s.Execute(
		"task-queue", "describe",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker.Options.TaskQueue,
	)
	s.NoError(res.Err)
	var jsonOut struct {
		Pollers []map[string]any `json:"pollers"`
		TaskQueues []map[string]any `json:"taskQueues"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.GreaterOrEqual(1, len(jsonOut.TaskQueues))
	// Check identity in the output
	s.Equal(s.DevServer.Options.ClientOptions.Identity, jsonOut.Pollers[0]["identity"])
}
