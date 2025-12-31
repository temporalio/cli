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

	"github.com/temporalio/cli/examples/ecommerce-nexus/commerce-ns/workflows"
	"github.com/temporalio/cli/examples/ecommerce-nexus/shared"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
)

var scenarios = map[string]string{
	"success":            "Complete order with all steps succeeding",
	"nexus-payment-fail": "Payment fails via Nexus (card declined)",
	"nexus-fraud":        "Fraud detection via Nexus chain",
	"child-shipping-fail":"Shipping fails via cross-NS child workflow",
	"inventory-fail":     "Inventory reservation fails",
	"saga-compensation":  "Order fails at shipping, triggers compensation",
	"deep-chain":         "4-level cross-NS chain that fails",
	"multi-fail":         "Multiple concurrent failures",
	"timeout":            "Payment timeout via Nexus",
	"all":                "Run all scenarios",
}

func main() {
	scenario := flag.String("scenario", "all", "Scenario to run")
	list := flag.Bool("list", false, "List available scenarios")
	flag.Parse()

	if *list {
		fmt.Println("Available scenarios:")
		for name, desc := range scenarios {
			fmt.Printf("  %-20s %s\n", name, desc)
		}
		return
	}

	// Get configuration
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	namespace := os.Getenv("COMMERCE_NS")
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

	log.Printf("Connected to Temporal at %s, namespace: %s", address, namespace)

	ctx := context.Background()
	timestamp := time.Now().Format("20060102-150405")

	switch *scenario {
	case "success":
		runSuccessScenario(ctx, c, timestamp)
	case "nexus-payment-fail":
		runNexusPaymentFailScenario(ctx, c, timestamp)
	case "nexus-fraud":
		runNexusFraudScenario(ctx, c, timestamp)
	case "child-shipping-fail":
		runChildShippingFailScenario(ctx, c, timestamp)
	case "inventory-fail":
		runInventoryFailScenario(ctx, c, timestamp)
	case "saga-compensation":
		runSagaCompensationScenario(ctx, c, timestamp)
	case "deep-chain":
		runDeepChainScenario(ctx, c, timestamp)
	case "multi-fail":
		runMultiFailScenario(ctx, c, timestamp)
	case "timeout":
		runTimeoutScenario(ctx, c, timestamp)
	case "all":
		runAllScenarios(ctx, c, timestamp)
	default:
		log.Fatalf("Unknown scenario: %s. Use -list to see available scenarios.", *scenario)
	}
}

func runAllScenarios(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running All Scenarios ===")
	
	runSuccessScenario(ctx, c, timestamp)
	time.Sleep(1 * time.Second)
	
	runNexusPaymentFailScenario(ctx, c, timestamp)
	time.Sleep(1 * time.Second)
	
	runNexusFraudScenario(ctx, c, timestamp)
	time.Sleep(1 * time.Second)
	
	runChildShippingFailScenario(ctx, c, timestamp)
	time.Sleep(1 * time.Second)
	
	runInventoryFailScenario(ctx, c, timestamp)
	time.Sleep(1 * time.Second)
	
	runSagaCompensationScenario(ctx, c, timestamp)
	time.Sleep(1 * time.Second)
	
	runDeepChainScenario(ctx, c, timestamp)
	
	log.Println("=== All Scenarios Started ===")
	printDebugCommands()
}

func runSuccessScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Success Scenario ===")
	
	orderID := fmt.Sprintf("order-success-%s", timestamp)
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-success",
		TotalPrice: 99.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-001", Name: "Widget", Quantity: 2, Price: 49.99},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow: %s", run.GetID())
}

func runNexusPaymentFailScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Nexus Payment Fail Scenario ===")
	log.Println("This order's payment will be declined via Nexus call to finance-ns")
	
	orderID := fmt.Sprintf("order-nexus-payment-fail-%s", timestamp)
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-DECLINED", // Triggers card declined
		TotalPrice: 199.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-002", Name: "Premium Widget", Quantity: 1, Price: 199.99},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (payment will fail via Nexus): %s", run.GetID())
}

func runNexusFraudScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Nexus Fraud Detection Scenario ===")
	log.Println("This order will be flagged as fraudulent by finance-ns via Nexus")
	log.Println("Chain: commerce-ns OrderSaga -> [Nexus] -> finance-ns Payment -> finance-ns FraudCheck")
	
	orderID := fmt.Sprintf("order-nexus-fraud-%s", timestamp)
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-FRAUD", // Triggers fraud detection
		TotalPrice: 9999.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-003", Name: "Expensive Widget", Quantity: 10, Price: 999.99},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (fraud detection via Nexus): %s", run.GetID())
}

func runChildShippingFailScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Child Workflow Shipping Fail Scenario ===")
	log.Println("This order will fail at shipping via cross-namespace child workflow")
	log.Println("Chain: commerce-ns OrderSaga -> [Child WF] -> logistics-ns ShipOrder")
	
	orderID := fmt.Sprintf("order-child-shipping-fail-%s", timestamp)
	// We'll manually set an address that triggers failure
	// The order workflow will extract this from the order
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-CARRIER_DOWN", // Address will contain CARRIER_DOWN
		TotalPrice: 49.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-004", Name: "Basic Widget", Quantity: 1, Price: 49.99},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (shipping fail via child WF): %s", run.GetID())
}

func runInventoryFailScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Inventory Fail Scenario ===")
	log.Println("This order will fail at inventory (same namespace)")
	
	orderID := fmt.Sprintf("order-inventory-fail-%s", timestamp)
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-inventory",
		TotalPrice: 149.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-OOS", Name: "Out of Stock Widget", Quantity: 1, Price: 149.99}, // OOS suffix triggers failure
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (inventory fail): %s", run.GetID())
}

func runSagaCompensationScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Saga Compensation Scenario ===")
	log.Println("This order will fail at shipping, triggering:")
	log.Println("  - Refund via Nexus to finance-ns")
	log.Println("  - Inventory release in commerce-ns")
	
	orderID := fmt.Sprintf("order-saga-compensation-%s", timestamp)
	// Use INVALID_ADDRESS to trigger shipping failure after payment succeeds
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-INVALID_ADDRESS",
		TotalPrice: 299.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-005", Name: "Saga Widget", Quantity: 3, Price: 99.99},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (saga compensation): %s", run.GetID())
}

func runDeepChainScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Deep Chain Scenario ===")
	log.Println("4-level failure chain across namespaces:")
	log.Println("  commerce-ns:OrderSaga -> commerce-ns:Inventory")
	log.Println("                        -> [Nexus] finance-ns:Payment -> finance-ns:FraudCheck")
	log.Println("                        -> [Child] logistics-ns:Ship -> logistics-ns:Track")
	
	// Run fraud scenario which creates the deepest chain
	orderID := fmt.Sprintf("order-deep-chain-%s", timestamp)
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-FRAUD",
		TotalPrice: 15000.00,
		Items: []shared.OrderItem{
			{SKU: "ITEM-006", Name: "Deep Chain Widget", Quantity: 100, Price: 150.00},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (deep chain via Nexus + child): %s", run.GetID())
}

func runMultiFailScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Multi-Failure Scenario ===")
	log.Println("Starting multiple orders that will fail in different ways")
	
	// Start multiple failures concurrently
	runNexusPaymentFailScenario(ctx, c, timestamp+"-multi1")
	runNexusFraudScenario(ctx, c, timestamp+"-multi2")
	runChildShippingFailScenario(ctx, c, timestamp+"-multi3")
	runInventoryFailScenario(ctx, c, timestamp+"-multi4")
}

func runTimeoutScenario(ctx context.Context, c client.Client, timestamp string) {
	log.Println("=== Running Timeout Scenario ===")
	log.Println("This order's payment will timeout via Nexus")
	
	orderID := fmt.Sprintf("order-timeout-%s", timestamp)
	input := shared.OrderInput{
		OrderID:    orderID,
		CustomerID: "customer-TIMEOUT", // Triggers payment timeout
		TotalPrice: 499.99,
		Items: []shared.OrderItem{
			{SKU: "ITEM-007", Name: "Timeout Widget", Quantity: 1, Price: 499.99},
		},
	}

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        orderID,
		TaskQueue: shared.CommerceTaskQueue,
	}, workflows.OrderSagaWorkflow, input)
	if err != nil {
		log.Printf("Failed to start workflow: %v", err)
		return
	}

	log.Printf("Started OrderSagaWorkflow (will timeout via Nexus): %s", run.GetID())
}

func printDebugCommands() {
	commerceNS := os.Getenv("COMMERCE_NS")
	if commerceNS == "" {
		commerceNS = "<commerce-namespace>"
	}

	log.Println("\n=== Debug Commands ===")
	log.Println("Find recent failures:")
	log.Printf("  temporal agent failures --namespace %s --since 1h --follow-children --format json", commerceNS)
	log.Println("\nWith leaf-only and compact errors:")
	log.Printf("  temporal agent failures --namespace %s --since 1h --follow-children --leaf-only --compact-errors --format json", commerceNS)
	log.Println("\nTrace a specific order:")
	log.Printf("  temporal agent trace --workflow-id order-<id> --namespace %s --format json", commerceNS)
	log.Println("\nCheck workflow state:")
	log.Printf("  temporal agent state --workflow-id order-<id> --namespace %s --format json", commerceNS)
}

