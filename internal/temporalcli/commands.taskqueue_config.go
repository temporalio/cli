package temporalcli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/temporalio/cli/internal/printer"
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

	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	namespace := c.Parent.Parent.Namespace
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

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

// parseFairnessKeyWeights parses "key=weight" format strings into a map
func parseFairnessKeyWeights(inputs []string) (map[string]float32, error) {
	weights := make(map[string]float32)
	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %s (expected key=weight)", input)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("empty key in: %s", input)
		}
		weight, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 32)
		if err != nil {
			return nil, fmt.Errorf("invalid weight in %s: %w", input, err)
		}
		if weight < 0 {
			return nil, fmt.Errorf("weight must be non-negative in: %s", input)
		}
		weights[key] = float32(weight)
	}
	return weights, nil
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

	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	namespace := c.Parent.Parent.Namespace
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

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

	// Handle fairness weight overrides
	// Validate mutual exclusivity
	if c.FairnessKeyWeightUnsetAll {
		if len(c.FairnessKeyWeightSet) > 0 || len(c.FairnessKeyWeightUnset) > 0 {
			return fmt.Errorf("--fairness-key-weight-unset-all cannot be used with --fairness-key-weight-set or --fairness-key-weight-unset")
		}
	}

	// Handle set operations
	if len(c.FairnessKeyWeightSet) > 0 {
		weights, err := parseFairnessKeyWeights(c.FairnessKeyWeightSet)
		if err != nil {
			return err
		}
		request.SetFairnessWeightOverrides = weights
	}

	// Handle unset operations
	if len(c.FairnessKeyWeightUnset) > 0 {
		request.UnsetFairnessWeightOverrides = c.FairnessKeyWeightUnset
	}

	// Handle unset all
	if c.FairnessKeyWeightUnsetAll {
		// Need to fetch current config to get all keys to unset
		descResp, err := cl.WorkflowService().DescribeTaskQueue(cctx, &workflowservice.DescribeTaskQueueRequest{
			Namespace: namespace,
			TaskQueue: &taskqueue.TaskQueue{
				Name: taskQueue,
				Kind: enums.TASK_QUEUE_KIND_NORMAL,
			},
			TaskQueueType: taskQueueType,
			ReportConfig:  true,
		})
		if err != nil {
			return fmt.Errorf("error fetching current config for unset-all: %w", err)
		}
		if descResp.Config != nil && descResp.Config.FairnessWeightOverrides != nil {
			keys := make([]string, 0, len(descResp.Config.FairnessWeightOverrides))
			for key := range descResp.Config.FairnessWeightOverrides {
				keys = append(keys, key)
			}
			request.UnsetFairnessWeightOverrides = keys
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
