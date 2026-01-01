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

