package workflows

import (
	"fmt"
	"time"

	"github.com/temporalio/cli/examples/debug-loop/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// OrderItem represents an item in the order
type OrderItem struct {
	SKU      string
	Quantity int
	Price    float64
}

// OrderInput is the input for the order processing workflow
type OrderInput struct {
	OrderID string
	Items   []OrderItem
}

// OrderResult is the result of the order processing workflow
type OrderResult struct {
	OrderID       string
	PaymentID     string
	ItemsReserved int
	TotalAmount   float64
	Status        string
}

// ProcessOrderWorkflow processes an order by:
// 1. Checking inventory for ALL items in parallel
// 2. Reserving inventory for each item sequentially
// 3. Processing payment
//
// BUG: The inventory check is done in parallel for all items, but reservation
// is done sequentially. Between check and reserve, another order can take
// the last item (race condition). The workflow doesn't re-check before reserving.
//
// The correct fix is to either:
// - Check and reserve atomically (in one activity)
// - Or re-check inventory just before each reservation
// - Or use a distributed lock
func ProcessOrderWorkflow(ctx workflow.Context, input OrderInput) (*OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Processing order", "orderID", input.OrderID, "items", len(input.Items))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // No retries - we want to see the error
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Check inventory for ALL items in PARALLEL
	// BUG: This creates a window for race conditions
	logger.Info("Checking inventory for all items in parallel")

	checkFutures := make([]workflow.Future, len(input.Items))
	for i, item := range input.Items {
		checkFutures[i] = workflow.ExecuteActivity(ctx, activities.CheckInventory, activities.InventoryCheckInput{
			OrderID:  input.OrderID,
			SKU:      item.SKU,
			Quantity: item.Quantity,
		})
	}

	// Wait for all checks to complete
	checkResults := make([]*activities.InventoryCheckResult, len(input.Items))
	for i, future := range checkFutures {
		var result activities.InventoryCheckResult
		if err := future.Get(ctx, &result); err != nil {
			return nil, fmt.Errorf("inventory check failed for %s: %w", input.Items[i].SKU, err)
		}
		if !result.Available {
			return nil, fmt.Errorf("item %s not available: requested %d, in stock %d",
				result.SKU, result.Requested, result.InStock)
		}
		checkResults[i] = &result
		logger.Info("Inventory check passed", "sku", result.SKU, "inStock", result.InStock)
	}

	logger.Info("All inventory checks passed, proceeding to reserve")

	// Simulate processing/validation delay before reserving
	// BUG: This gap creates a window for race conditions!
	// In real life, this happens due to network latency, DB queries, etc.
	workflow.Sleep(ctx, 200*time.Millisecond)

	// Step 2: Reserve inventory for each item SEQUENTIALLY
	// BUG: Between check and reserve, inventory state may have changed!
	reservedItems := []string{}
	var totalAmount float64

	for _, item := range input.Items {
		var result activities.ReserveInventoryResult
		err := workflow.ExecuteActivity(ctx, activities.ReserveInventory, activities.ReserveInventoryInput{
			OrderID:  input.OrderID,
			SKU:      item.SKU,
			Quantity: item.Quantity,
		}).Get(ctx, &result)

		if err != nil {
			// Compensation: release any items we already reserved
			logger.Error("Reservation failed, releasing reserved items", "failedSKU", item.SKU, "error", err)
			for _, sku := range reservedItems {
				_ = workflow.ExecuteActivity(ctx, activities.ReleaseInventory, activities.ReleaseInventoryInput{
					OrderID: input.OrderID,
					SKU:     sku,
				}).Get(ctx, nil)
			}
			return nil, err
		}

		reservedItems = append(reservedItems, item.SKU)
		totalAmount += item.Price * float64(item.Quantity)
		logger.Info("Reserved inventory", "sku", item.SKU, "quantity", item.Quantity)
	}

	// Step 3: Process payment
	var paymentResult activities.PaymentResult
	err := workflow.ExecuteActivity(ctx, activities.ProcessPayment, activities.PaymentInput{
		OrderID: input.OrderID,
		Amount:  totalAmount,
	}).Get(ctx, &paymentResult)
	if err != nil {
		// Release all reserved items
		for _, sku := range reservedItems {
			_ = workflow.ExecuteActivity(ctx, activities.ReleaseInventory, activities.ReleaseInventoryInput{
				OrderID: input.OrderID,
				SKU:     sku,
			}).Get(ctx, nil)
		}
		return nil, err
	}

	return &OrderResult{
		OrderID:       input.OrderID,
		PaymentID:     paymentResult.TransactionID,
		ItemsReserved: len(reservedItems),
		TotalAmount:   totalAmount,
		Status:        "completed",
	}, nil
}
