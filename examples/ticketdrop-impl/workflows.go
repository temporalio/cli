package ticketdrop

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TicketPurchase is the main workflow for purchasing a ticket.
func TicketPurchase(ctx workflow.Context, input PurchaseInput) (PurchaseResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ticket purchase", "user_id", input.UserID, "event_id", input.EventID)

	// Activity options with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var activities *Activities

	// Step 1: Reserve a seat
	var reservation ReserveSeatResult
	err := workflow.ExecuteActivity(ctx, activities.ReserveSeat, ReserveSeatInput{
		UserID:  input.UserID,
		EventID: input.EventID,
	}).Get(ctx, &reservation)
	if err != nil {
		return PurchaseResult{}, fmt.Errorf("failed to reserve seat: %w", err)
	}
	logger.Info("Seat reserved", "seat", reservation.SeatNumber, "expires_at", reservation.ExpiresAt)

	// Step 2: Process payment (with 10-second timeout)
	paymentOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    3,
		},
	}
	paymentCtx := workflow.WithActivityOptions(ctx, paymentOpts)

	var payment ProcessPaymentResult
	err = workflow.ExecuteActivity(paymentCtx, activities.ProcessPayment, ProcessPaymentInput{
		UserID:        input.UserID,
		ReservationID: reservation.ReservationID,
		Amount:        9999, // $99.99
	}).Get(paymentCtx, &payment)
	if err != nil {
		logger.Error("Payment failed, releasing seat", "error", err, "seat", reservation.SeatNumber)

		// Compensation: release the reserved seat back to inventory
		releaseErr := workflow.ExecuteActivity(ctx, activities.ReleaseSeat,
			input.EventID, input.UserID, reservation.SeatNumber,
		).Get(ctx, nil)
		if releaseErr != nil {
			logger.Error("Failed to release seat during compensation", "error", releaseErr)
		}

		return PurchaseResult{}, fmt.Errorf("payment failed: %w", err)
	}
	logger.Info("Payment processed", "transaction_id", payment.TransactionID)

	// Step 3: Issue ticket
	var ticket IssueTicketResult
	err = workflow.ExecuteActivity(ctx, activities.IssueTicket, IssueTicketInput{
		UserID:        input.UserID,
		EventID:       input.EventID,
		SeatNumber:    reservation.SeatNumber,
		TransactionID: payment.TransactionID,
	}).Get(ctx, &ticket)
	if err != nil {
		return PurchaseResult{}, fmt.Errorf("failed to issue ticket: %w", err)
	}
	logger.Info("Ticket issued", "ticket_id", ticket.TicketID)

	// Step 4: Send confirmation via child workflow
	confirmationID := fmt.Sprintf("conf-%s", ticket.TicketID)
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("confirmation-%s", confirmationID),
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	var confirmationResult SendConfirmationResult
	err = workflow.ExecuteChildWorkflow(childCtx, SendConfirmation, SendConfirmationInput{
		UserID:         input.UserID,
		EventID:        input.EventID,
		ConfirmationID: confirmationID,
		SeatNumber:     reservation.SeatNumber,
		QRCode:         ticket.QRCode,
	}).Get(ctx, &confirmationResult)
	if err != nil {
		// Log but don't fail the purchase if confirmation fails
		logger.Warn("Failed to send confirmation", "error", err)
	}

	return PurchaseResult{
		ConfirmationID: confirmationID,
		SeatNumber:     reservation.SeatNumber,
		QRCode:         ticket.QRCode,
		PurchasedAt:    workflow.Now(ctx),
	}, nil
}

// SendConfirmation is a child workflow that sends email and SMS confirmations.
func SendConfirmation(ctx workflow.Context, input SendConfirmationInput) (SendConfirmationResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Sending confirmations", "user_id", input.UserID, "confirmation_id", input.ConfirmationID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var activities *Activities
	result := SendConfirmationResult{}

	// Send email and SMS in parallel
	emailFuture := workflow.ExecuteActivity(ctx, activities.SendEmail, input.UserID, input.ConfirmationID, input.QRCode)
	smsFuture := workflow.ExecuteActivity(ctx, activities.SendSMS, input.UserID, input.ConfirmationID)

	if err := emailFuture.Get(ctx, nil); err != nil {
		logger.Warn("Failed to send email", "error", err)
	} else {
		result.EmailSent = true
	}

	if err := smsFuture.Get(ctx, nil); err != nil {
		logger.Warn("Failed to send SMS", "error", err)
	} else {
		result.SMSSent = true
	}

	return result, nil
}
