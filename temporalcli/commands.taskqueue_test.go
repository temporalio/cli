package temporalcli_test

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func (s *SharedServerSuite) TestTaskQueue_Describe_Simple() {
	// Wait until the poller appears
	s.Eventually(func() bool {
		desc, err := s.Client.DescribeTaskQueue(s.Context, s.Worker.Options.TaskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
		s.NoError(err)
		for _, poller := range desc.Pollers {
			if poller.Identity == s.DevServer.Options.ClientOptions.Identity {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "Worker never appeared")

	// Text
	res := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", s.Worker.Options.TaskQueue,
	)
	s.NoError(res.Err)
	// For text, just making sure our client identity is present is good enough
	s.Contains(res.Stdout.String(), s.DevServer.Options.ClientOptions.Identity)

	// JSON
	res = s.Execute(
		"task-queue", "describe",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker.Options.TaskQueue,
	)
	s.NoError(res.Err)
	var jsonOut struct {
		Pollers    []map[string]any `json:"pollers"`
		TaskQueues []map[string]any `json:"taskQueues"`
	}
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(1, len(jsonOut.TaskQueues))
	// Check identity in the output
	s.Equal(s.DevServer.Options.ClientOptions.Identity, jsonOut.Pollers[0]["identity"])

	// Multiple partitions
	res = s.Execute(
		"task-queue", "describe",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker.Options.TaskQueue,
		"--partitions", "10",
	)
	s.NoError(res.Err)
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.GreaterOrEqual(10, len(jsonOut.TaskQueues))
}

func (s *SharedServerSuite) TestTaskQueue_ListPartition() {
	testTaskQueue := uuid.NewString()
	res := s.Execute(
		"task-queue", "list-partition",
		"--address", s.Address(),
		"--task-queue", testTaskQueue,
	)
	s.Contains(res.Stdout.String(), testTaskQueue)
	s.Contains(res.Stdout.String(), "Workflow Task Queue Partitions")
	s.Contains(res.Stdout.String(), "Activity Task Queue Partitions")
	s.NoError(res.Err)
}

func (s *SharedServerSuite) TestTaskQueue_ListPartitionInvalidNamespace() {
	testTaskQueue := uuid.NewString()
	res := s.Execute(
		"task-queue", "list-partition",
		"--address", s.Address(),
		"--task-queue", testTaskQueue,
		"--namespace", "invalid_namespace",
	)
	s.Error(res.Err)
	s.Contains(res.Err.Error(), "Namespace invalid_namespace is not found")
}

func (s *SharedServerSuite) TestTaskQueue_ListPartitionJsonOutput() {
	testTaskQueue := uuid.NewString()
	res := s.Execute(
		"task-queue", "list-partition",
		"--address", s.Address(),
		"--task-queue", testTaskQueue,
		"--output", "json",
	)
	s.NoError(res.Err)
	var listResp workflowservice.ListTaskQueuePartitionsResponse
	s.NoError(temporalcli.UnmarshalProtoJSONWithOptions(res.Stdout.Bytes(), &listResp, true))
	s.NotEmpty(listResp.ActivityTaskQueuePartitions)
	s.NotEmpty(listResp.WorkflowTaskQueuePartitions)
}
