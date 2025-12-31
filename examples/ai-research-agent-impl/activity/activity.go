package activity

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
)

// Research performs research for a given question.
// Simulates processing by sleeping, then returns a formatted response.
func Research(ctx context.Context, req shared.ResearchRequest) (*shared.ResearchResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Research activity started", "question", req.Question)

	// Simulate processing time
	logger.Info("Processing question...")
	select {
	case <-time.After(2 * time.Second):
		// Processing complete
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Generate a formatted response
	answer := fmt.Sprintf("After analyzing the question '%s', here are the findings:\n\n"+
		"1. This question requires further breakdown into sub-questions\n"+
		"2. Each sub-question would be researched independently\n"+
		"3. Results would be synthesized into a final answer\n\n"+
		"[Processed at: %s]",
		req.Question,
		time.Now().Format(time.RFC3339))

	result := &shared.ResearchResult{
		Question: req.Question,
		Answer:   answer,
	}

	logger.Info("Research activity completed")
	return result, nil
}

