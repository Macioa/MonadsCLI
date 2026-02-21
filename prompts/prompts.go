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
