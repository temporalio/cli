package ticketdrop

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
)

type Activities struct{}

// ReserveSeat locks a seat for 5 minutes.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID)

	// Simulate seat reservation by sleeping 1 second
	time.Sleep(1 * time.Second)

	// Generate a seat number like 'A15'
	seatNumber := fmt.Sprintf("A%d", time.Now().UnixNano()%50+1)
	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())

	logger.Info("Seat reserved", "seat", seatNumber)

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
