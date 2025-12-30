package workflows

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/temporalio/cli/examples/ecommerce-nexus/chaos"
	"github.com/temporalio/cli/examples/ecommerce-nexus/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// OrderSagaWorkflow orchestrates the complete order process
// Uses Nexus for payment (cross-namespace) and child workflow for shipping
func OrderSagaWorkflow(ctx workflow.Context, input shared.OrderInput) (shared.OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderSagaWorkflow started", "orderID", input.OrderID)

	result := shared.OrderResult{
		OrderID: input.OrderID,
		Status:  "processing",
	}

	// Step 1: Reserve inventory (same namespace, child workflow)
	logger.Info("Step 1: Reserving inventory")
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("inventory-%s", input.OrderID),
	})

	var inventoryResult shared.InventoryResult
	err := workflow.ExecuteChildWorkflow(childCtx, ReserveInventoryWorkflow, shared.InventoryInput{
		OrderID: input.OrderID,
		Items:   input.Items,
	}).Get(ctx, &inventoryResult)
	if err != nil {
		result.Status = "failed"
		result.FailureStage = "inventory"
		result.Error = fmt.Sprintf("inventory reservation failed: %v", err)
		return result, err
	}
	logger.Info("Inventory reserved", "reservationID", inventoryResult.ReservationID)

	// Step 2: Process payment via Nexus (cross-namespace)
	logger.Info("Step 2: Processing payment via Nexus")

	// Get the Nexus endpoint name (configured in Temporal Cloud/Server)
	nexusEndpoint := os.Getenv("NEXUS_PAYMENT_ENDPOINT")
	if nexusEndpoint == "" {
		nexusEndpoint = "payment-endpoint" // Default endpoint name
	}

	// Create Nexus client for payment service
	nexusClient := workflow.NewNexusClient(nexusEndpoint, shared.NexusPaymentService)

	paymentInput := shared.PaymentInput{
		OrderID:    input.OrderID,
		CustomerID: input.CustomerID,
		Amount:     input.TotalPrice,
		CardToken:  getCardToken(input), // Extract from order or use default
	}

	// Execute payment via Nexus
	paymentFuture := nexusClient.ExecuteOperation(ctx, shared.NexusProcessPayment, paymentInput, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 2 * time.Minute,
	})

	var paymentResult shared.PaymentResult
	if err := paymentFuture.Get(ctx, &paymentResult); err != nil {
		// Payment failed - compensate by releasing inventory
		logger.Error("Payment failed, compensating", "error", err)

		// Compensation: release inventory
		compensateCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
		})
		_ = workflow.ExecuteActivity(compensateCtx, ReleaseInventoryActivity, inventoryResult.ReservationID).Get(ctx, nil)

		result.Status = "failed"
		result.FailureStage = "payment"
		result.Error = fmt.Sprintf("payment failed: %v", err)
		return result, fmt.Errorf("payment failed: %w", err)
	}

	result.PaymentID = paymentResult.PaymentID
	logger.Info("Payment processed", "paymentID", paymentResult.PaymentID)

	// Step 3: Ship order via cross-namespace child workflow
	logger.Info("Step 3: Shipping order via child workflow")

	// Get logistics namespace from environment
	logisticsNS := os.Getenv("LOGISTICS_NS")
	if logisticsNS == "" {
		logisticsNS = "default"
	}

	shippingChildCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("shipping-%s", input.OrderID),
		Namespace:  logisticsNS,
		TaskQueue:  shared.LogisticsTaskQueue,
	})

	var shippingResult shared.ShippingResult
	err = workflow.ExecuteChildWorkflow(shippingChildCtx, "ShipOrderWorkflow", shared.ShippingInput{
		OrderID:  input.OrderID,
		Address:  getShippingAddress(input),
		Carrier:  "UPS",
		Priority: "standard",
	}).Get(ctx, &shippingResult)
	if err != nil {
		// Shipping failed - compensate by refunding payment
		logger.Error("Shipping failed, compensating", "error", err)

		// Compensation: refund payment via Nexus
		refundInput := map[string]interface{}{
			"payment_id": paymentResult.PaymentID,
			"order_id":   input.OrderID,
			"amount":     input.TotalPrice,
			"reason":     "shipping_failed",
		}
		refundFuture := nexusClient.ExecuteOperation(ctx, shared.NexusRefundPayment, refundInput, workflow.NexusOperationOptions{
			ScheduleToCloseTimeout: 2 * time.Minute,
		})
		_ = refundFuture.Get(ctx, nil) // Best effort refund

		// Also release inventory
		compensateCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
		})
		_ = workflow.ExecuteActivity(compensateCtx, ReleaseInventoryActivity, inventoryResult.ReservationID).Get(ctx, nil)

		result.Status = "failed"
		result.FailureStage = "shipping"
		result.Error = fmt.Sprintf("shipping failed: %v", err)
		return result, fmt.Errorf("shipping failed: %w", err)
	}

	result.ShipmentID = shippingResult.ShipmentID
	result.Status = "completed"
	result.CompletedAt = time.Now()

	logger.Info("OrderSagaWorkflow completed successfully",
		"orderID", input.OrderID,
		"paymentID", result.PaymentID,
		"shipmentID", result.ShipmentID)

	return result, nil
}

// ReserveInventoryWorkflow handles inventory reservation
func ReserveInventoryWorkflow(ctx workflow.Context, input shared.InventoryInput) (shared.InventoryResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ReserveInventoryWorkflow started", "orderID", input.OrderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result shared.InventoryResult
	err := workflow.ExecuteActivity(ctx, ReserveInventoryActivity, input).Get(ctx, &result)
	if err != nil {
		return shared.InventoryResult{
			Status: "failed",
			Error:  err.Error(),
		}, err
	}

	logger.Info("ReserveInventoryWorkflow completed", "reservationID", result.ReservationID)
	return result, nil
}

// --- Activities ---

// ReserveInventoryActivity reserves items from inventory
func ReserveInventoryActivity(ctx context.Context, input shared.InventoryInput) (shared.InventoryResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ReserveInventoryActivity started", "orderID", input.OrderID, "itemCount", len(input.Items))

	// Chaos injection
	if err := chaos.MaybeInject(ctx, "commerce", "ReserveInventory"); err != nil {
		return shared.InventoryResult{}, err
	}

	// Simulate inventory check
	time.Sleep(300 * time.Millisecond)

	// Check for out-of-stock scenarios
	for _, item := range input.Items {
		if strings.HasSuffix(item.SKU, "OOS") {
			return shared.InventoryResult{
				Status: "unavailable",
				Error:  fmt.Sprintf("item %s (%s) is out of stock", item.Name, item.SKU),
			}, errors.New("inventory unavailable: " + item.SKU + " out of stock")
		}
	}

	return shared.InventoryResult{
		ReservationID: fmt.Sprintf("RES-%s-%d", input.OrderID, time.Now().Unix()),
		Status:        "reserved",
		ReservedAt:    time.Now(),
	}, nil
}

// ReleaseInventoryActivity releases a previous reservation (compensation)
func ReleaseInventoryActivity(ctx context.Context, reservationID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("ReleaseInventoryActivity started", "reservationID", reservationID)

	// Simulate release
	time.Sleep(200 * time.Millisecond)

	logger.Info("Inventory released", "reservationID", reservationID)
	return nil
}

// Helper functions
func getCardToken(input shared.OrderInput) string {
	// In a real app, this would come from the order or customer session
	// For testing, we use the customer ID suffix to trigger different scenarios
	if strings.HasSuffix(input.CustomerID, "FRAUD") {
		return "tok_FRAUD"
	}
	if strings.HasSuffix(input.CustomerID, "DECLINED") {
		return "tok_DECLINED"
	}
	if strings.HasSuffix(input.CustomerID, "TIMEOUT") {
		return "tok_TIMEOUT"
	}
	return "tok_valid_card"
}

func getShippingAddress(input shared.OrderInput) string {
	// In a real app, this would come from the order
	return fmt.Sprintf("123 Main St, Customer %s", input.CustomerID)
}
