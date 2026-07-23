package temporalcli

import (
	"testing"

	"github.com/stretchr/testify/require"
	computepb "go.temporal.io/api/compute/v1"
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
	// None set -> nil payload so WCI defaults (min 0, max 30, initial 0) apply.
	// The 0 values are ignored because the *Set booleans are false.
	p, err := gcpCloudRunScalerDetails("gcp-cloud-run", 0, false, 0, false, 0, false)
	require.NoError(t, err)
	require.Nil(t, p)

	// Any use alongside a non-GCP provider is rejected, even for value 0.
	_, err = gcpCloudRunScalerDetails("aws-lambda", 0, true, 0, false, 0, false)
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")

	// A partial set is rejected; all three must be set together so the
	// min<=initial<=max relationship never depends on WCI's defaults.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 5, true, 0, false, 0, false)
	require.ErrorContains(t, err, "must be set together")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 1, true, 3, true, 0, false) // missing initial
	require.ErrorContains(t, err, "must be set together")

	// All set: negative min rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", -1, true, 3, true, 0, true)
	require.ErrorContains(t, err, "cannot be negative")

	// All set: max < 1 rejected (WCI requires max_count >= 1).
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 0, true, 0, true, 0, true)
	require.ErrorContains(t, err, "--gcp-cloud-run-max-instances must be at least 1")

	// All set: min > max rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 5, true, 3, true, 4, true)
	require.ErrorContains(t, err, "cannot exceed")

	// All set: initial outside [min, max] rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", 2, true, 10, true, 15, true)
	require.ErrorContains(t, err, "must be between")

	// All set and valid -> payload decodes to the WCI rate-based config keys.
	p, err = gcpCloudRunScalerDetails("gcp-cloud-run", 1, true, 10, true, 5, true)
	require.NoError(t, err)
	require.NotNil(t, p)
	var details map[string]any
	require.NoError(t, converter.GetDefaultDataConverter().FromPayload(p, &details))
	// JSON round-trips numbers as float64; WCI's getInt64FromMap handles that.
	require.Equal(t, float64(1), details["min_count"])
	require.Equal(t, float64(10), details["max_count"])
	require.Equal(t, float64(5), details["initial_count"])

	// min and initial of 0 are valid (WCI's floor is 0) when all three are set.
	p, err = gcpCloudRunScalerDetails("gcp-cloud-run", 0, true, 5, true, 0, true)
	require.NoError(t, err)
	details = nil
	require.NoError(t, converter.GetDefaultDataConverter().FromPayload(p, &details))
	require.Equal(t, float64(0), details["min_count"])
	require.Equal(t, float64(5), details["max_count"])
	require.Equal(t, float64(0), details["initial_count"])
}

func TestFormatComputeConfigProto_ScalerBounds(t *testing.T) {
	// Build the scaler details the same way the run methods do.
	scalerDetails, err := gcpCloudRunScalerDetails("gcp-cloud-run", 0, true, 10, true, 5, true)
	require.NoError(t, err)
	require.NotNil(t, scalerDetails)

	cc := &computepb.ComputeConfig{
		ScalingGroups: map[string]*computepb.ComputeConfigScalingGroup{
			"default": {
				Provider: &computepb.ComputeProvider{Type: "gcp-cloud-run"},
				Scaler:   &computepb.ComputeScaler{Type: "rate-based", Details: scalerDetails},
			},
		},
	}

	// JSON/structured path surfaces min, max, and initial on the scaler.
	formatted := formatComputeConfigProto(cc)
	require.NotNil(t, formatted)
	sg, ok := formatted.ScalingGroups["default"]
	require.True(t, ok)
	require.NotNil(t, sg.Scaler)
	require.Equal(t, "rate-based", sg.Scaler.Type)
	require.NotNil(t, sg.Scaler.MinInstances)
	require.NotNil(t, sg.Scaler.MaxInstances)
	require.NotNil(t, sg.Scaler.InitialInstances)
	require.Equal(t, int64(0), *sg.Scaler.MinInstances)
	require.Equal(t, int64(10), *sg.Scaler.MaxInstances)
	require.Equal(t, int64(5), *sg.Scaler.InitialInstances)

	// Human-readable summary reflects the settings (ordered min, initial, max).
	require.Equal(t, "gcp-cloud-run (min 0, initial 5, max 10)", computeConfigSummaryStr(cc))

	// Without scaler details, the settings are nil and the summary is just the
	// provider (guards against printing zeroed-out values).
	ccNoBounds := &computepb.ComputeConfig{
		ScalingGroups: map[string]*computepb.ComputeConfigScalingGroup{
			"default": {
				Provider: &computepb.ComputeProvider{Type: "gcp-cloud-run"},
				Scaler:   &computepb.ComputeScaler{Type: "rate-based"},
			},
		},
	}
	formatted = formatComputeConfigProto(ccNoBounds)
	sg = formatted.ScalingGroups["default"]
	require.NotNil(t, sg.Scaler)
	require.Nil(t, sg.Scaler.MinInstances)
	require.Nil(t, sg.Scaler.MaxInstances)
	require.Nil(t, sg.Scaler.InitialInstances)
	require.Equal(t, "gcp-cloud-run", computeConfigSummaryStr(ccNoBounds))
}
