package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	"github.com/temporalio/cli/examples/agent-demo/workflows"
)

func main() {
	// Command line flags
	scenario := flag.String("scenario", "all", "Scenario to run: success, payment-fail, shipping-fail, nested-fail, all")
	flag.Parse()

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
	if apiKey != "" {
		clientProfile.APIKey = apiKey
	}

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

	// Create Temporal client
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	log.Printf("Connected to Temporal at %s, namespace: %s", address, namespace)

	ctx := context.Background()
	timestamp := time.Now().Format("150405")

	switch *scenario {
	case "success":
		runSuccessScenario(ctx, c, taskQueue, timestamp)
	case "payment-fail":
		runPaymentFailScenario(ctx, c, taskQueue, timestamp)
	case "shipping-fail":
		runShippingFailScenario(ctx, c, taskQueue, timestamp)
	case "nested-fail":
		runNestedFailScenario(ctx, c, taskQueue, timestamp)
	case "all":
		runSuccessScenario(ctx, c, taskQueue, timestamp)
		runPaymentFailScenario(ctx, c, taskQueue, timestamp)
		runShippingFailScenario(ctx, c, taskQueue, timestamp)
		runNestedFailScenario(ctx, c, taskQueue, timestamp)
	default:
		log.Fatalf("Unknown scenario: %s", *scenario)
	}
}

func runSuccessScenario(ctx context.Context, c client.Client, taskQueue, timestamp string) {
	log.Println("=== Running Success Scenario ===")

	// Simple success workflow
	workflowID := fmt.Sprintf("simple-success-%s", timestamp)
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, workflows.SimpleSuccessWorkflow, "hello-world")
	if err != nil {
		log.Printf("Failed to start SimpleSuccessWorkflow: %v", err)
		return
	}
	log.Printf("Started SimpleSuccessWorkflow: %s", run.GetID())

	// Order workflow with success
	workflowID = fmt.Sprintf("order-success-%s", timestamp)
	run, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, workflows.OrderWorkflow, fmt.Sprintf("ORD-%s-OK", timestamp))
	if err != nil {
		log.Printf("Failed to start OrderWorkflow: %v", err)
		return
	}
	log.Printf("Started OrderWorkflow (success): %s", run.GetID())
}

func runPaymentFailScenario(ctx context.Context, c client.Client, taskQueue, timestamp string) {
	log.Println("=== Running Payment Failure Scenario ===")

	// Order workflow that will fail at payment (orderID ends with X)
	workflowID := fmt.Sprintf("order-payment-fail-%s", timestamp)
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, workflows.OrderWorkflow, fmt.Sprintf("ORD-%s-X", timestamp))
	if err != nil {
		log.Printf("Failed to start OrderWorkflow: %v", err)
		return
	}
	log.Printf("Started OrderWorkflow (payment fail): %s", run.GetID())
}

func runShippingFailScenario(ctx context.Context, c client.Client, taskQueue, timestamp string) {
	log.Println("=== Running Shipping Failure Scenario ===")

	// Order workflow that will fail at shipping (orderID ends with Y)
	workflowID := fmt.Sprintf("order-shipping-fail-%s", timestamp)
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, workflows.OrderWorkflow, fmt.Sprintf("ORD-%s-Y", timestamp))
	if err != nil {
		log.Printf("Failed to start OrderWorkflow: %v", err)
		return
	}
	log.Printf("Started OrderWorkflow (shipping fail): %s", run.GetID())
}

func runNestedFailScenario(ctx context.Context, c client.Client, taskQueue, timestamp string) {
	log.Println("=== Running Nested Failure Scenario ===")

	// Nested workflow that will fail 3 levels deep
	workflowID := fmt.Sprintf("nested-failure-%s", timestamp)
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: taskQueue,
	}, workflows.NestedFailureWorkflow, 0, 3)
	if err != nil {
		log.Printf("Failed to start NestedFailureWorkflow: %v", err)
		return
	}
	log.Printf("Started NestedFailureWorkflow (3 levels deep): %s", run.GetID())
}
