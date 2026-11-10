package temporalcli

import (
	"fmt"

	computepb "go.temporal.io/api/compute/v1"
	"go.temporal.io/sdk/converter"
)

func (c *TemporalWorkerDeploymentVersionComputeConfigAwsAgentcoreCreateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := awsAgentcoreProvider(c.EndpointArn, &c.AwsAssumeRoleOptions)
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
	return createVersionWithScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, sg, awsAgentcoreProviderType)
}

func (c *TemporalWorkerDeploymentVersionComputeConfigAwsAgentcoreUpdateCommand) run(cctx *CommandContext, args []string) error {
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
	return updateVersionScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, update, awsAgentcoreProviderType)
}

// provider returns the AWS AgentCore provider to set, or nil to leave it
// unchanged. The provider flags are all or nothing, because the provider's
// details are replaced as a whole.
func (c *TemporalWorkerDeploymentVersionComputeConfigAwsAgentcoreUpdateCommand) provider() (*computepb.ComputeProvider, error) {
	endpointArnSet := c.Command.Flags().Changed("endpoint-arn")
	if !endpointArnSet && !c.AwsAssumeRoleOptions.passed() {
		return nil, nil
	}
	if !endpointArnSet {
		return nil, fmt.Errorf("--endpoint-arn is required when changing the provider")
	}
	return awsAgentcoreProvider(c.EndpointArn, &c.AwsAssumeRoleOptions)
}

// The AWS Bedrock AgentCore Runtime compute provider.
const awsAgentcoreProviderType = "aws-agentcore"

// awsAgentcoreProvider builds the AWS AgentCore compute provider from the
// provider flags. Provider config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/compute_provider/aws_agentcore.go
func awsAgentcoreProvider(endpointArn string, role *AwsAssumeRoleOptions) (*computepb.ComputeProvider, error) {
	if endpointArn == "" {
		return nil, fmt.Errorf("--endpoint-arn cannot be empty")
	}
	details := map[string]any{"endpoint_arn": endpointArn}
	if err := role.setDetails(details); err != nil {
		return nil, err
	}
	payload, err := converter.GetDefaultDataConverter().ToPayload(details)
	if err != nil {
		return nil, err
	}
	return &computepb.ComputeProvider{Type: awsAgentcoreProviderType, Details: payload}, nil
}
