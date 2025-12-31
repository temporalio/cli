package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
	"github.com/temporalio/cli/examples/ai-research-agent-impl/workflow"
)

func main() {
	// Create the Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Questions to research
	questions := []string{
		"What causes climate change?",
		"How does machine learning work?",
		"What is quantum computing?",
		"Why is the sky blue?",
		"How do vaccines work?",
		"What causes economic recessions?",
		"How does photosynthesis work?",
		"What is dark matter?",
		"How do neural networks learn?",
		"What causes earthquakes?",
	}

	log.Printf("Starting %d concurrent research workflows...\n", len(questions))
	startTime := time.Now()

	var wg sync.WaitGroup
	results := make(chan string, len(questions))

	for i, question := range questions {
		wg.Add(1)
		go func(idx int, q string) {
			defer wg.Done()

			workflowID := fmt.Sprintf("research-%s", uuid.New().String()[:8])
			options := client.StartWorkflowOptions{
				ID:        workflowID,
				TaskQueue: shared.TaskQueue,
			}

			req := shared.ResearchRequest{Question: q}

			log.Printf("[%d] Starting workflow %s: %s\n", idx+1, workflowID, q)
			we, err := c.ExecuteWorkflow(context.Background(), options, workflow.ResearchWorkflow, req)
			if err != nil {
				results <- fmt.Sprintf("[%d] ❌ Failed to start: %s - %v", idx+1, q, err)
				return
			}

			var result shared.ResearchResult
			err = we.Get(context.Background(), &result)
			if err != nil {
				results <- fmt.Sprintf("[%d] ❌ Failed: %s - %v", idx+1, q, err)
				return
			}

			results <- fmt.Sprintf("[%d] ✅ Completed: %s", idx+1, q)
		}(i, question)
	}

	// Wait for all workflows to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Print results as they come in
	successCount := 0
	failCount := 0
	for result := range results {
		log.Println(result)
		if strings.Contains(result, "✅") {
			successCount++
		} else {
			failCount++
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("\n=== Load Test Summary ===")
	log.Printf("Total workflows: %d", len(questions))
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failCount)
	log.Printf("Total time: %v", elapsed)
	log.Printf("Avg time per workflow: %v", elapsed/time.Duration(len(questions)))
}

