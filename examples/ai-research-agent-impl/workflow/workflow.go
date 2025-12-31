package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/activity"
	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
)

// ResearchWorkflow takes a question, breaks it into sub-questions,
// researches each one, and synthesizes the results.
func ResearchWorkflow(ctx workflow.Context, req shared.ResearchRequest) (*shared.ResearchResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ResearchWorkflow started", "question", req.Question)

	// Activity options with timeout
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Break down the question into sub-questions
	logger.Info("Breaking down question into sub-questions")
	var subQuestions []shared.SubQuestion
	err := workflow.ExecuteActivity(ctx, activity.BreakdownQuestion, req.Question).Get(ctx, &subQuestions)
	if err != nil {
		return nil, err
	}
	logger.Info("Got sub-questions", "count", len(subQuestions))

	// Step 2: Research all sub-questions in parallel
	logger.Info("Researching sub-questions in parallel")
	futures := make([]workflow.Future, len(subQuestions))
	for i, sq := range subQuestions {
		futures[i] = workflow.ExecuteActivity(ctx, activity.ResearchSubQuestion, sq)
	}

	// Wait for all research activities to complete
	researchedQuestions := make([]shared.SubQuestion, len(subQuestions))
	for i, future := range futures {
		var researched shared.SubQuestion
		if err := future.Get(ctx, &researched); err != nil {
			return nil, err
		}
		researchedQuestions[i] = researched
	}

	// Step 3: Synthesize the answers
	logger.Info("Synthesizing answers")
	var answer string
	err = workflow.ExecuteActivity(ctx, activity.SynthesizeAnswers, req.Question, researchedQuestions).Get(ctx, &answer)
	if err != nil {
		return nil, err
	}

	result := &shared.ResearchResult{
		Question:     req.Question,
		SubQuestions: researchedQuestions,
		Answer:       answer,
	}

	logger.Info("ResearchWorkflow completed")
	return result, nil
}

