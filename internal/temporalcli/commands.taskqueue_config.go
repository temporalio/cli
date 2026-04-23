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

	// minFairnessWeight matches server-side limit in temporal/service/matching/fairness_util.go
	// Client only enforces positive weight - server handles clamping to its configured range
	minFairnessWeight = 0.001
)

// parseFairnessKeyWeights parses "key=weight" format strings into a map
// Returns nil if inputs is empty, or an error if there are duplicate keys, invalid weights, or malformed input
func parseFairnessKeyWeights(inputs []string) (map[string]float32, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	weights := make(map[string]float32)
	seen := make(map[string]bool)

	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %q (expected key=weight)", input)
		}

		key := parts[0]
		if key == "" {
			return nil, fmt.Errorf("empty key in: %q", input)
		}

		// Check for duplicate keys
		if seen[key] {
			return nil, fmt.Errorf("duplicate fairness key %q specified multiple times", key)
		}
		seen[key] = true

		// Validate key length (server enforces 64 byte limit)
		if len(key) > maxFairnessKeyLength {
			return nil, fmt.Errorf("fairness key %q exceeds maximum length of %d bytes", key, maxFairnessKeyLength)
		}

		weightStr := parts[1]
		if weightStr == "" {
			return nil, fmt.Errorf("empty weight value for key %q", key)
		}

		weight, err := strconv.ParseFloat(weightStr, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid weight %q for key %q: must be a number", weightStr, key)
		}

		// Validate weight is positive - server handles clamping to its configured range
		if weight < minFairnessWeight {
			return nil, fmt.Errorf("weight %.3f for key %q is below minimum %.3f", weight, key, minFairnessWeight)
		}

		weights[key] = float32(weight)
	}
	return weights, nil
}

// prepareUnsetKeys validates fairness key names for unsetting
// Returns nil if keys is empty, or an error if validation fails
func prepareUnsetKeys(keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	seen := make(map[string]bool)

	for _, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("empty fairness key name")
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate fairness key %q specified multiple times", key)
		}
		seen[key] = true

		if len(key) > maxFairnessKeyLength {
			return nil, fmt.Errorf("fairness key %q exceeds maximum length of %d bytes", key, maxFairnessKeyLength)
		}
	}
	return keys, nil
}

// findConflictingKeys returns keys that appear in both set and unset lists
func findConflictingKeys(setWeights map[string]float32, unsetKeys []string) []string {
	var conflicts []string
	for _, key := range unsetKeys {
		if _, exists := setWeights[key]; exists {
			conflicts = append(conflicts, key)
		}
	}
	return conflicts
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
			cctx.Printer.Println("WARNING: Setting queue rate limit to 0 will STOP ALL TRAFFIC on this task queue.")
			cctx.Printer.Println("         This will prevent any tasks from being dispatched until the limit is changed.")
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
			cctx.Printer.Println("WARNING: Setting fairness key rate limit default to 0 will STOP ALL TRAFFIC on this task queue.")
			cctx.Printer.Println("         This will prevent any tasks from being dispatched until the limit is changed.")
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
		len(c.FairnessKeyWeightSet) > 0 ||
		len(c.FairnessKeyWeightUnset) > 0 ||
		c.FairnessKeyWeightUnsetAll

	if !hasAnyUpdate {
		return fmt.Errorf("at least one configuration update must be specified (use --help to see available options)")
	}

	// Handle fairness weight overrides
	// Validate mutual exclusivity of unset-all with other weight operations
	if c.FairnessKeyWeightUnsetAll {
		if len(c.FairnessKeyWeightSet) > 0 || len(c.FairnessKeyWeightUnset) > 0 {
			return fmt.Errorf("--fairness-key-weight-unset-all cannot be used with --fairness-key-weight-set or --fairness-key-weight-unset")
		}
	}

	// Parse and validate set operations
	setWeights, err := parseFairnessKeyWeights(c.FairnessKeyWeightSet)
	if err != nil {
		return err
	}
	request.SetFairnessWeightOverrides = setWeights

	// Validate and prepare unset operations
	unsetKeys, err := prepareUnsetKeys(c.FairnessKeyWeightUnset)
	if err != nil {
		return fmt.Errorf("invalid fairness key in unset list: %w", err)
	}

	// Check for conflicts between set and unset
	conflicts := findConflictingKeys(setWeights, unsetKeys)
	if len(conflicts) > 0 {
		return fmt.Errorf("fairness keys appear in both set and unset operations: %v", conflicts)
	}

	request.UnsetFairnessWeightOverrides = unsetKeys

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
		if descResp.Config != nil && descResp.Config.FairnessWeightOverrides != nil && len(descResp.Config.FairnessWeightOverrides) > 0 {
			keys := maps.Keys(descResp.Config.FairnessWeightOverrides)
			request.UnsetFairnessWeightOverrides = keys
			cctx.Printer.Println(fmt.Sprintf("Unsetting %d fairness weight override(s)", len(keys)))
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
