package nexus

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/temporalio/cli/examples/ecommerce-nexus/finance-ns/workflows"
	"github.com/temporalio/cli/examples/ecommerce-nexus/shared"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
)

// ProcessPaymentOperation is the Nexus operation for processing payments
// This maps the Nexus operation to the ProcessPaymentWorkflow
var ProcessPaymentOperation = temporalnexus.NewWorkflowRunOperation(
	shared.NexusProcessPayment,
	workflows.ProcessPaymentWorkflow,
	func(ctx context.Context, input shared.PaymentInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID:        fmt.Sprintf("payment-%s", input.OrderID),
			TaskQueue: shared.FinanceTaskQueue,
		}, nil
	},
)

// RefundPaymentOperation is the Nexus operation for refunding payments
var RefundPaymentOperation = temporalnexus.NewWorkflowRunOperation(
	shared.NexusRefundPayment,
	workflows.RefundPaymentWorkflow,
	func(ctx context.Context, input workflows.RefundInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID:        fmt.Sprintf("refund-%s", input.PaymentID),
			TaskQueue: shared.FinanceTaskQueue,
		}, nil
	},
)

// NewPaymentService creates the Nexus service with all payment operations
func NewPaymentService() *nexus.Service {
	svc := nexus.NewService(shared.NexusPaymentService)
	if err := svc.Register(ProcessPaymentOperation, RefundPaymentOperation); err != nil {
		panic(fmt.Sprintf("failed to register Nexus operations: %v", err))
	}
	return svc
}

