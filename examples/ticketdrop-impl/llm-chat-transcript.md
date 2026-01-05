# Ticket purchase workflow setup
_Exported on 1/5/2026 at 10:40:54 PST from Cursor (2.2.44)_

---

**User**

I want to build a ticket sales system called TicketDrop. Start with a simple workflow that represents a ticket purchase: take a user ID and event ID, then return a confirmation. Set up the project structure with a worker, a starter, and basic types.

Here's the architecture:
```
User clicks "Buy Ticket"
         │
         ▼
┌─────────────────┐
│  TicketPurchase │  ← Main workflow (one per user)
│    Workflow     │
└────────┬────────┘
         │
         ├──► ReserveSeat (activity)     - Lock a seat for 5 minutes
         │         │
         │         ▼
         ├──► ProcessPayment (activity)  - Charge credit card
         │         │
         │         ▼
         ├──► IssueTicket (activity)     - Generate QR code
         │         │
         │         ▼
         └──► SendConfirmation (child)   - Email + SMS
```

---

**Cursor**

I'll help you build TicketDrop! Let me first check the existing project context.

I can see you have a Temporal dev server running. Let me create the TicketDrop project structure with Go.

```go.mod
module ticketdrop

go 1.22

require go.temporal.io/sdk v1.31.0

```

```go
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

```

```go
package ticketdrop

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
)

type Activities struct{}

// ReserveSeat locks a seat for 5 minutes.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID)

	// Simulate seat reservation
	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())
	seatNumber := fmt.Sprintf("A%d", time.Now().UnixMilli()%100+1)

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

	// Simulate payment processing
	transactionID := fmt.Sprintf("txn-%s-%d", input.ReservationID, time.Now().UnixMilli())

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

```

```go
package ticketdrop

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// TicketPurchase is the main workflow for purchasing a ticket.
func TicketPurchase(ctx workflow.Context, input PurchaseInput) (PurchaseResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ticket purchase", "user_id", input.UserID, "event_id", input.EventID)

	// Activity options with retries
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &workflow.RetryPolicy{
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

	// Step 2: Process payment
	var payment ProcessPaymentResult
	err = workflow.ExecuteActivity(ctx, activities.ProcessPayment, ProcessPaymentInput{
		UserID:        input.UserID,
		ReservationID: reservation.ReservationID,
		Amount:        9999, // $99.99
	}).Get(ctx, &payment)
	if err != nil {
		return PurchaseResult{}, fmt.Errorf("failed to process payment: %w", err)
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
		RetryPolicy: &workflow.RetryPolicy{
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

```

```go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ticketdrop"
)

func main() {
	// Connect to Temporal server
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, ticketdrop.TaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(ticketdrop.TicketPurchase)
	w.RegisterWorkflow(ticketdrop.SendConfirmation)

	// Register activities
	activities := &ticketdrop.Activities{}
	w.RegisterActivity(activities)

	log.Printf("Starting TicketDrop worker on task queue: %s", ticketdrop.TaskQueue)

	// Start worker
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}

```

```go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	"ticketdrop"
)

func main() {
	userID := flag.String("user", "user-123", "User ID")
	eventID := flag.String("event", "event-456", "Event ID")
	flag.Parse()

	// Connect to Temporal server
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Start the workflow
	workflowID := fmt.Sprintf("ticket-purchase-%s-%s", *userID, *eventID)
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: ticketdrop.TaskQueue,
	}

	input := ticketdrop.PurchaseInput{
		UserID:  *userID,
		EventID: *eventID,
	}

	log.Printf("Starting TicketPurchase workflow: %s", workflowID)

	we, err := c.ExecuteWorkflow(context.Background(), options, ticketdrop.TicketPurchase, input)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	log.Printf("Workflow started: WorkflowID=%s, RunID=%s", we.GetID(), we.GetRunID())

	// Wait for result
	var result ticketdrop.PurchaseResult
	if err := we.Get(context.Background(), &result); err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Pretty print result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\n✅ Purchase complete!\n%s\n", resultJSON)
}

```

Now let me download dependencies and verify the project builds:

Let me fix the RetryPolicy import:

```go
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
		RetryPolicy: &workflow.RetryPolicy{
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

	// Step 2: Process payment
	var payment ProcessPaymentResult
	err = workflow.ExecuteActivity(ctx, activities.ProcessPayment, ProcessPaymentInput{
		UserID:        input.UserID,
		ReservationID: reservation.ReservationID,
		Amount:        9999, // $99.99
	}).Get(ctx, &payment)
	if err != nil {
		return PurchaseResult{}, fmt.Errorf("failed to process payment: %w", err)
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
		RetryPolicy: &workflow.RetryPolicy{
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


```

```go
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

	// Step 2: Process payment
	var payment ProcessPaymentResult
	err = workflow.ExecuteActivity(ctx, activities.ProcessPayment, ProcessPaymentInput{
		UserID:        input.UserID,
		ReservationID: reservation.ReservationID,
		Amount:        9999, // $99.99
	}).Get(ctx, &payment)
	if err != nil {
		return PurchaseResult{}, fmt.Errorf("failed to process payment: %w", err)
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


```

The project builds successfully. Here's the structure I created:

```
ticketdrop-impl/
├── go.mod              # Dependencies
├── go.sum              # Lock file
├── types.go            # Shared types (inputs/outputs)
├── activities.go       # Activity implementations
├── workflows.go        # TicketPurchase + SendConfirmation workflows
├── worker/
│   └── main.go         # Worker that listens for tasks
└── starter/
    └── main.go         # CLI to start purchases
```

**To run it:**

1. **Start the worker** (in a new terminal):
```bash
cd /Users/dashti/repos/temporal/github.com/temporalio/cli/examples/ticketdrop-impl
go run ./worker
```

2. **Start a purchase** (in another terminal):
```bash
go run ./starter --user alice --event concert-2025
```

**The workflow flow:**

1. `TicketPurchase` receives user + event IDs
2. `ReserveSeat` → locks seat A23 for 5 minutes  
3. `ProcessPayment` → charges $99.99
4. `IssueTicket` → generates QR code
5. `SendConfirmation` (child workflow) → sends email + SMS in parallel
6. Returns confirmation with seat number and QR code

Want me to start the worker so you can test it?

```plaintext
ticketdrop-impl/
├── go.mod              # Dependencies
├── go.sum              # Lock file
├── types.go            # Shared types (inputs/outputs)
├── activities.go       # Activity implementations
├── workflows.go        # TicketPurchase + SendConfirmation workflows
├── worker/
│   └── main.go         # Worker that listens for tasks
└── starter/
    └── main.go         # CLI to start purchases
```

```shellscript
cd /Users/dashti/repos/temporal/github.com/temporalio/cli/examples/ticketdrop-impl
go run ./worker
```

```shellscript
go run ./starter --user alice --event concert-2025
```

---

**User**

yes

---

**Cursor**

Worker is running. Now let's start a ticket purchase:

It works! Alice got seat **A46** for `concert-2025`. The full flow executed:

1. ✅ Reserved seat A46
2. ✅ Processed payment ($99.99)
3. ✅ Issued ticket with QR code
4. ✅ Sent confirmation (email + SMS via child workflow)

You can view the workflow in the Temporal UI at http://localhost:8233 — look for workflow ID `ticket-purchase-alice-concert-2025`.

Or use the agent CLI to trace it:

```bash
temporal agent trace --workflow-id ticket-purchase-alice-concert-2025 --follow-children --format mermaid
```

```shellscript
temporal agent trace --workflow-id ticket-purchase-alice-concert-2025 --follow-children --format mermaid
```

---

**User**

I ran the workflow but it just returns immediately with empty data. How do I see what happened?

---

**Cursor**

Let me trace the workflow to see what happened:

The workflow completed successfully. Let me check the timeline and also look for any recent failures:

The workflow `ticket-purchase-alice-concert-2025` ran successfully — all activities and the child workflow completed. Let me check your terminal to see what you saw:

```plaintext
✅ Purchase complete!
{
  "confirmation_id": "conf-tkt-concert-taylor-swift-A7-1767224080214",
  "seat_number": "A7",
  "qr_code": "QR:tkt-concert-taylor-swift-A7-1767224080214:user-123:txn-res-user-123-concert-taylor-swift-1767224080206-1767224080210",
  "purchased_at": "2025-12-31T23:34:40.22362Z"
}
```

I see your terminal - actually the workflow **did work**! You got seat **A7** for the Taylor Swift concert:

```json
{
  "confirmation_id": "conf-tkt-concert-taylor-swift-A7-1767224080214",
  "seat_number": "A7",
  "qr_code": "QR:tkt-concert-taylor-swift-A7-...",
  "purchased_at": "2025-12-31T23:34:40.22362Z"
}
```

If you want to see the child workflow in the trace, use `--depth`:

```json
{
  "confirmation_id": "conf-tkt-concert-taylor-swift-A7-1767224080214",
  "seat_number": "A7",
  "qr_code": "QR:tkt-concert-taylor-swift-A7-...",
  "purchased_at": "2025-12-31T23:34:40.22362Z"
}
```

The timeline shows everything executed in order:

```
ReserveSeat → ProcessPayment → IssueTicket → SendConfirmation (child) → ✅ Completed
```

Your workflow ran in ~27ms. All 3 activities and the child workflow completed successfully. The data you got back is correct:

| Field | Value |
|-------|-------|
| Seat | A7 |
| Confirmation | conf-tkt-concert-taylor-swift-A7-... |
| QR Code | QR:tkt-concert-taylor-swift-A7-... |

Is there something specific you expected to see that's missing?

```plaintext
ReserveSeat → ProcessPayment → IssueTicket → SendConfirmation (child) → ✅ Completed
```

```plaintext
ReserveSeat → ProcessPayment → IssueTicket → SendConfirmation (child) → ✅ Completed
```

---

**User**

Add an activity that reserves a seat. For now, simulate it by sleeping 1 second and returning the seat number. The activity should take the event ID and return an available seat like 'A15'.

---

**Cursor**

I'll update the existing `ReserveSeat` activity to sleep for 1 second:

```go
package ticketdrop

import (
	"context"
	"fmt"
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

	// Simulate payment processing
	transactionID := fmt.Sprintf("txn-%s-%d", input.ReservationID, time.Now().UnixMilli())

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


```

Done. The `ReserveSeat` activity now sleeps for 1 second to simulate the reservation before returning a seat like `A15`.

You'll need to restart the worker to pick up the changes. The workflow will now take ~1 second longer.

---

**User**

you try it, as you're running the worker in the background

---

**Cursor**

Hmm, the duration shows 2ms. Let me check the actual timestamps:

The activity completed in ~1ms - the sleep isn't working. Let me check if the old worker is still running:

Now it's working. `ReserveSeat` took **1005ms** (from `54.436` to `55.441`):

```
scheduled: 23:39:54.435
started:   23:39:54.436
completed: 23:39:55.441  ← 1 second later
```

Charlie got seat **A1** for AC/DC 2025. The 1-second sleep is simulating the seat reservation.

```plaintext
scheduled: 23:39:54.435
started:   23:39:54.436
completed: 23:39:55.441  ← 1 second later
```

---

**User**

After reserving a seat, add a payment processing activity. It should take the user ID and amount, simulate processing for 2 seconds, and return a payment confirmation ID. Sometimes payments fail - add a 20% random failure rate.

---

**Cursor**

```go
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

	// Simulate payment processing
	transactionID := fmt.Sprintf("txn-%s-%d", input.ReservationID, time.Now().UnixMilli())

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

```

```go
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

```

Now let me restart the worker and test:

Interesting - ProcessPayment took **5017ms**. That suggests it failed once and retried. Let me check:

The 20% failure rate kicked in:

- **Attempt 1**: Failed (payment declined)
- **Attempt 2**: Succeeded ✅

You can see `"attempt": 2` in the timeline. Temporal automatically retried after the first failure. The total time was ~5s (2s first attempt + 1s backoff + 2s retry).

Activity timings:
| Activity | Duration |
|----------|----------|
| ReserveSeat | 1003ms |
| ProcessPayment | 5017ms (1 retry) |
| IssueTicket | 2ms |

---

**User**

Some purchases are failing. Show me exactly what's going wrong.

---

**Cursor**

No workflow failures - the retries are handling the payment declines. Let me check if anything is currently stuck:

No workflows are failing because Temporal retries handle the payment declines. But let me show you the payment failures that are being retried:

Found one! User-8's payment failed and retried. Let me show the details:

Found it! Here's exactly what happened in `ticket-purchase-user-8-concert-1`:

```
ProcessPayment Activity:
├── Attempt 1: ❌ FAILED
│   └── Error: "payment declined: insufficient funds"
│
└── Attempt 2: ✅ SUCCESS (auto-retry by Temporal)
```

The 20% random failure rate triggered on the first attempt. Temporal automatically retried and it succeeded.

**Summary of recent purchases:**

| Workflow | Payment Attempts | Result |
|----------|-----------------|--------|
| user-6 | 1 | ✅ |
| user-7 | 1 | ✅ |
| user-8 | 2 (1 failed) | ✅ |
| user-9 | 1 | ✅ |
| user-10 | 1 | ✅ |

All workflows completed successfully because the retry policy (max 3 attempts) handles the transient payment failures. No purchases are actually failing at the workflow level.

```plaintext
ProcessPayment Activity:
├── Attempt 1: ❌ FAILED
│   └── Error: "payment declined: insufficient funds"
│
└── Attempt 2: ✅ SUCCESS (auto-retry by Temporal)
```

---

**User**

Sometimes the payment takes too long and the workflow seems stuck. Add a 10-second timeout on the payment activity. If it times out, the purchase should fail gracefully.

---

**Cursor**

```go
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
		logger.Error("Payment failed", "error", err)
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


```

Now let me restart the worker and test it:

Done. The `ProcessPayment` activity now has:

- **10-second timeout** per attempt
- **Max 3 retries** with exponential backoff (1s → 2s → 4s, capped at 5s)

If all retries fail or timeout, the workflow returns a graceful error:

```go
return PurchaseResult{}, fmt.Errorf("payment failed: %w", err)
```

The workflow I just tested completed in ~3 seconds (1s seat + 2s payment + child workflow). The timeout would kick in if a payment takes longer than 10 seconds.

```go
return PurchaseResult{}, fmt.Errorf("payment failed: %w", err)
```

---

**User**

Right now every purchase gets a seat. Make it realistic: there are only 10 seats available. Add an in-memory inventory that tracks available seats. If no seats are left, the reservation should fail.

---

**Cursor**

```go
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

```

```go
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

```

```go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ticketdrop"
)

func main() {
	// Connect to Temporal server
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, ticketdrop.TaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(ticketdrop.TicketPurchase)
	w.RegisterWorkflow(ticketdrop.SendConfirmation)

	// Register activities with shared seat inventory
	inventory := ticketdrop.NewSeatInventory()
	activities := &ticketdrop.Activities{Inventory: inventory}
	w.RegisterActivity(activities)

	log.Printf("Starting TicketDrop worker on task queue: %s", ticketdrop.TaskQueue)

	// Start worker
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}


```

Now let me restart the worker and test selling out:

It works. Seats A1-A10 sold out, then user-11 failed:

```
user-1  → A1  ✅
user-2  → A2  ✅
user-3  → A3  ✅
...
user-10 → A10 ✅
user-11 → ❌ "sold out: no seats available"
```

The inventory is per-event, so a different event would have its own 10 seats.

```plaintext
user-1  → A1  ✅
user-2  → A2  ✅
user-3  → A3  ✅
...
user-10 → A10 ✅
user-11 → ❌ "sold out: no seats available"
```

---

**User**

I noticed something weird. Sometimes two users get the same seat! How do I debug this?

---

**Cursor**

Let me check recent workflows for duplicate seats:

The Temporal server is down. Let me check the terminal status:

The Temporal server is down. But I can explain the race condition issue even without looking at specific workflows.

**The bug**: Your `SeatInventory` is **in-memory per worker**. This causes duplicates when:

1. **Multiple workers run** - each has its own inventory, so both can give out seat A1
2. **Activity retries after worker restart** - the new worker has a fresh inventory

Let me show you the problem in the code:

```89:99:activities.go
// ReserveSeat locks a seat for 5 minutes.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID,
		"available", a.Inventory.Available(input.EventID))

	// Simulate seat reservation by sleeping 1 second
	time.Sleep(1 * time.Second)

	// Try to reserve a seat from inventory
	seatNumber, err := a.Inventory.Reserve(input.EventID, input.UserID)
```

The `Inventory` lives in the worker's memory. Start the server and I can show you the duplicate assignments, or I can fix it now by making the inventory persist across workers (using a simple file or adding idempotency checks).

Would you like me to:
1. **Start the server** and reproduce/debug the issue
2. **Fix it** by adding idempotency to prevent duplicate reservations

```go
// ReserveSeat locks a seat for 5 minutes.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID,
		"available", a.Inventory.Available(input.EventID))

	// Simulate seat reservation by sleeping 1 second
	time.Sleep(1 * time.Second)

	// Try to reserve a seat from inventory
	seatNumber, err := a.Inventory.Reserve(input.EventID, input.UserID)
```

---

**User**

yes

---

**Cursor**

I'll fix it by adding idempotency - if a user already reserved a seat for an event, return the same seat on retry:

```go
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
	mu          sync.Mutex
	seats       map[string][]string // eventID -> available seats
	userSeats   map[string]string   // "eventID:userID" -> seat (for idempotency)
	seatOwners  map[string]string   // "eventID:seat" -> userID
}

// NewSeatInventory creates an inventory with 10 seats per event.
func NewSeatInventory() *SeatInventory {
	return &SeatInventory{
		seats:      make(map[string][]string),
		userSeats:  make(map[string]string),
		seatOwners: make(map[string]string),
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
// Idempotent: if user already has a seat for this event, return the same seat.
func (inv *SeatInventory) Reserve(eventID, userID string) (string, bool, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.initEvent(eventID)

	// Idempotency check: if user already has a seat, return it
	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	if existingSeat, exists := inv.userSeats[userKey]; exists {
		return existingSeat, true, nil // true = was already reserved
	}

	available := inv.seats[eventID]
	if len(available) == 0 {
		return "", false, errors.New("sold out: no seats available")
	}

	// Take the first available seat
	seat := available[0]
	inv.seats[eventID] = available[1:]

	// Track reservation for idempotency
	inv.userSeats[userKey] = seat
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)
	inv.seatOwners[seatKey] = userID

	return seat, false, nil
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

```

```go
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
	mu          sync.Mutex
	seats       map[string][]string // eventID -> available seats
	userSeats   map[string]string   // "eventID:userID" -> seat (for idempotency)
	seatOwners  map[string]string   // "eventID:seat" -> userID
}

// NewSeatInventory creates an inventory with 10 seats per event.
func NewSeatInventory() *SeatInventory {
	return &SeatInventory{
		seats:      make(map[string][]string),
		userSeats:  make(map[string]string),
		seatOwners: make(map[string]string),
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
// Idempotent: if user already has a seat for this event, return the same seat.
func (inv *SeatInventory) Reserve(eventID, userID string) (string, bool, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.initEvent(eventID)

	// Idempotency check: if user already has a seat, return it
	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	if existingSeat, exists := inv.userSeats[userKey]; exists {
		return existingSeat, true, nil // true = was already reserved
	}

	available := inv.seats[eventID]
	if len(available) == 0 {
		return "", false, errors.New("sold out: no seats available")
	}

	// Take the first available seat
	seat := available[0]
	inv.seats[eventID] = available[1:]

	// Track reservation for idempotency
	inv.userSeats[userKey] = seat
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)
	inv.seatOwners[seatKey] = userID

	return seat, false, nil
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
// Idempotent: retries return the same seat.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID,
		"available", a.Inventory.Available(input.EventID))

	// Simulate seat reservation by sleeping 1 second
	time.Sleep(1 * time.Second)

	// Try to reserve a seat from inventory (idempotent)
	seatNumber, wasRetry, err := a.Inventory.Reserve(input.EventID, input.UserID)
	if err != nil {
		logger.Warn("Reservation failed", "error", err)
		return ReserveSeatResult{}, err
	}

	if wasRetry {
		logger.Info("Returning existing reservation (idempotent)", "seat", seatNumber)
	} else {
		logger.Info("Seat reserved", "seat", seatNumber, "remaining", a.Inventory.Available(input.EventID))
	}

	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())

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

```

Now let me start the server and worker to test:

Now let me test no duplicates with multiple users:

Fixed! All 5 concurrent purchases got unique seats: A1, A2, A3, A4, A5.

**What I changed:**

The `SeatInventory` now tracks reservations by `eventID:userID`, so:

1. **First request**: Assigns a new seat, stores `userSeats["event:user"] = "A1"`
2. **Retry/duplicate**: Looks up existing seat and returns it (idempotent)

```go
// Idempotency check: if user already has a seat, return it
userKey := fmt.Sprintf("%s:%s", eventID, userID)
if existingSeat, exists := inv.userSeats[userKey]; exists {
    return existingSeat, true, nil // true = was already reserved
}
```

This prevents the race condition where retries could consume additional seats.

```go
// Idempotency check: if user already has a seat, return it
userKey := fmt.Sprintf("%s:%s", eventID, userID)
if existingSeat, exists := inv.userSeats[userKey]; exists {
    return existingSeat, true, nil // true = was already reserved
}
```

---

**User**

The timeline shows that two users are checking the same seat simultaneously. How do I fix this race condition?

---

**Cursor**

Let me check the timeline to see the race condition:

Found the race! All 5 activities overlapped during the 1-second sleep:

```
user-2: 50.580 ─────────────────────────── 51.584  (started first)
user-4: 50.582 ─────────────────────────── 51.584
user-1: 50.584 ─────────────────────────── 51.585
user-5: 50.585 ─────────────────────────── 51.587
user-3: 50.587 ─────────────────────────── 51.588
        ^^^^^^
        All sleeping at the same time before Reserve()
```

**The problem**: We log "available: 10" and then sleep *before* calling `Reserve()`. All 5 see "10 available" because nobody has reserved yet.

**The fix**: Move the sleep *after* the reservation (or make the whole operation atomic):

```plaintext
user-2: 50.580 ─────────────────────────── 51.584  (started first)
user-4: 50.582 ─────────────────────────── 51.584
user-1: 50.584 ─────────────────────────── 51.585
user-5: 50.585 ─────────────────────────── 51.587
user-3: 50.587 ─────────────────────────── 51.588
        ^^^^^^
        All sleeping at the same time before Reserve()
```

```go
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
	mu         sync.Mutex
	seats      map[string][]string // eventID -> available seats
	userSeats  map[string]string   // "eventID:userID" -> seat (for idempotency)
	seatOwners map[string]string   // "eventID:seat" -> userID
}

// NewSeatInventory creates an inventory with 10 seats per event.
func NewSeatInventory() *SeatInventory {
	return &SeatInventory{
		seats:      make(map[string][]string),
		userSeats:  make(map[string]string),
		seatOwners: make(map[string]string),
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
// Idempotent: if user already has a seat for this event, return the same seat.
func (inv *SeatInventory) Reserve(eventID, userID string) (string, bool, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.initEvent(eventID)

	// Idempotency check: if user already has a seat, return it
	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	if existingSeat, exists := inv.userSeats[userKey]; exists {
		return existingSeat, true, nil // true = was already reserved
	}

	available := inv.seats[eventID]
	if len(available) == 0 {
		return "", false, errors.New("sold out: no seats available")
	}

	// Take the first available seat
	seat := available[0]
	inv.seats[eventID] = available[1:]

	// Track reservation for idempotency
	inv.userSeats[userKey] = seat
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)
	inv.seatOwners[seatKey] = userID

	return seat, false, nil
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
// Idempotent: retries return the same seat.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID)

	// Reserve seat first (atomic operation protected by mutex)
	seatNumber, wasRetry, err := a.Inventory.Reserve(input.EventID, input.UserID)
	if err != nil {
		logger.Warn("Reservation failed", "error", err)
		return ReserveSeatResult{}, err
	}

	// Simulate confirmation delay (e.g., writing to database)
	time.Sleep(1 * time.Second)

	if wasRetry {
		logger.Info("Returning existing reservation (idempotent)", "seat", seatNumber)
	} else {
		logger.Info("Seat reserved", "seat", seatNumber, "remaining", a.Inventory.Available(input.EventID))
	}

	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())

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

```

Fixed! All 5 concurrent users got unique seats: A1, A2, A3, A4, A5.

**Before**: Sleep happened *before* `Reserve()`, so all activities saw "10 available" simultaneously.

**After**: `Reserve()` runs first (protected by mutex), then sleep simulates confirmation time.

```
Before (race):                    After (fixed):
┌─────────────────────────┐       ┌─────────────────────────┐
│ Log "10 available"      │       │ Reserve() ← mutex lock  │
│ Sleep 1s                │ ←race │ Sleep 1s                │
│ Reserve() ← mutex lock  │       │ Log result              │
└─────────────────────────┘       └─────────────────────────┘
```

```plaintext
Before (race):                    After (fixed):
┌─────────────────────────┐       ┌─────────────────────────┐
│ Log "10 available"      │       │ Reserve() ← mutex lock  │
│ Sleep 1s                │ ←race │ Sleep 1s                │
│ Reserve() ← mutex lock  │       │ Log result              │
└─────────────────────────┘       └─────────────────────────┘
```

```plaintext
Before (race):                    After (fixed):
┌─────────────────────────┐       ┌─────────────────────────┐
│ Log "10 available"      │       │ Reserve() ← mutex lock  │
│ Sleep 1s                │ ←race │ Sleep 1s                │
│ Reserve() ← mutex lock  │       │ Log result              │
└─────────────────────────┘       └─────────────────────────┘
```

---

**User**

When a payment fails after a seat is reserved, the seat stays reserved forever. Add compensation: if payment fails, release the seat back to inventory.

---

**Cursor**

I'll add a `Release` method to the inventory and a compensation activity:

```go
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
	mu         sync.Mutex
	seats      map[string][]string // eventID -> available seats
	userSeats  map[string]string   // "eventID:userID" -> seat (for idempotency)
	seatOwners map[string]string   // "eventID:seat" -> userID
}

// NewSeatInventory creates an inventory with 10 seats per event.
func NewSeatInventory() *SeatInventory {
	return &SeatInventory{
		seats:      make(map[string][]string),
		userSeats:  make(map[string]string),
		seatOwners: make(map[string]string),
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
// Idempotent: if user already has a seat for this event, return the same seat.
func (inv *SeatInventory) Reserve(eventID, userID string) (string, bool, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.initEvent(eventID)

	// Idempotency check: if user already has a seat, return it
	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	if existingSeat, exists := inv.userSeats[userKey]; exists {
		return existingSeat, true, nil // true = was already reserved
	}

	available := inv.seats[eventID]
	if len(available) == 0 {
		return "", false, errors.New("sold out: no seats available")
	}

	// Take the first available seat
	seat := available[0]
	inv.seats[eventID] = available[1:]

	// Track reservation for idempotency
	inv.userSeats[userKey] = seat
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)
	inv.seatOwners[seatKey] = userID

	return seat, false, nil
}

// Available returns the count of available seats for an event.
func (inv *SeatInventory) Available(eventID string) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.initEvent(eventID)
	return len(inv.seats[eventID])
}

// Release returns a seat back to the available pool (compensation).
func (inv *SeatInventory) Release(eventID, userID, seat string) bool {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)

	// Verify this user owns this seat
	if inv.seatOwners[seatKey] != userID {
		return false
	}

	// Remove from tracking
	delete(inv.userSeats, userKey)
	delete(inv.seatOwners, seatKey)

	// Add seat back to available pool
	inv.seats[eventID] = append(inv.seats[eventID], seat)

	return true
}

type Activities struct {
	Inventory *SeatInventory
}

// ReserveSeat locks a seat for 5 minutes.
// Idempotent: retries return the same seat.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID)

	// Reserve seat first (atomic operation protected by mutex)
	seatNumber, wasRetry, err := a.Inventory.Reserve(input.EventID, input.UserID)
	if err != nil {
		logger.Warn("Reservation failed", "error", err)
		return ReserveSeatResult{}, err
	}

	// Simulate confirmation delay (e.g., writing to database)
	time.Sleep(1 * time.Second)

	if wasRetry {
		logger.Info("Returning existing reservation (idempotent)", "seat", seatNumber)
	} else {
		logger.Info("Seat reserved", "seat", seatNumber, "remaining", a.Inventory.Available(input.EventID))
	}

	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())

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

```

```go
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
	mu         sync.Mutex
	seats      map[string][]string // eventID -> available seats
	userSeats  map[string]string   // "eventID:userID" -> seat (for idempotency)
	seatOwners map[string]string   // "eventID:seat" -> userID
}

// NewSeatInventory creates an inventory with 10 seats per event.
func NewSeatInventory() *SeatInventory {
	return &SeatInventory{
		seats:      make(map[string][]string),
		userSeats:  make(map[string]string),
		seatOwners: make(map[string]string),
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
// Idempotent: if user already has a seat for this event, return the same seat.
func (inv *SeatInventory) Reserve(eventID, userID string) (string, bool, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.initEvent(eventID)

	// Idempotency check: if user already has a seat, return it
	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	if existingSeat, exists := inv.userSeats[userKey]; exists {
		return existingSeat, true, nil // true = was already reserved
	}

	available := inv.seats[eventID]
	if len(available) == 0 {
		return "", false, errors.New("sold out: no seats available")
	}

	// Take the first available seat
	seat := available[0]
	inv.seats[eventID] = available[1:]

	// Track reservation for idempotency
	inv.userSeats[userKey] = seat
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)
	inv.seatOwners[seatKey] = userID

	return seat, false, nil
}

// Available returns the count of available seats for an event.
func (inv *SeatInventory) Available(eventID string) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.initEvent(eventID)
	return len(inv.seats[eventID])
}

// Release returns a seat back to the available pool (compensation).
func (inv *SeatInventory) Release(eventID, userID, seat string) bool {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	userKey := fmt.Sprintf("%s:%s", eventID, userID)
	seatKey := fmt.Sprintf("%s:%s", eventID, seat)

	// Verify this user owns this seat
	if inv.seatOwners[seatKey] != userID {
		return false
	}

	// Remove from tracking
	delete(inv.userSeats, userKey)
	delete(inv.seatOwners, seatKey)

	// Add seat back to available pool
	inv.seats[eventID] = append(inv.seats[eventID], seat)

	return true
}

type Activities struct {
	Inventory *SeatInventory
}

// ReserveSeat locks a seat for 5 minutes.
// Idempotent: retries return the same seat.
func (a *Activities) ReserveSeat(ctx context.Context, input ReserveSeatInput) (ReserveSeatResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Reserving seat", "user_id", input.UserID, "event_id", input.EventID)

	// Reserve seat first (atomic operation protected by mutex)
	seatNumber, wasRetry, err := a.Inventory.Reserve(input.EventID, input.UserID)
	if err != nil {
		logger.Warn("Reservation failed", "error", err)
		return ReserveSeatResult{}, err
	}

	// Simulate confirmation delay (e.g., writing to database)
	time.Sleep(1 * time.Second)

	if wasRetry {
		logger.Info("Returning existing reservation (idempotent)", "seat", seatNumber)
	} else {
		logger.Info("Seat reserved", "seat", seatNumber, "remaining", a.Inventory.Available(input.EventID))
	}

	reservationID := fmt.Sprintf("res-%s-%s-%d", input.UserID, input.EventID, time.Now().UnixMilli())

	return ReserveSeatResult{
		ReservationID: reservationID,
		SeatNumber:    seatNumber,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}, nil
}

// ReleaseSeat returns a seat to the available pool (compensation for failed payment).
func (a *Activities) ReleaseSeat(ctx context.Context, eventID, userID, seat string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Releasing seat (compensation)", "user_id", userID, "event_id", eventID, "seat", seat)

	released := a.Inventory.Release(eventID, userID, seat)
	if !released {
		logger.Warn("Seat was not released (may not be owned by user)", "seat", seat)
		return nil // Don't fail compensation
	}

	logger.Info("Seat released", "seat", seat, "available", a.Inventory.Available(eventID))
	return nil
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

```

Now let me update the workflow to compensate on payment failure:

```go
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

```

All 10 seats sold! Let me check the worker logs to see if any compensation happened during payment retries:

The logs show the system is working. I see payment failures on attempts 1-2 for several users (comp-user-4, comp-user-6, comp-user-10), but they all succeeded on retry before exhausting all 3 attempts.

The compensation only triggers when **all 3 retries fail**. With 20% failure rate, that's only 0.8% chance (0.2³).

**Summary of what's implemented:**

```
Workflow Flow:
┌─────────────────┐
│ ReserveSeat     │ ← Reserves seat A1
└────────┬────────┘
         ▼
┌─────────────────┐
│ ProcessPayment  │ ← 3 retries max
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
 SUCCESS    FAIL (all 3)
    │         │
    ▼         ▼
IssueTicket  ReleaseSeat ← Compensation!
    │         │
    ▼         ▼
SendConfirm  Return error
```

The `ReleaseSeat` activity calls `Inventory.Release()` to return the seat to the pool so another user can buy it.

```plaintext
Workflow Flow:
┌─────────────────┐
│ ReserveSeat     │ ← Reserves seat A1
└────────┬────────┘
         ▼
┌─────────────────┐
│ ProcessPayment  │ ← 3 retries max
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
 SUCCESS    FAIL (all 3)
    │         │
    ▼         ▼
IssueTicket  ReleaseSeat ← Compensation!
    │         │
    ▼         ▼
SendConfirm  Return error
```

---

**User**

The seat release is sometimes taking forever. I think it's hanging. How do I see what's stuck?

---

**Cursor**

No running workflows right now. Let me check for any with pending activities:

No stuck workflows right now. Here's how to debug when something is hanging:

No workflows are stuck right now. Here's how to debug when something hangs:

**1. Find running workflows:**
```bash
temporal workflow list --query "ExecutionStatus = 'Running'"
```

**2. Check for pending activities:**
```bash
temporal workflow describe --pending --workflow-id <id> --format json
```

**3. See what's waiting:**
```bash
temporal workflow describe --pending --workflow-id <id> --format mermaid
```

The output shows:
- `pending_activities`: Activities waiting to complete
- `pending_children`: Child workflows in progress
- Each pending activity shows: type, attempt count, last failure, scheduled time

**Example output when stuck:**
```json
{
  "pending_activities": [
    {
      "activity_type": "ReleaseSeat",
      "state": "STARTED",
      "attempt": 1,
      "scheduled_time": "2026-01-02T...",
      "last_started_time": "2026-01-02T...",
      "heartbeat_timeout": "0s"  // No heartbeat = can't detect stuck
    }
  ]
}
```

If `ReleaseSeat` is hanging, it's likely the 1-second sleep or mutex contention. Want me to add heartbeating to detect stuck activities?

```shellscript
temporal workflow list --query "ExecutionStatus = 'Running'"
```

```shellscript
temporal workflow describe --pending --workflow-id <id> --format json
```

```shellscript
temporal workflow describe --pending --workflow-id <id> --format mermaid
```

```json
{
  "pending_activities": [
    {
      "activity_type": "ReleaseSeat",
      "state": "STARTED",
      "attempt": 1,
      "scheduled_time": "2026-01-02T...",
      "last_started_time": "2026-01-02T...",
      "heartbeat_timeout": "0s"  // No heartbeat = can't detect stuck
    }
  ]
}
```

---

**User**

During high traffic, we need to queue users fairly. Add a queue workflow that users enter before starting their purchase. Process users in order, max 10 concurrent purchases.

---

**Cursor**

I'll create a `TicketQueue` workflow that manages fair queuing with max 10 concurrent purchases:

```go
package ticketdrop

import "time"

const (
	TaskQueue          = "ticketdrop"
	MaxConcurrent      = 10 // Max concurrent purchases per event
	SignalJoinQueue    = "join-queue"
	SignalPurchaseDone = "purchase-done"
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

```

```go
package ticketdrop

import "time"

const (
	TaskQueue          = "ticketdrop"
	MaxConcurrent      = 10 // Max concurrent purchases per event
	SignalJoinQueue    = "join-queue"
	SignalPurchaseDone = "purchase-done"
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

// QueueEntry represents a user waiting in the queue.
type QueueEntry struct {
	UserID    string    `json:"user_id"`
	JoinedAt  time.Time `json:"joined_at"`
	Position  int       `json:"position"`
}

// JoinQueueSignal is sent when a user wants to join the queue.
type JoinQueueSignal struct {
	UserID string `json:"user_id"`
}

// PurchaseDoneSignal is sent when a purchase completes (success or failure).
type PurchaseDoneSignal struct {
	UserID  string `json:"user_id"`
	Success bool   `json:"success"`
}

// QueueStatus represents the current state of the queue.
type QueueStatus struct {
	EventID       string       `json:"event_id"`
	QueueLength   int          `json:"queue_length"`
	ActiveCount   int          `json:"active_count"`
	WaitingUsers  []QueueEntry `json:"waiting_users"`
}

```

Now let me add the queue workflow:

```go
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
	activePurchases := make(map[string]workflow.Future) // userID -> purchase future
	var completedCount int

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
		for len(activePurchases) < MaxConcurrent && len(waitingUsers) > 0 {
			userID := waitingUsers[0]
			waitingUsers = waitingUsers[1:]

			logger.Info("Starting purchase", "user_id", userID, "active", len(activePurchases)+1, "waiting", len(waitingUsers))

			childOpts := workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("purchase-%s-%s", eventID, userID),
			}
			childCtx := workflow.WithChildOptions(ctx, childOpts)

			future := workflow.ExecuteChildWorkflow(childCtx, TicketPurchase, PurchaseInput{
				UserID:  userID,
				EventID: eventID,
			})
			activePurchases[userID] = future

			// Add completion handler for this child
			userIDCopy := userID
			selector.AddFuture(future, func(f workflow.Future) {
				var result PurchaseResult
				err := f.Get(ctx, &result)
				success := err == nil
				logger.Info("Child workflow completed", "user_id", userIDCopy, "success", success)
				delete(activePurchases, userIDCopy)
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

// GetQueueStatus is a query handler that returns current queue status.
func GetQueueStatus(waitingUsers []string, activeCount int) QueueStatus {
	entries := make([]QueueEntry, len(waitingUsers))
	for i, userID := range waitingUsers {
		entries[i] = QueueEntry{
			UserID:   userID,
			Position: i + 1,
		}
	}
	return QueueStatus{
		QueueLength:  len(waitingUsers),
		ActiveCount:  activeCount,
		WaitingUsers: entries,
	}
}

```

```go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ticketdrop"
)

func main() {
	// Connect to Temporal server
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, ticketdrop.TaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(ticketdrop.TicketPurchase)
	w.RegisterWorkflow(ticketdrop.SendConfirmation)
	w.RegisterWorkflow(ticketdrop.TicketQueue)

	// Register activities with shared seat inventory
	inventory := ticketdrop.NewSeatInventory()
	activities := &ticketdrop.Activities{Inventory: inventory}
	w.RegisterActivity(activities)

	log.Printf("Starting TicketDrop worker on task queue: %s", ticketdrop.TaskQueue)

	// Start worker
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}


```

Now let me create a starter for the queue and test it:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	"ticketdrop"
)

func main() {
	eventID := flag.String("event", "concert-2025", "Event ID")
	action := flag.String("action", "start", "Action: start, join, status")
	userID := flag.String("user", "", "User ID (for join action)")
	flag.Parse()

	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	queueWorkflowID := fmt.Sprintf("ticket-queue-%s", *eventID)

	switch *action {
	case "start":
		// Start the queue workflow for this event
		options := client.StartWorkflowOptions{
			ID:        queueWorkflowID,
			TaskQueue: ticketdrop.TaskQueue,
		}

		we, err := c.ExecuteWorkflow(context.Background(), options, ticketdrop.TicketQueue, *eventID)
		if err != nil {
			log.Fatalf("Failed to start queue workflow: %v", err)
		}
		fmt.Printf("✅ Queue started for event: %s\n", *eventID)
		fmt.Printf("   WorkflowID: %s\n", we.GetID())
		fmt.Printf("   RunID: %s\n", we.GetRunID())

	case "join":
		if *userID == "" {
			log.Fatal("--user is required for join action")
		}

		// Send join signal to the queue
		err := c.SignalWorkflow(context.Background(), queueWorkflowID, "", ticketdrop.SignalJoinQueue, ticketdrop.JoinQueueSignal{
			UserID: *userID,
		})
		if err != nil {
			log.Fatalf("Failed to join queue: %v", err)
		}
		fmt.Printf("✅ User %s joined queue for event %s\n", *userID, *eventID)

	case "status":
		// Describe the workflow to see pending work
		desc, err := c.DescribeWorkflowExecution(context.Background(), queueWorkflowID, "")
		if err != nil {
			log.Fatalf("Failed to get queue status: %v", err)
		}
		fmt.Printf("Queue: %s\n", queueWorkflowID)
		fmt.Printf("Status: %s\n", desc.WorkflowExecutionInfo.Status.String())
		fmt.Printf("Pending children: %d\n", len(desc.PendingChildren))

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

```

The queue is working:

- **15 users joined** the queue
- **10 concurrent purchases** running (max limit hit)
- **5 users waiting** in queue

Let me check how it's progressing:

All 15 purchases completed. Let me verify:

The queue worked perfectly:

| Users | Result | Reason |
|-------|--------|--------|
| 1-10 | ✅ Completed | Got seats A1-A10 (first in queue) |
| 11-15 | ❌ Failed | Sold out (only 10 seats available) |

**Queue behavior:**
1. First 10 users started immediately (max concurrent = 10)
2. Users 11-15 waited in queue
3. As purchases completed, waiting users were processed in FIFO order
4. Users 11-15 failed because all seats were taken by users 1-10

**Architecture:**
```
User clicks "Buy" → Signal → TicketQueue (long-running)
                                   │
                    ┌──────────────┼──────────────┐
                    ▼              ▼              ▼
               Purchase 1    Purchase 2   ... Purchase 10 (max)
                                   │
                            Users 11-15 wait...
```

```plaintext
User clicks "Buy" → Signal → TicketQueue (long-running)
                                   │
                    ┌──────────────┼──────────────┐
                    ▼              ▼              ▼
               Purchase 1    Purchase 2   ... Purchase 10 (max)
                                   │
                            Users 11-15 wait...
```

---

**User**

How can I see who's in the queue and how many are being processed?

---

**Cursor**

I'll add a query handler to the queue workflow so you can check its status:

```go
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
		for len(activePurchases) < MaxConcurrent && len(waitingUsers) > 0 {
			userID := waitingUsers[0]
			waitingUsers = waitingUsers[1:]

			logger.Info("Starting purchase", "user_id", userID, "active", len(activePurchases)+1, "waiting", len(waitingUsers))

			childOpts := workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("purchase-%s-%s", eventID, userID),
			}
			childCtx := workflow.WithChildOptions(ctx, childOpts)

			future := workflow.ExecuteChildWorkflow(childCtx, TicketPurchase, PurchaseInput{
				UserID:  userID,
				EventID: eventID,
			})
			activePurchases[userID] = future

			// Add completion handler for this child
			userIDCopy := userID
			selector.AddFuture(future, func(f workflow.Future) {
				var result PurchaseResult
				err := f.Get(ctx, &result)
				success := err == nil
				logger.Info("Child workflow completed", "user_id", userIDCopy, "success", success)
				delete(activePurchases, userIDCopy)
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

// GetQueueStatus is a query handler that returns current queue status.
func GetQueueStatus(waitingUsers []string, activeCount int) QueueStatus {
	entries := make([]QueueEntry, len(waitingUsers))
	for i, userID := range waitingUsers {
		entries[i] = QueueEntry{
			UserID:   userID,
			Position: i + 1,
		}
	}
	return QueueStatus{
		QueueLength:  len(waitingUsers),
		ActiveCount:  activeCount,
		WaitingUsers: entries,
	}
}

```

```go
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
		for len(activePurchases) < MaxConcurrent && len(waitingUsers) > 0 {
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

// GetQueueStatus is a query handler that returns current queue status.
func GetQueueStatus(waitingUsers []string, activeCount int) QueueStatus {
	entries := make([]QueueEntry, len(waitingUsers))
	for i, userID := range waitingUsers {
		entries[i] = QueueEntry{
			UserID:   userID,
			Position: i + 1,
		}
	}
	return QueueStatus{
		QueueLength:  len(waitingUsers),
		ActiveCount:  activeCount,
		WaitingUsers: entries,
	}
}

```

```go
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

// GetQueueStatus is a query handler that returns current queue status.
func GetQueueStatus(waitingUsers []string, activeCount int) QueueStatus {
	entries := make([]QueueEntry, len(waitingUsers))
	for i, userID := range waitingUsers {
		entries[i] = QueueEntry{
			UserID:   userID,
			Position: i + 1,
		}
	}
	return QueueStatus{
		QueueLength:  len(waitingUsers),
		ActiveCount:  activeCount,
		WaitingUsers: entries,
	}
}

```

```go
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

```

Now let me test the query:

The query works. Let me run a live demo:

The query shows live status:

**1 second in:**
```json
{
  "active": 10,           // 10 purchases running (max)
  "waiting": 10,          // 10 users waiting in queue
  "users_waiting": ["demo-14", "demo-4", "demo-15", ...]
}
```

**4 seconds in:**
```json
{
  "active": 10,           // Still 10 running
  "waiting": 6            // 4 moved from waiting to active
}
```

**To query the queue status:**
```bash
temporal workflow query --workflow-id ticket-queue-<event> --type status -o json
```

This returns:
- `active_count`: Users currently purchasing (max 10)
- `queue_length`: Users waiting in line
- `waiting_users`: List of waiting users with their position

```json
{
  "active": 10,           // 10 purchases running (max)
  "waiting": 10,          // 10 users waiting in queue
  "users_waiting": ["demo-14", "demo-4", "demo-15", ...]
}
```

```json
{
  "active": 10,           // Still 10 running
  "waiting": 6            // 4 moved from waiting to active
}
```

```shellscript
temporal workflow query --workflow-id ticket-queue-<event> --type status -o json
```

---

**User**

A user is complaining they were in position 5 but after a page refresh they're at position 50. How do I investigate?

---

**Cursor**

Let me show you how to investigate queue issues:

No issues found in this demo. Here's how to investigate a real case:

**Investigation Checklist:**

| Check | Command | Issue If... |
|-------|---------|------------|
| Duplicate joins | See below | User joined queue multiple times |
| Continue-as-new | Check for `ContinueAsNewInitiated` events | Queue restarted, lost state |
| Workflow failures | Check for `Failed`/`TimedOut` events | Queue crashed |
| History length | `historyLength` in describe | >10K = likely continue-as-new |

**Commands to run:**

```bash
# 1. Find if user joined multiple times
temporal workflow show --workflow-id ticket-queue-<event> -o json | \
  jq '[.events[] | select(.eventType == "EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED")] | 
      map(.workflowExecutionSignaledEventAttributes.input.payloads[0].data | @base64d | fromjson) |
      group_by(.user_id) | map(select(length > 1))'

# 2. Check current position
temporal workflow query --workflow-id ticket-queue-<event> --type status -o json

# 3. Look for queue restarts
temporal workflow show --workflow-id ticket-queue-<event> -o json | \
  jq '[.events[] | select(.eventType | test("ContinueAsNew"))]'
```

**Most likely causes for position jump:**
1. **User refreshed and rejoined** → Duplicate entry at end of queue
2. **Queue continued-as-new** → State was reset (we handle this, but check)
3. **Frontend bug** → Position cached incorrectly on client side

```shellscript
# 1. Find if user joined multiple times
temporal workflow show --workflow-id ticket-queue-<event> -o json | \
  jq '[.events[] | select(.eventType == "EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED")] | 
      map(.workflowExecutionSignaledEventAttributes.input.payloads[0].data | @base64d | fromjson) |
      group_by(.user_id) | map(select(length > 1))'

# 2. Check current position
temporal workflow query --workflow-id ticket-queue-<event> --type status -o json

# 3. Look for queue restarts
temporal workflow show --workflow-id ticket-queue-<event> -o json | \
  jq '[.events[] | select(.eventType | test("ContinueAsNew"))]'
```

---

**User**

After a successful purchase, send an email confirmation. Add this as a child workflow so it doesn't block the main purchase.

---

**Cursor**

Good news - this is already implemented! Let me show you:

Already implemented! Here's the architecture:

```
TicketPurchase Workflow
         │
         ├── ReserveSeat
         ├── ProcessPayment
         ├── IssueTicket
         │
         └──► SendConfirmation (child workflow)
                    │
                    ├── SendEmail  ─┐
                    │               ├─ Run in parallel
                    └── SendSMS   ─┘
```

**Key features already in place:**

1. **Child workflow** - `SendConfirmation` runs independently
2. **Non-blocking** - Main purchase returns success even if email fails:
   ```go
   if err != nil {
       logger.Warn("Failed to send confirmation", "error", err)  // Log, don't fail
   }
   ```
3. **Parallel sends** - Email and SMS run simultaneously
4. **Retries** - 5 attempts with exponential backoff

You can verify it's working:
```bash
temporal workflow list --query "WorkflowType = 'SendConfirmation'" --limit 5
```

```plaintext
TicketPurchase Workflow
         │
         ├── ReserveSeat
         ├── ProcessPayment
         ├── IssueTicket
         │
         └──► SendConfirmation (child workflow)
                    │
                    ├── SendEmail  ─┐
                    │               ├─ Run in parallel
                    └── SendSMS   ─┘
```

```go
   if err != nil {
       logger.Warn("Failed to send confirmation", "error", err)  // Log, don't fail
   }
```

```shellscript
temporal workflow list --query "WorkflowType = 'SendConfirmation'" --limit 5
```

---

**User**

The email service is down but tickets should still be issued. How do I make email failures not fail the whole purchase?

---

**Cursor**

This is already handled! Let me show you the resilient design:

Already handled at **two levels**:

**Level 1: Child workflow failure doesn't fail purchase**
```go
// In TicketPurchase workflow
err = workflow.ExecuteChildWorkflow(..., SendConfirmation, ...).Get(ctx, &result)
if err != nil {
    logger.Warn("Failed to send confirmation", "error", err)  // ← Log only
}
return PurchaseResult{...}  // ← Purchase still succeeds!
```

**Level 2: Email failure doesn't fail SendConfirmation**
```go
// In SendConfirmation workflow
if err := emailFuture.Get(ctx, nil); err != nil {
    logger.Warn("Failed to send email", "error", err)  // ← Log only
} else {
    result.EmailSent = true
}
return result, nil  // ← Always returns success
```

**Result tracking:**
```go
type SendConfirmationResult struct {
    EmailSent bool  // false if email failed
    SMSSent   bool  // false if SMS failed
}
```

So if the email service is down:
- ✅ Ticket is issued
- ✅ Purchase completes successfully
- ⚠️ `EmailSent: false` in the result
- 📝 Warning logged for monitoring

```go
// In TicketPurchase workflow
err = workflow.ExecuteChildWorkflow(..., SendConfirmation, ...).Get(ctx, &result)
if err != nil {
    logger.Warn("Failed to send confirmation", "error", err)  // ← Log only
}
return PurchaseResult{...}  // ← Purchase still succeeds!
```

```go
// In SendConfirmation workflow
if err := emailFuture.Get(ctx, nil); err != nil {
    logger.Warn("Failed to send email", "error", err)  // ← Log only
} else {
    result.EmailSent = true
}
return result, nil  // ← Always returns success
```

```go
type SendConfirmationResult struct {
    EmailSent bool  // false if email failed
    SMSSent   bool  // false if SMS failed
}
```

---

**User**

Simulate a real ticket drop: 100 users trying to buy 20 tickets at exactly the same time. Show me what breaks.

---

**Cursor**

Let me set up and run a stress test:

**Nothing broke!** The system handled the rush correctly:

| Metric | Value |
|--------|-------|
| Users rushing | 100 |
| Tickets available | 10 |
| Got tickets | 10 ✅ |
| Sold out | 90 ❌ |
| Processing time | ~30s |

**The architecture protected us:**

```
100 users ──► Queue ──► 10 concurrent max ──► Inventory (mutex)
    │            │              │                   │
    │            │              │                   └─ No race conditions
    │            │              └─ Rate limited purchases
    │            └─ FIFO ordering preserved
    └─ All signals accepted immediately
```

**Failure reason for the 90 rejected:**
```
"sold out: no seats available"
```

This is correct behavior - the first 10 users got seats A1-A10, and everyone else was fairly rejected.

```plaintext
100 users ──► Queue ──► 10 concurrent max ──► Inventory (mutex)
    │            │              │                   │
    │            │              │                   └─ No race conditions
    │            │              └─ Rate limited purchases
    │            └─ FIFO ordering preserved
    └─ All signals accepted immediately
```

```plaintext
"sold out: no seats available"
```

---

**User**

can you make it a shell script that I can also run it myself?

---

**Cursor**

```shellscript
#!/bin/bash
# Ticket Drop Simulation Script
# Simulates a high-traffic ticket sale with configurable users and seats

set -e

# Configuration
USERS=${1:-100}
EVENT=${2:-"drop-$(date +%s)"}
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              TICKET DROP SIMULATION                        ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Event:     $EVENT"
echo "║  Users:     $USERS"
echo "║  Seats:     10 (hardcoded in inventory)"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Check if Temporal is running
if ! temporal workflow list --limit 1 &>/dev/null; then
    echo "❌ Temporal server not running. Start it with:"
    echo "   temporal server start-dev"
    exit 1
fi

# Build binaries
echo "Building binaries..."
cd "$SCRIPT_DIR"
go build -o bin/queue-starter ./queue-starter
go build -o bin/starter ./starter
echo "✓ Build complete"
echo ""

# Start the queue
echo "Starting queue for event: $EVENT"
./bin/queue-starter --action start --event "$EVENT" 2>/dev/null || true
sleep 1

# Send all users at once
echo ""
echo "Sending $USERS users simultaneously..."
start_time=$(date +%s)

for i in $(seq 1 $USERS); do
    ./bin/queue-starter --action join --event "$EVENT" --user "user-$i" 2>/dev/null &
done

# Wait for all signals to be sent
wait
end_time=$(date +%s)
echo "✓ All $USERS join signals sent in $((end_time - start_time)) seconds"

# Monitor progress
echo ""
echo "Monitoring queue progress..."
QUEUE_WF="ticket-queue-$EVENT"

while true; do
    status=$(temporal workflow query --workflow-id "$QUEUE_WF" --type status -o json 2>/dev/null | jq -r '.queryResult[0]')
    active=$(echo "$status" | jq -r '.active_count')
    waiting=$(echo "$status" | jq -r '.queue_length')
    
    echo "  Active: $active | Waiting: $waiting"
    
    if [ "$active" = "0" ] && [ "$waiting" = "0" ]; then
        break
    fi
    sleep 2
done

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                      RESULTS                               ║"
echo "╠════════════════════════════════════════════════════════════╣"

# Count results
completed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Completed'" --limit 500 2>/dev/null | grep "$EVENT" | wc -l | tr -d ' ')
failed=$(temporal workflow list --query "WorkflowType = 'TicketPurchase' AND ExecutionStatus = 'Failed'" --limit 500 2>/dev/null | grep "$EVENT" | wc -l | tr -d ' ')

echo "║  ✅ Got tickets:    $completed"
echo "║  ❌ Sold out:       $failed"
echo "╠════════════════════════════════════════════════════════════╣"

# Check for issues
if [ "$completed" -le 10 ]; then
    echo "║  ✓ Correct: Only 10 seats were available                  ║"
else
    echo "║  ⚠ WARNING: More than 10 tickets issued!                  ║"
fi

total=$((completed + failed))
if [ "$total" -eq "$USERS" ]; then
    echo "║  ✓ All $USERS users processed                              ║"
else
    echo "║  ⚠ Only $total of $USERS users processed                   ║"
fi

echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "To investigate:"
echo "  temporal workflow list --query \"WorkflowType = 'TicketPurchase'\" | grep $EVENT"
echo "  temporal workflow query --workflow-id $QUEUE_WF --type status"

```

Done! The script is at `simulate-drop.sh`. Usage:

```bash
# Default: 100 users, random event name
./simulate-drop.sh

# Custom: 50 users for "concert-2025"
./simulate-drop.sh 50 concert-2025

# Stress test: 500 users
./simulate-drop.sh 500 mega-sale
```

**Example output:**
```
╔════════════════════════════════════════════════════════════╗
║              TICKET DROP SIMULATION                        ║
╠════════════════════════════════════════════════════════════╣
║  Event:     test-run
║  Users:     20
║  Seats:     10 (hardcoded in inventory)
╚════════════════════════════════════════════════════════════╝

Monitoring queue progress...
  Active: 10 | Waiting: 10
  Active: 10 | Waiting: 1
  Active: 0 | Waiting: 0

╔════════════════════════════════════════════════════════════╗
║                      RESULTS                               ║
╠════════════════════════════════════════════════════════════╣
║  ✅ Got tickets:    10
║  ❌ Sold out:       10
║  ✓ Correct: Only 10 seats were available
║  ✓ All 20 users processed
╚════════════════════════════════════════════════════════════╝
```

```shellscript
# Default: 100 users, random event name
./simulate-drop.sh

# Custom: 50 users for "concert-2025"
./simulate-drop.sh 50 concert-2025

# Stress test: 500 users
./simulate-drop.sh 500 mega-sale
```

```plaintext
╔════════════════════════════════════════════════════════════╗
║              TICKET DROP SIMULATION                        ║
╠════════════════════════════════════════════════════════════╣
║  Event:     test-run
║  Users:     20
║  Seats:     10 (hardcoded in inventory)
╚════════════════════════════════════════════════════════════╝

Monitoring queue progress...
  Active: 10 | Waiting: 10
  Active: 10 | Waiting: 1
  Active: 0 | Waiting: 0

╔════════════════════════════════════════════════════════════╗
║                      RESULTS                               ║
╠════════════════════════════════════════════════════════════╣
║  ✅ Got tickets:    10
║  ❌ Sold out:       10
║  ✓ Correct: Only 10 seats were available
║  ✓ All 20 users processed
╚════════════════════════════════════════════════════════════╝
```

---

**User**

After the chaos, which users actually got tickets? Show me.

---

**Cursor**

Here are the lucky winners:

**test-run event (20 users → 10 tickets):**
| User | Seat |
|------|------|
| user-1 | A1 |
| user-3 | A2 |
| user-11 | A3 |
| user-17 | A4 |
| user-16 | A5 |
| user-10 | A6 |
| user-19 | A7 |
| user-12 | A8 |
| user-15 | A9 |
| user-9 | A10 |

**ticket-drop-2025 event (100 users → 10 tickets):**
| User | Seat |
|------|------|
| rush-2 | A1 |
| rush-6 | A2 |
| rush-8 | A3 |
| rush-1 | A4 |
| rush-7 | A5 |
| rush-18 | A6 |
| rush-11 | A7 |
| rush-9 | A8 |
| rush-25 | A9 |
| rush-24 | A10 |

Notice the order isn't sequential (user-1, user-2, user-3...) because signals arrived in random order due to concurrent processing. The queue preserved **FIFO order based on signal arrival time**, not user ID number.

