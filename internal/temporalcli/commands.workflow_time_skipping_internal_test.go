package temporalcli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
)

func TestWorkflowTimeSkippingConfig(t *testing.T) {
	config, err := workflowTimeSkippingConfig(0, false, "", false, 0)
	require.NoError(t, err)
	require.True(t, config.Enabled)
	require.Nil(t, config.FastForwardConfig)
	require.False(t, config.DisablePropagation)

	config, err = workflowTimeSkippingConfig(5*time.Hour, true, "YourFastForwardId", true, 17)
	require.NoError(t, err)
	require.True(t, config.DisablePropagation)
	require.Equal(t, int32(17), config.MaxSessionSkipCount)
	require.Equal(t, "YourFastForwardId", config.FastForwardConfig.ID)
	require.Equal(t, 5*time.Hour, config.FastForwardConfig.Duration)

	_, err = workflowTimeSkippingConfig(0, false, "YourFastForwardId", false, 0)
	require.EqualError(t, err, "--fast-forward-id requires --fast-forward-duration")
	_, err = workflowTimeSkippingConfig(0, true, "", false, 0)
	require.EqualError(t, err, "--fast-forward-duration must be greater than zero")
}

func TestWorkflowTimeSkippingPollResult(t *testing.T) {
	for _, tc := range []struct {
		result enumspb.FastForwardPollingResult
		reason string
		status string
	}{
		{enumspb.FAST_FORWARD_POLLING_RESULT_FAST_FORWARD_COMPLETED, "", "completed"},
		{enumspb.FAST_FORWARD_POLLING_RESULT_POLL_TIMEOUT, "", "poll timed out"},
		{enumspb.FAST_FORWARD_POLLING_RESULT_FAST_FORWARD_FAILED, "time skipping disabled", "failed"},
	} {
		status, reason, err := workflowTimeSkippingPollResult(&workflowservice.PollWorkflowExecutionTimeSkippingResponse{
			FastForwardPollingResult: tc.result,
			FailedReason:             tc.reason,
		})
		require.NoError(t, err)
		require.Equal(t, tc.status, status)
		require.Equal(t, tc.reason, reason)
	}
}
