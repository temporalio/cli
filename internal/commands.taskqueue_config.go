package temporalcli

import (
	"fmt"
	"strconv"
	"strings"

	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"

	"github.com/temporalio/cli/internal/printer"
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
		if c.Command.Flags().Changed("queue-rps-limit") ||
			c.Command.Flags().Changed("queue-rps-limit-reason") {
			return fmt.Errorf("setting rate limit on workflow task queues is not allowed")
		}
	}

	// Helper to parse RPS values for a given flag name.
	// Accepts "default" or a non-negative float string.
	parseRPS := func(flagName string) (*taskqueue.RateLimit, error) {
		raw := strings.TrimSpace(c.Command.Flags().Lookup(flagName).Value.String())
		if raw == "" {
			return nil, fmt.Errorf("invalid value for --%s: must be a non-negative number or 'default'", flagName)
		}
		if strings.EqualFold(raw, "default") {
			// Unset: returning nil RateLimit removes the existing rate limit.
			return nil, nil
		}
		v, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid value for --%s: must be a non-negative number or 'default'", flagName)
		}
		if v < 0 {
			return nil, fmt.Errorf("invalid value for --%s: must be >= 0 or 'default'", flagName)
		}
		return &taskqueue.RateLimit{RequestsPerSecond: float32(v)}, nil
	}

	var queueRpsLimitParsed *taskqueue.RateLimit
	if c.Command.Flags().Changed("queue-rps-limit") {
		var err error
		if queueRpsLimitParsed, err = parseRPS("queue-rps-limit"); err != nil {
			return err
		}
	} else if c.Command.Flags().Changed("queue-rps-limit-reason") {
		return fmt.Errorf("queue-rps-limit-reason can only be set if queue-rps-limit is updated")
	}

	var fairnessKeyRpsLimitDefaultParsed *taskqueue.RateLimit
	if c.Command.Flags().Changed("fairness-key-rps-limit-default") {
		var err error
		if fairnessKeyRpsLimitDefaultParsed, err = parseRPS("fairness-key-rps-limit-default"); err != nil {
			return err
		}
	} else if c.Command.Flags().Changed("fairness-key-rps-limit-default-reason") {
		return fmt.Errorf("fairness-key-rps-limit-default-reason can only be set if fairness-key-rps-limit-default is updated")
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
	if c.Command.Flags().Changed("queue-rps-limit") {
		request.UpdateQueueRateLimit = &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
			RateLimit: queueRpsLimitParsed,
			Reason:    c.QueueRpsLimitReason,
		}
	}

	// Add fairness key rate limit default if specified (including unset)
	if c.Command.Flags().Changed("fairness-key-rps-limit-default") {
		request.UpdateFairnessKeyRateLimitDefault = &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
			RateLimit: fairnessKeyRpsLimitDefaultParsed,
			Reason:    c.FairnessKeyRpsLimitReason,
		}
	}

	// Call the API
	resp, err := cl.WorkflowService().UpdateTaskQueueConfig(cctx, request)
	if err != nil {
		return fmt.Errorf("error updating task queue config: %w", err)
	}

	cctx.Printer.Println("Successfully updated task queue configuration")
	return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
}
