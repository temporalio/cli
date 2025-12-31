package workflows

import (
	"time"

	"github.com/temporalio/cli/examples/debug-loop/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// OrderInput is the input for the order processing workflow
type OrderInput struct {
	OrderID string
	Items   []string // SKUs
	Amount  float64
}

// OrderResult is the result of the order processing workflow
type OrderResult struct {
	OrderID       string
	PaymentID     string
	ItemsVerified int
	Status        string
}

// ProcessOrderWorkflow processes an order by:
// 1. Processing payment
// 2. Checking inventory for each item
//
// BUG: The retry policy for inventory check has MaximumAttempts: 2,
// but the inventory service needs 3 attempts to succeed.
// This causes the workflow to fail with "retry limit exceeded".
func ProcessOrderWorkflow(ctx workflow.Context, input OrderInput) (*OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Processing order", "orderID", input.OrderID, "items", len(input.Items))

	// Step 1: Process payment (this always succeeds)
	paymentOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	paymentCtx := workflow.WithActivityOptions(ctx, paymentOpts)

	var paymentResult activities.PaymentResult
	err := workflow.ExecuteActivity(paymentCtx, activities.ProcessPayment, activities.PaymentInput{
		OrderID: input.OrderID,
		Amount:  input.Amount,
	}).Get(ctx, &paymentResult)
	if err != nil {
		return nil, err
	}

	logger.Info("Payment processed", "transactionID", paymentResult.TransactionID)

	// Step 2: Check inventory for each item
	// BUG: MaximumAttempts is 2, but the inventory service needs 3 attempts
	inventoryOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 1.5,
			MaximumAttempts:    2, // BUG: Should be 3 or more!
		},
	}
	inventoryCtx := workflow.WithActivityOptions(ctx, inventoryOpts)

	itemsVerified := 0
	for _, sku := range input.Items {
		var inventoryResult activities.InventoryCheckResult
		err := workflow.ExecuteActivity(inventoryCtx, activities.CheckInventory, activities.InventoryCheckInput{
			OrderID: input.OrderID,
			SKU:     sku,
		}).Get(ctx, &inventoryResult)
		if err != nil {
			logger.Error("Inventory check failed", "sku", sku, "error", err)
			return nil, err
		}
		itemsVerified++
		logger.Info("Inventory verified", "sku", sku, "available", inventoryResult.Available)
	}

	return &OrderResult{
		OrderID:       input.OrderID,
		PaymentID:     paymentResult.TransactionID,
		ItemsVerified: itemsVerified,
		Status:        "completed",
	}, nil
}
