package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/tqname"
)

func (c *TemporalTaskQueueDescribeCommand) run(cctx *CommandContext, args []string) error {
	// Call describe
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	var taskQueueType enums.TaskQueueType
	switch c.TaskQueueType.Value {
	case "workflow":
		taskQueueType = enums.TASK_QUEUE_TYPE_WORKFLOW
	case "activity":
		taskQueueType = enums.TASK_QUEUE_TYPE_ACTIVITY
	default:
		return fmt.Errorf("unrecognized task queue type: %q", c.TaskQueueType.Value)
	}

	taskQueueName, err := tqname.FromBaseName(c.TaskQueue)
	if err != nil {
		return fmt.Errorf("failed to parse task queue name: %w", err)
	}
	partitions := c.Partitions

	type statusWithPartition struct {
		Partition int `json:"partition"`
		taskqueue.TaskQueueStatus
	}
	type pollerWithPartition struct {
		Partition int `json:"partition"`
		taskqueue.PollerInfo
		// copy this out to display nicer in table or card, but not json
		Versioning *commonpb.WorkerVersionCapabilities `json:"-"`
	}

	var statuses []*statusWithPartition
	var pollers []*pollerWithPartition

	// TODO: remove this when the server does partition fan-out
	for p := 0; p < partitions; p++ {
		resp, err := cl.WorkflowService().DescribeTaskQueue(cctx, &workflowservice.DescribeTaskQueueRequest{
			Namespace: c.Parent.Namespace,
			TaskQueue: &taskqueue.TaskQueue{
				Name: taskQueueName.WithPartition(p).FullName(),
				Kind: enums.TASK_QUEUE_KIND_NORMAL,
			},
			TaskQueueType:          taskQueueType,
			IncludeTaskQueueStatus: true,
		})
		if err != nil {
			return fmt.Errorf("unable to describe task queue: %w", err)
		}
		statuses = append(statuses, &statusWithPartition{
			Partition:       p,
			TaskQueueStatus: *resp.TaskQueueStatus,
		})
		for _, pi := range resp.Pollers {
			pollers = append(pollers, &pollerWithPartition{
				Partition:  p,
				PollerInfo: *pi,
				Versioning: pi.WorkerVersionCapabilities,
			})
		}
	}

	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(map[string]any{
			"taskQueues": statuses,
			"pollers":    pollers,
		}, printer.StructuredOptions{})
	}

	// For text, we will use a table for pollers
	cctx.Printer.Println(color.MagentaString("Pollers:"))
	items := make([]struct {
		Identity       string
		LastAccessTime time.Time
		RatePerSecond  float64
	}, len(pollers))
	for i, poller := range pollers {
		items[i].Identity = poller.Identity
		items[i].LastAccessTime = poller.LastAccessTime.AsTime()
		items[i].RatePerSecond = poller.RatePerSecond
	}
	return cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})
}

func (c *TemporalTaskQueueListPartitionCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	request := &workflowservice.ListTaskQueuePartitionsRequest{
		Namespace: c.Parent.Namespace,
		TaskQueue: &taskqueue.TaskQueue{
			Name: c.TaskQueue,
			Kind: enums.TASK_QUEUE_KIND_NORMAL,
		},
	}

	resp, err := cl.WorkflowService().ListTaskQueuePartitions(cctx, request)
	if err != nil {
		return fmt.Errorf("unable to list task queues: %w", err)
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	var items []*taskqueue.TaskQueuePartitionMetadata
	cctx.Printer.Println(color.MagentaString("Workflow Task Queue Partitions\n"))
	for _, e := range resp.WorkflowTaskQueuePartitions {
		items = append(items, e)
	}
	_ = cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})

	items = items[:0]
	cctx.Printer.Println(color.MagentaString("\nActivity Task Queue Partitions\n"))
	for _, e := range resp.ActivityTaskQueuePartitions {
		items = append(items, e)
	}
	_ = cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})

	return nil
}
