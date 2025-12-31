package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/temporalio/cli/examples/debug-loop/workflows"
	"go.temporal.io/sdk/client"
)

const TaskQueue = "debug-loop-tasks"

func main() {
	// Flags
	wait := flag.Bool("wait", true, "Wait for workflow completion")
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
	orderID := fmt.Sprintf("order-%d", time.Now().Unix())

	// Start workflow
	input := workflows.OrderInput{
		OrderID: orderID,
		Items:   []string{"ITEM-001", "ITEM-002"}, // Two items to check
		Amount:  99.99,
	}

	opts := client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: TaskQueue,
	}

	log.Printf("Starting order workflow: %s", orderID)
	log.Printf("  Items: %v", input.Items)
	log.Printf("  Amount: $%.2f", input.Amount)

	run, err := c.ExecuteWorkflow(context.Background(), opts, workflows.ProcessOrderWorkflow, input)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	log.Printf("Workflow started: %s (run ID: %s)", run.GetID(), run.GetRunID())

	if *wait {
		log.Println("Waiting for workflow completion...")

		var result workflows.OrderResult
		err = run.Get(context.Background(), &result)
		if err != nil {
			log.Printf("Workflow FAILED: %v", err)
			log.Println("")
			log.Println("=== DEBUG INSTRUCTIONS ===")
			log.Printf("Use the temporal agent CLI to diagnose this failure:")
			log.Println("")
			log.Printf("  temporal agent failures --namespace %s --since 5m -o json", namespace)
			log.Printf("  temporal agent trace --workflow-id %s --namespace %s -o json", orderID, namespace)
			log.Printf("  temporal agent timeline --workflow-id %s --namespace %s --compact -o json", orderID, namespace)
			log.Println("")
			os.Exit(1)
		}

		log.Printf("Workflow completed successfully!")
		log.Printf("  Payment ID: %s", result.PaymentID)
		log.Printf("  Items verified: %d", result.ItemsVerified)
		log.Printf("  Status: %s", result.Status)
	} else {
		log.Println("Workflow started (not waiting for completion)")
	}
}
