package main

import (
	"log"
	"os"

	"github.com/temporalio/cli/examples/debug-loop/activities"
	"github.com/temporalio/cli/examples/debug-loop/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const TaskQueue = "debug-loop-tasks"

func main() {
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

	// Create worker
	w := worker.New(c, TaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.ProcessOrderWorkflow)

	// Register activities
	w.RegisterActivity(activities.ProcessPayment)
	w.RegisterActivity(activities.CheckInventory)

	log.Printf("Starting worker on task queue: %s", TaskQueue)
	log.Printf("Connected to Temporal at %s, namespace: %s", address, namespace)

	// Start worker
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
