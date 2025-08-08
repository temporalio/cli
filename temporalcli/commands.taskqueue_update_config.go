package temporalcli

import (
	"fmt"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	enums "go.temporal.io/api/enums/v1"
	workflowservice "go.temporal.io/api/workflowservice/v1"
)

const (
	UnsetRateLimit = -1
)

func (c *TemporalTaskQueueUpdateConfigCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	taskQueue := c.TaskQueue
	taskQueueType, err := parseTaskQueueType(c.TaskQueueType.Value)
	if err != nil {
		return err
	}

	// Build the request
	if taskQueue == "" {
		return fmt.Errorf("taskQueue name is required")
	}
	if taskQueueType == enums.TASK_QUEUE_TYPE_WORKFLOW {
		if c.QueueRateLimit != 0 && c.QueueRateLimit != UnsetRateLimit {
			return fmt.Errorf("setting rate limit on workflow task queues is not allowed")
		}
	}
	namespace := c.Parent.Namespace
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	request := &workflowservice.UpdateTaskQueueConfigRequest{
		Namespace:     namespace,
		Identity:      c.Parent.Identity,
		TaskQueue:     taskQueue,
		TaskQueueType: taskQueueType,
	}

	// Add queue rate limit if specified (including unset)
	request.UpdateQueueRateLimit = buildRateLimitUpdate(
		c.Command.Flags().Changed("queue-rate-limit"),
		c.QueueRateLimit,
		c.QueueRateLimitReason,
	)

	// Add fairness key rate limit default if specified (including unset)
	request.UpdateFairnessKeyRateLimitDefault = buildRateLimitUpdate(
		c.Command.Flags().Changed("fairness-key-rate-limit-default"),
		c.FairnessKeyRateLimitDefault,
		c.FairnessKeyRateLimitReason,
	)

	// Call the API
	resp, err := cl.WorkflowService().UpdateTaskQueueConfig(cctx, request)
	if err != nil {
		return fmt.Errorf("error updating task queue config: %w", err)
	}

	cctx.Printer.Println("Successfully updated task queue configuration")
	return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
}
