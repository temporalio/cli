#!/bin/bash
# Creates a "fresh" copy of the debug-loop example with all bug hints removed
# Perfect for testing AI agent diagnosis capabilities

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRESH_DIR="$SCRIPT_DIR/../debug-loop-fresh"

echo "Creating fresh debug-loop example at: $FRESH_DIR"

# Remove existing fresh directory if it exists
rm -rf "$FRESH_DIR"

# Create fresh directory
mkdir -p "$FRESH_DIR"
mkdir -p "$FRESH_DIR/activities"
mkdir -p "$FRESH_DIR/workflows"
mkdir -p "$FRESH_DIR/worker"
mkdir -p "$FRESH_DIR/starter"

# Copy go.mod and go.sum
cp "$SCRIPT_DIR/go.mod" "$FRESH_DIR/"
cp "$SCRIPT_DIR/go.sum" "$FRESH_DIR/" 2>/dev/null || true

# Create clean activities/inventory.go
cat > "$FRESH_DIR/activities/inventory.go" << 'GOEOF'
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
		"KEYBOARD-03": 1,
	}
	reservations = make(map[string]map[string]int)
	inventoryMu  sync.Mutex
)

type InventoryCheckInput struct {
	OrderID  string
	SKU      string
	Quantity int
}

type InventoryCheckResult struct {
	SKU       string
	Available bool
	InStock   int
	Requested int
}

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

type ReserveInventoryInput struct {
	OrderID  string
	SKU      string
	Quantity int
}

type ReserveInventoryResult struct {
	SKU         string
	Reserved    bool
	ReservedQty int
}

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

	inventory[input.SKU] = stock - input.Quantity

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

type ReleaseInventoryInput struct {
	OrderID string
	SKU     string
}

func ReleaseInventory(ctx context.Context, input ReleaseInventoryInput) error {
	inventoryMu.Lock()
	defer inventoryMu.Unlock()

	if reservations[input.OrderID] == nil {
		return nil
	}

	qty, exists := reservations[input.OrderID][input.SKU]
	if !exists {
		return nil
	}

	inventory[input.SKU] += qty
	delete(reservations[input.OrderID], input.SKU)

	return nil
}

type PaymentInput struct {
	OrderID string
	Amount  float64
}

type PaymentResult struct {
	TransactionID string
	Status        string
}

func ProcessPayment(ctx context.Context, input PaymentInput) (*PaymentResult, error) {
	return &PaymentResult{
		TransactionID: fmt.Sprintf("txn-%s", input.OrderID),
		Status:        "approved",
	}, nil
}
GOEOF

# Create clean workflows/order.go
cat > "$FRESH_DIR/workflows/order.go" << 'GOEOF'
package workflows

import (
	"fmt"
	"time"

	"github.com/temporalio/cli/examples/debug-loop-fresh/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type OrderItem struct {
	SKU      string
	Quantity int
	Price    float64
}

type OrderInput struct {
	OrderID string
	Items   []OrderItem
}

type OrderResult struct {
	OrderID       string
	PaymentID     string
	ItemsReserved int
	TotalAmount   float64
	Status        string
}

// ProcessOrderWorkflow processes an order by checking and reserving inventory
func ProcessOrderWorkflow(ctx workflow.Context, input OrderInput) (*OrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Processing order", "orderID", input.OrderID, "items", len(input.Items))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Check inventory for all items
	logger.Info("Checking inventory for all items")
	
	checkFutures := make([]workflow.Future, len(input.Items))
	for i, item := range input.Items {
		checkFutures[i] = workflow.ExecuteActivity(ctx, activities.CheckInventory, activities.InventoryCheckInput{
			OrderID:  input.OrderID,
			SKU:      item.SKU,
			Quantity: item.Quantity,
		})
	}

	checkResults := make([]*activities.InventoryCheckResult, len(input.Items))
	for i, future := range checkFutures {
		var result activities.InventoryCheckResult
		if err := future.Get(ctx, &result); err != nil {
			return nil, fmt.Errorf("inventory check failed for %s: %w", input.Items[i].SKU, err)
		}
		if !result.Available {
			return nil, fmt.Errorf("item %s not available: requested %d, in stock %d",
				result.SKU, result.Requested, result.InStock)
		}
		checkResults[i] = &result
		logger.Info("Inventory check passed", "sku", result.SKU, "inStock", result.InStock)
	}

	logger.Info("All inventory checks passed, proceeding to reserve")

	// Processing delay
	workflow.Sleep(ctx, 200*time.Millisecond)

	// Step 2: Reserve inventory for each item
	reservedItems := []string{}
	var totalAmount float64

	for _, item := range input.Items {
		var result activities.ReserveInventoryResult
		err := workflow.ExecuteActivity(ctx, activities.ReserveInventory, activities.ReserveInventoryInput{
			OrderID:  input.OrderID,
			SKU:      item.SKU,
			Quantity: item.Quantity,
		}).Get(ctx, &result)

		if err != nil {
			logger.Error("Reservation failed, releasing reserved items", "failedSKU", item.SKU, "error", err)
			for _, sku := range reservedItems {
				_ = workflow.ExecuteActivity(ctx, activities.ReleaseInventory, activities.ReleaseInventoryInput{
					OrderID: input.OrderID,
					SKU:     sku,
				}).Get(ctx, nil)
			}
			return nil, err
		}

		reservedItems = append(reservedItems, item.SKU)
		totalAmount += item.Price * float64(item.Quantity)
		logger.Info("Reserved inventory", "sku", item.SKU, "quantity", item.Quantity)
	}

	// Step 3: Process payment
	var paymentResult activities.PaymentResult
	err := workflow.ExecuteActivity(ctx, activities.ProcessPayment, activities.PaymentInput{
		OrderID: input.OrderID,
		Amount:  totalAmount,
	}).Get(ctx, &paymentResult)
	if err != nil {
		for _, sku := range reservedItems {
			_ = workflow.ExecuteActivity(ctx, activities.ReleaseInventory, activities.ReleaseInventoryInput{
				OrderID: input.OrderID,
				SKU:     sku,
			}).Get(ctx, nil)
		}
		return nil, err
	}

	return &OrderResult{
		OrderID:       input.OrderID,
		PaymentID:     paymentResult.TransactionID,
		ItemsReserved: len(reservedItems),
		TotalAmount:   totalAmount,
		Status:        "completed",
	}, nil
}
GOEOF

# Create clean worker/main.go
cat > "$FRESH_DIR/worker/main.go" << 'GOEOF'
package main

import (
	"log"
	"os"

	"github.com/temporalio/cli/examples/debug-loop-fresh/activities"
	"github.com/temporalio/cli/examples/debug-loop-fresh/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const TaskQueue = "debug-loop-tasks"

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	w := worker.New(c, TaskQueue, worker.Options{})

	w.RegisterWorkflow(workflows.ProcessOrderWorkflow)
	w.RegisterActivity(activities.CheckInventory)
	w.RegisterActivity(activities.ReserveInventory)
	w.RegisterActivity(activities.ReleaseInventory)
	w.RegisterActivity(activities.ProcessPayment)

	log.Printf("Starting worker on task queue: %s", TaskQueue)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
GOEOF

# Create clean starter/main.go
cat > "$FRESH_DIR/starter/main.go" << 'GOEOF'
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/temporalio/cli/examples/debug-loop-fresh/workflows"
	"go.temporal.io/sdk/client"
)

const TaskQueue = "debug-loop-tasks"

func main() {
	wait := flag.Bool("wait", true, "Wait for workflow completion")
	scenario := flag.String("scenario", "race", "Scenario: 'race' or 'success'")
	flag.Parse()

	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	ts := time.Now().UnixNano()
	orderID := fmt.Sprintf("order-%d", ts)

	switch *scenario {
	case "race":
		runRaceScenario(c, namespace, orderID, ts, *wait)
	case "success":
		runSuccessScenario(c, namespace, orderID, *wait)
	default:
		log.Fatalf("Unknown scenario: %s", *scenario)
	}
}

func runRaceScenario(c client.Client, namespace, orderID string, ts int64, wait bool) {
	log.Println("=== Running scenario: race ===")
	log.Println("Two orders will compete for the same item")
	log.Println("")

	mainInput := workflows.OrderInput{
		OrderID: orderID,
		Items: []workflows.OrderItem{
			{SKU: "LAPTOP-001", Quantity: 1, Price: 999.99},
			{SKU: "MOUSE-002", Quantity: 2, Price: 29.99},
			{SKU: "KEYBOARD-03", Quantity: 1, Price: 149.99},
		},
	}

	competingID := fmt.Sprintf("competing-%d", ts)
	competingInput := workflows.OrderInput{
		OrderID: competingID,
		Items: []workflows.OrderItem{
			{SKU: "KEYBOARD-03", Quantity: 1, Price: 149.99},
		},
	}

	var wg sync.WaitGroup
	var mainRun, competingRun client.WorkflowRun
	var mainErr, competingErr error

	log.Printf("Starting main order: %s", orderID)
	log.Printf("  Items: LAPTOP-001 x1, MOUSE-002 x2, KEYBOARD-03 x1")

	wg.Add(2)

	go func() {
		defer wg.Done()
		opts := client.StartWorkflowOptions{
			ID:        orderID,
			TaskQueue: TaskQueue,
		}
		mainRun, mainErr = c.ExecuteWorkflow(context.Background(), opts, workflows.ProcessOrderWorkflow, mainInput)
		if mainErr != nil {
			log.Printf("Failed to start main order: %v", mainErr)
			return
		}
		log.Printf("Main order started: %s", mainRun.GetID())
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		opts := client.StartWorkflowOptions{
			ID:        competingID,
			TaskQueue: TaskQueue,
		}
		competingRun, competingErr = c.ExecuteWorkflow(context.Background(), opts, workflows.ProcessOrderWorkflow, competingInput)
		if competingErr != nil {
			log.Printf("Failed to start competing order: %v", competingErr)
			return
		}
		log.Printf("Competing order started: %s", competingRun.GetID())
	}()

	wg.Wait()

	if mainErr != nil || competingErr != nil {
		log.Fatal("Failed to start workflows")
	}

	if !wait {
		log.Println("Workflows started (not waiting)")
		return
	}

	log.Println("")
	log.Println("Waiting for both orders to complete...")
	log.Println("")

	var wg2 sync.WaitGroup
	wg2.Add(2)

	var mainResult, competingResult workflows.OrderResult
	var mainFinalErr, competingFinalErr error

	go func() {
		defer wg2.Done()
		mainFinalErr = mainRun.Get(context.Background(), &mainResult)
	}()

	go func() {
		defer wg2.Done()
		competingFinalErr = competingRun.Get(context.Background(), &competingResult)
	}()

	wg2.Wait()

	log.Println("=== RESULTS ===")
	if mainFinalErr != nil {
		log.Printf("Main order FAILED: %v", mainFinalErr)
	} else {
		log.Printf("Main order SUCCEEDED")
	}

	if competingFinalErr != nil {
		log.Printf("Competing order FAILED: %v", competingFinalErr)
	} else {
		log.Printf("Competing order SUCCEEDED")
	}

	var failedID string
	if mainFinalErr != nil {
		failedID = orderID
	} else if competingFinalErr != nil {
		failedID = competingID
	}

	if failedID != "" {
		log.Println("")
		log.Println("=== DEBUG CHALLENGE ===")
		log.Println("One order failed. Use temporal workflow CLI to diagnose why.")
		log.Println("")
		log.Printf("  temporal workflow diagnose --workflow-id %s --namespace %s --format json", failedID, namespace)
		log.Printf("  temporal workflow show --compact --workflow-id %s --namespace %s --format json", failedID, namespace)
		log.Println("")
		log.Println("Question: Why did the inventory check pass but the reservation fail?")
	}
}

func runSuccessScenario(c client.Client, namespace, orderID string, wait bool) {
	input := workflows.OrderInput{
		OrderID: orderID,
		Items: []workflows.OrderItem{
			{SKU: "LAPTOP-001", Quantity: 1, Price: 999.99},
			{SKU: "MOUSE-002", Quantity: 2, Price: 29.99},
		},
	}

	opts := client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: TaskQueue,
	}

	log.Printf("Starting order: %s", orderID)

	run, err := c.ExecuteWorkflow(context.Background(), opts, workflows.ProcessOrderWorkflow, input)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	if wait {
		var result workflows.OrderResult
		err = run.Get(context.Background(), &result)
		if err != nil {
			log.Printf("Workflow FAILED: %v", err)
			os.Exit(1)
		}
		log.Printf("Workflow completed: %s", result.Status)
	}
}
GOEOF

# Update go.mod module path
sed -i '' 's|debug-loop|debug-loop-fresh|g' "$FRESH_DIR/go.mod"

# Create a clean README without solutions
cat > "$FRESH_DIR/README.md" << 'EOF'
# Debug Loop Challenge

An order processing workflow is failing with inventory errors. Diagnose the root cause using `temporal workflow` CLI.

## Setup

### 1. Start Dev Server

```bash
temporal server start-dev
```

### 2. Start Worker

```bash
cd examples/debug-loop-fresh
go run ./worker
```

### 3. Run the Failing Scenario

```bash
go run ./starter --scenario race
```

## The Problem

When running the `race` scenario, one order fails with:

```
insufficient inventory for KEYBOARD-03: requested 1, available 0
```

But the workflow checks inventory before reserving. Why does the check pass but the reservation fail?

## Your Task

Diagnose using `temporal workflow`:

```bash
temporal workflow failures --namespace default --since 5m --format json
temporal workflow diagnose --workflow-id <id> --namespace default --format json
temporal workflow show --compact --workflow-id <id> --namespace default --format json
```

## Questions

1. Did the inventory check pass?
2. What value did it show?
3. When did the reservation fail?
4. What is the root cause?
5. How would you fix it?

## Hints

<details>
<summary>Hint 1</summary>
The timeline shows precise timestamps for each event.
</details>

<details>
<summary>Hint 2</summary>
Look for gaps between related operations.
</details>

<details>
<summary>Hint 3</summary>
There are two workflows running. What might happen if they both want the same item?
</details>
EOF

echo ""
echo "Fresh example created at: $FRESH_DIR"
echo ""
echo "Hints removed:"
echo "  - All BUG/race comments from source code"
echo "  - RESULTS.md (solution document)"
echo "  - Solution sections from README.md"
echo ""
echo "To test:"
echo "  cd $FRESH_DIR"
echo "  go run ./worker  # terminal 1"
echo "  go run ./starter --scenario race  # terminal 2"
