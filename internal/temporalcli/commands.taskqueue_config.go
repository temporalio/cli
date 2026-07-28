package temporalcli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/temporalio/cli/internal/printer"
	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"golang.org/x/exp/maps"
)

// TaskQueueConfigGetCommand handles getting task queue configuration
func (c *TemporalTaskQueueConfigGetCommand) run(cctx *CommandContext, args []string) error {
	// Validate inputs before dialing client
	taskQueue := strings.TrimSpace(c.TaskQueue)
	if taskQueue == "" {
		return fmt.Errorf("task queue name is required and cannot be empty")
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

const (
	// maxFairnessKeyLength matches server-side limit in temporal/common/priorities/priority_util.go
	maxFairnessKeyLength = 64
)

// printZeroRateLimitWarning prints a warning when a rate limit is set to 0
func printZeroRateLimitWarning(cctx *CommandContext, limitType string) {
	fmt.Fprintf(cctx.Options.Stderr, "WARNING: Setting %s to 0 will STOP ALL TRAFFIC on this task queue.\n", limitType)
	fmt.Fprintln(cctx.Options.Stderr, "         This will prevent any tasks from being dispatched until the limit is changed.")
}

// parseFairnessKeyWeights parses "key=weight" or "key=default" format strings
// Returns separate maps for set and unset operations, or an error if there are duplicate keys, invalid weights, or malformed input
// If inputs is empty, returns nil for both maps
func parseFairnessKeyWeights(inputs []string) (setWeights map[string]float32, unsetKeys []string, err error) {
	if len(inputs) == 0 {
		return nil, nil, nil
	}

	setWeights = make(map[string]float32)
	unsetKeysMap := make(map[string]bool) // Track unset keys in a map to check for duplicates
	seen := make(map[string]bool)

	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid format: %q (expected key=weight or key=default)", input)
		}

		key := parts[0]
		if key == "" {
			return nil, nil, fmt.Errorf("empty key in: %q", input)
		}

		// Check for duplicate keys across both set and unset
		if seen[key] {
			return nil, nil, fmt.Errorf("duplicate fairness key %q specified multiple times", key)
		}
		seen[key] = true

		valueStr := parts[1]
		if valueStr == "" {
			return nil, nil, fmt.Errorf("empty value for key %q", key)
		}

		// Check if this is an unset operation (value is "default")
		// Do this before validating key length since we don't care about length when unsetting
		if strings.EqualFold(valueStr, "default") {
			unsetKeysMap[key] = true
			continue
		}

		// Validate key length only for set operations (server enforces 64 byte limit)
		if len(key) > maxFairnessKeyLength {
			return nil, nil, fmt.Errorf("fairness key %q exceeds maximum length of %d bytes", key, maxFairnessKeyLength)
		}

		// Parse as weight
		weight, err := strconv.ParseFloat(valueStr, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid weight %q for key %q: must be a number or 'default'", valueStr, key)
		}

		// Validate weight is positive - server handles clamping to its configured range
		if weight <= 0 {
			return nil, nil, fmt.Errorf("weight for key %q must be positive", key)
		}

		setWeights[key] = float32(weight)
	}

	// Convert unset map to slice
	if len(unsetKeysMap) > 0 {
		unsetKeys = maps.Keys(unsetKeysMap)
	}

	// Return nil instead of empty maps/slices
	if len(setWeights) == 0 {
		setWeights = nil
	}
	if len(unsetKeys) == 0 {
		unsetKeys = nil
	}

	return setWeights, unsetKeys, nil
}

// TaskQueueConfigSetCommand handles setting task queue configuration
func (c *TemporalTaskQueueConfigSetCommand) run(cctx *CommandContext, args []string) error {
	// Validate inputs before dialing client
	taskQueue := strings.TrimSpace(c.TaskQueue)
	if taskQueue == "" {
		return fmt.Errorf("task queue name is required and cannot be empty")
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
	// Returns (rateLimit, isZero, error)
	parseRPS := func(flagName string) (*taskqueue.RateLimit, bool, error) {
		raw := strings.TrimSpace(c.Command.Flags().Lookup(flagName).Value.String())
		if raw == "" {
			return nil, false, fmt.Errorf("invalid value for --%s: must be a non-negative number or 'default'", flagName)
		}
		if strings.EqualFold(raw, "default") {
			// Unset: returning nil RateLimit removes the existing rate limit.
			return nil, false, nil
		}
		v, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return nil, false, fmt.Errorf("invalid value for --%s: must be a non-negative number or 'default'", flagName)
		}
		if v < 0 {
			return nil, false, fmt.Errorf("invalid value for --%s: must be >= 0 or 'default'", flagName)
		}
		isZero := v == 0
		return &taskqueue.RateLimit{RequestsPerSecond: float32(v)}, isZero, nil
	}

	// Parse and validate queue rate limit
	var queueRpsLimitParsed *taskqueue.RateLimit
	var queueRateLimitIsZero bool
	if c.Command.Flags().Changed("queue-rps-limit") {
		var err error
		queueRpsLimitParsed, queueRateLimitIsZero, err = parseRPS("queue-rps-limit")
		if err != nil {
			return err
		}

		// Warn about zero rate limit (stops all traffic)
		if queueRateLimitIsZero {
			printZeroRateLimitWarning(cctx, "queue rate limit")
		}
	}

	// Parse and validate fairness key rate limit default
	var fairnessKeyRpsLimitDefaultParsed *taskqueue.RateLimit
	var fairnessRateLimitIsZero bool
	if c.Command.Flags().Changed("fairness-key-rps-limit-default") {
		var err error
		fairnessKeyRpsLimitDefaultParsed, fairnessRateLimitIsZero, err = parseRPS("fairness-key-rps-limit-default")
		if err != nil {
			return err
		}

		// Warn about zero rate limit
		if fairnessRateLimitIsZero {
			printZeroRateLimitWarning(cctx, "fairness key rate limit default")
		}
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

	// Validate at least one configuration change is requested
	hasAnyUpdate := c.Command.Flags().Changed("queue-rps-limit") ||
		c.Command.Flags().Changed("fairness-key-rps-limit-default") ||
		len(c.FairnessKeyWeight) > 0 ||
		c.FairnessKeyWeightClearAll

	if !hasAnyUpdate {
		return fmt.Errorf("at least one configuration update must be specified (use --help to see available options)")
	}

	// Handle fairness weight overrides
	// Validate mutual exclusivity of clear-all with other weight operations
	if c.FairnessKeyWeightClearAll {
		if len(c.FairnessKeyWeight) > 0 {
			return fmt.Errorf("--fairness-key-weight-clear-all cannot be used with --fairness-key-weight")
		}
	}

	// Parse fairness key weights (handles both set and unset operations)
	setWeights, unsetKeys, err := parseFairnessKeyWeights(c.FairnessKeyWeight)
	if err != nil {
		return err
	}
	request.SetFairnessWeightOverrides = setWeights
	request.UnsetFairnessWeightOverrides = unsetKeys

	// Handle clear all
	if c.FairnessKeyWeightClearAll {
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
			return fmt.Errorf("error fetching current config for clear-all: %w", err)
		}
		var overrides map[string]float32
		if descResp.Config != nil {
			overrides = descResp.Config.FairnessWeightOverrides
		}
		keys := maps.Keys(overrides)
		if len(keys) > 0 {
			request.UnsetFairnessWeightOverrides = keys
			cctx.Printer.Printlnf("Unsetting %d fairness weight override(s)", len(keys))
		} else {
			cctx.Printer.Println("No fairness weight overrides found to unset")
			// Don't return error, just proceed with no-op update
		}
	}

	// Call the API
	resp, err := cl.WorkflowService().UpdateTaskQueueConfig(cctx, request)
	if err != nil {
		// Provide more context in error message
		return fmt.Errorf("failed to update task queue config for %s/%s: %w", namespace, taskQueue, err)
	}

	cctx.Printer.Println("Successfully updated task queue configuration")
	return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
}
