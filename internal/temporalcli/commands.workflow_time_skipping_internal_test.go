package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
)

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
	for _, tc := range []struct {
		name     string
		response *workflowservice.PollWorkflowExecutionTimeSkippingResponse
	}{
		{name: "nil response"},
		{name: "empty response", response: &workflowservice.PollWorkflowExecutionTimeSkippingResponse{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, reason, err := workflowTimeSkippingPollResult(tc.response)
			require.ErrorContains(t, err, "unknown fast-forward polling result")
			require.Empty(t, status)
			require.Empty(t, reason)
		})
	}
}
