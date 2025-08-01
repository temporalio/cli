package temporalcli

import (
	"fmt"

	enums "go.temporal.io/api/enums/v1"
	taskqueue "go.temporal.io/api/taskqueue/v1"
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
	taskQueueType := enums.TASK_QUEUE_TYPE_WORKFLOW // default
	if c.TaskQueueType != "" {
		switch c.TaskQueueType {
		case "workflow":
			taskQueueType = enums.TASK_QUEUE_TYPE_WORKFLOW
		case "activity":
			taskQueueType = enums.TASK_QUEUE_TYPE_ACTIVITY
		case "nexus":
			taskQueueType = enums.TASK_QUEUE_TYPE_NEXUS
		default:
			return fmt.Errorf("invalid task queue type: %s. Must be one of: workflow, activity, nexus", c.TaskQueueType)
		}
	}

	// Build the request
	if taskQueue == "" {
		return fmt.Errorf("taskQueue name is required")
	}
	if taskQueueType == enums.TASK_QUEUE_TYPE_WORKFLOW {
		if c.QueueRateLimit != nil && c.QueueRateLimit.RequestsPerSecond != UnsetRateLimit {
			return fmt.Errorf("setting rate limit on workflow task queues is not allowed")
		}
		return fmt.Errorf("taskQueueType is required")
	}
	namespace := c.Namespace
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	request := &workflowservice.UpdateTaskQueueConfigRequest{
		Namespace:     namespace,
		Identity:      c.Identity,
		TaskQueue:     taskQueue,
		TaskQueueType: taskQueueType,
	}

	// Add queue rate limit if specified (including unset)
	if c.QueueRateLimit != nil {
		if c.QueueRateLimit.RequestsPerSecond == UnsetRateLimit {
			// For unset, we pass an empty RateLimitUpdate with no RateLimit
			request.UpdateQueueRateLimit = &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
				Reason: c.QueueRateLimit.Reason,
			}
		} else {
			// For setting a value
			request.UpdateQueueRateLimit = &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
				RateLimit: &taskqueue.RateLimit{
					RequestsPerSecond: float32(c.QueueRateLimit.RequestsPerSecond),
				},
				Reason: c.QueueRateLimit.Reason,
			}
		}
	}

	// Add fairness key rate limit default if specified (including unset)
	if c.FairnessKeyRateLimitDefault != nil {
		if c.FairnessKeyRateLimitDefault.RequestsPerSecond == UnsetRateLimit {
			// For unset, we pass an empty RateLimitUpdate with no RateLimit
			request.UpdateFairnessKeyRateLimitDefault = &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
				Reason: c.FairnessKeyRateLimitDefault.Reason,
			}
		} else {
			// For setting a value
			request.UpdateFairnessKeyRateLimitDefault = &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
				RateLimit: &taskqueue.RateLimit{
					RequestsPerSecond: float32(c.FairnessKeyRateLimitDefault.RequestsPerSecond),
				},
				Reason: c.FairnessKeyRateLimitDefault.Reason,
			}
		}
	}

	// Call the API
	_, err = cl.WorkflowService().UpdateTaskQueueConfig(cctx, request)
	if err != nil {
		return fmt.Errorf("error updating task queue config: %w", err)
	}

	cctx.Printer.Println("Successfully updated task queue configuration")

	// Print summary of what was updated
	if c.QueueRateLimit != nil {
		if c.QueueRateLimit.RequestsPerSecond == UnsetRateLimit {
			cctx.Printer.Println("Queue Rate Limit: Unset")
		} else {
			cctx.Printer.Printlnf("Queue Rate Limit: %.2f requests/second", c.QueueRateLimit.RequestsPerSecond)
		}
		if c.QueueRateLimit.Reason != "" {
			cctx.Printer.Printlnf("Queue Rate Limit Reason: %s", c.QueueRateLimit.Reason)
		}
	}

	if c.FairnessKeyRateLimitDefault != nil {
		if c.FairnessKeyRateLimitDefault.RequestsPerSecond == UnsetRateLimit {
			cctx.Printer.Println("Fairness Key Rate Limit Default: Unset")
		} else {
			cctx.Printer.Printlnf("Fairness Key Rate Limit Default: %.2f requests/second", c.FairnessKeyRateLimitDefault.RequestsPerSecond)
		}
		if c.FairnessKeyRateLimitDefault.Reason != "" {
			cctx.Printer.Printlnf("Fairness Key Rate Limit Reason: %s", c.FairnessKeyRateLimitDefault.Reason)
		}
	}

	return nil
}
