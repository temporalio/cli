package temporalcli

import (
	"fmt"
	"slices"
	"strconv"

	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	"go.temporal.io/sdk/converter"
)

func (c *TemporalWorkerDeploymentVersionComputeConfigModalCreateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := modalProvider(c.AppName, c.FunctionName, c.Environment, c.WorkspaceId, c.Version)
	if err != nil {
		return err
	}
	scalerDetails, err := c.NoSyncScalerOptions.details()
	if err != nil {
		return err
	}
	sg, err := noSyncScalingGroup(&c.ScalingGroupOptions, provider, scalerDetails)
	if err != nil {
		return err
	}
	return createVersionWithScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, sg, "modal")
}

func (c *TemporalWorkerDeploymentVersionComputeConfigModalUpdateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := c.provider()
	if err != nil {
		return err
	}
	scalerDetails, err := c.NoSyncScalerOptions.details()
	if err != nil {
		return err
	}
	update, err := noSyncScalingGroupUpdate(&c.ScalingGroupOptions, provider, scalerDetails)
	if err != nil {
		return err
	}
	return updateVersionScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, update, "modal")
}

// provider returns the Modal provider to set, or nil to leave it unchanged.
// The required provider flags are passed together because provider details are
// replaced as a whole; workspace and version remain optional.
func (c *TemporalWorkerDeploymentVersionComputeConfigModalUpdateCommand) provider() (*computepb.ComputeProvider, error) {
	flags := c.Command.Flags()
	providerSet := slices.ContainsFunc(
		[]string{"app-name", "function-name", "environment", "workspace-id", "version"},
		flags.Changed,
	)
	if !providerSet {
		return nil, nil
	}
	if !flags.Changed("app-name") || !flags.Changed("function-name") || !flags.Changed("environment") {
		return nil, fmt.Errorf("--app-name, --function-name, and --environment are required when changing the provider")
	}
	return modalProvider(c.AppName, c.FunctionName, c.Environment, c.WorkspaceId, c.Version)
}

// The Modal compute provider.
const modalProviderType = "modal"

// modalProvider builds the Modal compute provider from the provider flags.
func modalProvider(appName, functionName, environment, workspaceID, version string) (*computepb.ComputeProvider, error) {
	payload, err := modalProviderDetailsPayload(appName, functionName, environment, workspaceID, version)
	if err != nil {
		return nil, err
	}
	return &computepb.ComputeProvider{Type: modalProviderType, Details: payload}, nil
}

func validateModalProviderDetails(details map[string]any) error {
	for _, key := range []string{"app_name", "function_name", "environment"} {
		if v, ok := details[key].(string); !ok || v == "" {
			return fmt.Errorf("missing required Modal provider detail: %s", key)
		}
	}
	if rawVersion, ok := details["version"]; ok {
		version, ok := rawVersion.(string)
		parsed, err := strconv.Atoi(version)
		if !ok || err != nil || parsed < 0 {
			return fmt.Errorf("Modal provider detail %q must be a non-negative integer string", "version")
		}
	}
	return nil
}

// modalProviderDetailsPayload returns the encoded Payload expected by WCI's
// invoke-based Modal provider. The environment is part of the compute resource
// identity and must be explicit so provider updates cannot silently change it.
func modalProviderDetailsPayload(
	appName string,
	functionName string,
	environment string,
	workspaceID string,
	version string,
) (*commonpb.Payload, error) {
	// Map keys from temporal-auto-scaled-workers:
	// https://github.com/temporalio/temporal-auto-scaled-workers/blob/324950cf843e560ff0c6531070bb1d4e139d5000/wci/workflow/compute_provider/modal.go#L18-L24
	providerDetails := map[string]any{
		"app_name":      appName,
		"function_name": functionName,
		"environment":   environment,
	}
	if workspaceID != "" {
		providerDetails["workspace_id"] = workspaceID
	}
	if version != "" {
		providerDetails["version"] = version
	}
	if err := validateModalProviderDetails(providerDetails); err != nil {
		return nil, err
	}
	return converter.GetDefaultDataConverter().ToPayload(&providerDetails)
}
