package workflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// OrderWorkflow represents a main order processing workflow
func OrderWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderWorkflow started", "orderID", orderID)

	// Execute payment child workflow
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("payment-%s", orderID),
	})

	var paymentResult string
	err := workflow.ExecuteChildWorkflow(childCtx, PaymentWorkflow, orderID).Get(ctx, &paymentResult)
	if err != nil {
		logger.Error("Payment failed", "error", err)
		return fmt.Errorf("order failed: payment error: %w", err)
	}

	// Execute shipping child workflow
	childCtx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("shipping-%s", orderID),
	})

	var shippingResult string
	err = workflow.ExecuteChildWorkflow(childCtx, ShippingWorkflow, orderID).Get(ctx, &shippingResult)
	if err != nil {
		logger.Error("Shipping failed", "error", err)
		return fmt.Errorf("order failed: shipping error: %w", err)
	}

	logger.Info("OrderWorkflow completed successfully", "orderID", orderID)
	return nil
}

// PaymentWorkflow handles payment processing
func PaymentWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("PaymentWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, ProcessPaymentActivity, orderID).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	logger.Info("PaymentWorkflow completed", "orderID", orderID, "result", result)
	return result, nil
}

// ShippingWorkflow handles shipping logistics
func ShippingWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ShippingWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, ShipOrderActivity, orderID).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	logger.Info("ShippingWorkflow completed", "orderID", orderID, "result", result)
	return result, nil
}

// NestedFailureWorkflow demonstrates deep failure chains
func NestedFailureWorkflow(ctx workflow.Context, depth int, maxDepth int) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("NestedFailureWorkflow started", "depth", depth, "maxDepth", maxDepth)

	if depth >= maxDepth {
		// Deepest level - execute an activity that will fail
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 1,
			},
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		var result string
		err := workflow.ExecuteActivity(ctx, FailingActivity, depth).Get(ctx, &result)
		if err != nil {
			return fmt.Errorf("leaf workflow failed at depth %d: %w", depth, err)
		}
		return nil
	}

	// Not at max depth - spawn a child workflow
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("nested-level-%d", depth+1),
	})

	err := workflow.ExecuteChildWorkflow(childCtx, NestedFailureWorkflow, depth+1, maxDepth).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("child workflow at depth %d failed: %w", depth, err)
	}

	return nil
}

// SimpleSuccessWorkflow is a basic successful workflow
func SimpleSuccessWorkflow(ctx workflow.Context, input string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SimpleSuccessWorkflow started", "input", input)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SuccessActivity, input).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	logger.Info("SimpleSuccessWorkflow completed", "result", result)
	return result, nil
}

// TimeoutWorkflow demonstrates activity timeout failures
func TimeoutWorkflow(ctx workflow.Context, taskID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("TimeoutWorkflow started", "taskID", taskID)

	// Set a very short timeout that will be exceeded
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // No retries - fail immediately on timeout
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SlowActivity, taskID).Get(ctx, &result)
	if err != nil {
		return fmt.Errorf("timeout workflow failed: %w", err)
	}

	logger.Info("TimeoutWorkflow completed", "taskID", taskID, "result", result)
	return nil
}

// RetryExhaustionWorkflow demonstrates retry exhaustion failures
func RetryExhaustionWorkflow(ctx workflow.Context, taskID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("RetryExhaustionWorkflow started", "taskID", taskID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        5,
			InitialInterval:        100 * time.Millisecond,
			MaximumInterval:        500 * time.Millisecond,
			BackoffCoefficient:     1.5,
			NonRetryableErrorTypes: []string{}, // All errors are retryable
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, AlwaysFailsActivity, taskID).Get(ctx, &result)
	if err != nil {
		return fmt.Errorf("retry exhaustion: all %d attempts failed: %w", 5, err)
	}

	logger.Info("RetryExhaustionWorkflow completed", "taskID", taskID, "result", result)
	return nil
}

// MultiChildFailureWorkflow spawns multiple children, only one fails
// This tests the agent's ability to identify which branch failed
func MultiChildFailureWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("MultiChildFailureWorkflow started", "orderID", orderID)

	// Execute multiple child workflows in parallel
	// Only the "validation" child will fail

	// Child 1: Inventory check (succeeds)
	inventoryCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("inventory-check-%s", orderID),
	})
	inventoryFuture := workflow.ExecuteChildWorkflow(inventoryCtx, InventoryCheckWorkflow, orderID)

	// Child 2: Validation (fails)
	validationCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("validation-%s", orderID),
	})
	validationFuture := workflow.ExecuteChildWorkflow(validationCtx, ValidationWorkflow, orderID)

	// Child 3: Pricing (succeeds)
	pricingCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("pricing-%s", orderID),
	})
	pricingFuture := workflow.ExecuteChildWorkflow(pricingCtx, PricingWorkflow, orderID)

	// Wait for all children
	var inventoryResult, validationResult, pricingResult string

	if err := inventoryFuture.Get(ctx, &inventoryResult); err != nil {
		return fmt.Errorf("inventory check failed: %w", err)
	}
	logger.Info("Inventory check completed", "result", inventoryResult)

	if err := validationFuture.Get(ctx, &validationResult); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	logger.Info("Validation completed", "result", validationResult)

	if err := pricingFuture.Get(ctx, &pricingResult); err != nil {
		return fmt.Errorf("pricing failed: %w", err)
	}
	logger.Info("Pricing completed", "result", pricingResult)

	logger.Info("MultiChildFailureWorkflow completed successfully", "orderID", orderID)
	return nil
}

// InventoryCheckWorkflow - always succeeds
func InventoryCheckWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("InventoryCheckWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SuccessActivity, "inventory-"+orderID).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// ValidationWorkflow - fails with a specific validation error
func ValidationWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ValidationWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, ValidationActivity, orderID).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// PricingWorkflow - always succeeds
func PricingWorkflow(ctx workflow.Context, orderID string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("PricingWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, SuccessActivity, "pricing-"+orderID).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// --- Activities ---

func ProcessPaymentActivity(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Processing payment", "orderID", orderID)

	// Simulate some work
	time.Sleep(500 * time.Millisecond)

	// Simulate failure for certain order IDs
	if len(orderID) > 0 && orderID[len(orderID)-1] == 'X' {
		return "", errors.New("payment gateway connection timeout")
	}

	return fmt.Sprintf("payment-confirmed-%s", orderID), nil
}

func ShipOrderActivity(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Shipping order", "orderID", orderID)

	time.Sleep(300 * time.Millisecond)

	// Simulate failure for certain order IDs
	if len(orderID) > 0 && orderID[len(orderID)-1] == 'Y' {
		return "", errors.New("warehouse inventory depleted")
	}

	return fmt.Sprintf("shipped-%s", orderID), nil
}

func FailingActivity(ctx context.Context, depth int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("FailingActivity executing at deepest level", "depth", depth)

	time.Sleep(200 * time.Millisecond)

	return "", fmt.Errorf("critical failure at depth %d: database connection refused", depth)
}

func SuccessActivity(ctx context.Context, input string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SuccessActivity executing", "input", input)

	time.Sleep(100 * time.Millisecond)

	return fmt.Sprintf("processed: %s", input), nil
}

// SlowActivity takes longer than typical timeouts - used to trigger timeout failures
func SlowActivity(ctx context.Context, taskID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SlowActivity started - will take 5 seconds", "taskID", taskID)

	// Sleep for 5 seconds - longer than the 2 second timeout in TimeoutWorkflow
	time.Sleep(5 * time.Second)

	return fmt.Sprintf("slow-completed-%s", taskID), nil
}

// AlwaysFailsActivity always returns an error - used for retry exhaustion testing
func AlwaysFailsActivity(ctx context.Context, taskID string) (string, error) {
	logger := activity.GetLogger(ctx)
	info := activity.GetInfo(ctx)
	attempt := info.Attempt

	logger.Info("AlwaysFailsActivity executing", "taskID", taskID, "attempt", attempt)

	time.Sleep(50 * time.Millisecond)

	return "", fmt.Errorf("transient error on attempt %d: service temporarily unavailable", attempt)
}

// ValidationActivity always fails with a validation error
func ValidationActivity(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ValidationActivity executing", "orderID", orderID)

	time.Sleep(100 * time.Millisecond)

	return "", errors.New("validation failed: order contains invalid product SKU 'INVALID-123'")
}

// --- Long-running workflows for state demo ---

// LongRunningWorkflow runs for ~30 seconds with visible pending activities
// Use this to demo `temporal agent state` showing pending activities
func LongRunningWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("LongRunningWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Long activity (10 seconds)
	logger.Info("Starting Step 1: LongActivity")
	var result string
	if err := workflow.ExecuteActivity(ctx, LongActivity, "step1-"+orderID, 10).Get(ctx, &result); err != nil {
		return fmt.Errorf("step 1 failed: %w", err)
	}
	logger.Info("Step 1 completed", "result", result)

	// Step 2: Another long activity (10 seconds)
	logger.Info("Starting Step 2: LongActivity")
	if err := workflow.ExecuteActivity(ctx, LongActivity, "step2-"+orderID, 10).Get(ctx, &result); err != nil {
		return fmt.Errorf("step 2 failed: %w", err)
	}
	logger.Info("Step 2 completed", "result", result)

	// Step 3: Spawn child workflow that also takes time
	logger.Info("Starting Step 3: LongChildWorkflow")
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("long-child-%s", orderID),
	})
	if err := workflow.ExecuteChildWorkflow(childCtx, LongChildWorkflow, orderID).Get(ctx, nil); err != nil {
		return fmt.Errorf("step 3 child workflow failed: %w", err)
	}
	logger.Info("Step 3 completed")

	logger.Info("LongRunningWorkflow completed successfully", "orderID", orderID)
	return nil
}

// LongChildWorkflow is a child workflow that runs for ~10 seconds
func LongChildWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("LongChildWorkflow started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	if err := workflow.ExecuteActivity(ctx, LongActivity, "child-"+orderID, 8).Get(ctx, &result); err != nil {
		return err
	}

	logger.Info("LongChildWorkflow completed", "orderID", orderID)
	return nil
}

// LongActivity runs for a configurable number of seconds
func LongActivity(ctx context.Context, taskID string, durationSeconds int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("LongActivity started", "taskID", taskID, "durationSeconds", durationSeconds)

	// Heartbeat every 2 seconds
	for i := 0; i < durationSeconds; i++ {
		time.Sleep(1 * time.Second)
		activity.RecordHeartbeat(ctx, fmt.Sprintf("progress: %d/%d", i+1, durationSeconds))
	}

	logger.Info("LongActivity completed", "taskID", taskID)
	return fmt.Sprintf("completed-%s-after-%ds", taskID, durationSeconds), nil
}

// --- Deep chain workflows for trace demo ---

// DeepChainWorkflow creates a 4-level deep failure chain
func DeepChainWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("DeepChainWorkflow (Level 1) started", "orderID", orderID)

	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("deep-level2-%s", orderID),
	})

	if err := workflow.ExecuteChildWorkflow(childCtx, DeepLevel2Workflow, orderID).Get(ctx, nil); err != nil {
		return fmt.Errorf("level 1 failed: %w", err)
	}

	return nil
}

// DeepLevel2Workflow is level 2 of the deep chain
func DeepLevel2Workflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("DeepLevel2Workflow started", "orderID", orderID)

	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("deep-level3-%s", orderID),
	})

	if err := workflow.ExecuteChildWorkflow(childCtx, DeepLevel3Workflow, orderID).Get(ctx, nil); err != nil {
		return fmt.Errorf("level 2 failed: %w", err)
	}

	return nil
}

// DeepLevel3Workflow is level 3 of the deep chain
func DeepLevel3Workflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("DeepLevel3Workflow started", "orderID", orderID)

	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("deep-level4-%s", orderID),
	})

	if err := workflow.ExecuteChildWorkflow(childCtx, DeepLevel4Workflow, orderID).Get(ctx, nil); err != nil {
		return fmt.Errorf("level 3 failed: %w", err)
	}

	return nil
}

// DeepLevel4Workflow is the leaf - fails with a database error
func DeepLevel4Workflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("DeepLevel4Workflow (LEAF) started", "orderID", orderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	if err := workflow.ExecuteActivity(ctx, DatabaseFailureActivity, orderID).Get(ctx, &result); err != nil {
		return fmt.Errorf("database operation failed: %w", err)
	}

	return nil
}

// DatabaseFailureActivity simulates a database connection failure
func DatabaseFailureActivity(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("DatabaseFailureActivity executing", "orderID", orderID)

	time.Sleep(100 * time.Millisecond)

	return "", errors.New("FATAL: connection to database 'orders_db' failed: ECONNREFUSED 10.0.1.5:5432")
}
