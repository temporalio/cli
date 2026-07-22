package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
)

func TestScalerTypeForProvider(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		expected  string
		expectErr bool
	}{
		{"aws-lambda is invoke-based -> no-sync", "aws-lambda", "no-sync", false},
		{"gcp-cloud-run is worker-set-based -> rate-based", "gcp-cloud-run", "rate-based", false},
		{"unknown provider errors", "azure-container-apps", "", true},
		{"empty provider errors", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scaler, err := scalerTypeForProvider(tt.provider)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, scaler)
		})
	}
}

// Every provider type computeProviderConfig can emit must have an explicit
// scaler mapping; a missing entry makes scalerTypeForProvider error before the
// request is sent, so this guards against forgetting to map a newly-added provider.
func TestScalerTypeByProviderCoversAllProviders(t *testing.T) {
	for _, providerType := range []string{"aws-lambda", "gcp-cloud-run"} {
		_, ok := scalerTypeByProvider[providerType]
		require.Truef(t, ok, "provider %q has no scaler mapping", providerType)
	}
}

func TestGCPCloudRunScalerDetails(t *testing.T) {
	// Neither flag set -> nil payload so WCI defaults (min 0, max 30) apply.
	// The 0 values are ignored because minSet/maxSet are false.
	p, err := gcpCloudRunScalerDetails("gcp-cloud-run", 0, false, 0, false)
	require.NoError(t, err)
	require.Nil(t, p)

	// Any use alongside a non-GCP provider is rejected, even for value 0.
	_, err = gcpCloudRunScalerDetails("aws-lambda", 0, true, 0, false)
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")

	// Only one of the two set -> rejected; they must be set together so the
	// min<=max relationship never depends on WCI's default for the other bound.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 5, true, 0, false)
	require.ErrorContains(t, err, "must be set together")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 0, false, 5, true)
	require.ErrorContains(t, err, "must be set together")

	// Both set: negative min rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", -1, true, 3, true)
	require.ErrorContains(t, err, "cannot be negative")

	// Both set: max < 1 rejected (WCI requires max_count >= 1).
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 0, true, 0, true)
	require.ErrorContains(t, err, "--gcp-cloud-run-max-instances must be at least 1")

	// Both set: min > max rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 5, true, 3, true)
	require.ErrorContains(t, err, "cannot exceed")

	// Both set and valid -> payload decodes to the WCI rate-based config keys.
	p, err = gcpCloudRunScalerDetails("gcp-cloud-run", 1, true, 3, true)
	require.NoError(t, err)
	require.NotNil(t, p)
	var details map[string]any
	require.NoError(t, converter.GetDefaultDataConverter().FromPayload(p, &details))
	// JSON round-trips numbers as float64; WCI's getInt64FromMap handles that.
	require.Equal(t, float64(1), details["min_count"])
	require.Equal(t, float64(3), details["max_count"])

	// min of 0 is valid (WCI's min_count floor is 0) when max is also set.
	p, err = gcpCloudRunScalerDetails("gcp-cloud-run", 0, true, 5, true)
	require.NoError(t, err)
	details = nil
	require.NoError(t, converter.GetDefaultDataConverter().FromPayload(p, &details))
	require.Equal(t, float64(0), details["min_count"])
	require.Equal(t, float64(5), details["max_count"])
}
