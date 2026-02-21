package types

import (
	"github.com/ryanmontgomery/MonadsCLI/prompts"
)

// TagNoValidation is a functional tag. When present on a node, the processed node's
// ValidatePrompt is left empty (validation is skipped).
const TagNoValidation = "NoValidation"

// AvailableTags provides behavioral descriptions for each supported functional tag.
// Implementations hold the set of known tags and return a description for any tag name.
type AvailableTags interface {
	TagDescription(tag string) string
}

// AvailableVariables provides behavioral descriptions for each supported functional variable.
// Implementations hold the set of known variables and return a description for any variable name.
type AvailableVariables interface {
	VariableDescription(name string) string
}

// ProcessedNode is a recursive linked-list style type for processed document
// structure. Children maps route names to child ProcessedNodes.
type ProcessedNode struct {
	Prompt          string `json:"prompt,omitempty"`
	RetryPrompt     string `json:"retry_prompt,omitempty"`
	ValidatePrompt  string `json:"validate_prompt,omitempty"`
	CLI             string `json:"cli,omitempty"`

	// Retries is the max retry count (conventional default 3).
	Retries int `json:"retries,omitempty"`
	// Retried is the number of retries so far (default 0).
	Retried int `json:"retried,omitempty"`

	// Children: route name -> child processed node.
	Children map[string]*ProcessedNode `json:"children,omitempty"`
}

// hasTag returns whether the node has the given tag.
func hasTag(n *Node, tag string) bool {
	for _, t := range n.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// NodeToProcessedNode converts a Node into a ProcessedNode. Prompt is set from the node's Text.
// Unless the node has the NoValidation tag, ValidatePrompt is set to the default validation prompt.
// Children are converted recursively. More functionality (e.g. tags, variables) can be wired in later.
func NodeToProcessedNode(n *Node) *ProcessedNode {
	if n == nil {
		return nil
	}
	out := &ProcessedNode{
		Prompt: n.Text,
	}
	if !hasTag(n, TagNoValidation) {
		out.ValidatePrompt = prompts.DefaultValidatePrompt()
	}
	if len(n.Children) > 0 {
		out.Children = make(map[string]*ProcessedNode, len(n.Children))
		for route, child := range n.Children {
			out.Children[route] = NodeToProcessedNode(child)
		}
	}
	return out
}
