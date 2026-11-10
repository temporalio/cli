package temporalcli_test

import (
	"github.com/google/uuid"
	computepb "go.temporal.io/api/compute/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
)

func agentcoreComputeConfigArgs(verb, address, deploymentName string, extra ...string) []string {
	return append([]string{
		"worker", "deployment", "version", "compute-config", "aws-agentcore", verb,
		"--address", address,
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
		"--endpoint-arn", "arn:aws:bedrock-agentcore:us-east-1:123:runtime/payments-worker/runtime-endpoint/v17",
		"--assume-role-arn", "arn:aws:iam::123:role/temporal-invoke-use1",
		"--assume-role-external-id", "ext-id",
	}, extra...)
}

func (s *SharedServerSuite) TestComputeConfig_AwsAgentcore_Create() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	deploymentName := uuid.NewString()

	res := s.Execute(agentcoreComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully created worker deployment version")
	s.Empty(fake.takeUpdates())
	creates := fake.takeCreates()
	s.Len(creates, 1)
	s.Equal(deploymentName, creates[0].GetDeploymentVersion().GetDeploymentName())
	s.Equal("build-17", creates[0].GetDeploymentVersion().GetBuildId())
	sgs := creates[0].GetComputeConfig().GetScalingGroups()
	s.Len(sgs, 1)
	sg := sgs["default"]
	s.Empty(sg.GetTaskQueueTypes())
	s.Equal("aws-agentcore", sg.GetProvider().GetType())
	s.Equal(map[string]any{
		"endpoint_arn":     "arn:aws:bedrock-agentcore:us-east-1:123:runtime/payments-worker/runtime-endpoint/v17",
		"role":             "arn:aws:iam::123:role/temporal-invoke-use1",
		"role_external_id": "ext-id",
	}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Nil(sg.GetScaler().GetDetails())

	// The Version now exists, so creating it again fails and points to update.
	res = s.Execute(agentcoreComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.ErrorContains(res.Err, "error creating worker deployment version: Worker Deployment Version already exists; "+
		"to set its compute config, use `temporal worker deployment version compute-config aws-agentcore update`")
	s.Len(fake.takeCreates(), 1)
	s.Empty(fake.takeUpdates())
}

func (s *SharedServerSuite) TestComputeConfig_AwsAgentcore_Update() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(nil)
	deploymentName := uuid.NewString()

	// The Version exists but the scaling group does not, so it is created.
	res := s.Execute(agentcoreComputeConfigArgs("update", s.Address(), deploymentName,
		"--scaling-group", "aws-us-east-1-ac",
		"--scale-up-cooloff-ms", "500",
	)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	s.Empty(fake.takeCreates())
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	update := updates[0].GetComputeConfigScalingGroups()["aws-us-east-1-ac"]
	s.NotNil(update)
	s.Equal([]string{"provider.type", "provider.details", "scaler.type", "scaler.details"}, update.GetUpdateMask().GetPaths())
	sg := update.GetScalingGroup()
	s.Equal("aws-agentcore", sg.GetProvider().GetType())
	s.Equal(
		"arn:aws:bedrock-agentcore:us-east-1:123:runtime/payments-worker/runtime-endpoint/v17",
		s.decodePayloadMap(sg.GetProvider().GetDetails())["endpoint_arn"],
	)
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Equal(map[string]any{"scale_up_cooloff_ms": float64(500)}, s.decodePayloadMap(sg.GetScaler().GetDetails()))
}

func (s *SharedServerSuite) TestComputeConfig_AwsAgentcore_UpdateMissingVersion() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true

	res := s.Execute(agentcoreComputeConfigArgs("update", s.Address(), uuid.NewString())...)
	s.ErrorContains(res.Err, "error updating worker deployment version compute config: build ID 'build-17' not found in Worker Deployment; "+
		"to create it, use `temporal worker deployment version compute-config aws-agentcore create`")
	s.Empty(fake.takeCreates())
	s.Len(fake.takeUpdates(), 1)
}

func (s *SharedServerSuite) TestComputeConfig_AwsAgentcore_DefaultScalingGroup() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	res := s.Execute(
		"worker", "deployment", "version", "compute-config", "aws-agentcore", "create",
		"--address", s.Address(),
		"--deployment-name", uuid.NewString(),
		"--build-id", "build-17",
		"--endpoint-arn", "arn:aws:bedrock-agentcore:us-east-1:123:runtime/R/runtime-endpoint/E",
		"--skip-role-and-external-id",
	)
	s.NoError(res.Err)
	creates := fake.takeCreates()
	s.Len(creates, 1)
	sg := creates[0].GetComputeConfig().GetScalingGroups()["default"]
	s.NotNil(sg)
	s.Equal(map[string]any{"endpoint_arn": "arn:aws:bedrock-agentcore:us-east-1:123:runtime/R/runtime-endpoint/E"},
		s.decodePayloadMap(sg.GetProvider().GetDetails()))
}

func (s *SharedServerSuite) TestComputeConfig_AwsAgentcore_PartialUpdate() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(twoScalingGroupComputeConfig())
	deploymentName := uuid.NewString()
	args := func(extra ...string) []string {
		return append([]string{
			"worker", "deployment", "version", "compute-config", "aws-agentcore", "update",
			"--address", s.Address(),
			"--deployment-name", deploymentName,
			"--build-id", "build-17",
		}, extra...)
	}
	onlyUpdate := func(key string) *computepb.ComputeConfigScalingGroupUpdate {
		updates := fake.takeUpdates()
		s.Len(updates, 1)
		s.Len(updates[0].GetComputeConfigScalingGroups(), 1)
		return updates[0].GetComputeConfigScalingGroups()[key]
	}
	stored := func(key string) *computepb.ComputeConfigScalingGroup {
		return fake.computeConfig.GetScalingGroups()[key]
	}

	// Scaler flags on their own change only the scaler details.
	res := s.Execute(args("--scaling-group", "aws-us-west-2-ac", "--scale-up-cooloff-ms", "500")...)
	s.NoError(res.Err)
	update := onlyUpdate("aws-us-west-2-ac")
	s.Equal([]string{"scaler.details"}, update.GetUpdateMask().GetPaths())
	s.Nil(update.GetScalingGroup().GetProvider())
	s.Equal(map[string]any{"scale_up_cooloff_ms": float64(500)}, s.decodePayloadMap(stored("aws-us-west-2-ac").GetScaler().GetDetails()))

	// The provider flags replace the provider and reset the scaler to the
	// scaler flags passed with them, which here are none.
	res = s.Execute(args("--scaling-group", "aws-us-west-2-ac",
		"--endpoint-arn", "arn:aws:bedrock-agentcore:us-west-2:123:runtime/R/runtime-endpoint/E", "--skip-role-and-external-id")...)
	s.NoError(res.Err)
	s.Equal([]string{"provider.type", "provider.details", "scaler.type", "scaler.details"},
		onlyUpdate("aws-us-west-2-ac").GetUpdateMask().GetPaths())
	sg := stored("aws-us-west-2-ac")
	s.Equal("aws-agentcore", sg.GetProvider().GetType())
	s.Equal(map[string]any{"endpoint_arn": "arn:aws:bedrock-agentcore:us-west-2:123:runtime/R/runtime-endpoint/E"},
		s.decodePayloadMap(sg.GetProvider().GetDetails()))
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Nil(sg.GetScaler().GetDetails())
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_ACTIVITY}, sg.GetTaskQueueTypes())

	// The provider flags are passed together, and something must be passed.
	for _, tc := range []struct {
		name    string
		args    []string
		errText string
	}{
		{
			name:    "role without endpoint ARN",
			args:    args("--assume-role-arn", "role", "--assume-role-external-id", "ext-id"),
			errText: "--endpoint-arn is required when changing the provider",
		},
		{
			name:    "skip role without endpoint ARN",
			args:    args("--skip-role-and-external-id"),
			errText: "--endpoint-arn is required when changing the provider",
		},
		{
			name:    "endpoint ARN without role",
			args:    args("--endpoint-arn", "arn"),
			errText: "--assume-role-arn and --assume-role-external-id are required unless --skip-role-and-external-id is passed",
		},
		{
			name:    "empty endpoint ARN",
			args:    args("--endpoint-arn", "", "--skip-role-and-external-id"),
			errText: "--endpoint-arn cannot be empty",
		},
		{
			name:    "nothing to update",
			args:    args("--scaling-group", "aws-us-east-1-wf"),
			errText: "no compute configuration provided to update",
		},
	} {
		res := s.Execute(tc.args...)
		s.EqualError(res.Err, tc.errText, tc.name)
	}
	s.Empty(fake.takeUpdates())
}

func (s *SharedServerSuite) TestComputeConfig_AwsAgentcore_Errors() {
	fake := s.newFakeComputeConfigServer()
	create := []string{
		"worker", "deployment", "version", "compute-config", "aws-agentcore", "create",
		"--address", s.Address(),
		"--deployment-name", uuid.NewString(),
		"--build-id", "build-17",
	}

	for _, tc := range []struct {
		name    string
		args    []string
		errText string
	}{
		{
			name:    "missing endpoint ARN",
			args:    append(append([]string{}, create...), "--skip-role-and-external-id"),
			errText: `required flag(s) "endpoint-arn" not set`,
		},
		{
			name:    "empty endpoint ARN",
			args:    append(append([]string{}, create...), "--endpoint-arn", "", "--skip-role-and-external-id"),
			errText: "--endpoint-arn cannot be empty",
		},
		{
			name:    "missing external ID",
			args:    append(append([]string{}, create...), "--endpoint-arn", "arn", "--assume-role-arn", "role"),
			errText: "--assume-role-arn and --assume-role-external-id are required unless --skip-role-and-external-id is passed",
		},
		{
			name: "assume role with skip",
			args: append(append([]string{}, create...),
				"--endpoint-arn", "arn", "--assume-role-arn", "role", "--skip-role-and-external-id"),
			errText: "must be omitted when --skip-role-and-external-id is passed",
		},
		{
			name: "empty scaling group",
			args: append(append([]string{}, create...),
				"--endpoint-arn", "arn", "--skip-role-and-external-id", "--scaling-group", ""),
			errText: "--scaling-group cannot be empty",
		},
	} {
		res := s.Execute(tc.args...)
		s.ErrorContains(res.Err, tc.errText, tc.name)
	}

	// A server error is returned as is.
	fake.setComputeConfig(nil)
	fake.updateErr = serviceerror.NewInvalidArgument("invalid compute config")
	res := s.Execute(agentcoreComputeConfigArgs("update", s.Address(), uuid.NewString())...)
	s.ErrorContains(res.Err, "error updating worker deployment version compute config: invalid compute config")

	s.Empty(fake.takeCreates())
}
