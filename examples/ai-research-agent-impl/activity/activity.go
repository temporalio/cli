package activity

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/temporalio/cli/examples/ai-research-agent-impl/shared"
)

// BreakdownQuestion takes a question and returns 3 sub-questions.
// Simulates AI processing by sleeping, then returns generated sub-questions.
func BreakdownQuestion(ctx context.Context, question string) ([]shared.SubQuestion, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("BreakdownQuestion activity started", "question", question)

	// Simulate AI processing time
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Generate sub-questions based on the main question
	// In a real implementation, this would call an LLM
	subQuestions := []shared.SubQuestion{
		{Question: fmt.Sprintf("What are the key concepts in: %s", question)},
		{Question: fmt.Sprintf("What evidence or data supports: %s", question)},
		{Question: fmt.Sprintf("What are different perspectives on: %s", question)},
	}

	logger.Info("BreakdownQuestion activity completed", "count", len(subQuestions))
	return subQuestions, nil
}

// ResearchSubQuestion researches a single sub-question.
// Simulates processing by sleeping, then returns an answer.
func ResearchSubQuestion(ctx context.Context, subQuestion shared.SubQuestion) (shared.SubQuestion, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ResearchSubQuestion activity started", "question", subQuestion.Question)

	// Simulate research time with random duration between 5-15 seconds
	// Some will timeout (over 10s), some will succeed (under 10s)
	sleepDuration := time.Duration(5+rand.Intn(11)) * time.Second
	logger.Info("Simulating research", "duration", sleepDuration)

	select {
	case <-time.After(sleepDuration):
	case <-ctx.Done():
		return shared.SubQuestion{}, ctx.Err()
	}

	// Generate a simulated answer
	result := shared.SubQuestion{
		Question: subQuestion.Question,
		Answer: fmt.Sprintf("Research findings for '%s': This sub-question has been analyzed. "+
			"[Researched at: %s]",
			truncate(subQuestion.Question, 50),
			time.Now().Format(time.RFC3339)),
	}

	logger.Info("ResearchSubQuestion activity completed")
	return result, nil
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SynthesizeAnswers combines sub-question answers into a final answer.
func SynthesizeAnswers(ctx context.Context, question string, subQuestions []shared.SubQuestion) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SynthesizeAnswers activity started", "subQuestionCount", len(subQuestions))

	// Simulate synthesis time
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Build the synthesized answer
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Research Summary for: %s\n\n", question))

	for i, sq := range subQuestions {
		sb.WriteString(fmt.Sprintf("## Finding %d\n", i+1))
		sb.WriteString(fmt.Sprintf("Q: %s\n", sq.Question))
		sb.WriteString(fmt.Sprintf("A: %s\n\n", sq.Answer))
	}

	sb.WriteString(fmt.Sprintf("[Synthesized at: %s]", time.Now().Format(time.RFC3339)))

	logger.Info("SynthesizeAnswers activity completed")
	return sb.String(), nil
}
