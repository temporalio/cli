package workflows

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/temporalio/cli/examples/ecommerce-nexus/chaos"
	"github.com/temporalio/cli/examples/ecommerce-nexus/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ShipOrderWorkflow handles order shipping with tracking
func ShipOrderWorkflow(ctx workflow.Context, input shared.ShippingInput) (shared.ShippingResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ShipOrderWorkflow started", "orderID", input.OrderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Create shipment
	var shipmentResult shared.ShippingResult
	err := workflow.ExecuteActivity(ctx, CreateShipmentActivity, input).Get(ctx, &shipmentResult)
	if err != nil {
		return shared.ShippingResult{
			Status: "failed",
			Error:  fmt.Sprintf("failed to create shipment: %v", err),
		}, err
	}

	logger.Info("Shipment created", "shipmentID", shipmentResult.ShipmentID)

	// Step 2: Start tracking workflow as child
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("tracking-%s", shipmentResult.ShipmentID),
	})

	var trackingResult TrackingResult
	err = workflow.ExecuteChildWorkflow(childCtx, TrackShipmentWorkflow, TrackingInput{
		ShipmentID:  shipmentResult.ShipmentID,
		TrackingNum: shipmentResult.TrackingNum,
		Carrier:     shipmentResult.Carrier,
	}).Get(ctx, &trackingResult)
	if err != nil {
		// Tracking failure doesn't fail the shipment, just log it
		logger.Warn("Tracking workflow failed", "error", err)
	}

	shipmentResult.Status = "shipped"
	shipmentResult.ShippedAt = time.Now()

	logger.Info("ShipOrderWorkflow completed", "shipmentID", shipmentResult.ShipmentID)
	return shipmentResult, nil
}

// TrackingInput is the input for shipment tracking
type TrackingInput struct {
	ShipmentID  string `json:"shipment_id"`
	TrackingNum string `json:"tracking_number"`
	Carrier     string `json:"carrier"`
}

// TrackingResult is the result of shipment tracking
type TrackingResult struct {
	ShipmentID   string   `json:"shipment_id"`
	Status       string   `json:"status"`
	Locations    []string `json:"locations"`
	DeliveredAt  string   `json:"delivered_at,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// TrackShipmentWorkflow tracks a shipment until delivery
func TrackShipmentWorkflow(ctx workflow.Context, input TrackingInput) (TrackingResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TrackShipmentWorkflow started", "shipmentID", input.ShipmentID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := TrackingResult{
		ShipmentID: input.ShipmentID,
		Status:     "tracking",
		Locations:  []string{},
	}

	// Poll for tracking updates (simplified - in real app would use signals)
	for i := 0; i < 3; i++ {
		var location string
		err := workflow.ExecuteActivity(ctx, GetTrackingUpdateActivity, input).Get(ctx, &location)
		if err != nil {
			result.Error = err.Error()
			result.Status = "tracking_failed"
			return result, err
		}
		result.Locations = append(result.Locations, location)

		// Sleep between updates
		workflow.Sleep(ctx, 5*time.Second)
	}

	result.Status = "delivered"
	result.DeliveredAt = time.Now().Format(time.RFC3339)

	logger.Info("TrackShipmentWorkflow completed", "shipmentID", input.ShipmentID)
	return result, nil
}

// --- Activities ---

// CreateShipmentActivity creates a shipment with the carrier
func CreateShipmentActivity(ctx context.Context, input shared.ShippingInput) (shared.ShippingResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateShipmentActivity started", "orderID", input.OrderID)

	// Chaos injection
	if err := chaos.MaybeInject(ctx, "logistics", "CreateShipment"); err != nil {
		return shared.ShippingResult{}, err
	}

	// Simulate carrier API call
	time.Sleep(500 * time.Millisecond)

	// Check for failure scenarios
	if strings.Contains(input.Address, "CARRIER_DOWN") {
		return shared.ShippingResult{}, errors.New("carrier API unavailable: connection refused to ups.com:443")
	}

	if strings.Contains(input.Address, "INVALID_ADDRESS") {
		return shared.ShippingResult{}, errors.New("address validation failed: address not deliverable")
	}

	if strings.Contains(input.Address, "SLOW") {
		// Simulate very slow processing
		time.Sleep(65 * time.Second) // Trigger timeout
		return shared.ShippingResult{}, errors.New("carrier API timeout")
	}

	shipmentID := fmt.Sprintf("SHIP-%s-%d", input.OrderID, time.Now().Unix())

	return shared.ShippingResult{
		ShipmentID:  shipmentID,
		TrackingNum: fmt.Sprintf("1Z%s%d", input.Carrier, time.Now().UnixNano()%1000000),
		Status:      "created",
		Carrier:     input.Carrier,
	}, nil
}

// GetTrackingUpdateActivity gets the current location of a shipment
func GetTrackingUpdateActivity(ctx context.Context, input TrackingInput) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GetTrackingUpdateActivity started", "trackingNum", input.TrackingNum)

	// Chaos injection
	if err := chaos.MaybeInject(ctx, "logistics", "GetTrackingUpdate"); err != nil {
		return "", err
	}

	// Simulate carrier tracking API
	time.Sleep(200 * time.Millisecond)

	if strings.Contains(input.TrackingNum, "TRACKING_ERROR") {
		return "", errors.New("tracking API error: package not found in system")
	}

	locations := []string{
		"Picked up at origin facility",
		"In transit to regional hub",
		"Arrived at regional hub",
		"Out for delivery",
		"Delivered",
	}

	// Random location for simulation
	idx := time.Now().Nanosecond() % len(locations)
	return locations[idx], nil
}

// CancelShipmentActivity cancels a shipment (for compensation)
func CancelShipmentActivity(ctx context.Context, shipmentID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CancelShipmentActivity started", "shipmentID", shipmentID)

	// Simulate cancellation
	time.Sleep(300 * time.Millisecond)

	logger.Info("Shipment cancelled", "shipmentID", shipmentID)
	return nil
}

