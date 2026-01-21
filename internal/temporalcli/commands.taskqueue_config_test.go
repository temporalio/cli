package temporalcli_test

import (
	"encoding/json"
	"fmt"
	"os"
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
		"--queue-rps-limit", "20.0",
		"--queue-rps-limit-reason", "queue limit reason",
		"--fairness-key-rps-limit-default", "10.0",
		"--fairness-key-rps-limit-reason", "fairness limit reason",
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

func (s *SharedServerSuite) TestTaskQueue_Config_Unset_Rate_Limits() {
	taskQueue := "test-config-queue-" + s.T().Name()
	testIdentity := "test-identity-" + s.T().Name()
	var config taskQueueConfigType
	// Set initial configuration
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rps-limit", "10.0",
		"--fairness-key-rps-limit-default", "5.0",
	)
	s.NoError(res.Err)

	res = s.Execute(
		"task-queue", "config", "get",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))
	s.NotNil(config.QueueRateLimit)
	s.NotNil(config.QueueRateLimit.RateLimit)
	s.Equal(float32(10.0), config.QueueRateLimit.RateLimit.RequestsPerSecond)
	s.NotNil(config.FairnessKeysRateLimitDefault)
	s.NotNil(config.FairnessKeysRateLimitDefault.RateLimit)
	s.Equal(float32(5.0), config.FairnessKeysRateLimitDefault.RateLimit.RequestsPerSecond)

	// Unset queue rate limit (set to default)
	res = s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--queue-rps-limit", "default",
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

	var unsetQrlConfig taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &unsetQrlConfig))
	s.NotNil(unsetQrlConfig.QueueRateLimit)
	s.Nil(unsetQrlConfig.QueueRateLimit.RateLimit)
	s.NotNil(unsetQrlConfig.FairnessKeysRateLimitDefault)
	s.Equal(float32(5.0), unsetQrlConfig.FairnessKeysRateLimitDefault.RateLimit.RequestsPerSecond)

	// Unset fairness key rate limit
	res = s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--identity", testIdentity,
		"--fairness-key-rps-limit-default", "default",
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

	var unsetFkrlConfig taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &unsetFkrlConfig))
	s.NotNil(unsetFkrlConfig.FairnessKeysRateLimitDefault)
	s.Nil(unsetFkrlConfig.FairnessKeysRateLimitDefault.RateLimit)
}

func (s *SharedServerSuite) TestTaskQueue_Config_Workflow_Task_Queue_Restrictions() {
	taskQueue := "test-config-queue-" + s.T().Name()

	// Try to set queue rate limit on workflow task queue (should fail)
	res := s.Execute(
		"task-queue", "config", "set",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type", "workflow",
		"--queue-rps-limit", "10.0",
	)
	s.Error(res.Err)

	// TODO : add test to check if setting fairness key rate limit on workflow task queue is allowed
	// Will be done after the server PR (pending) to allow setting fairness key rate limit on workflow task queues is merged.
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
		"--queue-rps-limit", "12.5",
		"--queue-rps-limit-reason", "describe test",
	)
	s.NoError(res.Err)

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
	var result map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &result))

	cfg, ok := result["config"].(map[string]any)
	s.True(ok, "config should be an object")
	s.NotEmpty(cfg)

	qrl, ok := cfg["queue_rate_limit"].(map[string]any)
	s.True(ok, "config.queueRateLimit should be an object")
	s.NotEmpty(qrl)

	rl, ok := qrl["rate_limit"].(map[string]any)
	s.True(ok, "config.queueRateLimit.RateLimit should be an object")

	rps, ok := rl["requests_per_second"].(float64)
	s.True(ok, "requests_per_second should be a number")
	s.Equal(12.5, rps)

	md, ok := qrl["metadata"].(map[string]any)
	s.True(ok, "metadata should be an object")
	s.NotEmpty(md)

	reason, ok := md["reason"].(string)
	s.True(ok)
	s.Equal("describe test", reason)

	updID, ok := md["update_identity"].(string)
	s.True(ok)
	s.Equal(testIdentity, updID)

	updTime, _ := md["update_time"].(map[string]any)
	s.NotEmpty(updTime)
}

func (s *SharedServerSuite) TestTaskQueueConfig_EnvConfigNamespace() {
	// Create test namespace
	testNS := "tq-config-envconfig-test"
	res := s.Execute(
		"operator", "namespace", "create",
		"--address", s.Address(),
		"-n", testNS,
	)
	s.NoError(res.Err)

	// Create temp config file
	f, err := os.CreateTemp("", "temporal-test-*.toml")
	s.NoError(err)
	defer os.Remove(f.Name())

	_, err = fmt.Fprintf(f, `
[profile.default]
address = "%s"
namespace = "%s"
`, s.Address(), testNS)
	s.NoError(err)
	f.Close()

	// Set environment
	s.CommandHarness.Options.EnvLookup = EnvLookupMap{
		"TEMPORAL_CONFIG_FILE": f.Name(),
	}

	taskQueue := "test-tq-envconfig-" + s.T().Name()

	// Test 1: Set config without -n flag (should use envconfig)
	res = s.Execute(
		"task-queue", "config", "set",
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"--queue-rps-limit", "15.0",
		"--queue-rps-limit-reason", "envconfig test",
	)
	s.NoError(res.Err)

	// Test 2: Get config without -n flag (should use envconfig)
	res = s.Execute(
		"task-queue", "config", "get",
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-o", "json",
	)
	s.NoError(res.Err)

	var config taskQueueConfigType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &config))
	s.NotNil(config.QueueRateLimit)
	s.NotNil(config.QueueRateLimit.RateLimit)
	s.Equal(float32(15.0), config.QueueRateLimit.RateLimit.RequestsPerSecond)
	s.Equal("envconfig test", config.QueueRateLimit.Metadata.Reason)

	// Test 3: CLI flag should override envconfig - verify config doesn't exist in default namespace
	res = s.Execute(
		"task-queue", "config", "get",
		"--task-queue", taskQueue,
		"--task-queue-type", "activity",
		"-n", "default",
	)
	s.NoError(res.Err)
	// In default namespace, no config was set for this task queue
	s.Contains(res.Stdout.String(), "No configuration found for task queue",
		"CLI flag should override envconfig and query default namespace")
}
