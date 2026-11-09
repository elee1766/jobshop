package domain

// MaxQuestionOptions bounds the options on one question.
const MaxQuestionOptions = 5

// QuestionOption is a suggested answer to a question (JOB-101).
// Recommended marks the asker's pick. A question ticket stores its options
// in meta, and the RPC reports carry them, so it is one type for both.
type QuestionOption struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Recommended bool   `json:"recommended,omitempty"`
}
