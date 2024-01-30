package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/enums/v1"
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
	resp, err := cl.DescribeTaskQueue(cctx, c.TaskQueue, taskQueueType)
	if err != nil {
		return fmt.Errorf("failed describing task queue")
	}

	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}

	// For text, we will use a table for pollers
	cctx.Printer.Println(color.MagentaString("Pollers:"))
	items := make([]struct {
		Identity       string
		LastAccessTime time.Time
		RatePerSecond  float64
	}, len(resp.Pollers))
	for i, poller := range resp.Pollers {
		items[i].Identity = poller.Identity
		items[i].LastAccessTime = poller.LastAccessTime.AsTime()
		items[i].RatePerSecond = poller.RatePerSecond
	}
	return cctx.Printer.PrintStructured(items, printer.StructuredOptions{Table: &printer.TableOptions{}})
}
