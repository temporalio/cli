package temporalcli

import (
	"fmt"

	computepb "go.temporal.io/api/compute/v1"
	"go.temporal.io/sdk/converter"
)

func (c *TemporalWorkerDeploymentVersionComputeConfigAwsLambdaCreateCommand) run(cctx *CommandContext, args []string) error {
	provider, err := awsLambdaProvider(c.FunctionArn, &c.AwsAssumeRoleOptions)
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
	return createVersionWithScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, sg, "aws-lambda")
}

func (c *TemporalWorkerDeploymentVersionComputeConfigAwsLambdaUpdateCommand) run(cctx *CommandContext, args []string) error {
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
	return updateVersionScalingGroup(cctx, c.Parent.Parent.Parent.Parent.Parent, c.deploymentVersion(), c.ScalingGroup, update, "aws-lambda")
}

// provider returns the AWS Lambda provider to set, or nil to leave it unchanged.
// The provider flags are all or nothing, because the provider's details are
// replaced as a whole.
func (c *TemporalWorkerDeploymentVersionComputeConfigAwsLambdaUpdateCommand) provider() (*computepb.ComputeProvider, error) {
	functionArnSet := c.Command.Flags().Changed("function-arn")
	if !functionArnSet && !c.AwsAssumeRoleOptions.passed() {
		return nil, nil
	}
	if !functionArnSet {
		return nil, fmt.Errorf("--function-arn is required when changing the provider")
	}
	return awsLambdaProvider(c.FunctionArn, &c.AwsAssumeRoleOptions)
}

// The AWS Lambda compute provider.
const awsLambdaProviderType = "aws-lambda"

// awsLambdaProvider builds the AWS Lambda compute provider from the provider
// flags. Provider config keys:
// https://github.com/temporalio/temporal-auto-scaled-workers/blob/main/wci/workflow/compute_provider/aws_lambda.go
func awsLambdaProvider(functionArn string, role *AwsAssumeRoleOptions) (*computepb.ComputeProvider, error) {
	if functionArn == "" {
		return nil, fmt.Errorf("--function-arn cannot be empty")
	}
	details := map[string]any{"arn": functionArn}
	if err := role.setDetails(details); err != nil {
		return nil, err
	}
	payload, err := converter.GetDefaultDataConverter().ToPayload(details)
	if err != nil {
		return nil, err
	}
	return &computepb.ComputeProvider{Type: awsLambdaProviderType, Details: payload}, nil
}
