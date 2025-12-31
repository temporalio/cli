package shared

const TaskQueue = "research-agent-task-queue"

// ResearchRequest is the input to the research workflow.
type ResearchRequest struct {
	Question string `json:"question"`
}

// ResearchResult is the output from the research workflow.
type ResearchResult struct {
	Question     string        `json:"question"`
	SubQuestions []SubQuestion `json:"sub_questions,omitempty"`
	Answer       string        `json:"answer"`
}

// SubQuestion represents a breakdown of the main question.
type SubQuestion struct {
	Question string `json:"question"`
	Answer   string `json:"answer,omitempty"`
}

// QualityCheckResult contains the quality score and feedback.
type QualityCheckResult struct {
	Score    float64 `json:"score"`    // 0.0 to 1.0
	Feedback string  `json:"feedback"` // Explanation of the score
}

