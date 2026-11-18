package temporalcli_test

import (
	"github.com/google/uuid"
	computepb "go.temporal.io/api/compute/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
)

func cloudRunComputeConfigArgs(verb, address, deploymentName string, extra ...string) []string {
	return append([]string{
		"worker", "deployment", "version", "compute-config", "gcp-cloud-run", verb,
		"--address", address,
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
		"--project", "payments-prod",
		"--worker-pool-region", "us-central1",
		"--worker-pool", "payments-worker",
		"--service-account", "temporal@payments-prod.iam.gserviceaccount.com",
	}, extra...)
}

func cloudRunScalerArgs(minInstances, maxInstances, initialInstances, utilizationTarget, stabilizationMs, scaleUpCooldownMs, scaleDownCooldownMs string) []string {
	return []string{
		"--min-instances", minInstances,
		"--max-instances", maxInstances,
		"--initial-instances", initialInstances,
		"--utilization-target", utilizationTarget,
		"--scale-down-stabilization-ms", stabilizationMs,
		"--scale-up-cooldown-ms", scaleUpCooldownMs,
		"--scale-down-cooldown-ms", scaleDownCooldownMs,
	}
}

const cloudRunScalerFlagsTogetherErr = "--min-instances, --max-instances, --initial-instances, --utilization-target, " +
	"--scale-down-stabilization-ms, --scale-up-cooldown-ms, and --scale-down-cooldown-ms must be set together"

func (s *SharedServerSuite) TestComputeConfig_GcpCloudRun_Create() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	deploymentName := uuid.NewString()

	res := s.Execute(cloudRunComputeConfigArgs("create", s.Address(), deploymentName)...)
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
	s.Equal("gcp-cloud-run", sg.GetProvider().GetType())
	s.Equal(map[string]any{
		"project":         "payments-prod",
		"region":          "us-central1",
		"worker_pool":     "payments-worker",
		"service_account": "temporal@payments-prod.iam.gserviceaccount.com",
	}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
	// Without scaler flags the server defaults apply.
	s.Equal("rate-based", sg.GetScaler().GetType())
	s.Nil(sg.GetScaler().GetDetails())

	// The Version now exists, so creating it again fails and points to update.
	res = s.Execute(cloudRunComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.ErrorContains(res.Err, "error creating worker deployment version: Worker Deployment Version already exists; "+
		"to set its compute config, use `temporal worker deployment version compute-config gcp-cloud-run update`")
	s.Len(fake.takeCreates(), 1)
	s.Empty(fake.takeUpdates())
}

func (s *SharedServerSuite) TestComputeConfig_GcpCloudRun_ScalerSettings() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true

	res := s.Execute(append(cloudRunComputeConfigArgs("create", s.Address(), uuid.NewString(), "--scaling-group", "gcp-us-central1"),
		cloudRunScalerArgs("1", "3", "2", "0.75", "300000", "5000", "30000")...)...)
	s.NoError(res.Err)
	creates := fake.takeCreates()
	s.Len(creates, 1)
	scaler := creates[0].GetComputeConfig().GetScalingGroups()["gcp-us-central1"].GetScaler()
	s.Equal("rate-based", scaler.GetType())
	s.Equal(map[string]any{
		"min_count":              float64(1),
		"max_count":              float64(3),
		"initial_count":          float64(2),
		"utilization_target":     0.75,
		"no_sync_quiet_ms":       float64(300000),
		"scale_up_cooldown_ms":   float64(5000),
		"scale_down_cooldown_ms": float64(30000),
	}, s.decodePayloadMap(scaler.GetDetails()))
}

func (s *SharedServerSuite) TestComputeConfig_GcpCloudRun_UpdateMissingVersion() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true

	res := s.Execute(cloudRunComputeConfigArgs("update", s.Address(), uuid.NewString())...)
	s.ErrorContains(res.Err, "error updating worker deployment version compute config: build ID 'build-17' not found in Worker Deployment; "+
		"to create it, use `temporal worker deployment version compute-config gcp-cloud-run create`")
	s.Empty(fake.takeCreates())
	s.Len(fake.takeUpdates(), 1)
}

func (s *SharedServerSuite) TestComputeConfig_GcpCloudRun_PartialUpdate() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(twoScalingGroupComputeConfig())
	deploymentName := uuid.NewString()
	args := func(extra ...string) []string {
		return append([]string{
			"worker", "deployment", "version", "compute-config", "gcp-cloud-run", "update",
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

	scalerSettings := map[string]any{
		"min_count":              float64(0),
		"max_count":              float64(5),
		"initial_count":          float64(0),
		"utilization_target":     0.5,
		"no_sync_quiet_ms":       float64(0),
		"scale_up_cooldown_ms":   float64(0),
		"scale_down_cooldown_ms": float64(0),
	}

	// The provider flags replace the provider and the scaler type, and the
	// scaler flags passed with them replace the scaler settings.
	res := s.Execute(append(cloudRunComputeConfigArgs("update", s.Address(), deploymentName, "--scaling-group", "aws-us-west-2-ac"),
		cloudRunScalerArgs("0", "5", "0", "0.5", "0", "0", "0")...)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	s.Equal([]string{"provider.type", "provider.details", "scaler.type", "scaler.details"},
		onlyUpdate("aws-us-west-2-ac").GetUpdateMask().GetPaths())
	sg := stored("aws-us-west-2-ac")
	s.Equal("gcp-cloud-run", sg.GetProvider().GetType())
	s.Equal("rate-based", sg.GetScaler().GetType())
	s.Equal(scalerSettings, s.decodePayloadMap(sg.GetScaler().GetDetails()))
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_ACTIVITY}, sg.GetTaskQueueTypes())

	// Without the scaler flags, the provider flags keep the scaler settings.
	res = s.Execute(cloudRunComputeConfigArgs("update", s.Address(), deploymentName,
		"--scaling-group", "aws-us-west-2-ac", "--worker-pool", "payments-worker-v2")...)
	s.NoError(res.Err)
	s.Equal([]string{"provider.type", "provider.details", "scaler.type"},
		onlyUpdate("aws-us-west-2-ac").GetUpdateMask().GetPaths())
	sg = stored("aws-us-west-2-ac")
	s.Equal("payments-worker-v2", s.decodePayloadMap(sg.GetProvider().GetDetails())["worker_pool"])
	s.Equal("rate-based", sg.GetScaler().GetType())
	s.Equal(scalerSettings, s.decodePayloadMap(sg.GetScaler().GetDetails()))

	// Scaler flags on their own change only the scaler details.
	res = s.Execute(args(append([]string{"--scaling-group", "aws-us-west-2-ac"}, cloudRunScalerArgs("1", "3", "1", "0.75", "300000", "2000", "60000")...)...)...)
	s.NoError(res.Err)
	update := onlyUpdate("aws-us-west-2-ac")
	s.Equal([]string{"scaler.details"}, update.GetUpdateMask().GetPaths())
	s.Nil(update.GetScalingGroup().GetProvider())
	sg = stored("aws-us-west-2-ac")
	s.Equal("gcp-cloud-run", sg.GetProvider().GetType())
	s.Equal("rate-based", sg.GetScaler().GetType())
	s.Equal(map[string]any{
		"min_count":              float64(1),
		"max_count":              float64(3),
		"initial_count":          float64(1),
		"utilization_target":     0.75,
		"no_sync_quiet_ms":       float64(300000),
		"scale_up_cooldown_ms":   float64(2000),
		"scale_down_cooldown_ms": float64(60000),
	}, s.decodePayloadMap(sg.GetScaler().GetDetails()))

	// The provider flags are passed together, the scaler flags are passed
	// together, and something must be passed.
	for _, tc := range []struct {
		name    string
		args    []string
		errText string
	}{
		{
			name:    "worker pool without the other provider flags",
			args:    args("--worker-pool", "payments-worker"),
			errText: "--project is required when changing the provider",
		},
		{
			name:    "missing service account",
			args:    args("--project", "p", "--worker-pool-region", "r", "--worker-pool", "w"),
			errText: "--service-account is required when changing the provider",
		},
		{
			name:    "empty worker pool",
			args:    args("--project", "p", "--worker-pool-region", "r", "--worker-pool", "", "--service-account", "sa"),
			errText: "--worker-pool cannot be empty",
		},
		{
			name:    "some scaler flags",
			args:    args("--min-instances", "1", "--max-instances", "3"),
			errText: cloudRunScalerFlagsTogetherErr,
		},
		{
			name:    "cooldown alone",
			args:    args("--scale-down-cooldown-ms", "45000"),
			errText: cloudRunScalerFlagsTogetherErr,
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

func (s *SharedServerSuite) TestComputeConfig_GcpCloudRun_Errors() {
	fake := s.newFakeComputeConfigServer()
	create := func(extra ...string) []string {
		return cloudRunComputeConfigArgs("create", s.Address(), uuid.NewString(), extra...)
	}

	for _, tc := range []struct {
		name    string
		args    []string
		errText string
	}{
		{
			name: "missing provider flag",
			args: []string{
				"worker", "deployment", "version", "compute-config", "gcp-cloud-run", "create",
				"--address", s.Address(),
				"--deployment-name", uuid.NewString(),
				"--build-id", "build-17",
				"--project", "p", "--worker-pool-region", "r", "--worker-pool", "w",
			},
			errText: `required flag(s) "service-account" not set`,
		},
		{
			name:    "empty project",
			args:    create("--project", ""),
			errText: "--project cannot be empty",
		},
		{
			name:    "empty worker pool region",
			args:    create("--worker-pool-region", ""),
			errText: "--worker-pool-region cannot be empty",
		},
		{
			name:    "empty service account",
			args:    create("--service-account", ""),
			errText: "--service-account cannot be empty",
		},
		{
			name:    "some scaler flags",
			args:    create("--utilization-target", "0.5"),
			errText: cloudRunScalerFlagsTogetherErr,
		},
		{
			name:    "size flags without cooldowns",
			args:    create(cloudRunScalerArgs("1", "3", "1", "0.75", "0", "0", "0")[:10]...),
			errText: cloudRunScalerFlagsTogetherErr,
		},
		{
			name:    "negative min",
			args:    create(cloudRunScalerArgs("-1", "3", "1", "0.75", "0", "2000", "60000")...),
			errText: "--min-instances cannot be negative",
		},
		{
			name:    "zero max",
			args:    create(cloudRunScalerArgs("0", "0", "0", "0.75", "0", "2000", "60000")...),
			errText: "--max-instances must be at least 1",
		},
		{
			name:    "min above max",
			args:    create(cloudRunScalerArgs("4", "3", "3", "0.75", "0", "2000", "60000")...),
			errText: "--min-instances cannot exceed --max-instances",
		},
		{
			name:    "initial above max",
			args:    create(cloudRunScalerArgs("1", "3", "4", "0.75", "0", "2000", "60000")...),
			errText: "--initial-instances must be between --min-instances and --max-instances",
		},
		{
			name:    "zero utilization",
			args:    create(cloudRunScalerArgs("1", "3", "1", "0", "0", "2000", "60000")...),
			errText: "--utilization-target must be greater than 0 and at most 1",
		},
		{
			name:    "utilization above 1",
			args:    create(cloudRunScalerArgs("1", "3", "1", "1.5", "0", "2000", "60000")...),
			errText: "--utilization-target must be greater than 0 and at most 1",
		},
		{
			name:    "negative stabilization",
			args:    create(cloudRunScalerArgs("1", "3", "1", "0.75", "-1", "2000", "60000")...),
			errText: "--scale-down-stabilization-ms cannot be negative",
		},
		{
			name:    "negative scale-up cooldown",
			args:    create(cloudRunScalerArgs("1", "3", "1", "0.75", "0", "-1", "0")...),
			errText: "--scale-up-cooldown-ms cannot be negative",
		},
		{
			name:    "negative scale-down cooldown",
			args:    create(cloudRunScalerArgs("1", "3", "1", "0.75", "0", "0", "-1")...),
			errText: "--scale-down-cooldown-ms cannot be negative",
		},
		{
			name:    "empty scaling group",
			args:    create("--scaling-group", ""),
			errText: "--scaling-group cannot be empty",
		},
	} {
		res := s.Execute(tc.args...)
		s.ErrorContains(res.Err, tc.errText, tc.name)
	}
	s.Empty(fake.takeCreates())

	// A server error is returned as is.
	fake.setComputeConfig(nil)
	fake.updateErr = serviceerror.NewInvalidArgument("invalid compute config")
	res := s.Execute(cloudRunComputeConfigArgs("update", s.Address(), uuid.NewString())...)
	s.ErrorContains(res.Err, "error updating worker deployment version compute config: invalid compute config")
}
