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

