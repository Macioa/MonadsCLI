package types

// ProcessResponse holds the result of a process step.
type ProcessResponse struct {
	Completed   bool     `json:"completed"`
	SecsTaken   float64  `json:"secs_taken"`
	TokensUsed float64  `json:"tokens_used"`
	Comments    []string `json:"comments"`
}

// DecisionResponse holds the result of a decision step.
type DecisionResponse struct {
	Choices []string `json:"choices"`
	Answer  string   `json:"answer"`
	Reasons []string `json:"reasons"`
}

// ValidationResponse holds the result of a validation step.
type ValidationResponse struct {
	FullyCompleted    bool     `json:"fully_completed"`
	PartiallyCompleted bool    `json:"partially_completed"`
	ShouldRetry      bool     `json:"should_retry"`
	Warnings         []string `json:"warnings"`
}
