package temporalcli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/cli/internal/printer"
	computepb "go.temporal.io/api/compute/v1"
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestScalerTypeForProvider(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		expected  string
		expectErr bool
	}{
		{"aws-lambda is invoke-based -> no-sync", "aws-lambda", "no-sync", false},
		{"aws-agentcore is invoke-based -> no-sync", "aws-agentcore", "no-sync", false},
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
	for _, providerType := range []string{"aws-lambda", "aws-agentcore", "gcp-cloud-run"} {
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

func TestPrintWorkerDeploymentVersionInfoProto_AllFields(t *testing.T) {
	ts := func(sec int64) *timestamppb.Timestamp { return timestamppb.New(time.Unix(sec, 0).UTC()) }
	info := &deploymentpb.WorkerDeploymentVersionInfo{
		Status:               enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_DRAINING,
		DeploymentVersion:    &deploymentpb.WorkerDeploymentVersion{DeploymentName: "my-deployment", BuildId: "v1"},
		CreateTime:           ts(1000),
		RoutingChangedTime:   ts(2000),
		FirstActivationTime:  ts(3000),
		LastCurrentTime:      ts(4000),
		LastDeactivationTime: ts(5000),
		DrainageInfo: &deploymentpb.VersionDrainageInfo{
			Status:          enumspb.VERSION_DRAINAGE_STATUS_DRAINING,
			LastChangedTime: ts(6000),
			LastCheckedTime: ts(7000),
		},
		LastModifierIdentity: "some-identity",
	}

	t.Run("text", func(t *testing.T) {
		var buf bytes.Buffer
		cctx := &CommandContext{Printer: &printer.Printer{Output: &buf}}
		require.NoError(t, printWorkerDeploymentVersionInfoProto(cctx, info, nil, "Worker Deployment Version:", printVersionInfoOptions{}))
		out := buf.String()
		for _, field := range []string{
			"Status", "FirstActivationTime", "LastCurrentTime", "LastDeactivationTime",
			"DrainageStatus", "DrainageLastChangedTime", "DrainageLastCheckedTime", "LastModifierIdentity",
		} {
			require.Contains(t, out, field)
		}
		require.Contains(t, out, "draining")
		require.Contains(t, out, "some-identity")
		require.Contains(t, out, time.Unix(3000, 0).UTC().Format(time.RFC3339))
		// Never current or ramping, so these must not be rendered.
		require.NotContains(t, out, "CurrentSinceTime")
		require.NotContains(t, out, "RampingSinceTime")
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		cctx := &CommandContext{JSONOutput: true, Printer: &printer.Printer{Output: &buf, JSON: true}}
		require.NoError(t, printWorkerDeploymentVersionInfoProto(cctx, info, nil, "", printVersionInfoOptions{}))
		var out formattedWorkerDeploymentVersionInfoType
		require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
		require.Equal(t, "draining", out.Status)
		require.Equal(t, time.Unix(3000, 0).UTC(), *out.FirstActivationTime)
		require.Equal(t, time.Unix(4000, 0).UTC(), *out.LastCurrentTime)
		require.Equal(t, time.Unix(5000, 0).UTC(), *out.LastDeactivationTime)
		require.Equal(t, "some-identity", out.LastModifierIdentity)
		require.NotNil(t, out.DrainageInfo)
		require.Equal(t, "draining", out.DrainageInfo.DrainageStatus)
		require.Equal(t, time.Unix(6000, 0).UTC(), *out.DrainageInfo.LastChangedTime)
		require.Equal(t, time.Unix(7000, 0).UTC(), *out.DrainageInfo.LastCheckedTime)
		require.Nil(t, out.CurrentSinceTime)
		require.Nil(t, out.RampingSinceTime)
	})

	t.Run("json omits unset drainage info", func(t *testing.T) {
		var buf bytes.Buffer
		cctx := &CommandContext{JSONOutput: true, Printer: &printer.Printer{Output: &buf, JSON: true}}
		current := &deploymentpb.WorkerDeploymentVersionInfo{
			Status:            enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_CURRENT,
			DeploymentVersion: &deploymentpb.WorkerDeploymentVersion{DeploymentName: "my-deployment", BuildId: "v2"},
			CreateTime:        ts(1000),
		}
		require.NoError(t, printWorkerDeploymentVersionInfoProto(cctx, current, nil, "", printVersionInfoOptions{}))
		var raw map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))
		require.Equal(t, "current", raw["status"])
		require.NotContains(t, raw, "drainageInfo")
	})
}

func TestVersionStatusProtoToStr(t *testing.T) {
	require.Equal(t, "unspecified", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_UNSPECIFIED))
	require.Equal(t, "inactive", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_INACTIVE))
	require.Equal(t, "current", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_CURRENT))
	require.Equal(t, "ramping", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_RAMPING))
	require.Equal(t, "draining", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_DRAINING))
	require.Equal(t, "drained", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_DRAINED))
	require.Equal(t, "created", versionStatusProtoToStr(enumspb.WORKER_DEPLOYMENT_VERSION_STATUS_CREATED))
}
