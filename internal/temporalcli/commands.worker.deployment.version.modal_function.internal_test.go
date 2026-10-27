package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

func TestModalProviderDetailsPayload(t *testing.T) {
	decode := func(t *testing.T, payload *commonpb.Payload) map[string]any {
		t.Helper()
		var details map[string]any
		require.NoError(t, converter.GetDefaultDataConverter().FromPayload(payload, &details))
		return details
	}

	p, err := modalProviderDetailsPayload("payments", "worker", "main", "", "")
	require.NoError(t, err)
	details := decode(t, p)
	require.Equal(t, "payments", details["app_name"])
	require.Equal(t, "worker", details["function_name"])
	require.Equal(t, "main", details["environment"])
	require.NotContains(t, details, "workspace_id")
	require.NotContains(t, details, "version")

	p, err = modalProviderDetailsPayload("payments", "worker", "staging", "ws-123", "42")
	require.NoError(t, err)
	details = decode(t, p)
	require.Equal(t, "staging", details["environment"])
	require.Equal(t, "ws-123", details["workspace_id"])
	require.Equal(t, "42", details["version"])

	_, err = modalProviderDetailsPayload("", "worker", "main", "", "")
	require.EqualError(t, err, "--app-name cannot be empty")
	_, err = modalProviderDetailsPayload("payments", "", "main", "", "")
	require.EqualError(t, err, "--function-name cannot be empty")
	_, err = modalProviderDetailsPayload("payments", "worker", "", "", "")
	require.EqualError(t, err, "--environment cannot be empty")
	_, err = modalProviderDetailsPayload("payments", "worker", "main", "", "-1")
	require.EqualError(t, err, "--app-version must be a non-negative integer")
	_, err = modalProviderDetailsPayload("payments", "worker", "main", "", "latest")
	require.EqualError(t, err, "--app-version must be a non-negative integer")
}
