package temporalcli_test

import (
	"encoding/json"

	"github.com/stretchr/testify/require"
)

// Helper function to get task queue config via describe command
func (s *SharedServerSuite) getTaskQueueConfig(taskQueue, taskQueueType string) map[string]interface{} {
	res := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", taskQueue,
		"--task-queue-type-legacy", taskQueueType,
		"--legacy-mode",
		"--report-config",
		"-o", "json",
	)
	require.NoError(s.T(), res.Err)

	var result map[string]any
	err := json.Unmarshal(res.Stdout.Bytes(), &result)
	require.NoError(s.T(), err)

	return result
}

// Helper function to verify rate limit in config
func (s *SharedServerSuite) verifyRateLimit(config map[string]interface{}, expectedRateLimit float64, expectedReason string) {
	configData, exists := config["config"]
	require.True(s.T(), exists, "Config should exist in response")

	configMap, ok := configData.(map[string]any)
	require.True(s.T(), ok, "Config should be a map")

	queueRateLimit, exists := configMap["queue_rate_limit"]
	require.True(s.T(), exists, "Queue rate limit should exist")

	rateLimitMap, ok := queueRateLimit.(map[string]any)
	require.True(s.T(), ok, "Queue rate limit should be a map")

	rateLimit, exists := rateLimitMap["rate_limit"]
	require.True(s.T(), exists, "Rate limit should exist")

	rateLimitData, ok := rateLimit.(map[string]any)
	require.True(s.T(), ok, "Rate limit should be a map")

	requestsPerSecond, exists := rateLimitData["requests_per_second"]
	if !exists {
		// If requests_per_second is missing, it means the rate limit is 0
		require.Equal(s.T(), float64(0), expectedRateLimit, "Rate limit should be 0 when requests_per_second is missing")
	} else {
		actualRateLimit, ok := requestsPerSecond.(float64)
		require.True(s.T(), ok, "Requests per second should be a float64")
		require.Equal(s.T(), expectedRateLimit, actualRateLimit, "Rate limit should match expected value")
	}

	if expectedReason != "" {
		metadata, exists := rateLimitMap["metadata"]
		require.True(s.T(), exists, "Metadata should exist")

		metadataMap, ok := metadata.(map[string]any)
		require.True(s.T(), ok, "Metadata should be a map")

		reason, exists := metadataMap["reason"]
		require.True(s.T(), exists, "Reason should exist")

		actualReason, ok := reason.(string)
		require.True(s.T(), ok, "Reason should be a string")

		require.Equal(s.T(), expectedReason, actualReason, "Reason should match expected value")
	}
}

// Helper function to verify rate limit is unset in config
func (s *SharedServerSuite) verifyRateLimitUnset(config map[string]interface{}) {
	configData, exists := config["config"]
	require.True(s.T(), exists, "Config should exist in response")

	configMap, ok := configData.(map[string]interface{})
	require.True(s.T(), ok, "Config should be a map")

	queueRateLimit, exists := configMap["queue_rate_limit"]
	require.True(s.T(), exists, "Queue rate limit field should exist")

	rateLimitMap, ok := queueRateLimit.(map[string]any)
	require.True(s.T(), ok, "Queue rate limit should be a map")

	// If rate_limit subfield is missing, it means the rate limit is unset
	_, exists = rateLimitMap["rate_limit"]
	require.False(s.T(), exists, "Rate limit should not exist (unset)")
}

// Helper function to verify fairness key rate limit in config
func (s *SharedServerSuite) verifyFairnessKeyRateLimit(config map[string]interface{}, expectedRateLimit float64, expectedReason string) {
	configData, exists := config["config"]
	require.True(s.T(), exists, "Config should exist in response")

	configMap, ok := configData.(map[string]any)
	require.True(s.T(), ok, "Config should be a map")

	fairnessKeyRateLimit, exists := configMap["fairness_keys_rate_limit_default"]
	require.True(s.T(), exists, "Fairness key rate limit should exist")

	rateLimitMap, ok := fairnessKeyRateLimit.(map[string]interface{})
	require.True(s.T(), ok, "Fairness key rate limit should be a map")

	rateLimit, exists := rateLimitMap["rate_limit"]
	require.True(s.T(), exists, "Rate limit should exist")

	rateLimitData, ok := rateLimit.(map[string]interface{})
	require.True(s.T(), ok, "Rate limit should be a map")

	requestsPerSecond, exists := rateLimitData["requests_per_second"]
	if !exists {
		// If requests_per_second is missing, it means the rate limit is 0
		require.Equal(s.T(), float64(0), expectedRateLimit, "Fairness key rate limit should be 0 when requests_per_second is missing")
	} else {
		actualRateLimit, ok := requestsPerSecond.(float64)
		require.True(s.T(), ok, "Requests per second should be a float64")
		require.Equal(s.T(), expectedRateLimit, actualRateLimit, "Fairness key rate limit should match expected value")
	}

	if expectedReason != "" {
		metadata, exists := rateLimitMap["metadata"]
		require.True(s.T(), exists, "Metadata should exist")

		metadataMap, ok := metadata.(map[string]interface{})
		require.True(s.T(), ok, "Metadata should be a map")

		reason, exists := metadataMap["reason"]
		require.True(s.T(), exists, "Reason should exist")

		actualReason, ok := reason.(string)
		require.True(s.T(), ok, "Reason should be a string")

		require.Equal(s.T(), expectedReason, actualReason, "Reason should match expected value")
	}
}

// Helper function to verify fairness key rate limit is unset in config
func (s *SharedServerSuite) verifyFairnessKeyRateLimitUnset(config map[string]interface{}) {
	configData, exists := config["config"]
	require.True(s.T(), exists, "Config should exist in response")

	configMap, ok := configData.(map[string]interface{})
	require.True(s.T(), ok, "Config should be a map")

	fairnessKeyRateLimit, exists := configMap["fairness_keys_rate_limit_default"]
	require.True(s.T(), exists, "Fairness key rate limit field should exist")

	rateLimitMap, ok := fairnessKeyRateLimit.(map[string]interface{})
	require.True(s.T(), ok, "Fairness key rate limit should be a map")

	// If rate_limit subfield is missing, it means the rate limit is unset
	_, exists = rateLimitMap["rate_limit"]
	require.False(s.T(), exists, "Rate limit should not exist (unset)")
}

func (s *SharedServerSuite) TestTaskQueue_UpdateConfig_AllFields() {
	// Test updating all available fields
	res := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-all-fields-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--queue-rate-limit", "100",
		"--queue-rate-limit-reason", "Test queue rate limit",
		"--fairness-key-rate-limit-default", "45",
		"--fairness-key-rate-limit-reason", "Test fairness key rate limit",
	)
	require.NoError(s.T(), res.Err)

	// Verify the update by describing the task queue
	config := s.getTaskQueueConfig("test-all-fields-queue", "activity")
	s.verifyRateLimit(config, 100, "Test queue rate limit")
	s.verifyFairnessKeyRateLimit(config, 45, "Test fairness key rate limit")
}

func (s *SharedServerSuite) TestTaskQueue_UpdateConfig_ZeroRateLimit() {
	// Test setting rate limit to 0 (valid value)
	res := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-zero-rate-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--queue-rate-limit", "0",
		"--queue-rate-limit-reason", "Setting to zero",
	)
	require.NoError(s.T(), res.Err)

	// Verify the update by describing the task queue
	config := s.getTaskQueueConfig("test-zero-rate-queue", "activity")
	s.verifyRateLimit(config, 0, "Setting to zero")
}

func (s *SharedServerSuite) TestTaskQueue_UpdateConfig_UnsetQueueRateLimit() {
	// First, set a rate limit
	res := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-unset-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--queue-rate-limit", "50",
		"--queue-rate-limit-reason", "Initial setting",
	)
	require.NoError(s.T(), res.Err)

	// Verify it was set
	config := s.getTaskQueueConfig("test-unset-queue", "activity")
	s.verifyRateLimit(config, 50, "Initial setting")

	// Now unset the rate limit using sentinel value -1
	res2 := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-unset-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--queue-rate-limit", "-1",
		"--queue-rate-limit-reason", "Unsetting rate limit",
	)
	require.NoError(s.T(), res2.Err)

	// Verify it was unset
	config2 := s.getTaskQueueConfig("test-unset-queue", "activity")
	s.verifyRateLimitUnset(config2)
}

func (s *SharedServerSuite) TestTaskQueue_UpdateConfig_UnsetFairnessKeyRateLimit() {
	// First, set a fairness key rate limit
	res := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-unset-fairness-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--fairness-key-rate-limit-default", "30",
		"--fairness-key-rate-limit-reason", "Initial fairness setting",
	)
	require.NoError(s.T(), res.Err)

	// Verify it was set
	config := s.getTaskQueueConfig("test-unset-fairness-queue", "activity")
	s.verifyFairnessKeyRateLimit(config, 30, "Initial fairness setting")

	// Now unset the fairness key rate limit using sentinel value -1
	res2 := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-unset-fairness-queue",
		"--task-queue-type", "activity",
		"--namespace", "default",
		"--identity", "test-identity",
		"--fairness-key-rate-limit-default", "-1",
		"--fairness-key-rate-limit-reason", "Unsetting fairness rate limit",
	)
	require.NoError(s.T(), res2.Err)

	// Verify it was unset
	config2 := s.getTaskQueueConfig("test-unset-fairness-queue", "activity")
	s.verifyFairnessKeyRateLimitUnset(config2)
}

func (s *SharedServerSuite) TestTaskQueue_UpdateConfig_WorkflowTaskQueueError() {
	// Test that setting queue rate limit on workflow task queues is not allowed
	res := s.Execute(
		"task-queue", "update-config",
		"--address", s.Address(),
		"--task-queue", "test-workflow-queue",
		"--task-queue-type", "workflow",
		"--namespace", "default",
		"--identity", "test-identity",
		"--queue-rate-limit", "100",
	)
	require.ErrorContains(s.T(), res.Err, "setting rate limit on workflow task queues is not allowed")
}
