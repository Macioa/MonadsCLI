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

// parseJSONFromStdout unmarshals trimmed stdout into v; if that fails, strips markdown
// code fences (```json ... ```) and tries again, then tries from last "{" to end.
func parseJSONFromStdout(trimmed string, v interface{}) error {
	try := trimmed
	for _, prefix := range []string{"```json\n", "```json\r\n", "```\n", "```\r\n"} {
		if strings.HasPrefix(try, prefix) {
			try = try[len(prefix):]
			break
		}
	}
	if strings.HasSuffix(try, "```") {
		try = strings.TrimSuffix(try, "```")
		try = strings.TrimSpace(try)
	}
	if err := json.Unmarshal([]byte(try), v); err == nil {
		return nil
	}
	start := strings.LastIndex(try, "{")
	if start < 0 {
		return json.Unmarshal([]byte(try), v)
	}
	return json.Unmarshal([]byte(try[start:]), v)
}

// ParseProcessResponse extracts and parses a ProcessResponse from CLI stdout.
func ParseProcessResponse(stdout string) (ProcessResponse, error) {
	var p ProcessResponse
	trimmed := strings.TrimSpace(stdout)
	if err := parseJSONFromStdout(trimmed, &p); err != nil {
		return p, err
	}
	return p, nil
}

// ParseDecisionResponse extracts and parses a DecisionResponse from CLI stdout.
func ParseDecisionResponse(stdout string) (DecisionResponse, error) {
	var d DecisionResponse
	trimmed := strings.TrimSpace(stdout)
	if err := parseJSONFromStdout(trimmed, &d); err != nil {
		return d, err
	}
	return d, nil
}

// ParseValidationResponse extracts and parses a ValidationResponse from CLI stdout.
// It tries to unmarshal the trimmed output first; if that fails, it looks for the last
// JSON object in the output (e.g. when the CLI echoes other text).
func ParseValidationResponse(stdout string) (ValidationResponse, error) {
	var v ValidationResponse
	trimmed := strings.TrimSpace(stdout)
	if err := parseJSONFromStdout(trimmed, &v); err != nil {
		return v, err
	}
	return v, nil
}
