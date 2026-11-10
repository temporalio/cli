package temporalcli

import (
	"fmt"
	"slices"

	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func (c *TemporalWorkerDeploymentVersionComputeConfigGcpCloudRunCreateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := gcpCloudRunProvider(c.Project, c.WorkerPoolRegion, c.WorkerPool, c.ServiceAccount)
	if err != nil {
		return err
	}
	scalerDetails, err := c.RateBasedScalerOptions.details()
	if err != nil {
		return err
	}
	sg, err := c.ScalingGroupOptions.scalingGroup()
	if err != nil {
		return err
	}
	sg.Provider = provider
	sg.Scaler = &computepb.ComputeScaler{Type: rateBasedScalerType, Details: scalerDetails}
	return createVersionWithScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, sg, gcpCloudRunProviderType)
}

func (c *TemporalWorkerDeploymentVersionComputeConfigGcpCloudRunUpdateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := c.provider()
	if err != nil {
		return err
	}
	scalerDetails, err := c.RateBasedScalerOptions.details()
	if err != nil {
		return err
	}
	sg, paths, err := c.ScalingGroupOptions.update()
	if err != nil {
		return err
	}
	switch {
	case provider != nil:
		// Provider (re)configuration: rebuild the provider and scaler type.
		// As in update-version-compute-config, the scaler's settings are only
		// replaced when the scaler flags are passed; otherwise they are kept.
		sg.Provider = provider
		sg.Scaler = &computepb.ComputeScaler{Type: rateBasedScalerType, Details: scalerDetails}
		paths = append(paths, "provider.type", "provider.details", "scaler.type")
		if scalerDetails != nil {
			paths = append(paths, "scaler.details")
		}
	case scalerDetails != nil:
		// Scaler-only update: change scaler.details without touching the
		// provider or scaler type.
		sg.Scaler = &computepb.ComputeScaler{Details: scalerDetails}
		paths = append(paths, "scaler.details")
	}
	if len(paths) == 0 {
		return fmt.Errorf("no compute configuration provided to update")
	}
	update := &computepb.ComputeConfigScalingGroupUpdate{
		ScalingGroup: sg,
		UpdateMask:   &fieldmaskpb.FieldMask{Paths: paths},
	}
	return updateVersionScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, update, gcpCloudRunProviderType)
}

// provider returns the GCP Cloud Run provider to set, or nil to leave it
// unchanged. The provider flags are all or nothing, because the provider's
// details are replaced as a whole.
func (c *TemporalWorkerDeploymentVersionComputeConfigGcpCloudRunUpdateCommand) provider() (*computepb.ComputeProvider, error) {
	flags := []string{"project", "worker-pool-region", "worker-pool", "service-account"}
	if !slices.ContainsFunc(flags, c.Command.Flags().Changed) {
		return nil, nil
	}
	for _, flag := range flags {
		if !c.Command.Flags().Changed(flag) {
			return nil, fmt.Errorf("--%s is required when changing the provider", flag)
		}
	}
	return gcpCloudRunProvider(c.Project, c.WorkerPoolRegion, c.WorkerPool, c.ServiceAccount)
}

// The GCP Cloud Run compute provider.
const gcpCloudRunProviderType = "gcp-cloud-run"

// gcpCloudRunProvider builds the GCP Cloud Run compute provider from the
// provider flags. Provider config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/compute_provider/gcp_cloudrun.go
func gcpCloudRunProvider(project, region, workerPool, serviceAccount string) (*computepb.ComputeProvider, error) {
	if project == "" {
		return nil, fmt.Errorf("--project cannot be empty")
	}
	if region == "" {
		return nil, fmt.Errorf("--worker-pool-region cannot be empty")
	}
	if workerPool == "" {
		return nil, fmt.Errorf("--worker-pool cannot be empty")
	}
	if serviceAccount == "" {
		return nil, fmt.Errorf("--service-account cannot be empty")
	}
	payload, err := converter.GetDefaultDataConverter().ToPayload(map[string]any{
		"project":         project,
		"region":          region,
		"worker_pool":     workerPool,
		"service_account": serviceAccount,
	})
	if err != nil {
		return nil, err
	}
	return &computepb.ComputeProvider{Type: gcpCloudRunProviderType, Details: payload}, nil
}

// details builds the details of the worker controller's rate-based scaler, used
// by compute providers that size a set of Workers, such as GCP Cloud Run, from
// the scaler flags. The flags are all or nothing, because the scaler's details
// are replaced as a whole and the server checks them against each other. It
// returns nil when none are set, so the server defaults apply. Config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/scaling_algorithm/rate_based.go
func (o *RateBasedScalerOptions) details() (*commonpb.Payload, error) {
	flags := []string{
		"min-instances", "max-instances", "initial-instances", "utilization-target",
		"scale-down-stabilization-ms", "scale-up-cooldown-ms", "scale-down-cooldown-ms",
	}
	if !slices.ContainsFunc(flags, o.FlagSet.Changed) {
		return nil, nil
	}
	for _, flag := range flags {
		if !o.FlagSet.Changed(flag) {
			return nil, fmt.Errorf("--min-instances, --max-instances, --initial-instances, --utilization-target, " +
				"--scale-down-stabilization-ms, --scale-up-cooldown-ms, and --scale-down-cooldown-ms must be set together")
		}
	}
	if o.MinInstances < 0 {
		return nil, fmt.Errorf("--min-instances cannot be negative")
	}
	if o.MaxInstances < 1 {
		return nil, fmt.Errorf("--max-instances must be at least 1")
	}
	if o.MinInstances > o.MaxInstances {
		return nil, fmt.Errorf("--min-instances cannot exceed --max-instances")
	}
	if o.InitialInstances < o.MinInstances || o.InitialInstances > o.MaxInstances {
		return nil, fmt.Errorf("--initial-instances must be between --min-instances and --max-instances")
	}
	// The (0, 1] range is intrinsic to utilization_target's meaning, not a
	// default that could drift, so mirroring the check here is safe.
	if o.UtilizationTarget <= 0 || o.UtilizationTarget > 1 {
		return nil, fmt.Errorf("--utilization-target must be greater than 0 and at most 1")
	}
	if o.ScaleDownStabilizationMs < 0 {
		return nil, fmt.Errorf("--scale-down-stabilization-ms cannot be negative")
	}
	if o.ScaleUpCooldownMs < 0 {
		return nil, fmt.Errorf("--scale-up-cooldown-ms cannot be negative")
	}
	if o.ScaleDownCooldownMs < 0 {
		return nil, fmt.Errorf("--scale-down-cooldown-ms cannot be negative")
	}
	return converter.GetDefaultDataConverter().ToPayload(map[string]any{
		scalerKeyMinCount:          o.MinInstances,
		scalerKeyMaxCount:          o.MaxInstances,
		scalerKeyInitialCount:      o.InitialInstances,
		scalerKeyUtilizationTarget: o.UtilizationTarget,
		scalerKeyNoSyncQuietMs:     o.ScaleDownStabilizationMs,
		"scale_up_cooldown_ms":     o.ScaleUpCooldownMs,
		"scale_down_cooldown_ms":   o.ScaleDownCooldownMs,
	})
}
