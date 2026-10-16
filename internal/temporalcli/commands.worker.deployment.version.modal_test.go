package temporalcli_test

import (
	"github.com/google/uuid"
	enumspb "go.temporal.io/api/enums/v1"
)

func modalComputeConfigArgs(verb, address, deploymentName string, extra ...string) []string {
	return append([]string{
		"worker", "deployment", "version", "compute-config", "modal", verb,
		"--address", address,
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
		"--app-name", "payments",
		"--function-name", "worker",
		"--environment", "main",
	}, extra...)
}

func (s *SharedServerSuite) TestComputeConfig_Modal_Create() {
	fake := s.newFakeComputeConfigServer()
	fake.versionMissing = true
	deploymentName := uuid.NewString()

	res := s.Execute(modalComputeConfigArgs("create", s.Address(), deploymentName,
		"--workspace-id", "ws-123",
		"--version", "7",
		"--region-id", "modal-us",
	)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully created worker deployment version")
	s.Empty(fake.takeUpdates())
	creates := fake.takeCreates()
	s.Len(creates, 1)
	s.Equal(deploymentName, creates[0].GetDeploymentVersion().GetDeploymentName())
	s.Equal("build-17", creates[0].GetDeploymentVersion().GetBuildId())
	s.NotEmpty(creates[0].GetRequestId())
	sg := creates[0].GetComputeConfig().GetScalingGroups()["default"]
	s.NotNil(sg)
	s.Equal([]string{"modal-us"}, sg.GetRegionIds())
	s.Equal("modal", sg.GetProvider().GetType())
	s.Equal(map[string]any{
		"app_name":      "payments",
		"function_name": "worker",
		"environment":   "main",
		"workspace_id":  "ws-123",
		"version":       "7",
	}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Nil(sg.GetScaler().GetDetails())

	res = s.Execute(modalComputeConfigArgs("create", s.Address(), deploymentName)...)
	s.ErrorContains(res.Err, "error creating worker deployment version: Worker Deployment Version already exists; "+
		"to set its compute config, use `temporal worker deployment version compute-config modal update`")
	s.Len(fake.takeCreates(), 1)
}

func (s *SharedServerSuite) TestComputeConfig_Modal_Update() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(nil)
	deploymentName := uuid.NewString()

	res := s.Execute(modalComputeConfigArgs("update", s.Address(), deploymentName,
		"--scaling-group", "modal-workflows",
		"--task-queue-type", "workflow",
		"--region-id", "modal-us",
		"--scale-up-backlog-threshold", "5",
	)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully updated worker deployment version compute config")
	s.Empty(fake.takeCreates())
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	update := updates[0].GetComputeConfigScalingGroups()["modal-workflows"]
	s.NotNil(update)
	s.Equal([]string{"task_queue_types", "region_ids", "provider.type", "provider.details", "scaler.type", "scaler.details"}, update.GetUpdateMask().GetPaths())
	sg := update.GetScalingGroup()
	s.Equal([]enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_WORKFLOW}, sg.GetTaskQueueTypes())
	s.Equal([]string{"modal-us"}, sg.GetRegionIds())
	s.Equal("modal", sg.GetProvider().GetType())
	s.Equal(map[string]any{
		"app_name":      "payments",
		"function_name": "worker",
		"environment":   "main",
	}, s.decodePayloadMap(sg.GetProvider().GetDetails()))
	s.Equal("no-sync", sg.GetScaler().GetType())
	s.Equal(map[string]any{"scale_up_backlog_threshold": float64(5)}, s.decodePayloadMap(sg.GetScaler().GetDetails()))

	// Scaling settings can be changed without resending the provider fields.
	res = s.Execute(
		"worker", "deployment", "version", "compute-config", "modal", "update",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
		"--scaling-group", "modal-workflows",
		"--max-worker-lifetime-ms", "300000",
	)
	s.NoError(res.Err)
	updates = fake.takeUpdates()
	s.Len(updates, 1)
	update = updates[0].GetComputeConfigScalingGroups()["modal-workflows"]
	s.Equal([]string{"scaler.details"}, update.GetUpdateMask().GetPaths())
	s.Nil(update.GetScalingGroup().GetProvider())
	s.Equal(map[string]any{"max_worker_lifetime_ms": float64(300000)},
		s.decodePayloadMap(update.GetScalingGroup().GetScaler().GetDetails()))
	s.Equal([]string{"modal-us"}, fake.computeConfig.GetScalingGroups()["modal-workflows"].GetRegionIds())
}

func (s *SharedServerSuite) TestComputeConfig_Modal_Errors() {
	fake := s.newFakeComputeConfigServer()
	fake.setComputeConfig(nil)
	deploymentName := uuid.NewString()
	base := []string{
		"worker", "deployment", "version", "compute-config", "modal", "update",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
	}

	for _, tc := range []struct {
		name    string
		args    []string
		errText string
	}{
		{
			name:    "provider fields must be passed together",
			args:    append(append([]string{}, base...), "--app-name", "payments"),
			errText: "--app-name, --function-name, and --environment are required when changing the provider",
		},
		{
			name: "environment cannot be empty",
			args: append(append([]string{}, base...),
				"--app-name", "payments", "--function-name", "worker", "--environment", ""),
			errText: "missing required Modal provider detail: environment",
		},
		{
			name: "version must be non-negative",
			args: append(append([]string{}, base...),
				"--app-name", "payments", "--function-name", "worker", "--environment", "main", "--version", "-1"),
			errText: `Modal provider detail "version" must be a non-negative integer string`,
		},
		{
			name:    "nothing to update",
			args:    append([]string{}, base...),
			errText: "no compute configuration provided to update",
		},
	} {
		res := s.Execute(tc.args...)
		s.ErrorContains(res.Err, tc.errText, tc.name)
	}
	s.Empty(fake.takeUpdates())
}
