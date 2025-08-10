package temporalcli

import (
	"fmt"

	"github.com/temporalio/cli/temporalcli/internal/printer"
	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
)

// TaskQueueConfigGetCommand handles getting task queue configuration
func (c *TemporalTaskQueueConfigGetCommand) run(cctx *CommandContext, args []string) error {
	// Validate inputs before dialing client
	taskQueue := c.TaskQueue
	if taskQueue == "" {
		return fmt.Errorf("taskQueue name is required")
	}

	taskQueueType, err := parseTaskQueueType(c.TaskQueueType.Value)
	if err != nil {
		return err
	}

	namespace := c.Parent.Parent.Namespace
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	// Get the task queue configuration
	resp, err := cl.WorkflowService().DescribeTaskQueue(cctx, &workflowservice.DescribeTaskQueueRequest{
		Namespace: namespace,
		TaskQueue: &taskqueue.TaskQueue{
			Name: taskQueue,
			Kind: enums.TASK_QUEUE_KIND_NORMAL,
		},
		TaskQueueType: taskQueueType,
		ReportConfig:  true,
	})
	if err != nil {
		return fmt.Errorf("error getting task queue config: %w", err)
	}
	if resp.Config == nil {
		cctx.Printer.Println("No configuration found for task queue")
		return nil
	}
	// Print the configuration using the shared function
	return printTaskQueueConfig(cctx, resp.Config)
}

// TaskQueueConfigSetCommand handles setting task queue configuration
func (c *TemporalTaskQueueConfigSetCommand) run(cctx *CommandContext, args []string) error {
	// Validate inputs before dialing client
	taskQueue := c.TaskQueue
	if taskQueue == "" {
		return fmt.Errorf("taskQueue name is required")
	}

	taskQueueType, err := parseTaskQueueType(c.TaskQueueType.Value)
	if err != nil {
		return err
	}

	namespace := c.Parent.Parent.Namespace
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	// Check workflow task queue restrictions
	if taskQueueType == enums.TASK_QUEUE_TYPE_WORKFLOW {
		if c.Command.Flags().Changed("queue-rate-limit") ||
			c.Command.Flags().Changed("queue-rate-limit-reason") {
			return fmt.Errorf("setting rate limit on workflow task queues is not allowed")
		}
	}

	cl, err := c.Parent.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	request := &workflowservice.UpdateTaskQueueConfigRequest{
		Namespace:     namespace,
		Identity:      c.Parent.Parent.Identity,
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
