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
	// Nothing set -> nil payload so WCI defaults apply (min 0, max 30,
	// initial 0, utilization_target 0.8).
	p, err := gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{})
	require.NoError(t, err)
	require.Nil(t, p)

	// Any scaler flag alongside a non-GCP provider is rejected. Covers both an
	// instance-count flag and the utilization flag.
	_, err = gcpCloudRunScalerDetails("aws-lambda", gcpScalerFlags{minSet: true})
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")
	_, err = gcpCloudRunScalerDetails("aws-lambda", gcpScalerFlags{utilization: 0.5, utilizationSet: true})
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")

	// All four settings are one all-or-none group: any partial set is rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 5, minSet: true})
	require.ErrorContains(t, err, "must be set together")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{utilization: 0.5, utilizationSet: true}) // utilization alone
	require.ErrorContains(t, err, "must be set together")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{ // trio set, utilization missing
		min: 1, minSet: true, max: 3, maxSet: true, initial: 2, initialSet: true,
	})
	require.ErrorContains(t, err, "must be set together")

	// Value checks, with all four set so the group check passes first.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: -1, minSet: true, max: 3, maxSet: true, initial: 0, initialSet: true, utilization: 0.5, utilizationSet: true})
	require.ErrorContains(t, err, "cannot be negative")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 0, minSet: true, max: 0, maxSet: true, initial: 0, initialSet: true, utilization: 0.5, utilizationSet: true})
	require.ErrorContains(t, err, "--gcp-cloud-run-max-instances must be at least 1")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 5, minSet: true, max: 3, maxSet: true, initial: 4, initialSet: true, utilization: 0.5, utilizationSet: true})
	require.ErrorContains(t, err, "cannot exceed")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 2, minSet: true, max: 10, maxSet: true, initial: 15, initialSet: true, utilization: 0.5, utilizationSet: true})
	require.ErrorContains(t, err, "must be between")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 0, minSet: true, max: 10, maxSet: true, initial: 5, initialSet: true, utilization: 0, utilizationSet: true})
	require.ErrorContains(t, err, "must be greater than 0 and at most 1")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 0, minSet: true, max: 10, maxSet: true, initial: 5, initialSet: true, utilization: 1.5, utilizationSet: true})
	require.ErrorContains(t, err, "must be greater than 0 and at most 1")

	// All four set and valid -> payload decodes to the WCI rate-based keys.
	// JSON round-trips numbers as float64; WCI handles that on read.
	p, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 1, minSet: true, max: 10, maxSet: true, initial: 5, initialSet: true, utilization: 0.5, utilizationSet: true})
	require.NoError(t, err)
	require.NotNil(t, p)
	var details map[string]any
	require.NoError(t, converter.GetDefaultDataConverter().FromPayload(p, &details))
	require.Equal(t, float64(1), details[scalerKeyMinCount])
	require.Equal(t, float64(10), details[scalerKeyMaxCount])
	require.Equal(t, float64(5), details[scalerKeyInitialCount])
	require.Equal(t, float64(0.5), details[scalerKeyUtilizationTarget])
}

func TestFormatComputeConfigProto_ScalerBounds(t *testing.T) {
	// Build the scaler details the same way the run methods do.
	scalerDetails, err := gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{
		min: 0, minSet: true,
		max: 10, maxSet: true,
		initial: 5, initialSet: true,
		utilization: 0.75, utilizationSet: true,
	})
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

	// JSON/structured path surfaces min, max, initial, and utilization.
	formatted := formatComputeConfigProto(cc)
	require.NotNil(t, formatted)
	sg, ok := formatted.ScalingGroups["default"]
	require.True(t, ok)
	require.NotNil(t, sg.Scaler)
	require.Equal(t, "rate-based", sg.Scaler.Type)
	require.NotNil(t, sg.Scaler.MinInstances)
	require.NotNil(t, sg.Scaler.MaxInstances)
	require.NotNil(t, sg.Scaler.InitialInstances)
	require.NotNil(t, sg.Scaler.UtilizationTarget)
	require.Equal(t, int64(0), *sg.Scaler.MinInstances)
	require.Equal(t, int64(10), *sg.Scaler.MaxInstances)
	require.Equal(t, int64(5), *sg.Scaler.InitialInstances)
	require.Equal(t, float64(0.75), *sg.Scaler.UtilizationTarget)

	// Human-readable summary reflects the settings (min, initial, max, utilization).
	require.Equal(t, "gcp-cloud-run (min 0, initial 5, max 10, utilization 0.75)", computeConfigSummaryStr(cc))

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
	require.Nil(t, sg.Scaler.UtilizationTarget)
	require.Equal(t, "gcp-cloud-run", computeConfigSummaryStr(ccNoBounds))
}
