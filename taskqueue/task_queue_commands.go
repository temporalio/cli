package taskqueue

import (
	"fmt"
	"strings"

	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
	enumspb "go.temporal.io/api/enums/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
)

// DescribeTaskQueue show pollers info of a given taskqueue
func DescribeTaskQueue(c *cli.Context) error {
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}
	taskQueue := c.String(common.FlagTaskQueue)
	taskQueueType := strToTaskQueueType(c.String(common.FlagTaskQueueType))

	ctx, cancel := common.NewContext(c)
	defer cancel()
	resp, err := sdkClient.DescribeTaskQueue(ctx, taskQueue, taskQueueType)
	if err != nil {
		return fmt.Errorf("unable to describe task queue: %w", err)
	}

	opts := &output.PrintOptions{
		// TODO enable when versioning feature is out
		// Fields: []string{"Identity", "LastAccessTime", "RatePerSecond", "WorkerVersioningId"},
		Fields: []string{"Identity", "LastAccessTime", "RatePerSecond"},
	}
	var items []interface{}
	for _, e := range resp.Pollers {
		items = append(items, e)
	}
	return output.PrintItems(c, items, opts)
}

// ListTaskQueuePartitions gets all the taskqueue partition and host information.
func ListTaskQueuePartitions(c *cli.Context) error {
	frontendClient := client.CFactory.FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	taskQueue := c.String(common.FlagTaskQueue)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	request := &workflowservice.ListTaskQueuePartitionsRequest{
		Namespace: namespace,
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: taskQueue,
			Kind: enumspb.TASK_QUEUE_KIND_NORMAL,
		},
	}

	resp, err := frontendClient.ListTaskQueuePartitions(ctx, request)
	if err != nil {
		return fmt.Errorf("unable to list task queues: %w", err)
	}

	optsW := &output.PrintOptions{
		Fields: []string{"Key", "OwnerHostName"},
	}

	var items []interface{}
	fmt.Println(color.Magenta(c, "Workflow Task Queue Partitions\n"))
	for _, e := range resp.WorkflowTaskQueuePartitions {
		items = append(items, e)
	}
	err = output.PrintItems(c, items, optsW)
	if err != nil {
		return err
	}

	optsA := &output.PrintOptions{
		Fields: []string{"Key", "OwnerHostName"},
	}
	items = items[:0]
	fmt.Println(color.Magenta(c, "\nActivity Task Queue Partitions\n"))
	for _, e := range resp.ActivityTaskQueuePartitions {
		items = append(items, e)
	}
	return output.PrintItems(c, items, optsA)
}

func strToTaskQueueType(str string) enumspb.TaskQueueType {
	if strings.ToLower(str) == "activity" {
		return enumspb.TASK_QUEUE_TYPE_ACTIVITY
	}
	return enumspb.TASK_QUEUE_TYPE_WORKFLOW
}
