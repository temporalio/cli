package temporalcli_test

import (
	"fmt"

	"github.com/stretchr/testify/require"
)

func (s *SharedServerSuite) TestTaskQueue_Update_ReportConfig() {
	// First, update a task queue configuration (activity task queue with queue rate limit)
	res := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-describe-config-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--queue-rate-limit", "100",
	)
	fmt.Println(res.Stdout.String())
	require.NoError(s.T(), res.Err)

	// Now describe the task queue with report-config flag
	res2 := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", "test-describe-config-queue",
		"--task-queue-type-legacy", "activity",
		"--legacy-mode",
		"--report-config",
		"-o", "json",
	)
	fmt.Println(res2.Stdout.String())
	require.NoError(s.T(), res2.Err)

	// Parse the JSON output to verify config is included
	output := res2.Stdout.String()
	require.Contains(s.T(), output, "config")
	require.Contains(s.T(), output, "queue_rate_limit")
	require.Contains(s.T(), output, "100")
}
