package activities

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
)

// InventoryCheckInput is the input for the inventory check activity
type InventoryCheckInput struct {
	OrderID string
	SKU     string
}

// InventoryCheckResult is the result of the inventory check
type InventoryCheckResult struct {
	SKU       string
	Available bool
	Quantity  int
}

// CheckInventory checks if an item is available in inventory.
// This activity simulates a transient failure that succeeds on the 3rd attempt.
// The bug in the workflow is that MaximumAttempts is set to 2, so it never succeeds.
//
// It uses Temporal's Activity.GetInfo().Attempt to track attempts reliably across
// worker restarts.
func CheckInventory(ctx context.Context, input InventoryCheckInput) (*InventoryCheckResult, error) {
	// Get the current attempt from Temporal (1-based)
	info := activity.GetInfo(ctx)
	attempt := int(info.Attempt)

	// Simulate external inventory service that has transient failures
	// It takes 3 attempts to succeed (simulating a flaky service)
	if attempt < 3 {
		return nil, fmt.Errorf("inventory service unavailable (attempt %d/3)", attempt)
	}

	// On the 3rd attempt, succeed
	return &InventoryCheckResult{
		SKU:       input.SKU,
		Available: true,
		Quantity:  100,
	}, nil
}

// PaymentInput is the input for the payment activity
type PaymentInput struct {
	OrderID string
	Amount  float64
}

// PaymentResult is the result of the payment activity
type PaymentResult struct {
	TransactionID string
	Status        string
}

// ProcessPayment processes a payment (always succeeds for this test)
func ProcessPayment(ctx context.Context, input PaymentInput) (*PaymentResult, error) {
	return &PaymentResult{
		TransactionID: fmt.Sprintf("txn-%s", input.OrderID),
		Status:        "approved",
	}, nil
}
