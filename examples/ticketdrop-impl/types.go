package ticketdrop

import "time"

const (
	TaskQueue = "ticketdrop"
)

// PurchaseInput is the input to the TicketPurchase workflow.
type PurchaseInput struct {
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
}

// PurchaseResult is the output from the TicketPurchase workflow.
type PurchaseResult struct {
	ConfirmationID string    `json:"confirmation_id"`
	SeatNumber     string    `json:"seat_number"`
	QRCode         string    `json:"qr_code"`
	PurchasedAt    time.Time `json:"purchased_at"`
}

// ReserveSeatInput is the input to the ReserveSeat activity.
type ReserveSeatInput struct {
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
}

// ReserveSeatResult is the output from the ReserveSeat activity.
type ReserveSeatResult struct {
	ReservationID string    `json:"reservation_id"`
	SeatNumber    string    `json:"seat_number"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// ProcessPaymentInput is the input to the ProcessPayment activity.
type ProcessPaymentInput struct {
	UserID        string `json:"user_id"`
	ReservationID string `json:"reservation_id"`
	Amount        int64  `json:"amount"` // cents
}

// ProcessPaymentResult is the output from the ProcessPayment activity.
type ProcessPaymentResult struct {
	TransactionID string `json:"transaction_id"`
	ChargedAmount int64  `json:"charged_amount"`
}

// IssueTicketInput is the input to the IssueTicket activity.
type IssueTicketInput struct {
	UserID        string `json:"user_id"`
	EventID       string `json:"event_id"`
	SeatNumber    string `json:"seat_number"`
	TransactionID string `json:"transaction_id"`
}

// IssueTicketResult is the output from the IssueTicket activity.
type IssueTicketResult struct {
	TicketID string `json:"ticket_id"`
	QRCode   string `json:"qr_code"`
}

// SendConfirmationInput is the input to the SendConfirmation child workflow.
type SendConfirmationInput struct {
	UserID         string `json:"user_id"`
	EventID        string `json:"event_id"`
	ConfirmationID string `json:"confirmation_id"`
	SeatNumber     string `json:"seat_number"`
	QRCode         string `json:"qr_code"`
}

// SendConfirmationResult is the output from the SendConfirmation child workflow.
type SendConfirmationResult struct {
	EmailSent bool `json:"email_sent"`
	SMSSent   bool `json:"sms_sent"`
}
