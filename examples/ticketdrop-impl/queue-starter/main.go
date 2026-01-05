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

