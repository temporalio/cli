package temporalcli_test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/temporalio/cli/temporalcli"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
)

type statsRowType struct {
	BuildID                 string        `json:"buildId"`
	TaskQueueType           string        `json:"taskQueueType"`
	ApproximateBacklogCount int64         `json:"approximateBacklogCount"`
	ApproximateBacklogAge   time.Duration `json:"approximateBacklogAge"`
	BacklogIncreaseRate     float32       `json:"backlogIncreaseRate"`
	TasksAddRate            float32       `json:"tasksAddRate"`
	TasksDispatchRate       float32       `json:"tasksDispatchRate"`
}

type taskQueueStatsType struct {
	Stats []statsRowType `json:"stats"`
}

func (s *SharedServerSuite) TestTaskQueue_Describe_Task_Queue_Stats_Empty() {
	// text
	res := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "workflow", "0", "0s", "0", "0", "0")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "activity", "0", "0s", "0", "0", "0")

	// json
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"-o", "json",
	)
	s.NoError(res.Err)

	var jsonOut taskQueueStatsType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal(statsRowType{
		BuildID: "UNVERSIONED",
		// ordering of workflow/activity pollers is random
		TaskQueueType:           jsonOut.Stats[0].TaskQueueType,
		ApproximateBacklogCount: 0,
		ApproximateBacklogAge:   0,
		BacklogIncreaseRate:     0,
		TasksAddRate:            0,
		TasksDispatchRate:       0,
	}, jsonOut.Stats[0])
}

func (s *SharedServerSuite) TestTaskQueue_Describe_Task_Queue_Stats_NonEmpty() {
	s.Worker().OnDevWorkflow(func(ctx workflow.Context, input any) (any, error) {
		return map[string]string{"foo": "bar"}, nil
	})

	// starting a new workflow execution
	res := s.Execute(
		"workflow", "start",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--type", "DevWorkflow",
		"--workflow-id", "my-id1",
	)
	s.NoError(res.Err)

	result := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(result.Err)
	out := result.Stdout.String()
	tqMetricsValidator := true
	s.EventuallyWithT(func(collect *assert.CollectT) {
		lines := strings.Split(out, "\n")
		tqStat := false
		for _, line := range lines {
			// Separating Task Queue Statistics output from Pollers output since they are similar
			if strings.Contains(line, "Task Queue Statistics:") {
				tqStat = true
				continue
			} else {
				tqStat = false
			}

			if tqStat {
				fields := strings.Fields(line)
				if len(fields) < 7 {
					// lesser fields than expected in the output, skip this line
					continue
				}

				tqType := fields[1]
				if tqType == "activity" {
					// all metrics should be 0
					for _, metric := range fields[2:] {
						if metric != "0" && metric != "0s" {
							tqMetricsValidator = false
						}
					}
				} else if tqType == "workflow" {
					backlogIncreaseRate := fields[4]
					tasksAddRate := fields[5]
					tasksDispatchRate := fields[6]
					if tasksAddRate == "0" && tasksDispatchRate == "0" && backlogIncreaseRate == "0" {
						// instead of checking each individual attribute, the following check has been added since:

						// 1. backlogIncreaseRate can be 0 when tasksAddRate and tasksDispatchRate is the same
						// 2. tasksDispatchRate can be 0

						// however, there won't be a case when all three of these metrics in this test have the null value
						tqMetricsValidator = false
					}
				}
			}
		}
		assert.True(collect, tqMetricsValidator, "expected 'tqMetricsValidator' to be true")
	}, time.Second*5, time.Millisecond*200)

	// json
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"-o", "json",
	)
	s.NoError(res.Err)

	var jsonOut taskQueueStatsType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	nullStatsRow := statsRowType{
		BuildID:                 "UNVERSIONED",
		TaskQueueType:           jsonOut.Stats[0].TaskQueueType,
		ApproximateBacklogCount: 0,
		ApproximateBacklogAge:   0,
		BacklogIncreaseRate:     0,
		TasksAddRate:            0,
		TasksDispatchRate:       0,
	}

	// The workflow queue should have non-zero task queue statistics
	if jsonOut.Stats[0].TaskQueueType == "workflow" {
		s.NotEqual(nullStatsRow, jsonOut.Stats[0])
	} else {
		s.NotEqual(nullStatsRow, jsonOut.Stats[1])
	}
}

func (s *SharedServerSuite) TestTaskQueue_Describe_Simple() {
	type reachabilityRowType struct {
		BuildID      string `json:"buildId"`
		Reachability string `json:"reachability"`
	}

	type pollerRowType struct {
		BuildID        string    `json:"buildId"`
		TaskQueueType  string    `json:"taskQueueType"`
		Identity       string    `json:"identity"`
		LastAccessTime time.Time `json:"-"`
		RatePerSecond  float64   `json:"ratePerSecond"`
	}

	type taskQueueDescriptionType struct {
		Reachability []reachabilityRowType `json:"reachability"`
		Pollers      []pollerRowType       `json:"pollers"`
	}

	// Wait until the poller appears
	s.Eventually(func() bool {
		desc, err := s.Client.DescribeTaskQueue(s.Context, s.Worker().Options.TaskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
		s.NoError(err)
		for _, poller := range desc.Pollers {
			if poller.Identity == s.DevServer.Options.ClientOptions.Identity {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "Worker never appeared")

	// Text

	// No task reachability
	res := s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)

	s.NotContains(res.Stdout.String(), "reachable")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "workflow", s.DevServer.Options.ClientOptions.Identity, "now", "100000")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "activity", s.DevServer.Options.ClientOptions.Identity, "now", "100000")

	// With task reachability info
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--report-reachability",
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "reachable")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "workflow", s.DevServer.Options.ClientOptions.Identity, "now", "100000")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "activity", s.DevServer.Options.ClientOptions.Identity, "now", "100000")

	// json
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--report-reachability",
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
		"-o", "json",
	)
	s.NoError(res.Err)

	var jsonOut taskQueueDescriptionType
	s.NoError(json.Unmarshal(res.Stdout.Bytes(), &jsonOut))
	s.Equal([]reachabilityRowType{
		{
			BuildID:      "UNVERSIONED",
			Reachability: "reachable",
		},
	}, jsonOut.Reachability)
	// The number of pollers can change
	s.Equal(pollerRowType{
		BuildID: "UNVERSIONED",
		// ordering of workflow/activity pollers is random
		TaskQueueType: jsonOut.Pollers[0].TaskQueueType,
		Identity:      s.DevServer.Options.ClientOptions.Identity,
		RatePerSecond: 100000,
	}, jsonOut.Pollers[0])

	// Adding a default build ID
	res = s.Execute(
		"task-queue", "versioning", "insert-assignment-rule",
		"--build-id", "id1",
		"-y",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)

	// Text
	res = s.Execute(
		"task-queue", "describe",
		"--address", s.Address(),
		"--report-reachability",
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "id1", "reachable")
	// No pollers on id1
	s.NotContains(res.Stdout.String(), "now")

	res = s.Execute(
		"task-queue", "describe",
		"--select-unversioned",
		"--address", s.Address(),
		"--report-reachability",
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "unreachable")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "workflow", s.DevServer.Options.ClientOptions.Identity, "now", "100000")
	s.ContainsOnSameLine(res.Stdout.String(), "UNVERSIONED", "activity", s.DevServer.Options.ClientOptions.Identity, "now", "100000")

	res = s.Execute(
		"task-queue", "describe",
		"--select-build-id", "id2",
		"--address", s.Address(),
		"--report-reachability",
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)

	s.ContainsOnSameLine(res.Stdout.String(), "id2", "unreachable")
	// No pollers on id2
	s.NotContains(res.Stdout.String(), "now")

	res = s.Execute(
		"task-queue", "describe",
		"--select-all-active",
		"--address", s.Address(),
		"--report-reachability",
		"--disable-stats",
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)
	// id1 has no pollers so it is not active
	s.NotContains(res.Stdout.String(), "id1")
	s.NotContains(res.Stdout.String(), "now")
}

func (s *SharedServerSuite) TestTaskQueue_Describe_Simple_Legacy() {
	// Wait until the poller appears
	s.Eventually(func() bool {
		desc, err := s.Client.DescribeTaskQueue(s.Context, s.Worker().Options.TaskQueue, enums.TASK_QUEUE_TYPE_WORKFLOW)
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
		"--legacy-mode",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
	)
	s.NoError(res.Err)
	// For text, just making sure our client identity is present is good enough
	s.Contains(res.Stdout.String(), s.DevServer.Options.ClientOptions.Identity)

	// JSON
	res = s.Execute(
		"task-queue", "describe",
		"-o", "json",
		"--legacy-mode",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
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
		"--legacy-mode",
		"-o", "json",
		"--address", s.Address(),
		"--task-queue", s.Worker().Options.TaskQueue,
		"--partitions-legacy", "10",
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
