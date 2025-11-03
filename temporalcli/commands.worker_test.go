package temporalcli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	deploymentpb "go.temporal.io/api/deployment/v1"
	enumspb "go.temporal.io/api/enums/v1"
	workerpb "go.temporal.io/api/worker/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type heartbeatJSON struct {
	Heartbeat workerHeartbeat `json:"workerHeartbeat"`
}

type workerHeartbeat struct {
	WorkerInstanceKey         string            `json:"workerInstanceKey"`
	WorkerIdentity            string            `json:"workerIdentity"`
	HostInfo                  hostInfo          `json:"hostInfo"`
	TaskQueue                 string            `json:"taskQueue"`
	DeploymentVersion         deploymentVersion `json:"deploymentVersion"`
	SdkName                   string            `json:"sdkName"`
	SdkVersion                string            `json:"sdkVersion"`
	Status                    string            `json:"status"`
	StartTime                 time.Time         `json:"startTime"`
	HeartbeatTime             time.Time         `json:"heartbeatTime"`
	ElapsedSinceLastHeartbeat string            `json:"elapsedSinceLastHeartbeat"`
}
type hostInfo struct {
	HostName string `json:"hostName"`
}

type deploymentVersion struct {
	DeploymentName string `json:"deploymentName"`
	BuildId        string `json:"buildId"`
}

func (s *SharedServerSuite) TestWorkerHeartbeat_List() {
	heartbeat, err := s.recordWorkerHeartbeat()
	s.NoError(err)

	sentHeartbeatJSON := s.waitForWorkerListJSON(heartbeat.TaskQueue, heartbeat.WorkerInstanceKey)

	s.Equal(heartbeat.WorkerIdentity, sentHeartbeatJSON.Heartbeat.WorkerIdentity)
	s.Equal(heartbeat.TaskQueue, sentHeartbeatJSON.Heartbeat.TaskQueue)
	s.Equal(heartbeat.WorkerInstanceKey, sentHeartbeatJSON.Heartbeat.WorkerInstanceKey)
	s.False(sentHeartbeatJSON.Heartbeat.HeartbeatTime.IsZero())

	res := s.Execute(
		"worker", "list",
		"--address", s.Address(),
		"--query", fmt.Sprintf("TaskQueue=\"%s\"", heartbeat.TaskQueue),
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), heartbeat.WorkerInstanceKey, heartbeat.TaskQueue, heartbeat.WorkerIdentity)
}

func (s *SharedServerSuite) TestWorkerHeartbeat_Describe() {
	heartbeat, err := s.recordWorkerHeartbeat()
	s.NoError(err)

	listJSON := s.waitForWorkerListJSON(heartbeat.TaskQueue, heartbeat.WorkerInstanceKey)

	var detail heartbeatJSON
	res := s.Execute(
		"worker", "describe",
		"--address", s.Address(),
		"--worker-instance-key", heartbeat.WorkerInstanceKey,
		"--output", "json",
	)
	s.NoError(res.Err)

	var parsed heartbeatJSON
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &parsed))

	s.Equal(heartbeat.WorkerInstanceKey, parsed.Heartbeat.WorkerInstanceKey)
	s.Equal(heartbeat.TaskQueue, parsed.Heartbeat.TaskQueue)
	s.Equal(heartbeat.WorkerIdentity, parsed.Heartbeat.WorkerIdentity)
	s.False(parsed.Heartbeat.StartTime.IsZero())
	s.False(parsed.Heartbeat.HeartbeatTime.IsZero())

	detail = parsed

	s.Equal(heartbeat.WorkerInstanceKey, detail.Heartbeat.WorkerInstanceKey)
	s.Equal(heartbeat.TaskQueue, detail.Heartbeat.TaskQueue)
	s.Equal(heartbeat.WorkerIdentity, detail.Heartbeat.WorkerIdentity)
	// Ensure that JSON output for list and describe are identical
	s.Equal(listJSON, detail)

	res = s.Execute(
		"worker", "describe",
		"--address", s.Address(),
		"--worker-instance-key", heartbeat.WorkerInstanceKey,
	)
	s.NoError(res.Err)
	s.Contains(res.Stdout.String(), heartbeat.WorkerIdentity)
}

func (s *SharedServerSuite) recordWorkerHeartbeat() (*workerpb.WorkerHeartbeat, error) {
	workerIdentity := "heartbeat-list-" + uuid.NewString()
	taskQueue := "heartbeat-tq-" + uuid.NewString()
	instanceKey := "heartbeat-instance-" + uuid.NewString()
	hostName := "heartbeat-host-" + uuid.NewString()
	deploymentName := "heartbeat-deployment-" + uuid.NewString()
	buildID := "heartbeat-build-" + uuid.NewString()
	status := enumspb.WORKER_STATUS_RUNNING
	startTime := time.Now().UTC().Add(-1 * time.Minute).Truncate(time.Second)
	heartbeatTime := time.Now().UTC().Truncate(time.Second)

	hb := &workerpb.WorkerHeartbeat{
		WorkerInstanceKey: instanceKey,
		WorkerIdentity:    workerIdentity,
		TaskQueue:         taskQueue,
		DeploymentVersion: &deploymentpb.WorkerDeploymentVersion{
			DeploymentName: deploymentName,
			BuildId:        buildID,
		},
		HostInfo: &workerpb.WorkerHostInfo{
			HostName: hostName,
		},
		SdkName:                   "temporal-go-sdk",
		SdkVersion:                "v0.0.0-test",
		Status:                    status,
		StartTime:                 timestamppb.New(startTime),
		HeartbeatTime:             timestamppb.New(heartbeatTime),
		ElapsedSinceLastHeartbeat: durationpb.New(5 * time.Second),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.Client.WorkflowService().RecordWorkerHeartbeat(ctx, &workflowservice.RecordWorkerHeartbeatRequest{
		Namespace:       s.Namespace(),
		Identity:        identity,
		WorkerHeartbeat: []*workerpb.WorkerHeartbeat{hb},
	})
	s.NoError(err)

	return hb, nil
}

func (s *SharedServerSuite) waitForWorkerListJSON(taskQueue, instanceKey string) heartbeatJSON {
	var row heartbeatJSON

	s.EventuallyWithT(func(t *assert.CollectT) {
		res := s.Execute(
			"worker", "list",
			"--address", s.Address(),
			"--query", fmt.Sprintf("TaskQueue=\"%s\"", taskQueue),
			"--output", "json",
		)
		assert.NoError(t, res.Err)

		var parsed []heartbeatJSON
		if !assert.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &parsed)) {
			return
		}
		if len(parsed) == 0 {
			assert.Fail(t, "no workers returned yet")
			return
		}

		for _, candidate := range parsed {
			if candidate.Heartbeat.WorkerInstanceKey == instanceKey {
				row = candidate
				return
			}
		}

		assert.Failf(t, "worker instance key not yet found", "%s", instanceKey)
	}, 5*time.Second, 200*time.Millisecond)

	s.NotEmpty(row.Heartbeat.WorkerInstanceKey)
	return row
}
