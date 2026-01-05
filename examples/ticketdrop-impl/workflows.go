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

// TicketQueue manages a fair queue for ticket purchases.
// Users join via signal, max 10 concurrent purchases at a time.
func TicketQueue(ctx workflow.Context, eventID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ticket queue", "event_id", eventID)

	// Queue state
	var waitingUsers []string
	activeUsers := make(map[string]bool) // userID -> true if active
	activePurchases := make(map[string]workflow.Future) // userID -> purchase future
	var completedCount int

	// Register query handler for queue status
	err := workflow.SetQueryHandler(ctx, "status", func() (QueueStatus, error) {
		activeList := make([]string, 0, len(activeUsers))
		for userID := range activeUsers {
			activeList = append(activeList, userID)
		}
		return QueueStatus{
			EventID:      eventID,
			QueueLength:  len(waitingUsers),
			ActiveCount:  len(activeUsers),
			WaitingUsers: buildQueueEntries(waitingUsers),
		}, nil
	})
	if err != nil {
		return err
	}

	// Signal channels
	joinChan := workflow.GetSignalChannel(ctx, SignalJoinQueue)
	doneChan := workflow.GetSignalChannel(ctx, SignalPurchaseDone)

	// Selector for handling multiple signals and child completions
	selector := workflow.NewSelector(ctx)

	// Handle join signals
	selector.AddReceive(joinChan, func(c workflow.ReceiveChannel, more bool) {
		var signal JoinQueueSignal
		c.Receive(ctx, &signal)
		logger.Info("User joined queue", "user_id", signal.UserID, "position", len(waitingUsers)+1)
		waitingUsers = append(waitingUsers, signal.UserID)
	})

	// Handle purchase done signals
	selector.AddReceive(doneChan, func(c workflow.ReceiveChannel, more bool) {
		var signal PurchaseDoneSignal
		c.Receive(ctx, &signal)
		logger.Info("Purchase completed", "user_id", signal.UserID, "success", signal.Success)
		delete(activePurchases, signal.UserID)
		completedCount++
	})

	// Process the queue
	for {
		// Start purchases for waiting users if we have capacity
		for len(activeUsers) < MaxConcurrent && len(waitingUsers) > 0 {
			userID := waitingUsers[0]
			waitingUsers = waitingUsers[1:]

			logger.Info("Starting purchase", "user_id", userID, "active", len(activeUsers)+1, "waiting", len(waitingUsers))

			childOpts := workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("purchase-%s-%s", eventID, userID),
			}
			childCtx := workflow.WithChildOptions(ctx, childOpts)

			future := workflow.ExecuteChildWorkflow(childCtx, TicketPurchase, PurchaseInput{
				UserID:  userID,
				EventID: eventID,
			})
			activePurchases[userID] = future
			activeUsers[userID] = true

			// Add completion handler for this child
			userIDCopy := userID
			selector.AddFuture(future, func(f workflow.Future) {
				var result PurchaseResult
				err := f.Get(ctx, &result)
				success := err == nil
				logger.Info("Child workflow completed", "user_id", userIDCopy, "success", success)
				delete(activePurchases, userIDCopy)
				delete(activeUsers, userIDCopy)
				completedCount++
			})
		}

		// Wait for signals or child completions
		// Use a timeout to periodically check state
		timerFuture := workflow.NewTimer(ctx, 5*time.Second)
		selector.AddFuture(timerFuture, func(f workflow.Future) {
			// Timer fired, just continue the loop
		})

		selector.Select(ctx)

		// Log status periodically
		logger.Debug("Queue status", "waiting", len(waitingUsers), "active", len(activePurchases), "completed", completedCount)

		// Continue as new if history gets too long (every 1000 completions)
		if completedCount >= 1000 {
			logger.Info("Continuing as new workflow", "completed", completedCount)
			return workflow.NewContinueAsNewError(ctx, TicketQueue, eventID)
		}
	}
}

// buildQueueEntries creates queue entries from a list of user IDs.
func buildQueueEntries(waitingUsers []string) []QueueEntry {
	entries := make([]QueueEntry, len(waitingUsers))
	for i, userID := range waitingUsers {
		entries[i] = QueueEntry{
			UserID:   userID,
			Position: i + 1,
		}
	}
	return entries
}
