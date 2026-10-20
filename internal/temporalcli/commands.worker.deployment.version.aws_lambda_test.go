package temporalcli_test

import (
	"github.com/google/uuid"
	computepb "go.temporal.io/api/compute/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"google.golang.org/protobuf/proto"
)

func lambdaComputeConfigArgs(verb, address, deploymentName string, extra ...string) []string {
	return append([]string{
		"worker", "deployment", "version", "compute-config", "aws-lambda", verb,
		"--address", address,
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
		"--function-arn", "arn:aws:lambda:us-east-1:123:function:payments-worker:17",
		"--assume-role-arn", "arn:aws:iam::123:role/temporal-invoke-use1",
		"--assume-role-external-id", "ext-id",
	}, extra...)
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_Update() {
	fake := s.newFakeComputeConfigServer()
	deploymentName := uuid.NewString()
	lambdaArgs := lambdaComputeConfigArgs("update", s.Address(), deploymentName,
		"--scaling-group", "aws-us-east-1-wf",
		"--task-queue-type", "activity",
		"--task-queue-type", "workflow",
		"--task-queue-type", "activity",
	)

	// The Version exists but the scaling group does not, so it is created.
	fake.setComputeConfig(nil)
	res := s.Execute(lambdaArgs...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	s.Empty(fake.takeCreates())
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	req := updates[0]
	s.Equal(deploymentName, req.GetDeploymentVersion().GetDeploymentName())
	s.Equal("build-17", req.GetDeploymentVersion().GetBuildId())
	s.Empty(req.GetRemoveComputeConfigScalingGroups())
	s.Len(req.GetComputeConfigScalingGroups(), 1)
	update := req.GetComputeConfigScalingGroups()["aws-us-east-1-wf"]
	s.NotNil(update)
	s.Equal([]string{"task_queue_types", "provider.type", "provider.details", "scaler.type", "scaler.details"}, update.GetUpdateMask().GetPaths())
	sg := update.GetScalingGroup()
	// De-duplicated and sorted.
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_WORKFLOW, enumspb.TASK_QUEUE_TYPE_ACTIVITY}, sg.GetTaskQueueTypes())
	s.Equal("aws-lambda", sg.GetProvider().GetType())
	s.Equal(map[string]any{
		"arn":              "arn:aws:lambda:us-east-1:123:function:payments-worker:17",
		"role":             "arn:aws:iam::123:role/temporal-invoke-use1",
		"role_external_id": "ext-id",
	}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Nil(sg.GetScaler().GetDetails())

	// Running the same command again sends the same update again.
	res = s.Execute(lambdaArgs...)
	s.NoError(res.Err)
	updates = fake.takeUpdates()
	s.Len(updates, 1)
	s.True(proto.Equal(update, updates[0].GetComputeConfigScalingGroups()["aws-us-east-1-wf"]))

	// A different function ARN replaces the provider.
	res = s.Execute(append(lambdaArgs, "--function-arn", "arn:aws:lambda:us-east-1:123:function:payments-worker:18")...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	updates = fake.takeUpdates()
	s.Len(updates, 1)
	update = updates[0].GetComputeConfigScalingGroups()["aws-us-east-1-wf"]
	s.Equal([]string{"task_queue_types", "provider.type", "provider.details", "scaler.type", "scaler.details"}, update.GetUpdateMask().GetPaths())
	s.Equal(
		"arn:aws:lambda:us-east-1:123:function:payments-worker:18",
		s.decodePayloadMap(update.GetScalingGroup().GetProvider().GetDetails())["arn"],
	)
	s.Empty(fake.takeCreates())
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_Create() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	deploymentName := uuid.NewString()

	res := s.Execute(lambdaComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully created worker deployment version")
	s.Empty(fake.takeUpdates())
	creates := fake.takeCreates()
	s.Len(creates, 1)
	s.Equal(deploymentName, creates[0].GetDeploymentVersion().GetDeploymentName())
	s.Equal("build-17", creates[0].GetDeploymentVersion().GetBuildId())
	s.NotEmpty(creates[0].GetRequestId())
	sgs := creates[0].GetComputeConfig().GetScalingGroups()
	s.Len(sgs, 1)
	s.Equal("aws-lambda", sgs["default"].GetProvider().GetType())
	s.Equal("no-sync", sgs["default"].GetScaler().GetType())

	// The Version now exists, so creating it again fails and points to update.
	res = s.Execute(lambdaComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.ErrorContains(res.Err, "error creating worker deployment version: Worker Deployment Version already exists; "+
		"to set its compute config, use `temporal worker deployment version compute-config aws-lambda update`")
	s.Len(fake.takeCreates(), 1)
	s.Empty(fake.takeUpdates())

	// Update sets the scaling group on the existing Version.
	res = s.Execute(lambdaComputeConfigArgs("update", s.Address(), deploymentName)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	s.Empty(fake.takeCreates())
	s.Len(fake.takeUpdates(), 1)
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_UpdateMissingVersion() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true

	res := s.Execute(lambdaComputeConfigArgs("update", s.Address(), uuid.NewString())...)
	s.ErrorContains(res.Err, "error updating worker deployment version compute config: build ID 'build-17' not found in Worker Deployment; "+
		"to create it, use `temporal worker deployment version compute-config aws-lambda create`")
	s.Empty(fake.takeCreates())
	s.Len(fake.takeUpdates(), 1)
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_DeploymentMissing() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	fake.createErr = serviceerror.NewNotFound("Worker Deployment not found")
	deploymentName := uuid.NewString()

	res := s.Execute(lambdaComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.ErrorContains(res.Err, "error creating worker deployment version: Worker Deployment not found")
	s.Len(fake.takeCreates(), 1)
	s.Empty(fake.takeUpdates())
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_DefaultScalingGroup() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	res := s.Execute(
		"worker", "deployment", "version", "compute-config", "aws-lambda", "create",
		"--address", s.Address(),
		"--deployment-name", uuid.NewString(),
		"--build-id", "build-17",
		"--function-arn", "arn:aws:lambda:us-east-1:123:function:F",
		"--skip-role-and-external-id",
	)
	s.NoError(res.Err)
	creates := fake.takeCreates()
	s.Len(creates, 1)
	sg := creates[0].GetComputeConfig().GetScalingGroups()["default"]
	s.NotNil(sg)
	s.Empty(sg.GetTaskQueueTypes())
	s.Equal(map[string]any{"arn": "arn:aws:lambda:us-east-1:123:function:F"}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_UpdateError() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(nil)
	fake.updateErr = serviceerror.NewInvalidArgument("invalid compute config")

	res := s.Execute(lambdaComputeConfigArgs("update", s.Address(), uuid.NewString())...)
	s.ErrorContains(res.Err, "error updating worker deployment version compute config: invalid compute config")
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_TaskQueueTypesAndRegions() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(nil)

	res := s.Execute(lambdaComputeConfigArgs("update", s.Address(), uuid.NewString(),
		"--scaling-group", "use1",
		"--task-queue-type", "nexus",
		"--task-queue-type", "workflow",
		"--region-id", "aws-us-east-1",
	)...)
	s.NoError(res.Err)
	s.Empty(res.Stderr.String())
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	sg := updates[0].GetComputeConfigScalingGroups()["use1"].GetScalingGroup()
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_WORKFLOW, enumspb.TASK_QUEUE_TYPE_NEXUS}, sg.GetTaskQueueTypes())
	s.Equal([]string{"aws-us-east-1"}, sg.GetRegionIds())
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_RegionIDs() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(nil)
	deploymentName := uuid.NewString()
	regionArgs := lambdaComputeConfigArgs("update", s.Address(), deploymentName,
		"--scaling-group", "aws-us-west-2",
		"--region-id", "aws-us-west-2",
		"--region-id", "aws-us-east-1",
		"--region-id", "aws-us-west-2",
	)

	// Region IDs are de-duplicated and sorted, and the mask covers them.
	res := s.Execute(regionArgs...)
	s.NoError(res.Err)
	s.Empty(res.Stderr.String())
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	update := updates[0].GetComputeConfigScalingGroups()["aws-us-west-2"]
	s.Equal([]string{"aws-us-east-1", "aws-us-west-2"}, update.GetScalingGroup().GetRegionIds())
	s.Equal([]string{"region_ids", "provider.type", "provider.details", "scaler.type", "scaler.details"}, update.GetUpdateMask().GetPaths())

	// Leaving out --region-id keeps the region IDs.
	res = s.Execute(lambdaComputeConfigArgs("update", s.Address(), deploymentName, "--scaling-group", "aws-us-west-2")...)
	s.NoError(res.Err)
	updates = fake.takeUpdates()
	s.Len(updates, 1)
	update = updates[0].GetComputeConfigScalingGroups()["aws-us-west-2"]
	s.Empty(update.GetScalingGroup().GetRegionIds())
	s.NotContains(update.GetUpdateMask().GetPaths(), "region_ids")
	s.Equal([]string{"aws-us-east-1", "aws-us-west-2"}, fake.computeConfig.GetScalingGroups()["aws-us-west-2"].GetRegionIds())
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_ScalerSettings() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	deploymentName := uuid.NewString()

	// Every scaling flag sets the scaler details key of the same name.
	res := s.Execute(lambdaComputeConfigArgs("create", s.Address(), deploymentName,
		"--scale-up-backlog-threshold", "5",
		"--scale-up-cooloff-ms", "500",
		"--max-worker-lifetime-ms", "300000",
		"--metrics-poll-interval-ms", "30000",
		"--scale-up-dispatch-rate-epsilon", "0.05",
		"--scale-up-dispatch-rate-epsilon-confirm-ms", "90000",
		"--suppress-scale-up-ms", "120000",
		"--suppress-poll-interval-ms", "60000",
	)...)
	s.NoError(res.Err)
	creates := fake.takeCreates()
	s.Len(creates, 1)
	scaler := creates[0].GetComputeConfig().GetScalingGroups()["default"].GetScaler()
	s.Equal("no-sync", scaler.GetType())
	s.Equal(map[string]any{
		"scale_up_backlog_threshold":                float64(5),
		"scale_up_cooloff_ms":                       float64(500),
		"max_worker_lifetime_ms":                    float64(300000),
		"metrics_poll_interval_ms":                  float64(30000),
		"scale_up_dispatch_rate_epsilon":            0.05,
		"scale_up_dispatch_rate_epsilon_confirm_ms": float64(90000),
		"suppress_scale_up_ms":                      float64(120000),
		"suppress_poll_interval_ms":                 float64(60000),
	}, s.decodePayloadMap(scaler.GetDetails()))

	// Only the flags that are set are sent; the others keep the server default.
	res = s.Execute(lambdaComputeConfigArgs("update", s.Address(), deploymentName,
		"--max-worker-lifetime-ms", "0",
	)...)
	s.NoError(res.Err)
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	update := updates[0].GetComputeConfigScalingGroups()["default"]
	s.Contains(update.GetUpdateMask().GetPaths(), "scaler.details")
	s.Equal(map[string]any{"max_worker_lifetime_ms": float64(0)},
		s.decodePayloadMap(update.GetScalingGroup().GetScaler().GetDetails()))
}

func (s *SharedServerSuite) TestComputeConfig_AwsLambda_PartialUpdate() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(twoScalingGroupComputeConfig())
	deploymentName := uuid.NewString()
	args := func(extra ...string) []string {
		return append([]string{
			"worker", "deployment", "version", "compute-config", "aws-lambda", "update",
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

	// Scaling flags on their own change only the scaling settings. The ones that
	// are not passed go back to the server default.
	res := s.Execute(args("--scaling-group", "aws-us-west-2-ac", "--scale-up-cooloff-ms", "500")...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	update := onlyUpdate("aws-us-west-2-ac")
	s.Equal([]string{"scaler.details"}, update.GetUpdateMask().GetPaths())
	s.Nil(update.GetScalingGroup().GetProvider())
	sg := stored("aws-us-west-2-ac")
	s.Equal("aws-lambda", sg.GetProvider().GetType())
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Equal(map[string]any{"scale_up_cooloff_ms": float64(500)}, s.decodePayloadMap(sg.GetScaler().GetDetails()))
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_ACTIVITY}, sg.GetTaskQueueTypes())
	s.Equal([]string{"aws-us-west-2"}, sg.GetRegionIds())

	// --region-id on its own changes only the region IDs.
	res = s.Execute(args("--scaling-group", "aws-us-east-1-wf", "--region-id", "aws-us-west-2", "--region-id", "aws-us-east-1")...)
	s.NoError(res.Err)
	s.Equal([]string{"region_ids"}, onlyUpdate("aws-us-east-1-wf").GetUpdateMask().GetPaths())
	sg = stored("aws-us-east-1-wf")
	s.Equal([]string{"aws-us-east-1", "aws-us-west-2"}, sg.GetRegionIds())
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_WORKFLOW}, sg.GetTaskQueueTypes())
	s.Equal("aws-lambda", sg.GetProvider().GetType())

	// --task-queue-type on its own changes only the Task Queue types.
	res = s.Execute(args("--scaling-group", "aws-us-east-1-wf", "--task-queue-type", "nexus")...)
	s.NoError(res.Err)
	s.Equal([]string{"task_queue_types"}, onlyUpdate("aws-us-east-1-wf").GetUpdateMask().GetPaths())
	sg = stored("aws-us-east-1-wf")
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_NEXUS}, sg.GetTaskQueueTypes())
	s.Equal([]string{"aws-us-east-1", "aws-us-west-2"}, sg.GetRegionIds())
	s.Equal("aws-lambda", sg.GetProvider().GetType())
	s.Equal("no-sync", sg.GetScaler().GetType())

	// The provider flags replace the provider and reset the scaling settings to
	// the scaling flags passed with them, which here are none.
	res = s.Execute(args("--scaling-group", "aws-us-west-2-ac",
		"--function-arn", "arn:aws:lambda:us-west-2:123:function:F", "--skip-role-and-external-id")...)
	s.NoError(res.Err)
	s.Equal([]string{"provider.type", "provider.details", "scaler.type", "scaler.details"},
		onlyUpdate("aws-us-west-2-ac").GetUpdateMask().GetPaths())
	sg = stored("aws-us-west-2-ac")
	s.Equal(map[string]any{"arn": "arn:aws:lambda:us-west-2:123:function:F"}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Nil(sg.GetScaler().GetDetails())
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_ACTIVITY}, sg.GetTaskQueueTypes())
	s.Equal([]string{"aws-us-west-2"}, sg.GetRegionIds())

	// The provider flags are passed together, and something must be passed.
	for _, tc := range []struct {
		name    string
		args    []string
		errText string
	}{
		{
			name:    "role without function ARN",
			args:    args("--assume-role-arn", "role", "--assume-role-external-id", "ext-id"),
			errText: "--function-arn is required when changing the provider",
		},
		{
			name:    "function ARN without role",
			args:    args("--function-arn", "arn"),
			errText: "--assume-role-arn and --assume-role-external-id are required unless --skip-role-and-external-id is passed",
		},
		{
			name:    "empty function ARN",
			args:    args("--function-arn", "", "--skip-role-and-external-id"),
			errText: "--function-arn cannot be empty",
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

func (s *SharedServerSuite) TestComputeConfig_Provider_Errors() {
	fake := s.newFakeComputeConfigServer()
	lambda := []string{
		"worker", "deployment", "version", "compute-config", "aws-lambda", "create",
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
			name:    "missing function ARN",
			args:    append(append([]string{}, lambda...), "--skip-role-and-external-id"),
			errText: `required flag(s) "function-arn" not set`,
		},
		{
			name:    "empty function ARN",
			args:    append(append([]string{}, lambda...), "--function-arn", "", "--skip-role-and-external-id"),
			errText: "--function-arn cannot be empty",
		},
		{
			name:    "missing assume role",
			args:    append(append([]string{}, lambda...), "--function-arn", "arn"),
			errText: "--assume-role-arn and --assume-role-external-id are required unless --skip-role-and-external-id is passed",
		},
		{
			name:    "missing external ID",
			args:    append(append([]string{}, lambda...), "--function-arn", "arn", "--assume-role-arn", "role"),
			errText: "--assume-role-arn and --assume-role-external-id are required unless --skip-role-and-external-id is passed",
		},
		{
			name: "assume role with skip",
			args: append(append([]string{}, lambda...),
				"--function-arn", "arn", "--assume-role-arn", "role", "--skip-role-and-external-id"),
			errText: "must be omitted when --skip-role-and-external-id is passed",
		},
		{
			name: "empty scaling group",
			args: append(append([]string{}, lambda...),
				"--function-arn", "arn", "--skip-role-and-external-id", "--scaling-group", ""),
			errText: "--scaling-group cannot be empty",
		},
		{
			name: "empty region ID",
			args: append(append([]string{}, lambda...),
				"--function-arn", "arn", "--skip-role-and-external-id", "--region-id", ""),
			errText: "--region-id cannot be empty",
		},
	} {
		res := s.Execute(tc.args...)
		s.ErrorContains(res.Err, tc.errText, tc.name)
	}
	s.Empty(fake.takeCreates())
	s.Empty(fake.takeUpdates())
}
