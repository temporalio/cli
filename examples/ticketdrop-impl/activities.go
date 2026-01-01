package ticketdrop

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
)

// SeatInventory tracks available seats per event.
type SeatInventory struct {
	mu       sync.Mutex
	seats    map[string][]string // eventID -> available seats
	reserved map[string]string   // seatKey -> userID
}

// NewSeatInventory creates an inventory with 10 seats per event.
func NewSeatInventory() *SeatInventory {
	return &SeatInventory{
		seats:    make(map[string][]string),
		reserved: make(map[string]string),
	}
}

func (inv *SeatInventory) initEvent(eventID string) {
	if _, exists := inv.seats[eventID]; !exists {
		// Initialize 10 seats: A1-A10
		seats := make([]string, 10)
		for i := 0; i < 10; i++ {
			seats[i] = fmt.Sprintf("A%d", i+1)
		}
		inv.seats[eventID] = seats
	}
}

// Reserve attempts to reserve a seat for an event.
func (inv *SeatInventory) Reserve(eventID, userID string) (string, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.initEvent(eventID)

	available := inv.seats[eventID]
	if len(available) == 0 {
		return "", errors.New("sold out: no seats available")
	}

	// Take the first available seat
	seat := available[0]
	inv.seats[eventID] = available[1:]

	// Track reservation
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)
	inv.reserved[seatKey] = userID

	return seat, nil
}

// Available returns the count of available seats for an event.
func (inv *SeatInventory) Available(eventID string) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.initEvent(eventID)
	return len(inv.seats[eventID])
}

type Activities struct {
	Inventory *SeatInventory
}

// ReserveSeat locks a seat for 5 minutes.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID,
		"available", a.Inventory.Available(input.EventID))

	// Simulate seat reservation by sleeping 1 second
	time.Sleep(1 * time.Second)

	// Try to reserve a seat from inventory
	seatNumber, err := a.Inventory.Reserve(input.EventID, input.UserID)
	if err != nil {
		logger.Warn("Reservation failed", "error", err)
		return ReserveSeatResult{}, err
	}

	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())
	logger.Info("Seat reserved", "seat", seatNumber, "remaining", a.Inventory.Available(input.EventID))

	return ReserveSeatResult{
		ReservationID: reservationID,
		SeatNumber:    seatNumber,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}, nil
}

// ProcessPayment charges the credit card.
func (a *Activities) ProcessPayment(ctx context.Context, input ProcessPaymentInput) (ProcessPaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Processing payment", "user_id", input.UserID, "amount", input.Amount)

	// Simulate payment processing (2 seconds)
	time.Sleep(2 * time.Second)

	// 20% random failure rate
	if rand.Float64() < 0.2 {
		logger.Warn("Payment failed", "user_id", input.UserID)
		return ProcessPaymentResult{}, errors.New("payment declined: insufficient funds")
	}

	transactionID := fmt.Sprintf("pay-%s-%d", input.UserID, time.Now().UnixMilli())
	logger.Info("Payment successful", "transaction_id", transactionID)

	return ProcessPaymentResult{
		TransactionID: transactionID,
		ChargedAmount: input.Amount,
	}, nil
}

// IssueTicket generates a QR code for the ticket.
func (a *Activities) IssueTicket(ctx context.Context, input IssueTicketInput) (IssueTicketResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Issuing ticket", "user_id", input.UserID, "seat", input.SeatNumber)

	// Simulate ticket issuance with QR code generation
	ticketID := fmt.Sprintf("tkt-%s-%s-%d", input.EventID, input.SeatNumber, time.Now().UnixMilli())
	qrCode := fmt.Sprintf("QR:%s:%s:%s", ticketID, input.UserID, input.TransactionID)

	return IssueTicketResult{
		TicketID: ticketID,
		QRCode:   qrCode,
	}, nil
}

// SendEmail sends a confirmation email.
func (a *Activities) SendEmail(ctx context.Context, userID, confirmationID, qrCode string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending confirmation email", "user_id", userID, "confirmation_id", confirmationID)

	// Simulate email sending
	return nil
}

// SendSMS sends a confirmation SMS.
func (a *Activities) SendSMS(ctx context.Context, userID, confirmationID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending confirmation SMS", "user_id", userID, "confirmation_id", confirmationID)

	// Simulate SMS sending
	return nil
}
