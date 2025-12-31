package activities

import (
	"context"
	"fmt"
	"sync"
)

// Simulated inventory database
var (
	inventory = map[string]int{
		"LAPTOP-001":  5,
		"MOUSE-002":   10,
		"KEYBOARD-03": 1, // Only 1 in stock - will cause race condition
	}
	reservations = make(map[string]map[string]int) // orderID -> sku -> quantity
	inventoryMu  sync.Mutex
)

// InventoryCheckInput is the input for the inventory check activity
type InventoryCheckInput struct {
	OrderID  string
	SKU      string
	Quantity int
}

// InventoryCheckResult is the result of the inventory check
type InventoryCheckResult struct {
	SKU       string
	Available bool
	InStock   int
	Requested int
}

// CheckInventory checks if an item is available in inventory.
// This is a point-in-time check - inventory can change before reservation.
func CheckInventory(ctx context.Context, input InventoryCheckInput) (*InventoryCheckResult, error) {
	inventoryMu.Lock()
	defer inventoryMu.Unlock()

	stock, exists := inventory[input.SKU]
	if !exists {
		return &InventoryCheckResult{
			SKU:       input.SKU,
			Available: false,
			InStock:   0,
			Requested: input.Quantity,
		}, nil
	}

	return &InventoryCheckResult{
		SKU:       input.SKU,
		Available: stock >= input.Quantity,
		InStock:   stock,
		Requested: input.Quantity,
	}, nil
}

// ReserveInventoryInput is the input for the reserve inventory activity
type ReserveInventoryInput struct {
	OrderID  string
	SKU      string
	Quantity int
}

// ReserveInventoryResult is the result of the reservation
type ReserveInventoryResult struct {
	SKU         string
	Reserved    bool
	ReservedQty int
}

// ReserveInventory attempts to reserve inventory for an order.
// This can fail if inventory was depleted between check and reserve.
func ReserveInventory(ctx context.Context, input ReserveInventoryInput) (*ReserveInventoryResult, error) {
	inventoryMu.Lock()
	defer inventoryMu.Unlock()

	stock, exists := inventory[input.SKU]
	if !exists {
		return nil, fmt.Errorf("product %s not found", input.SKU)
	}

	if stock < input.Quantity {
		return nil, fmt.Errorf("insufficient inventory for %s: requested %d, available %d",
			input.SKU, input.Quantity, stock)
	}

	// Deduct from inventory
	inventory[input.SKU] = stock - input.Quantity

	// Track reservation for potential rollback
	if reservations[input.OrderID] == nil {
		reservations[input.OrderID] = make(map[string]int)
	}
	reservations[input.OrderID][input.SKU] = input.Quantity

	return &ReserveInventoryResult{
		SKU:         input.SKU,
		Reserved:    true,
		ReservedQty: input.Quantity,
	}, nil
}

// ReleaseInventoryInput is the input for releasing inventory
type ReleaseInventoryInput struct {
	OrderID string
	SKU     string
}

// ReleaseInventory releases reserved inventory back to stock
func ReleaseInventory(ctx context.Context, input ReleaseInventoryInput) error {
	inventoryMu.Lock()
	defer inventoryMu.Unlock()

	if reservations[input.OrderID] == nil {
		return nil // Nothing to release
	}

	qty, exists := reservations[input.OrderID][input.SKU]
	if !exists {
		return nil
	}

	// Return to stock
	inventory[input.SKU] += qty
	delete(reservations[input.OrderID], input.SKU)

	return nil
}

// SimulateExternalReservation simulates another order taking inventory
// This is called to create the race condition
func SimulateExternalReservation(ctx context.Context, sku string, quantity int) error {
	inventoryMu.Lock()
	defer inventoryMu.Unlock()

	stock := inventory[sku]
	if stock >= quantity {
		inventory[sku] = stock - quantity
	}
	return nil
}

// ResetInventory resets inventory to initial state (for testing)
func ResetInventory() {
	inventoryMu.Lock()
	defer inventoryMu.Unlock()
	inventory = map[string]int{
		"LAPTOP-001":  5,
		"MOUSE-002":   10,
		"KEYBOARD-03": 1,
	}
	reservations = make(map[string]map[string]int)
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
