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
