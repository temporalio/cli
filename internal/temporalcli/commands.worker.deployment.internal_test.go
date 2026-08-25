package temporalcli

import (
	"testing"
	"time"

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
	// A fully-set, valid group; each case clones this and overrides one field so
	// the all-or-none check passes and the case isolates a single value check.
	valid := func() gcpScalerFlags {
		return gcpScalerFlags{
			min: 1, minSet: true,
			max: 10, maxSet: true,
			initial: 5, initialSet: true,
			utilization: 0.5, utilizationSet: true,
			scaleDownStabilization: 90 * time.Second, scaleDownStabilizationSet: true,
		}
	}

	// Nothing set -> nil payload so WCI defaults apply (min 0, max 30,
	// initial 0, utilization_target 0.8, no_sync_quiet_ms 90000).
	p, err := gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{})
	require.NoError(t, err)
	require.Nil(t, p)

	// Any scaler flag alongside a non-GCP provider is rejected. Covers an
	// instance-count flag, the utilization flag, and the no-sync flag.
	_, err = gcpCloudRunScalerDetails("aws-lambda", gcpScalerFlags{minSet: true})
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")
	_, err = gcpCloudRunScalerDetails("aws-lambda", gcpScalerFlags{utilization: 0.5, utilizationSet: true})
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")
	_, err = gcpCloudRunScalerDetails("aws-lambda", gcpScalerFlags{scaleDownStabilization: time.Second, scaleDownStabilizationSet: true})
	require.ErrorContains(t, err, "only valid with --gcp-cloud-run-worker-pool")

	// All five settings are one all-or-none group: any partial set is rejected.
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{min: 5, minSet: true})
	require.ErrorContains(t, err, "must be set together")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{utilization: 0.5, utilizationSet: true}) // utilization alone
	require.ErrorContains(t, err, "must be set together")
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{scaleDownStabilization: time.Second, scaleDownStabilizationSet: true}) // scale-down-stabilization-duration alone
	require.ErrorContains(t, err, "must be set together")
	// The four instance/utilization flags without scale-down-stabilization-duration are also
	// rejected: scale-down-stabilization-duration is part of the same all-or-none group.
	missingStabilization := valid()
	missingStabilization.scaleDownStabilization, missingStabilization.scaleDownStabilizationSet = 0, false
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", missingStabilization)
	require.ErrorContains(t, err, "must be set together")

	// Value checks, with the whole group set so the group check passes first.
	neg := valid()
	neg.min, neg.initial = -1, 0
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", neg)
	require.ErrorContains(t, err, "--gcp-cloud-run-min-instances cannot be negative")

	maxTooLow := valid()
	maxTooLow.min, maxTooLow.max, maxTooLow.initial = 0, 0, 0
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", maxTooLow)
	require.ErrorContains(t, err, "--gcp-cloud-run-max-instances must be at least 1")

	minGtMax := valid()
	minGtMax.min, minGtMax.max, minGtMax.initial = 5, 3, 4
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", minGtMax)
	require.ErrorContains(t, err, "cannot exceed")

	initialOOR := valid()
	initialOOR.min, initialOOR.max, initialOOR.initial = 2, 10, 15
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", initialOOR)
	require.ErrorContains(t, err, "must be between")

	utilZero := valid()
	utilZero.utilization = 0
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", utilZero)
	require.ErrorContains(t, err, "must be greater than 0 and at most 1")

	utilHigh := valid()
	utilHigh.utilization = 1.5
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", utilHigh)
	require.ErrorContains(t, err, "must be greater than 0 and at most 1")

	negStabilization := valid()
	negStabilization.scaleDownStabilization = -time.Second
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", negStabilization)
	require.ErrorContains(t, err, "--gcp-cloud-run-scale-down-stabilization-duration cannot be negative")

	// A negative sub-millisecond value must be caught before Milliseconds()
	// truncates it toward zero (which would send 0 and silently disable the wait).
	negSubMs := valid()
	negSubMs.scaleDownStabilization = -time.Microsecond
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", negSubMs)
	require.ErrorContains(t, err, "--gcp-cloud-run-scale-down-stabilization-duration cannot be negative")

	// A positive sub-millisecond value is rejected rather than silently rounded.
	subMs := valid()
	subMs.scaleDownStabilization = 500 * time.Microsecond
	_, err = gcpCloudRunScalerDetails("gcp-cloud-run", subMs)
	require.ErrorContains(t, err, "--gcp-cloud-run-scale-down-stabilization-duration must be a whole number of milliseconds")

	// Whole group set and valid -> payload decodes to the WCI rate-based keys.
	// JSON round-trips numbers as float64; WCI handles that on read.
	ok := valid()
	ok.scaleDownStabilization = 120 * time.Second
	p, err = gcpCloudRunScalerDetails("gcp-cloud-run", ok)
	require.NoError(t, err)
	require.NotNil(t, p)
	var details map[string]any
	require.NoError(t, converter.GetDefaultDataConverter().FromPayload(p, &details))
	require.Equal(t, float64(1), details[scalerKeyMinCount])
	require.Equal(t, float64(10), details[scalerKeyMaxCount])
	require.Equal(t, float64(5), details[scalerKeyInitialCount])
	require.Equal(t, float64(0.5), details[scalerKeyUtilizationTarget])
	require.Equal(t, float64(120000), details[scalerKeyNoSyncQuietMs])
}

func TestFormatComputeConfigProto_ScalerBounds(t *testing.T) {
	// Build the scaler details the same way the run methods do.
	scalerDetails, err := gcpCloudRunScalerDetails("gcp-cloud-run", gcpScalerFlags{
		min: 0, minSet: true,
		max: 10, maxSet: true,
		initial: 5, initialSet: true,
		utilization: 0.75, utilizationSet: true,
		scaleDownStabilization: 120 * time.Second, scaleDownStabilizationSet: true,
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

	// JSON/structured path surfaces min, max, initial, utilization, and scale-down-stabilization.
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
	require.Equal(t, "2m 0s", sg.Scaler.ScaleDownStabilization)

	// Human-readable summary reflects the settings (min, initial, max, utilization, scale-down-stabilization).
	require.Equal(t, "gcp-cloud-run (min 0, initial 5, max 10, utilization 0.75, scale-down-stabilization 2m 0s)", computeConfigSummaryStr(cc))

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
	require.Empty(t, sg.Scaler.ScaleDownStabilization)
	require.Equal(t, "gcp-cloud-run", computeConfigSummaryStr(ccNoBounds))
}
