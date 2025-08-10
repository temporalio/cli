package temporalcli_test

import (
	"encoding/json"
)

type taskQueueConfigType struct {
	QueueRateLimit               *rateLimitConfigType `json:"queueRateLimit,omitempty"`
	FairnessKeysRateLimitDefault *rateLimitConfigType `json:"fairnessKeysRateLimitDefault,omitempty"`
}

type rateLimitConfigType struct {
	RateLimit *rateLimitType `json:"rateLimit,omitempty"`
	Metadata  *metadataType  `json:"metadata,omitempty"`
}

type rateLimitType struct {
	RequestsPerSecond float32 `json:"requestsPerSecond"`
}

type metadataType struct {
	Reason         string `json:"reason,omitempty"`
	UpdateIdentity string `json:"updateIdentity,omitempty"`
	UpdateTime     string `json:"updateTime,omitempty"`
}

func (s *SharedServerSuite) TestTaskQueue_Config_Get_Empty() {
	// Test getting config for a task queue with no configuration
	res := s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--task-queue-type", "activity",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "No configuration found for task queue")
}

func (s *SharedServerSuite) TestTaskQueue_Config_Set_And_Get_Queue_Rate_Limit() {
	taskQueue := "test-config-queue-" + s.T().Name()
	testIdentity := "test-identity-" + s.T().Name()

	// Set queue rate limit with explicit identity
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rate-limit", "15.5",
		"--queue-rate-limit-reason", "test rate limit",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated task queue configuration")

	// Get the configuration and verify it was set using JSON output
	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var config taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))
	s.NotNil(config.QueueRateLimit)
	s.NotNil(config.QueueRateLimit.RateLimit)
	s.Equal(float32(15.5), config.QueueRateLimit.RateLimit.RequestsPerSecond)
	s.NotNil(config.QueueRateLimit.Metadata)
	s.Equal("test rate limit", config.QueueRateLimit.Metadata.Reason)
	s.Equal(testIdentity, config.QueueRateLimit.Metadata.UpdateIdentity)
	s.NotEmpty(config.QueueRateLimit.Metadata.UpdateTime)
}

func (s *SharedServerSuite) TestTaskQueue_Config_Set_And_Get_Fairness_Key_Rate_Limit() {
	taskQueue := "test-config-queue-" + s.T().Name()
	testIdentity := "test-identity-" + s.T().Name()

	// Set fairness key rate limit default with explicit identity
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--fairness-key-rate-limit-default", "8.0",
		"--fairness-key-rate-limit-reason", "test fairness limit",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated task queue configuration")

	// Get the configuration and verify it was set using JSON output
	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var config taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))
	s.NotNil(config.FairnessKeysRateLimitDefault)
	s.NotNil(config.FairnessKeysRateLimitDefault.RateLimit)
	s.Equal(float32(8.0), config.FairnessKeysRateLimitDefault.RateLimit.RequestsPerSecond)
	s.NotNil(config.FairnessKeysRateLimitDefault.Metadata)
	s.Equal("test fairness limit", config.FairnessKeysRateLimitDefault.Metadata.Reason)
	s.Equal(testIdentity, config.FairnessKeysRateLimitDefault.Metadata.UpdateIdentity)
	s.NotEmpty(config.FairnessKeysRateLimitDefault.Metadata.UpdateTime)
}

func (s *SharedServerSuite) TestTaskQueue_Config_Set_And_Get_Both_Limits() {
	taskQueue := "test-config-queue-" + s.T().Name()
	testIdentity := "test-identity-" + s.T().Name()

	// Set both queue rate limit and fairness key rate limit with explicit identity
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rate-limit", "20.0",
		"--queue-rate-limit-reason", "queue limit reason",
		"--fairness-key-rate-limit-default", "10.0",
		"--fairness-key-rate-limit-reason", "fairness limit reason",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated task queue configuration")

	// Get the configuration and verify both were set using JSON output
	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var config taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))

	// Verify queue rate limit
	s.NotNil(config.QueueRateLimit)
	s.NotNil(config.QueueRateLimit.RateLimit)
	s.Equal(float32(20.0), config.QueueRateLimit.RateLimit.RequestsPerSecond)
	s.NotNil(config.QueueRateLimit.Metadata)
	s.Equal("queue limit reason", config.QueueRateLimit.Metadata.Reason)
	s.Equal(testIdentity, config.QueueRateLimit.Metadata.UpdateIdentity)
	s.NotEmpty(config.QueueRateLimit.Metadata.UpdateTime)

	// Verify fairness key rate limit
	s.NotNil(config.FairnessKeysRateLimitDefault)
	s.NotNil(config.FairnessKeysRateLimitDefault.RateLimit)
	s.Equal(float32(10.0), config.FairnessKeysRateLimitDefault.RateLimit.RequestsPerSecond)
	s.NotNil(config.FairnessKeysRateLimitDefault.Metadata)
	s.Equal("fairness limit reason", config.FairnessKeysRateLimitDefault.Metadata.Reason)
	s.Equal(testIdentity, config.FairnessKeysRateLimitDefault.Metadata.UpdateIdentity)
	s.NotEmpty(config.FairnessKeysRateLimitDefault.Metadata.UpdateTime)
}

func (s *SharedServerSuite) TestTaskQueue_Config_Update_Existing_Config() {
	taskQueue := "test-config-queue-" + s.T().Name()
	initialIdentity := "initial-identity-" + s.T().Name()
	updatedIdentity := "updated-identity-" + s.T().Name()

	// Set initial configuration with first identity
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", initialIdentity,
		"--queue-rate-limit", "5.0",
		"--queue-rate-limit-reason", "initial queue limit",
	)
	s.NoError(res.Err)

	// Update the configuration with different identity
	res = s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", updatedIdentity,
		"--queue-rate-limit", "25.0",
		"--queue-rate-limit-reason", "updated queue limit",
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated task queue configuration")

	// Get the configuration and verify it was updated using JSON output
	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var config taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))
	s.NotNil(config.QueueRateLimit)
	s.Equal(float32(25.0), config.QueueRateLimit.RateLimit.RequestsPerSecond)
	s.Equal("updated queue limit", config.QueueRateLimit.Metadata.Reason)
	s.Equal(updatedIdentity, config.QueueRateLimit.Metadata.UpdateIdentity)
	s.NotEmpty(config.QueueRateLimit.Metadata.UpdateTime)
}

func (s *SharedServerSuite) TestTaskQueue_Config_Unset_Rate_Limits() {
	taskQueue := "test-config-queue-" + s.T().Name()
	testIdentity := "test-identity-" + s.T().Name()

	// Set initial configuration
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rate-limit", "10.0",
		"--fairness-key-rate-limit-default", "5.0",
	)
	s.NoError(res.Err)

	// Unset queue rate limit (set to -1)
	res = s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rate-limit", "-1",
	)
	s.NoError(res.Err)

	// Get configuration and verify queue rate limit is unset using JSON output
	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var config taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))
	s.Nil(config.QueueRateLimit)
	s.NotNil(config.FairnessKeysRateLimitDefault)
	s.Equal(float32(5.0), config.FairnessKeysRateLimitDefault.RateLimit.RequestsPerSecond)

	// Unset fairness key rate limit
	res = s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--fairness-key-rate-limit-default", "-1",
	)
	s.NoError(res.Err)

	// Get configuration and verify both are unset using JSON output
	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var finalConfig taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &finalConfig))
	s.Nil(finalConfig.QueueRateLimit)
	s.Nil(finalConfig.FairnessKeysRateLimitDefault)
}

func (s *SharedServerSuite) TestTaskQueue_Config_Workflow_Task_Queue_Restrictions() {
	taskQueue := "test-config-queue-" + s.T().Name()

	// Try to set queue rate limit on workflow task queue (should fail)
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "workflow",
		"--queue-rate-limit", "10.0",
	)
	s.Error(res.Err)

	// TODO : add test to check if setting fairness key rate limit on workflow task queue is allowed
	// Will be done after we merge the server PR (pending) to allow setting fairness key rate limit on workflow task queues
}

func (s *SharedServerSuite) TestTaskQueue_Config_Describe_With_Report_Config() {
	taskQueue := "test-config-queue-" + s.T().Name()
	testIdentity := "test-identity-" + s.T().Name()

	// Set configuration
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rate-limit", "12.5",
		"--queue-rate-limit-reason", "describe test",
	)
	s.NoError(res.Err)

	// Use describe command with --report-config flag
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type-legacy", "activity",
		"--report-config",
		"--legacy-mode",
	)
	s.NoError(res.Err)
	out := res.Stdout.String()
	s.Contains(out, "Task Queue Configuration:")
	s.Contains(out, "Queue Rate Limit")
	s.Contains(out, "12.50 rps")
	s.Contains(out, "describe test")

	// Test JSON output with describe
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type-legacy", "activity",
		"--report-config",
		"--legacy-mode",
		"-o", "json",
	)
	s.NoError(res.Err)

	// The JSON output should contain the config section
	var result map[string]interface{}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &result))
	s.Contains(result, "config")
}
