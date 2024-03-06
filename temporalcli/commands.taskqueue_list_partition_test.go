package temporalcli_test

import (
	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/workflowservice/v1"
)

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
