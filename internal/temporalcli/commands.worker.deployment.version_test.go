package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	commonpb "go.temporal.io/api/common/v1"
	computepb "go.temporal.io/api/compute/v1"
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// fakeComputeConfigServer intercepts the describe, create, and update compute
// config RPCs: describe returns the configured version, and create and update
// requests are captured instead of being sent. Updates are applied the way the
// worker controller applies them: a new scaling group is stored whole, and an
// existing one only gets the fields in the update mask. This lets the tests
// cover multiple scaling groups without real cloud resources behind the server.
type fakeComputeConfigServer struct {
	mu sync.Mutex
	// versionMissing makes describe return NotFound.
	versionMissing bool
	computeConfig  *computepb.ComputeConfig
	// createErr is returned by create. Otherwise create fails with
	// AlreadyExists if the version exists, and makes the version exist with the
	// created compute config if not.
	createErr error
	// updateErr is returned by update.
	updateErr error
	// describeTemplate, if set, is the describe response before the compute
	// config is filled in.
	describeTemplate *workflowservice.DescribeWorkerDeploymentVersionResponse
	creates          []*workflowservice.CreateWorkerDeploymentVersionRequest
	updates          []*workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest
}

func (f *fakeComputeConfigServer) setComputeConfig(cc *computepb.ComputeConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.versionMissing = false
	f.computeConfig = cc
}

func (f *fakeComputeConfigServer) takeCreates() []*workflowservice.CreateWorkerDeploymentVersionRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	creates := f.creates
	f.creates = nil
	return creates
}

func (f *fakeComputeConfigServer) takeUpdates() []*workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	updates := f.updates
	f.updates = nil
	return updates
}

func (f *fakeComputeConfigServer) store(key string, sg *computepb.ComputeConfigScalingGroup) {
	sg = proto.Clone(sg).(*computepb.ComputeConfigScalingGroup)
	if f.computeConfig == nil {
		f.computeConfig = &computepb.ComputeConfig{}
	}
	if f.computeConfig.ScalingGroups == nil {
		f.computeConfig.ScalingGroups = map[string]*computepb.ComputeConfigScalingGroup{}
	}
	f.computeConfig.ScalingGroups[key] = sg
}

// update applies one scaling group update like the worker controller does.
func (f *fakeComputeConfigServer) update(key string, update *computepb.ComputeConfigScalingGroupUpdate) error {
	existing, ok := f.computeConfig.GetScalingGroups()[key]
	if !ok {
		f.store(key, update.GetScalingGroup())
		return nil
	}
	merged := proto.Clone(existing).(*computepb.ComputeConfigScalingGroup)
	src := update.GetScalingGroup()
	for _, path := range update.GetUpdateMask().GetPaths() {
		switch path {
		case "provider.type", "provider.details":
			if merged.Provider == nil {
				merged.Provider = &computepb.ComputeProvider{}
			}
			if path == "provider.type" {
				merged.Provider.Type = src.GetProvider().GetType()
			} else {
				merged.Provider.Details = src.GetProvider().GetDetails()
			}
		case "scaler.type", "scaler.details":
			if src.GetScaler() == nil {
				return serviceerror.NewInvalidArgumentf("%s set in field mask, but not provided", path)
			}
			if merged.Scaler == nil {
				merged.Scaler = &computepb.ComputeScaler{}
			}
			if path == "scaler.type" {
				merged.Scaler.Type = src.GetScaler().GetType()
			} else {
				merged.Scaler.Details = src.GetScaler().GetDetails()
			}
		default:
			return serviceerror.NewInvalidArgumentf("unsupported field mask path: %q", path)
		}
	}
	f.store(key, merged)
	return nil
}

func (f *fakeComputeConfigServer) interceptor(
	ctx context.Context, method string, req, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	switch r := req.(type) {
	case *workflowservice.DescribeWorkerDeploymentVersionRequest:
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.versionMissing {
			return serviceerror.ToStatus(serviceerror.NewNotFound("Worker Deployment Version not found")).Err()
		}
		resp := &workflowservice.DescribeWorkerDeploymentVersionResponse{}
		if f.describeTemplate != nil {
			resp = proto.Clone(f.describeTemplate).(*workflowservice.DescribeWorkerDeploymentVersionResponse)
		}
		if resp.WorkerDeploymentVersionInfo == nil {
			resp.WorkerDeploymentVersionInfo = &deploymentpb.WorkerDeploymentVersionInfo{}
		}
		resp.WorkerDeploymentVersionInfo.DeploymentVersion = r.GetDeploymentVersion()
		if f.computeConfig != nil {
			resp.WorkerDeploymentVersionInfo.ComputeConfig = proto.Clone(f.computeConfig).(*computepb.ComputeConfig)
		}
		proto.Merge(reply.(proto.Message), resp)
		return nil
	case *workflowservice.CreateWorkerDeploymentVersionRequest:
		f.mu.Lock()
		defer f.mu.Unlock()
		f.creates = append(f.creates, r)
		if f.createErr != nil {
			return serviceerror.ToStatus(f.createErr).Err()
		}
		if !f.versionMissing {
			return serviceerror.ToStatus(serviceerror.NewAlreadyExists("Worker Deployment Version already exists")).Err()
		}
		f.versionMissing = false
		f.computeConfig = nil
		for key, sg := range r.GetComputeConfig().GetScalingGroups() {
			f.store(key, sg)
		}
		return nil
	case *workflowservice.UpdateWorkerDeploymentVersionComputeConfigRequest:
		f.mu.Lock()
		defer f.mu.Unlock()
		f.updates = append(f.updates, r)
		if f.updateErr != nil {
			return serviceerror.ToStatus(f.updateErr).Err()
		}
		if f.versionMissing {
			return serviceerror.ToStatus(serviceerror.NewNotFound("build ID 'build-17' not found in Worker Deployment")).Err()
		}
		for key, update := range r.GetComputeConfigScalingGroups() {
			if err := f.update(key, update); err != nil {
				return serviceerror.ToStatus(err).Err()
			}
		}
		for _, key := range r.GetRemoveComputeConfigScalingGroups() {
			delete(f.computeConfig.GetScalingGroups(), key)
		}
		return nil
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (s *SharedServerSuite) newFakeComputeConfigServer() *fakeComputeConfigServer {
	f := &fakeComputeConfigServer{}
	s.CommandHarness.Options.AdditionalClientGRPCDialOptions = append(
		s.CommandHarness.Options.AdditionalClientGRPCDialOptions,
		grpc.WithChainUnaryInterceptor(f.interceptor),
	)
	return f
}

func (s *SharedServerSuite) decodePayloadMap(p *commonpb.Payload) map[string]any {
	var m map[string]any
	s.NoError(converter.GetDefaultDataConverter().FromPayload(p, &m))
	return m
}

func twoScalingGroupComputeConfig() *computepb.ComputeConfig {
	scalerDetails, _ := converter.GetDefaultDataConverter().ToPayload(map[string]any{
		"scale_up_backlog_threshold":     2,
		"max_worker_lifetime_ms":         300000,
		"scale_up_dispatch_rate_epsilon": 0.05,
	})
	return &computepb.ComputeConfig{
		ScalingGroups: map[string]*computepb.ComputeConfigScalingGroup{
			"aws-us-west-2-ac": {
				TaskQueueTypes: []enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_ACTIVITY},
				Provider:       &computepb.ComputeProvider{Type: "aws-lambda"},
				Scaler:         &computepb.ComputeScaler{Type: "no-sync", Details: scalerDetails},
			},
			"aws-us-east-1-wf": {
				TaskQueueTypes: []enumspb.TaskQueueType{enumspb.TASK_QUEUE_TYPE_WORKFLOW},
				Provider:       &computepb.ComputeProvider{Type: "aws-lambda"},
				Scaler:         &computepb.ComputeScaler{Type: "no-sync"},
			},
		},
	}
}

func (s *SharedServerSuite) TestComputeConfig_Describe() {
	fake := s.newFakeComputeConfigServer()
	describeArgs := []string{
		"worker", "deployment", "version", "compute-config", "describe",
		"--address", s.Address(),
		"--deployment-name", uuid.NewString(),
		"--build-id", "build-17",
	}

	// No compute config.
	res := s.Execute(describeArgs...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Worker deployment version has no compute config")
	res = s.Execute(append(describeArgs, "-o", "json")...)
	s.NoError(res.Err)
	s.JSONEq(`{"scalingGroups": {}}`, res.Stdout.String())

	fake.setComputeConfig(twoScalingGroupComputeConfig())

	// One card per scaling group, sorted by key. The Scaler line lists the
	// settings that are set.
	res = s.Execute(describeArgs...)
	s.NoError(res.Err)
	_, groups, found := strings.Cut(res.Stdout.String(), "Compute Config Scaling Groups:")
	s.True(found)
	cards := strings.Split(strings.TrimSpace(groups), "\n\n")
	s.Len(cards, 2)
	// Like describe-version, the Task Queue types are not shown.
	s.NotContains(groups, "TaskQueueTypes")
	s.ContainsOnSameLine(cards[0], "ScalingGroup", "aws-us-east-1-wf")
	s.ContainsOnSameLine(cards[0], "Provider", "aws-lambda")
	s.ContainsOnSameLine(cards[0], "Scaler", "no-sync")
	s.NotContains(cards[0], "(")
	s.ContainsOnSameLine(cards[1], "ScalingGroup", "aws-us-west-2-ac")
	s.ContainsOnSameLine(cards[1], "Scaler",
		"no-sync (scale_up_backlog_threshold 2, max_worker_lifetime_ms 300000, scale_up_dispatch_rate_epsilon 0.05)")

	res = s.Execute(append(describeArgs, "-o", "json")...)
	s.NoError(res.Err)
	var described struct {
		ScalingGroups map[string]struct {
			Provider struct {
				Type string `json:"type"`
			} `json:"provider"`
			Scaler struct {
				Type                       string   `json:"type"`
				ScaleUpBacklogThreshold    *int64   `json:"scaleUpBacklogThreshold"`
				MaxWorkerLifetimeMs        *int64   `json:"maxWorkerLifetimeMs"`
				ScaleUpDispatchRateEpsilon *float64 `json:"scaleUpDispatchRateEpsilon"`
				ScaleUpCooloffMs           *int64   `json:"scaleUpCooloffMs"`
			} `json:"scaler"`
		} `json:"scalingGroups"`
	}
	s.NotContains(res.Stdout.String(), "taskQueueTypes")
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &described))
	s.Len(described.ScalingGroups, 2)
	wf := described.ScalingGroups["aws-us-east-1-wf"]
	s.Equal("aws-lambda", wf.Provider.Type)
	s.Equal("no-sync", wf.Scaler.Type)
	ac := described.ScalingGroups["aws-us-west-2-ac"]
	s.Equal("aws-lambda", ac.Provider.Type)
	s.Equal("no-sync", ac.Scaler.Type)
	// The no-sync settings that are set are shown; the others are left out.
	s.Equal(int64(2), *ac.Scaler.ScaleUpBacklogThreshold)
	s.Equal(int64(300000), *ac.Scaler.MaxWorkerLifetimeMs)
	s.Equal(0.05, *ac.Scaler.ScaleUpDispatchRateEpsilon)
	s.Nil(ac.Scaler.ScaleUpCooloffMs)
	s.Nil(wf.Scaler.ScaleUpBacklogThreshold)
	s.Empty(fake.takeUpdates())
}

func (s *SharedServerSuite) TestComputeConfig_Remove() {
	fake := s.newFakeComputeConfigServer()
	deploymentName := uuid.NewString()
	removeArgs := []string{
		"worker", "deployment", "version", "compute-config", "remove",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "build-17",
	}

	// With no scaling groups there is nothing to confirm or remove.
	res := s.Execute(removeArgs...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Worker deployment version has no compute config to remove")
	s.Empty(fake.takeUpdates())

	// Only the named scaling group is removed, without a prompt.
	fake.setComputeConfig(twoScalingGroupComputeConfig())
	res = s.Execute(append(removeArgs, "--scaling-group", "aws-us-west-2-ac")...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully removed worker deployment version compute config")
	updates := fake.takeUpdates()
	s.Len(updates, 1)
	s.Equal([]string{"aws-us-west-2-ac"}, updates[0].GetRemoveComputeConfigScalingGroups())
	s.Empty(updates[0].GetComputeConfigScalingGroups())
	s.Equal([]string{"aws-us-east-1-wf"}, slices.Collect(maps.Keys(fake.computeConfig.GetScalingGroups())))

	// Removing a scaling group that does not exist fails without an update.
	res = s.Execute(append(removeArgs, "--scaling-group", "aws-us-west-2-ac")...)
	s.ErrorContains(res.Err, `scaling group "aws-us-west-2-ac" not found; run `+
		"`temporal worker deployment version compute-config describe`")
	s.Empty(fake.takeUpdates())

	// Without --scaling-group, every scaling group is removed after the prompt
	// lists them.
	fake.setComputeConfig(twoScalingGroupComputeConfig())
	s.CommandHarness.Stdin.WriteString("n\n")
	res = s.Execute(removeArgs...)
	s.ErrorContains(res.Err, "user denied confirmation")
	s.Contains(res.Stdout.String(), fmt.Sprintf(
		`Remove all 2 scaling groups (aws-us-east-1-wf, aws-us-west-2-ac) from Worker Deployment Version %q with Build ID "build-17"? y/N`,
		deploymentName))
	s.Empty(fake.takeUpdates())

	s.CommandHarness.Stdin.WriteString("y\n")
	res = s.Execute(removeArgs...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully removed worker deployment version compute config")
	updates = fake.takeUpdates()
	s.Len(updates, 1)
	s.Equal([]string{"aws-us-east-1-wf", "aws-us-west-2-ac"}, updates[0].GetRemoveComputeConfigScalingGroups())
	s.Empty(updates[0].GetComputeConfigScalingGroups())

	// With a single scaling group, the prompt names it.
	single := twoScalingGroupComputeConfig()
	delete(single.ScalingGroups, "aws-us-west-2-ac")
	fake.setComputeConfig(single)
	s.CommandHarness.Stdin.WriteString("n\n")
	res = s.Execute(removeArgs...)
	s.ErrorContains(res.Err, "user denied confirmation")
	s.Contains(res.Stdout.String(), fmt.Sprintf(
		`Remove scaling group "aws-us-east-1-wf" from Worker Deployment Version %q with Build ID "build-17"? y/N`,
		deploymentName))
	s.Empty(fake.takeUpdates())

	// --yes skips the prompt, and JSON output requires it.
	fake.setComputeConfig(twoScalingGroupComputeConfig())
	res = s.Execute(append(removeArgs, "-o", "json")...)
	s.ErrorContains(res.Err, "must bypass prompts when using JSON output")
	s.Empty(fake.takeUpdates())
	res = s.Execute(append(removeArgs, "--yes")...)
	s.NoError(res.Err)
	s.Len(fake.takeUpdates(), 1)
}

func (s *SharedServerSuite) TestComputeConfig_Remove_EmptyScalingGroup() {
	fake := s.newFakeComputeConfigServer()
	res := s.Execute(
		"worker", "deployment", "version", "compute-config", "remove",
		"--address", s.Address(),
		"--deployment-name", uuid.NewString(),
		"--build-id", "build-17",
		"--scaling-group", "",
	)
	s.ErrorContains(res.Err, "--scaling-group cannot be empty")

	// --yes only skips the prompt for removing all scaling groups.
	res = s.Execute(
		"worker", "deployment", "version", "compute-config", "remove",
		"--address", s.Address(),
		"--deployment-name", uuid.NewString(),
		"--build-id", "build-17",
		"--scaling-group", "default",
		"--yes",
	)
	s.EqualError(res.Err, "--yes cannot be used with --scaling-group")
	s.Empty(fake.takeUpdates())
}

// TestComputeConfig_NonExistentDeployment runs against the real server, where
// neither the Worker Deployment nor the Version exists.
func (s *SharedServerSuite) TestComputeConfig_NonExistentDeployment() {
	deploymentName := uuid.NewString()
	res := s.Execute(
		"worker", "deployment", "version", "compute-config", "aws-lambda", "create",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "non-existing",
		"--function-arn", "arn:aws:lambda:us-east-1:123:function:F",
		"--skip-role-and-external-id",
	)
	s.ErrorContains(res.Err, "error creating worker deployment version")
	s.ErrorContains(res.Err, deploymentName)

	res = s.Execute(
		"worker", "deployment", "version", "compute-config", "aws-lambda", "update",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "non-existing",
		"--function-arn", "arn:aws:lambda:us-east-1:123:function:F",
		"--skip-role-and-external-id",
	)
	s.ErrorContains(res.Err, "to create it, use `temporal worker deployment version compute-config aws-lambda create`")

	res = s.Execute(
		"worker", "deployment", "version", "compute-config", "remove",
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "non-existing",
	)
	s.ErrorContains(res.Err, "error describing worker deployment version")
}

// TestVersion_Describe_MatchesDescribeVersion checks that version describe prints
// exactly what describe-version prints.
func (s *SharedServerSuite) TestVersion_Describe_MatchesDescribeVersion() {
	fake := s.newFakeComputeConfigServer()
	at := func(minute int) *timestamppb.Timestamp {
		return timestamppb.New(time.Date(2026, 1, 2, 3, minute, 0, 0, time.UTC))
	}
	stats := func(backlog int64) *taskqueuepb.TaskQueueStats {
		return &taskqueuepb.TaskQueueStats{
			ApproximateBacklogCount: backlog,
			ApproximateBacklogAge:   durationpb.New(90 * time.Second),
			TasksAddRate:            2.5,
			TasksDispatchRate:       1.5,
		}
	}
	owner, err := converter.GetDefaultDataConverter().ToPayload("payments-team")
	s.NoError(err)
	fake.describeTemplate = &workflowservice.DescribeWorkerDeploymentVersionResponse{
		WorkerDeploymentVersionInfo: &deploymentpb.WorkerDeploymentVersionInfo{
			CreateTime:         at(1),
			RoutingChangedTime: at(2),
			CurrentSinceTime:   at(3),
			RampingSinceTime:   at(4),
			RampPercentage:     25,
			DrainageInfo: &deploymentpb.VersionDrainageInfo{
				Status:          enumspb.VERSION_DRAINAGE_STATUS_DRAINING,
				LastChangedTime: at(5),
				LastCheckedTime: at(6),
			},
			Metadata: &deploymentpb.VersionMetadata{Entries: map[string]*commonpb.Payload{"owner": owner}},
		},
		VersionTaskQueues: []*workflowservice.DescribeWorkerDeploymentVersionResponse_VersionTaskQueue{
			{
				Name:               "tq-a",
				Type:               enumspb.TASK_QUEUE_TYPE_WORKFLOW,
				Stats:              stats(7),
				StatsByPriorityKey: map[int32]*taskqueuepb.TaskQueueStats{1: stats(4), 3: stats(3)},
			},
			{Name: "tq-b", Type: enumspb.TASK_QUEUE_TYPE_ACTIVITY, Stats: stats(0)},
		},
	}
	fake.setComputeConfig(twoScalingGroupComputeConfig())

	for _, statsFlags := range [][]string{nil, {"--report-task-queue-stats"}} {
		for _, outputFlags := range [][]string{nil, {"-o", "json"}} {
			flags := append([]string{"--address", s.Address(), "--deployment-name", "d", "--build-id", "b"}, statsFlags...)
			flags = append(flags, outputFlags...)
			existing := s.Execute(append([]string{"worker", "deployment", "describe-version"}, flags...)...)
			s.NoError(existing.Err)
			described := s.Execute(append([]string{"worker", "deployment", "version", "describe"}, flags...)...)
			s.NoError(described.Err)
			if outputFlags != nil {
				s.JSONEq(existing.Stdout.String(), described.Stdout.String(), "flags %v", flags)
			} else {
				s.Equal(existing.Stdout.String(), described.Stdout.String(), "flags %v", flags)
			}
		}
	}
	s.Empty(fake.takeUpdates())
}

// TestVersion_CreateDescribeDelete runs against the real server, with a Version
// that has no compute config.
func (s *SharedServerSuite) TestVersion_CreateDescribeDelete() {
	deploymentName := uuid.NewString()
	version := []string{
		"--address", s.Address(),
		"--deployment-name", deploymentName,
		"--build-id", "build-1",
	}

	// The Worker Deployment must exist first.
	res := s.Execute(append([]string{"worker", "deployment", "version", "create"}, version...)...)
	s.ErrorContains(res.Err, "error creating worker deployment version")

	res = s.Execute("worker", "deployment", "create", "--address", s.Address(), "--name", deploymentName)
	s.NoError(res.Err)
	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(append([]string{"worker", "deployment", "version", "create"}, version...)...)
		assert.NoError(t, res.Err)
		assert.Contains(t, res.Stdout.String(), "Successfully created worker deployment version")
	}, 30*time.Second, 100*time.Millisecond)

	res = s.Execute(append([]string{"worker", "deployment", "version", "describe", "-o", "json"}, version...)...)
	s.NoError(res.Err)
	var described map[string]any
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &described))
	s.Equal(deploymentName, described["deploymentName"])
	s.Equal("build-1", described["BuildID"])
	s.NotContains(described, "computeConfig")

	// Creating the same Version again fails.
	res = s.Execute(append([]string{"worker", "deployment", "version", "create"}, version...)...)
	s.ErrorContains(res.Err, "already exists")

	res = s.Execute(append([]string{"worker", "deployment", "version", "delete"}, version...)...)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), "Successfully deleted worker deployment version")

	res = s.Execute(append([]string{"worker", "deployment", "version", "describe"}, version...)...)
	s.ErrorContains(res.Err, "error describing worker deployment version")
}
