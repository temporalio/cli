package temporalcli_test

import "encoding/json"

func (s *SharedServerSuite) TestTaskQueue_Describe_Simple() {
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
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	// Check identity in the output
	s.Equal(s.DevServer.Options.ClientOptions.Identity, jsonOut.Pollers[0]["identity"])
}
