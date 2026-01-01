package main

import (
	"crypto/tls"
	"log"
	"os"
	"strings"

	"github.com/temporalio/cli/examples/ecommerce-nexus/chaos"
	financenexus "github.com/temporalio/cli/examples/ecommerce-nexus/finance-ns/nexus"
	"github.com/temporalio/cli/examples/ecommerce-nexus/finance-ns/workflows"
	"github.com/temporalio/cli/examples/ecommerce-nexus/shared"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Initialize chaos injection from environment
	chaos.Init()

	// Get configuration
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	namespace := os.Getenv("FINANCE_NS")
	if namespace == "" {
		namespace = os.Getenv("TEMPORAL_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}
	}

	apiKey := os.Getenv("TEMPORAL_API_KEY")

	// Build client options
	clientProfile := envconfig.ClientConfigProfile{
		Address:   address,
		Namespace: namespace,
	}
	if apiKey != "" {
		clientProfile.APIKey = apiKey
	}

	clientOptions, err := clientProfile.ToClientOptions(envconfig.ToClientOptionsRequest{})
	if err != nil {
		log.Fatalf("Failed to create client options: %v", err)
	}

	// Handle TLS based on environment
	if strings.Contains(address, "tmprl-test.cloud") {
		// Staging: use TLS with self-signed cert
		clientOptions.ConnectionOptions.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else if strings.Contains(address, "localhost") || strings.Contains(address, "127.0.0.1") {
		// Local dev server: no TLS
		clientOptions.ConnectionOptions.TLS = nil
	}

	// Create Temporal client
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	log.Printf("[finance-ns] Connected to Temporal at %s, namespace: %s", address, namespace)

	// Create worker
	w := worker.New(c, shared.FinanceTaskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.ProcessPaymentWorkflow)
	w.RegisterWorkflow(workflows.FraudCheckWorkflow)
	w.RegisterWorkflow(workflows.RefundPaymentWorkflow)

	// Register activities
	w.RegisterActivity(workflows.ProcessPaymentActivity)
	w.RegisterActivity(workflows.FraudCheckActivity)
	w.RegisterActivity(workflows.RefundPaymentActivity)

	// Register Nexus service with payment operations
	paymentService := financenexus.NewPaymentService()
	w.RegisterNexusService(paymentService)

	log.Printf("[finance-ns] Starting worker on task queue: %s", shared.FinanceTaskQueue)
	log.Printf("[finance-ns] Nexus service registered: %s", shared.NexusPaymentService)

	// Run worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}

