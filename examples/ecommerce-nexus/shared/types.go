package shared

import "time"

// OrderInput is the input for the order saga workflow
type OrderInput struct {
	OrderID    string      `json:"order_id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// OrderResult is the result of the order saga
type OrderResult struct {
	OrderID      string    `json:"order_id"`
	Status       string    `json:"status"`
	PaymentID    string    `json:"payment_id,omitempty"`
	ShipmentID   string    `json:"shipment_id,omitempty"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
	FailureStage string    `json:"failure_stage,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// PaymentInput is the input for payment processing
type PaymentInput struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
	CardToken  string  `json:"card_token"`
}

// PaymentResult is the result of payment processing
type PaymentResult struct {
	PaymentID     string    `json:"payment_id"`
	Status        string    `json:"status"` // approved, declined, fraud_detected
	TransactionID string    `json:"transaction_id,omitempty"`
	ProcessedAt   time.Time `json:"processed_at"`
	Error         string    `json:"error,omitempty"`
}

// InventoryInput is the input for inventory reservation
type InventoryInput struct {
	OrderID string      `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

// InventoryResult is the result of inventory operations
type InventoryResult struct {
	ReservationID string    `json:"reservation_id"`
	Status        string    `json:"status"` // reserved, partial, unavailable
	ReservedAt    time.Time `json:"reserved_at"`
	Error         string    `json:"error,omitempty"`
}

// ShippingInput is the input for shipping
type ShippingInput struct {
	OrderID  string `json:"order_id"`
	Address  string `json:"address"`
	Carrier  string `json:"carrier"`
	Priority string `json:"priority"`
}

// ShippingResult is the result of shipping operations
type ShippingResult struct {
	ShipmentID  string    `json:"shipment_id"`
	TrackingNum string    `json:"tracking_number"`
	Status      string    `json:"status"` // scheduled, picked_up, in_transit, delivered
	Carrier     string    `json:"carrier"`
	ShippedAt   time.Time `json:"shipped_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// FraudCheckInput is the input for fraud detection
type FraudCheckInput struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
	CardToken  string  `json:"card_token"`
}

// FraudCheckResult is the result of fraud detection
type FraudCheckResult struct {
	RiskScore float64 `json:"risk_score"` // 0.0 - 1.0
	IsFraud   bool    `json:"is_fraud"`
	Reason    string  `json:"reason,omitempty"`
	CheckedAt string  `json:"checked_at"`
}

// Task Queue names
const (
	CommerceTaskQueue  = "commerce-tasks"
	FinanceTaskQueue   = "finance-tasks"
	LogisticsTaskQueue = "logistics-tasks"
)

// Nexus service and operation names
const (
	NexusPaymentService = "payment-service"
	NexusProcessPayment = "ProcessPayment"
	NexusRefundPayment  = "RefundPayment"
)
