package temporalcli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
)

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
