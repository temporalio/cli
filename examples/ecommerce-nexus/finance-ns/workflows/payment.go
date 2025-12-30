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

// ProcessPaymentWorkflow handles payment processing with fraud detection
func ProcessPaymentWorkflow(ctx workflow.Context, input shared.PaymentInput) (shared.PaymentResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProcessPaymentWorkflow started", "orderID", input.OrderID, "amount", input.Amount)

	// First, run fraud check as a child workflow
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("fraud-check-%s", input.OrderID),
	})

	var fraudResult shared.FraudCheckResult
	err := workflow.ExecuteChildWorkflow(childCtx, FraudCheckWorkflow, shared.FraudCheckInput{
		OrderID:    input.OrderID,
		CustomerID: input.CustomerID,
		Amount:     input.Amount,
		CardToken:  input.CardToken,
	}).Get(ctx, &fraudResult)
	if err != nil {
		return shared.PaymentResult{
			Status: "fraud_check_failed",
			Error:  fmt.Sprintf("fraud check failed: %v", err),
		}, err
	}

	if fraudResult.IsFraud {
		return shared.PaymentResult{
			Status: "fraud_detected",
			Error:  fmt.Sprintf("transaction blocked: %s", fraudResult.Reason),
		}, errors.New("fraud detected: " + fraudResult.Reason)
	}

	logger.Info("Fraud check passed", "riskScore", fraudResult.RiskScore)

	// Now process the actual payment
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result shared.PaymentResult
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, input).Get(ctx, &result)
	if err != nil {
		return shared.PaymentResult{
			Status: "payment_failed",
			Error:  err.Error(),
		}, err
	}

	logger.Info("ProcessPaymentWorkflow completed", "paymentID", result.PaymentID)
	return result, nil
}

// FraudCheckWorkflow performs fraud detection
func FraudCheckWorkflow(ctx workflow.Context, input shared.FraudCheckInput) (shared.FraudCheckResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("FraudCheckWorkflow started", "orderID", input.OrderID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result shared.FraudCheckResult
	err := workflow.ExecuteActivity(ctx, FraudCheckActivity, input).Get(ctx, &result)
	if err != nil {
		return shared.FraudCheckResult{}, err
	}

	logger.Info("FraudCheckWorkflow completed", "riskScore", result.RiskScore, "isFraud", result.IsFraud)
	return result, nil
}

// RefundInput is the input for refund operations (matches nexus package)
type RefundInput struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

// RefundResult is the result of refund operations
type RefundResult struct {
	RefundID string `json:"refund_id"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

// RefundPaymentWorkflow handles payment refunds
func RefundPaymentWorkflow(ctx workflow.Context, input RefundInput) (RefundResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("RefundPaymentWorkflow started", "paymentID", input.PaymentID, "amount", input.Amount)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var refundID string
	err := workflow.ExecuteActivity(ctx, RefundPaymentActivity, input.PaymentID, input.Amount, input.Reason).Get(ctx, &refundID)
	if err != nil {
		return RefundResult{Status: "failed", Error: err.Error()}, err
	}

	logger.Info("RefundPaymentWorkflow completed", "refundID", refundID)
	return RefundResult{RefundID: refundID, Status: "completed"}, nil
}

// --- Activities ---

// ProcessPaymentActivity processes a payment through the payment gateway
func ProcessPaymentActivity(ctx context.Context, input shared.PaymentInput) (shared.PaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ProcessPaymentActivity started", "orderID", input.OrderID)

	// Chaos injection
	if err := chaos.MaybeInject(ctx, "finance", "ProcessPayment"); err != nil {
		return shared.PaymentResult{}, err
	}

	// Simulate payment gateway processing
	time.Sleep(500 * time.Millisecond)

	// Simulate various failure scenarios based on card token
	if strings.HasSuffix(input.CardToken, "DECLINED") {
		return shared.PaymentResult{
			Status: "declined",
			Error:  "card declined: insufficient funds",
		}, errors.New("card declined: insufficient funds")
	}

	if strings.HasSuffix(input.CardToken, "TIMEOUT") {
		time.Sleep(35 * time.Second) // Trigger timeout
		return shared.PaymentResult{}, errors.New("payment gateway timeout")
	}

	if strings.HasSuffix(input.CardToken, "ERROR") {
		return shared.PaymentResult{}, errors.New("payment gateway connection refused: ECONNREFUSED 10.0.2.10:443")
	}

	return shared.PaymentResult{
		PaymentID:     fmt.Sprintf("PAY-%s-%d", input.OrderID, time.Now().Unix()),
		Status:        "approved",
		TransactionID: fmt.Sprintf("TXN-%d", time.Now().UnixNano()),
		ProcessedAt:   time.Now(),
	}, nil
}

// FraudCheckActivity checks for fraudulent transactions
func FraudCheckActivity(ctx context.Context, input shared.FraudCheckInput) (shared.FraudCheckResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("FraudCheckActivity started", "orderID", input.OrderID)

	// Chaos injection
	if err := chaos.MaybeInject(ctx, "finance", "FraudCheck"); err != nil {
		return shared.FraudCheckResult{}, err
	}

	// Simulate fraud detection
	time.Sleep(300 * time.Millisecond)

	// Simulate fraud detection based on card token
	if strings.HasSuffix(input.CardToken, "FRAUD") {
		return shared.FraudCheckResult{
			RiskScore: 0.95,
			IsFraud:   true,
			Reason:    "high risk transaction: velocity check failed",
			CheckedAt: time.Now().Format(time.RFC3339),
		}, nil
	}

	if strings.HasSuffix(input.CardToken, "FRAUD_API_ERROR") {
		return shared.FraudCheckResult{}, errors.New("fraud detection API unavailable: service temporarily unavailable")
	}

	return shared.FraudCheckResult{
		RiskScore: 0.1,
		IsFraud:   false,
		CheckedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// RefundPaymentActivity processes a refund
func RefundPaymentActivity(ctx context.Context, paymentID string, amount float64, reason string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("RefundPaymentActivity started", "paymentID", paymentID, "amount", amount)

	// Chaos injection
	if err := chaos.MaybeInject(ctx, "finance", "RefundPayment"); err != nil {
		return "", err
	}

	// Simulate refund processing
	time.Sleep(400 * time.Millisecond)

	refundID := fmt.Sprintf("REF-%s-%d", paymentID, time.Now().Unix())
	logger.Info("RefundPaymentActivity completed", "refundID", refundID)

	return refundID, nil
}

