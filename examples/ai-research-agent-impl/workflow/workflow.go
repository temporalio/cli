package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
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
	// Use a shorter timeout and limited retries for research activities
	researchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	logger.Info("Researching sub-questions in parallel")
	futures := make([]workflow.Future, len(subQuestions))
	for i, sq := range subQuestions {
		futures[i] = workflow.ExecuteActivity(researchCtx, activity.ResearchSubQuestion, sq)
	}

	// Wait for all research activities to complete
	// Tolerate partial failures: only fail if more than half fail
	var successfulResults []shared.SubQuestion
	var failedCount int

	for i, future := range futures {
		var researched shared.SubQuestion
		if err := future.Get(ctx, &researched); err != nil {
			logger.Warn("Research activity failed", "subQuestion", subQuestions[i].Question, "error", err)
			failedCount++
		} else {
			successfulResults = append(successfulResults, researched)
		}
	}

	// Check if we have enough successful results (more than half must succeed)
	totalCount := len(subQuestions)
	if failedCount > totalCount/2 {
		return nil, fmt.Errorf("too many research activities failed: %d out of %d", failedCount, totalCount)
	}

	logger.Info("Research completed with partial results", "successful", len(successfulResults), "failed", failedCount)

	// Step 3: Synthesize the answers from successful results
	logger.Info("Synthesizing answers")
	var answer string
	err = workflow.ExecuteActivity(ctx, activity.SynthesizeAnswers, req.Question, successfulResults).Get(ctx, &answer)
	if err != nil {
		return nil, err
	}

	// Step 4: Check quality of the synthesized answer
	logger.Info("Checking answer quality")
	var qualityResult shared.QualityCheckResult
	err = workflow.ExecuteActivity(ctx, activity.CheckQuality, req.Question, answer).Get(ctx, &qualityResult)
	if err != nil {
		return nil, err
	}

	logger.Info("Quality check completed", "score", qualityResult.Score, "feedback", qualityResult.Feedback)

	// Fail if quality score is below threshold
	const qualityThreshold = 0.7
	if qualityResult.Score < qualityThreshold {
		return nil, fmt.Errorf("answer quality too low: score %.2f (threshold %.2f). Feedback: %s",
			qualityResult.Score, qualityThreshold, qualityResult.Feedback)
	}

	result := &shared.ResearchResult{
		Question:     req.Question,
		SubQuestions: successfulResults,
		Answer:       answer,
	}

	logger.Info("ResearchWorkflow completed")
	return result, nil
}
