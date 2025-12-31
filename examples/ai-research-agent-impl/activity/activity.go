package activity

import (
	"context"

	"go.temporal.io/sdk/activity"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
)

// Research performs research for a given question.
// Currently returns a hardcoded answer - will be expanded to perform actual research.
func Research(ctx context.Context, req shared.ResearchRequest) (*shared.ResearchResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Research activity started", "question", req.Question)

	// Hardcoded response for now
	answer := "This is a hardcoded answer. The research agent will be expanded to break down " +
		"complex questions into sub-questions, research each one, and synthesize the results."

	result := &shared.ResearchResult{
		Question: req.Question,
		Answer:   answer,
	}

	logger.Info("Research activity completed")
	return result, nil
}

