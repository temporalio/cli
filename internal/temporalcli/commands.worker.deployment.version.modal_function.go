package temporalcli

import (
	"fmt"
	"slices"
	"strconv"

	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	"go.temporal.io/sdk/converter"
)

func (c *TemporalWorkerDeploymentVersionComputeConfigModalFunctionCreateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := modalProvider(c.AppName, c.FunctionName, c.Environment, c.WorkspaceId, c.AppVersion)
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
	return createVersionWithScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, sg, "modal-function")
}

func (c *TemporalWorkerDeploymentVersionComputeConfigModalFunctionUpdateCommand) run(cctx *CommandContext, args []string) error {
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
	return updateVersionScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, update, "modal-function")
}

// provider returns the Modal Function provider to set, or nil to leave it unchanged.
// The required provider flags are passed together because provider details are
// replaced as a whole; workspace and application version remain optional.
func (c *TemporalWorkerDeploymentVersionComputeConfigModalFunctionUpdateCommand) provider() (*computepb.ComputeProvider, error) {
	flags := c.Command.Flags()
	providerSet := slices.ContainsFunc(
		[]string{"app-name", "function-name", "environment", "workspace-id", "app-version"},
		flags.Changed,
	)
	if !providerSet {
		return nil, nil
	}
	if !flags.Changed("app-name") || !flags.Changed("function-name") || !flags.Changed("environment") {
		return nil, fmt.Errorf("--app-name, --function-name, and --environment are required when changing the provider")
	}
	return modalProvider(c.AppName, c.FunctionName, c.Environment, c.WorkspaceId, c.AppVersion)
}

// The Modal Function compute provider.
const modalProviderType = "modal-function"

// modalProvider builds the Modal Function compute provider from the provider
// flags. Provider config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/compute_provider/modal_function.go
func modalProvider(appName, functionName, environment, workspaceID, version string) (*computepb.ComputeProvider, error) {
	payload, err := modalProviderDetailsPayload(appName, functionName, environment, workspaceID, version)
	if err != nil {
		return nil, err
	}
	return &computepb.ComputeProvider{Type: modalProviderType, Details: payload}, nil
}

func validateModalProviderDetails(details map[string]any) error {
	for _, field := range []struct {
		key  string
		flag string
	}{
		{key: "app_name", flag: "app-name"},
		{key: "function_name", flag: "function-name"},
		{key: "environment", flag: "environment"},
	} {
		if v, ok := details[field.key].(string); !ok || v == "" {
			return fmt.Errorf("--%s cannot be empty", field.flag)
		}
	}
	if rawVersion, ok := details["version"]; ok {
		version, ok := rawVersion.(string)
		parsed, err := strconv.Atoi(version)
		if !ok || err != nil || parsed < 0 {
			return fmt.Errorf("--app-version must be a non-negative integer")
		}
	}
	return nil
}

// modalProviderDetailsPayload returns the encoded Payload expected by WCI's
// invoke-based Modal Function provider. The environment is part of the compute
// resource identity and must be explicit so provider updates cannot silently
// change it.
func modalProviderDetailsPayload(
	appName string,
	functionName string,
	environment string,
	workspaceID string,
	version string,
) (*commonpb.Payload, error) {
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
