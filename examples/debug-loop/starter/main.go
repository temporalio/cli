package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/temporalio/cli/examples/debug-loop/workflows"
	"go.temporal.io/sdk/client"
)

const TaskQueue = "debug-loop-tasks"

func main() {
	// Flags
	wait := flag.Bool("wait", true, "Wait for workflow completion")
	scenario := flag.String("scenario", "race", "Scenario: 'race' (race condition) or 'success' (no race)")
	flag.Parse()

	// Get Temporal address from environment or use default
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// Create client
	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Create unique order ID
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
	log.Println("=== RACE CONDITION SIMULATION ===")
	log.Println("Two orders will compete for the same item (KEYBOARD-03, only 1 in stock)")
	log.Println("")

	// Order 1: Main order with multiple items including the keyboard
	mainInput := workflows.OrderInput{
		OrderID: orderID,
		Items: []workflows.OrderItem{
			{SKU: "LAPTOP-001", Quantity: 1, Price: 999.99},
			{SKU: "MOUSE-002", Quantity: 2, Price: 29.99},
			{SKU: "KEYBOARD-03", Quantity: 1, Price: 149.99}, // Contested item
		},
	}

	// Order 2: Competing order that just wants the keyboard
	competingID := fmt.Sprintf("competing-%d", ts)
	competingInput := workflows.OrderInput{
		OrderID: competingID,
		Items: []workflows.OrderItem{
			{SKU: "KEYBOARD-03", Quantity: 1, Price: 149.99}, // Contested item
		},
	}

	// Start BOTH orders nearly simultaneously
	var wg sync.WaitGroup
	var mainRun, competingRun client.WorkflowRun
	var mainErr, competingErr error

	log.Printf("Starting main order: %s", orderID)
	log.Printf("  Items: LAPTOP-001 x1, MOUSE-002 x2, KEYBOARD-03 x1")

	wg.Add(2)

	// Start main order
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
		log.Printf("Main order started: %s (run ID: %s)", mainRun.GetID(), mainRun.GetRunID())
	}()

	// Start competing order with tiny delay
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Tiny delay to ensure order
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
		log.Println("Workflows started (not waiting for completion)")
		return
	}

	// Wait for both to complete
	log.Println("")
	log.Println("Waiting for both orders to complete...")
	log.Println("(One will succeed, one will fail due to insufficient inventory)")
	log.Println("")

	var wg2 sync.WaitGroup
	wg2.Add(2)

	var mainResult workflows.OrderResult
	var competingResult workflows.OrderResult
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

	// Report results
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

	// Determine which order to debug
	var failedID string
	if mainFinalErr != nil {
		failedID = orderID
	} else if competingFinalErr != nil {
		failedID = competingID
	}

	if failedID != "" {
		log.Println("")
		log.Println("=== DEBUG CHALLENGE ===")
		log.Println("One order's inventory check PASSED but reservation FAILED.")
		log.Println("This is a classic TOCTOU (Time-of-Check to Time-of-Use) race condition!")
		log.Println("")
		log.Println("Use temporal workflow CLI to analyze the failed order:")
		log.Println("")
		log.Printf("  temporal workflow describe --trace-root-cause --workflow-id %s --namespace %s --output json", failedID, namespace)
		log.Printf("  temporal workflow show --compact --workflow-id %s --namespace %s --output json", failedID, namespace)
		log.Println("")
		log.Println("Key insight: Look at the timeline to see that:")
		log.Println("  1. The inventory check PASSED (item was available)")
		log.Println("  2. There was a delay (200ms sleep in workflow)")
		log.Println("  3. The reservation FAILED (item no longer available)")
		log.Println("")
		log.Println("The other order claimed the item during the delay!")
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

	log.Printf("Starting order workflow: %s (scenario: success)", orderID)
	log.Printf("  Items:")
	for _, item := range input.Items {
		log.Printf("    - %s x%d @ $%.2f", item.SKU, item.Quantity, item.Price)
	}

	run, err := c.ExecuteWorkflow(context.Background(), opts, workflows.ProcessOrderWorkflow, input)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	log.Printf("Workflow started: %s (run ID: %s)", run.GetID(), run.GetRunID())

	if wait {
		log.Println("Waiting for workflow completion...")

		var result workflows.OrderResult
		err = run.Get(context.Background(), &result)
		if err != nil {
			log.Printf("Workflow FAILED: %v", err)
			os.Exit(1)
		}

		log.Printf("Workflow completed successfully!")
		log.Printf("  Payment ID: %s", result.PaymentID)
		log.Printf("  Items reserved: %d", result.ItemsReserved)
		log.Printf("  Total: $%.2f", result.TotalAmount)
		log.Printf("  Status: %s", result.Status)
	}
}
