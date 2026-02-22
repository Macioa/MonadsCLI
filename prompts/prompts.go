package prompts

import (
	_ "embed"
	"strings"
)

//go:embed validate.txt
var defaultValidatePrompt string

// DefaultValidatePrompt returns the default validation prompt used for processed nodes
// when the NoValidation tag is not present.
func DefaultValidatePrompt() string {
	return strings.TrimSpace(defaultValidatePrompt)
}

// ProcessResponseInstruction returns prompt text that instructs the CLI to respond
// with a JSON object matching ProcessResponse (completed, secs_taken, tokens_used, comments).
// Used for childless (leaf) nodes.
func ProcessResponseInstruction() string {
	return "Respond with only a JSON object with keys: completed (boolean), secs_taken (number), tokens_used (number), comments (array of strings)."
}

// DecisionResponseInstruction returns prompt text that instructs the CLI to respond
// with a JSON object matching DecisionResponse (choices, answer, reasons).
// Used for nodes that have children (decision nodes).
func DecisionResponseInstruction() string {
	return "Respond with only a JSON object with keys: choices (array of strings), answer (string), reasons (array of strings)."
}
