package types

import (
	"encoding/json"
	"strings"
)

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
	FullyCompleted     bool     `json:"fully_completed"`
	PartiallyCompleted bool    `json:"partially_completed"`
	ShouldRetry       bool     `json:"should_retry"`
	Warnings          []string `json:"warnings"`
}

// ParseValidationResponse extracts and parses a ValidationResponse from CLI stdout.
// It tries to unmarshal the trimmed output first; if that fails, it looks for the last
// JSON object in the output (e.g. when the CLI echoes other text).
func ParseValidationResponse(stdout string) (ValidationResponse, error) {
	var v ValidationResponse
	trimmed := strings.TrimSpace(stdout)
	if err := json.Unmarshal([]byte(trimmed), &v); err == nil {
		return v, nil
	}
	// Try to find a JSON object: last '{' to end
	start := strings.LastIndex(trimmed, "{")
	if start < 0 {
		err := json.Unmarshal([]byte(trimmed), &v)
		return v, err
	}
	if err := json.Unmarshal([]byte(trimmed[start:]), &v); err != nil {
		return v, err
	}
	return v, nil
}
