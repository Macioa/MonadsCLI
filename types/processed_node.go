package types

import (
	"strings"

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

// ProcessedNodeDefaults holds default values (e.g. from settings) used when a node
// does not set a value via tag/metadata. One field per "default" setting; see
// readme/settings.md. Retries of 0 means "use conventional default 3". Timeout of 0
// means use runner default.
type ProcessedNodeDefaults struct {
	CLI         string // DEFAULT_CLI, e.g. "CURSOR"
	ValidateCLI string // DEFAULT_VALIDATE_CLI
	RetryCLI    string // DEFAULT_RETRY_CLI
	Retries     int    // DEFAULT_RETRY_COUNT; 0 = use 3
	Timeout     int    // DEFAULT_TIMEOUT (seconds); 0 = use runner default
}

// ProcessedNode is a recursive linked-list style type for processed document
// structure. Children maps route names to child ProcessedNodes.
// All instruction fields are set by internal→processed conversion; the runner
// uses them as-is (defaults applied only when a field is empty).
type ProcessedNode struct {
	Prompt         string `json:"prompt,omitempty"`
	RetryPrompt    string `json:"retry_prompt,omitempty"`
	ValidatePrompt string `json:"validate_prompt,omitempty"`
	CLI            string `json:"cli,omitempty"`            // Codename for running the node (DEFAULT_CLI).
	ValidateCLI    string `json:"validate_cli,omitempty"`  // Codename for validation (DEFAULT_VALIDATE_CLI).
	RetryCLI       string `json:"retry_cli,omitempty"`    // Codename for retries (DEFAULT_RETRY_CLI).
	Retries        int    `json:"retries,omitempty"`       // Max retry count (DEFAULT_RETRY_COUNT).
	Timeout        int    `json:"timeout,omitempty"`      // Timeout in seconds for CLI ops (DEFAULT_TIMEOUT); 0 = default.
	Retried        int    `json:"retried,omitempty"`      // Number of retries so far (runtime).

	// Children: route name -> child processed node.
	Children map[string]*ProcessedNode `json:"children,omitempty"`
}

// hasTag returns whether the node has the given tag.
// Tag matching is case-insensitive and snake/camel-safe.
func hasTag(n *Node, tag string) bool {
	want := canonicalTag(tag)
	for _, t := range n.Tags {
		if canonicalTag(t) == want {
			return true
		}
	}
	return false
}

// NodeToProcessedNode converts a Node into a ProcessedNode with no settings defaults.
// When node metadata is not present, CLI and ValidateCLI stay empty (runner applies
// settings) and Retries is 3. Use NodeToProcessedNodeWithDefaults to pass defaults
// from settings so they are applied when metadata is absent.
func NodeToProcessedNode(n *Node) *ProcessedNode {
	return NodeToProcessedNodeWithDefaults(n, nil)
}

// NodeToProcessedNodeWithDefaults converts a Node into a ProcessedNode. When defaults
// is non-nil, any CLI/ValidateCLI/Retries not set by the node (tag or metadata) are
// filled from defaults (e.g. DEFAULT_CLI, DEFAULT_VALIDATE_CLI, DEFAULT_RETRY_COUNT).
// When defaults is nil, behavior matches NodeToProcessedNode (empty CLI/ValidateCLI, Retries 3).
func NodeToProcessedNodeWithDefaults(n *Node, defaults *ProcessedNodeDefaults) *ProcessedNode {
	if n == nil {
		return nil
	}
	defaultValidate := prompts.DefaultValidatePrompt()
	codenames := KnownCLICodenames()
	return nodeToProcessedNodeWith(n, defaultValidate, codenames, defaults)
}

func nodeToProcessedNodeWith(n *Node, defaultValidate string, codenames map[string]struct{}, defaults *ProcessedNodeDefaults) *ProcessedNode {
	if n == nil {
		return nil
	}
	res := resolveNodeValues(n, defaultValidate, codenames, defaults)
	out := &ProcessedNode{
		Prompt:         strings.TrimSpace(n.Text),
		ValidatePrompt: res.ValidatePrompt,
		CLI:            res.CLI,
		ValidateCLI:    res.ValidateCLI,
		RetryCLI:       res.RetryCLI,
		Retries:        res.Retries,
		Timeout:        res.Timeout,
	}
	if len(n.Children) > 0 {
		out.Children = make(map[string]*ProcessedNode, len(n.Children))
		for route, child := range n.Children {
			out.Children[route] = nodeToProcessedNodeWith(child, defaultValidate, codenames, defaults)
		}
	}
	return out
}
