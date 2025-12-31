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

// SynthesizeAnswers combines sub-question answers into a coherent final answer.
// In a real implementation, this would use an LLM to create a narrative summary.
func SynthesizeAnswers(ctx context.Context, question string, subQuestions []shared.SubQuestion) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SynthesizeAnswers activity started", "subQuestionCount", len(subQuestions))

	// Simulate synthesis time
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Build a coherent synthesized answer
	var sb strings.Builder

	// Executive summary
	sb.WriteString(fmt.Sprintf("# Research Report: %s\n\n", question))
	sb.WriteString(fmt.Sprintf("## Executive Summary\n"))
	sb.WriteString(fmt.Sprintf("This report synthesizes findings from %d research threads ", len(subQuestions)))
	sb.WriteString("to provide a comprehensive answer to the question above.\n\n")

	// Key findings section
	sb.WriteString("## Key Findings\n\n")
	for i, sq := range subQuestions {
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, extractTopic(sq.Question)))
		sb.WriteString(fmt.Sprintf("%s\n\n", sq.Answer))
	}

	// Conclusion
	sb.WriteString("## Conclusion\n")
	sb.WriteString(fmt.Sprintf("Based on the analysis of %d sub-questions, ", len(subQuestions)))
	sb.WriteString("the research provides multiple perspectives on the topic. ")
	sb.WriteString("The findings above represent the key insights gathered from each research thread.\n\n")

	sb.WriteString(fmt.Sprintf("---\n*Report generated at: %s*", time.Now().Format(time.RFC3339)))

	logger.Info("SynthesizeAnswers activity completed")
	return sb.String(), nil
}

// extractTopic extracts a short topic from a sub-question.
func extractTopic(question string) string {
	// Remove common prefixes to get the core topic
	prefixes := []string{
		"What are the key concepts in: ",
		"What evidence or data supports: ",
		"What are different perspectives on: ",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(question, prefix) {
			return strings.TrimPrefix(question, prefix)
		}
	}
	return truncate(question, 60)
}

// CheckQuality evaluates the quality of a synthesized answer.
// Returns a score between 0.0 and 1.0, with feedback.
func CheckQuality(ctx context.Context, question string, answer string) (shared.QualityCheckResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CheckQuality activity started")

	// Simulate quality check processing
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return shared.QualityCheckResult{}, ctx.Err()
	}

	// Simulate quality scoring based on content analysis
	// In a real implementation, this would use an LLM or other quality metrics
	score := 0.5 // Base score

	// Check for key sections
	if strings.Contains(answer, "## Executive Summary") {
		score += 0.1
	}
	if strings.Contains(answer, "## Key Findings") {
		score += 0.1
	}
	if strings.Contains(answer, "## Conclusion") {
		score += 0.1
	}

	// Check answer length (longer answers tend to be more thorough)
	if len(answer) > 500 {
		score += 0.1
	}
	if len(answer) > 1000 {
		score += 0.1
	}

	// Add some randomness to simulate real-world variability
	score += (rand.Float64() - 0.5) * 0.2

	// Clamp score to valid range
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	// Generate feedback based on score
	var feedback string
	switch {
	case score >= 0.9:
		feedback = "Excellent quality. The answer is comprehensive and well-structured."
	case score >= 0.7:
		feedback = "Good quality. The answer addresses the main points adequately."
	case score >= 0.5:
		feedback = "Moderate quality. The answer could use more depth or structure."
	default:
		feedback = "Low quality. The answer lacks key sections or sufficient detail."
	}

	result := shared.QualityCheckResult{
		Score:    score,
		Feedback: feedback,
	}

	logger.Info("CheckQuality activity completed", "score", score)
	return result, nil
}
