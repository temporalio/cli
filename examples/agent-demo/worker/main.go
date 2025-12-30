package main

import (
	"crypto/tls"
	"log"
	"os"
	"strings"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/cli/examples/agent-demo/workflows"
)

func main() {
	// Get configuration from environment
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	apiKey := os.Getenv("TEMPORAL_API_KEY")
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "agent-demo"
	}

	// Check if we should skip TLS verification (for staging environments)
	insecureSkipVerify := os.Getenv("TEMPORAL_TLS_INSECURE") == "true"

	// Use envconfig to build client options like the CLI does
	clientProfile := envconfig.ClientConfigProfile{
		Address:   address,
		Namespace: namespace,
	}

	// If API key is provided, configure it
	if apiKey != "" {
		clientProfile.APIKey = apiKey
		// TLS is automatically enabled when API key is set
	}

	// Convert to client options
	clientOptions, err := clientProfile.ToClientOptions(envconfig.ToClientOptionsRequest{})
	if err != nil {
		log.Fatalf("Failed to create client options: %v", err)
	}

	// For staging environments with self-signed certs
	if insecureSkipVerify || strings.Contains(address, "tmprl-test.cloud") {
		clientOptions.ConnectionOptions.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// Create Temporal client using Dial
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	log.Printf("Connected to Temporal at %s, namespace: %s", address, namespace)

	// Create worker
	w := worker.New(c, taskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.OrderWorkflow)
	w.RegisterWorkflow(workflows.PaymentWorkflow)
	w.RegisterWorkflow(workflows.ShippingWorkflow)
	w.RegisterWorkflow(workflows.NestedFailureWorkflow)
	w.RegisterWorkflow(workflows.SimpleSuccessWorkflow)

	// Register activities
	w.RegisterActivity(workflows.ProcessPaymentActivity)
	w.RegisterActivity(workflows.ShipOrderActivity)
	w.RegisterActivity(workflows.FailingActivity)
	w.RegisterActivity(workflows.SuccessActivity)

	log.Printf("Starting worker on task queue: %s", taskQueue)

	// Run worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
