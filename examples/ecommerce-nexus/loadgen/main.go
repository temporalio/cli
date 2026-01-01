package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/temporalio/cli/examples/ecommerce-nexus/shared"
	"go.temporal.io/sdk/client"
)

// Scenario types with relative weights
type Scenario struct {
	Name       string
	Weight     int // Relative probability
	CustomerID string
	Amount     float64
	Items      []string
	FailChance float64 // Additional failure probability via customer ID patterns
}

var scenarios = []Scenario{
	{Name: "normal-small", Weight: 40, CustomerID: "customer-%d", Amount: 50.0, Items: []string{"ITEM-001"}},
	{Name: "normal-medium", Weight: 25, CustomerID: "customer-%d", Amount: 150.0, Items: []string{"ITEM-001", "ITEM-002"}},
	{Name: "normal-large", Weight: 10, CustomerID: "customer-%d", Amount: 500.0, Items: []string{"ITEM-001", "ITEM-002", "ITEM-003"}},
	{Name: "fraud-risk", Weight: 8, CustomerID: "customer-FRAUD", Amount: 9999.99, Items: []string{"ITEM-001"}},
	{Name: "out-of-stock", Weight: 7, CustomerID: "customer-%d", Amount: 100.0, Items: []string{"ITEM-OOS"}},
	{Name: "invalid-address", Weight: 5, CustomerID: "customer-%d", Amount: 100.0, Items: []string{"ITEM-001"}},
	{Name: "high-value", Weight: 5, CustomerID: "customer-%d", Amount: 25000.0, Items: []string{"ITEM-PREMIUM"}},
}

// Stats tracks load generation statistics
type Stats struct {
	Started     int64
	Completed   int64
	Failed      int64
	InFlight    int64
	ByScenario  sync.Map
	StartedIDs  sync.Map
	FailedIDs   sync.Map
	CompletedAt time.Time
}

func (s *Stats) RecordStart(id, scenario string) {
	atomic.AddInt64(&s.Started, 1)
	atomic.AddInt64(&s.InFlight, 1)
	s.StartedIDs.Store(id, scenario)

	if val, ok := s.ByScenario.Load(scenario); ok {
		s.ByScenario.Store(scenario, val.(int)+1)
	} else {
		s.ByScenario.Store(scenario, 1)
	}
}

func (s *Stats) RecordComplete(id string) {
	atomic.AddInt64(&s.Completed, 1)
	atomic.AddInt64(&s.InFlight, -1)
	s.StartedIDs.Delete(id)
}

func (s *Stats) RecordFailed(id string) {
	atomic.AddInt64(&s.Failed, 1)
	atomic.AddInt64(&s.InFlight, -1)
	if scenario, ok := s.StartedIDs.Load(id); ok {
		s.FailedIDs.Store(id, scenario)
	}
	s.StartedIDs.Delete(id)
}

func main() {
	// Flags
	duration := flag.Duration("duration", 5*time.Minute, "Duration to run load generation")
	rate := flag.Float64("rate", 2.0, "Orders per second")
	maxConcurrent := flag.Int("max-concurrent", 50, "Maximum concurrent workflows")
	dryRun := flag.Bool("dry-run", false, "Print what would be done without starting workflows")
	statsInterval := flag.Duration("stats-interval", 10*time.Second, "Interval to print stats")
	flag.Parse()

	// Environment
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}
	namespace := os.Getenv("COMMERCE_NS")
	if namespace == "" {
		namespace = "default"
	}
	apiKey := os.Getenv("TEMPORAL_API_KEY")

	log.Printf("Load Generator Configuration:")
	log.Printf("  Duration: %v", *duration)
	log.Printf("  Rate: %.1f orders/sec", *rate)
	log.Printf("  Max Concurrent: %d", *maxConcurrent)
	log.Printf("  Target: %s / %s", address, namespace)

	if *dryRun {
		log.Println("DRY RUN MODE - no workflows will be started")
		simulateDryRun(*duration, *rate)
		return
	}

	// Create client
	clientOpts := client.Options{
		HostPort:  address,
		Namespace: namespace,
	}

	if apiKey != "" {
		clientOpts.Credentials = client.NewAPIKeyStaticCredentials(apiKey)
		clientOpts.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{InsecureSkipVerify: true},
		}
	}

	c, err := client.Dial(clientOpts)
	if err != nil {
		log.Fatalf("Failed to connect to Temporal: %v", err)
	}
	defer c.Close()

	log.Printf("Connected to Temporal at %s, namespace: %s", address, namespace)

	// Run load generation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("\nReceived shutdown signal, stopping load generation...")
		cancel()
	}()

	stats := &Stats{}
	runLoadGeneration(ctx, c, *duration, *rate, *maxConcurrent, *statsInterval, stats)

	// Print final stats
	printFinalStats(stats)
}

func simulateDryRun(duration time.Duration, rate float64) {
	totalOrders := int(duration.Seconds() * rate)
	log.Printf("Would generate approximately %d orders over %v", totalOrders, duration)

	// Show scenario distribution
	totalWeight := 0
	for _, s := range scenarios {
		totalWeight += s.Weight
	}

	log.Println("\nScenario distribution:")
	for _, s := range scenarios {
		pct := float64(s.Weight) / float64(totalWeight) * 100
		count := int(float64(totalOrders) * pct / 100)
		log.Printf("  %s: %.1f%% (~%d orders)", s.Name, pct, count)
	}
}

func runLoadGeneration(ctx context.Context, c client.Client, duration time.Duration, rate float64, maxConcurrent int, statsInterval time.Duration, stats *Stats) {
	interval := time.Duration(float64(time.Second) / rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	deadline := time.Now().Add(duration)
	semaphore := make(chan struct{}, maxConcurrent)

	// Stats printer
	statsTicker := time.NewTicker(statsInterval)
	defer statsTicker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderNum := 0

	log.Printf("Starting load generation for %v...", duration)

	for {
		select {
		case <-ctx.Done():
			log.Println("Load generation cancelled")
			waitForInFlight(stats, 30*time.Second)
			return

		case <-statsTicker.C:
			printStats(stats, deadline)

		case <-ticker.C:
			if time.Now().After(deadline) {
				log.Println("Load generation complete")
				stats.CompletedAt = time.Now()
				waitForInFlight(stats, 30*time.Second)
				return
			}

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
			default:
				// At max concurrency, skip this tick
				continue
			}

			orderNum++
			scenario := pickScenario(rng)
			workflowID := fmt.Sprintf("loadgen-%s-%d-%d", scenario.Name, time.Now().Unix(), orderNum)

			go func(id string, s Scenario) {
				defer func() { <-semaphore }()
				startWorkflow(ctx, c, id, s, stats)
			}(workflowID, scenario)
		}
	}
}

func pickScenario(rng *rand.Rand) Scenario {
	totalWeight := 0
	for _, s := range scenarios {
		totalWeight += s.Weight
	}

	roll := rng.Intn(totalWeight)
	cumulative := 0
	for _, s := range scenarios {
		cumulative += s.Weight
		if roll < cumulative {
			return s
		}
	}
	return scenarios[0]
}

func startWorkflow(ctx context.Context, c client.Client, workflowID string, scenario Scenario, stats *Stats) {
	// Build order
	customerID := scenario.CustomerID
	if customerID == "customer-%d" {
		customerID = fmt.Sprintf("customer-%d", rand.Intn(10000))
	}

	// Build order items from scenario items
	items := make([]shared.OrderItem, len(scenario.Items))
	pricePerItem := scenario.Amount / float64(len(scenario.Items))
	for i, sku := range scenario.Items {
		items[i] = shared.OrderItem{
			SKU:      sku,
			Name:     fmt.Sprintf("Product %s", sku),
			Quantity: 1,
			Price:    pricePerItem,
		}
	}

	order := shared.OrderInput{
		OrderID:    workflowID,
		CustomerID: customerID,
		Items:      items,
		TotalPrice: scenario.Amount,
	}

	stats.RecordStart(workflowID, scenario.Name)

	opts := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "commerce-tasks",
	}

	run, err := c.ExecuteWorkflow(ctx, opts, "OrderSagaWorkflow", order)
	if err != nil {
		log.Printf("Failed to start workflow %s: %v", workflowID, err)
		stats.RecordFailed(workflowID)
		return
	}

	// Wait for completion (with timeout)
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	var result shared.OrderResult
	err = run.Get(waitCtx, &result)
	if err != nil {
		// Expected failures are still recorded
		stats.RecordFailed(workflowID)
	} else {
		stats.RecordComplete(workflowID)
	}
}

func waitForInFlight(stats *Stats, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for atomic.LoadInt64(&stats.InFlight) > 0 && time.Now().Before(deadline) {
		log.Printf("Waiting for %d in-flight workflows...", atomic.LoadInt64(&stats.InFlight))
		time.Sleep(2 * time.Second)
	}
}

func printStats(stats *Stats, deadline time.Time) {
	remaining := time.Until(deadline).Round(time.Second)
	log.Printf("[STATS] Started: %d | Completed: %d | Failed: %d | InFlight: %d | Remaining: %v",
		atomic.LoadInt64(&stats.Started),
		atomic.LoadInt64(&stats.Completed),
		atomic.LoadInt64(&stats.Failed),
		atomic.LoadInt64(&stats.InFlight),
		remaining,
	)
}

func printFinalStats(stats *Stats) {
	log.Println("\n========== LOAD GENERATION COMPLETE ==========")
	log.Printf("Total Started:   %d", atomic.LoadInt64(&stats.Started))
	log.Printf("Total Completed: %d", atomic.LoadInt64(&stats.Completed))
	log.Printf("Total Failed:    %d", atomic.LoadInt64(&stats.Failed))
	log.Printf("Still InFlight:  %d", atomic.LoadInt64(&stats.InFlight))

	failRate := float64(0)
	if stats.Started > 0 {
		failRate = float64(stats.Failed) / float64(stats.Started) * 100
	}
	log.Printf("Failure Rate:    %.1f%%", failRate)

	log.Println("\nScenario breakdown:")
	stats.ByScenario.Range(func(key, value interface{}) bool {
		log.Printf("  %s: %d", key, value)
		return true
	})

	// List failed workflow IDs for debugging
	failCount := 0
	log.Println("\nFailed workflow IDs (for workflow diagnose testing):")
	stats.FailedIDs.Range(func(key, value interface{}) bool {
		failCount++
		if failCount <= 10 {
			log.Printf("  %s (%s)", key, value)
		}
		return true
	})
	if failCount > 10 {
		log.Printf("  ... and %d more", failCount-10)
	}
}
