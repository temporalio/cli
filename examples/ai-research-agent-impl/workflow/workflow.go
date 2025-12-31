package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
)

// ResearchWorkflow takes a question and returns an answer.
// Currently returns a hardcoded response - will be expanded to break down questions,
// research sub-questions, and synthesize answers.
func ResearchWorkflow(ctx workflow.Context, req shared.ResearchRequest) (*shared.ResearchResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ResearchWorkflow started", "question", req.Question)

	// Activity options with timeout
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute the research activity
	var result shared.ResearchResult
	err := workflow.ExecuteActivity(ctx, "Research", req).Get(ctx, &result)
	if err != nil {
		return nil, err
	}

	logger.Info("ResearchWorkflow completed", "answer", result.Answer)
	return &result, nil
}

