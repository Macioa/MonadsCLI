package types

import (
	"strconv"
	"strings"
	"unicode"
)

// canonicalMetadataKey converts a variable name to snake_case lowercase for lookup.
// Handles camelCase (validateCli) and snake_case (validate_cli) inputs.
func canonicalMetadataKey(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if (unicode.IsLower(prev) || unicode.IsDigit(prev)) && prev != '_' {
				b.WriteByte('_')
			}
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// canonicalTag normalizes a tag for comparison: removes underscores and lowercases.
// Allows "NoValidation", "no_validation", "noValidation" to match.
func canonicalTag(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if r != '_' {
			b.WriteRune(r)
		}
	}
	return strings.ToLower(b.String())
}

// NodeVariableField identifies which ProcessedNode field a metadata variable sets.
type NodeVariableField int

const (
	FieldCLI            NodeVariableField = iota // CLI codename for running the node (DEFAULT_CLI)
	FieldValidatePrompt                          // Custom validation prompt text (not a default setting)
	FieldValidateCLI                             // CLI codename for validation (DEFAULT_VALIDATE_CLI)
	FieldRetries                                 // Max retry count (DEFAULT_RETRY_COUNT)
	FieldRetryCLI                                // CLI codename for retries (DEFAULT_RETRY_CLI)
	FieldTimeout                                 // Timeout in seconds for CLI operations (DEFAULT_TIMEOUT)
)

// NodeVariableRegistry is the single map of all node metadata variable names
// that affect ProcessedNode. There is one variable per "default" setting in
// readme/settings.md (cli, validate_cli, retries, retry_cli, timeout), plus
// validate_prompt and the cli alias "codename". Keys are canonical (lowercase);
// lookup from Node.Metadata is case-insensitive.
var NodeVariableRegistry = map[string]NodeVariableField{
	"cli":             FieldCLI,
	"codename":        FieldCLI,
	"validate_prompt": FieldValidatePrompt,
	"validate_cli":    FieldValidateCLI,
	"retries":         FieldRetries,
	"retry_cli":       FieldRetryCLI,
	"timeout":         FieldTimeout,
}

// KnownCLICodenames returns the set of all known CLI codenames (uppercase).
// Used to detect when a node tag is a CLI override.
func KnownCLICodenames() map[string]struct{} {
	out := make(map[string]struct{}, len(AllCLIs))
	for _, c := range AllCLIs {
		if c.Codename != "" {
			out[c.Codename] = struct{}{}
		}
	}
	return out
}

// metadataGet returns the value for key from metadata using case-insensitive,
// snake/camel-safe key match. It only considers keys that exist in NodeVariableRegistry.
func metadataGet(metadata map[string]string, key string) (string, bool) {
	if metadata == nil {
		return "", false
	}
	canon := canonicalMetadataKey(key)
	if _, ok := NodeVariableRegistry[canon]; !ok {
		return "", false
	}
	for k, v := range metadata {
		if canonicalMetadataKey(k) == canon {
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}

// resolvedNodeValues holds the resolved values for a Node used to build a ProcessedNode.
type resolvedNodeValues struct {
	CLI            string
	ValidatePrompt string
	ValidateCLI    string
	Retries        int
	RetryCLI       string
	Timeout        int   // seconds; 0 = use runner default
	NoValidation   bool
}

func resolveNodeValues(n *Node, defaultValidatePrompt string, knownCodenames map[string]struct{}, defaults *ProcessedNodeDefaults) resolvedNodeValues {
	retriesDefault := 3
	timeoutDefault := 0
	if defaults != nil {
		if defaults.Retries > 0 {
			retriesDefault = defaults.Retries
		}
		if defaults.Timeout > 0 {
			timeoutDefault = defaults.Timeout
		}
	}
	out := resolvedNodeValues{
		ValidatePrompt: defaultValidatePrompt,
		Retries:        retriesDefault,
		Timeout:        timeoutDefault,
	}
	if n == nil {
		return out
	}

	// NoValidation tag → skip validation (clear ValidatePrompt)
	if hasTag(n, TagNoValidation) {
		out.NoValidation = true
		out.ValidatePrompt = ""
	}

	// CLI from tag: any tag that is a known codename overrides default
	for _, tag := range n.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		upper := strings.ToUpper(tag)
		if _, ok := knownCodenames[upper]; ok {
			out.CLI = upper
			break
		}
	}

	// Metadata variables (single map: NodeVariableRegistry)
	if n.Metadata != nil {
		for canonKey, field := range NodeVariableRegistry {
			val, ok := metadataGet(n.Metadata, canonKey)
			if !ok || val == "" {
				continue
			}
			switch field {
			case FieldCLI:
				out.CLI = strings.ToUpper(strings.TrimSpace(val))
			case FieldValidatePrompt:
				if !out.NoValidation {
					out.ValidatePrompt = val
				}
			case FieldValidateCLI:
				out.ValidateCLI = strings.ToUpper(strings.TrimSpace(val))
			case FieldRetries:
				if i, err := strconv.Atoi(strings.TrimSpace(val)); err == nil && i >= 0 {
					out.Retries = i
				}
			case FieldRetryCLI:
				out.RetryCLI = strings.ToUpper(strings.TrimSpace(val))
			case FieldTimeout:
				if i, err := strconv.Atoi(strings.TrimSpace(val)); err == nil && i >= 0 {
					out.Timeout = i
				}
			}
		}
	}

	// Apply defaults from settings when not set by node (tag or metadata)
	if defaults != nil {
		if out.CLI == "" && defaults.CLI != "" {
			out.CLI = strings.ToUpper(strings.TrimSpace(defaults.CLI))
		}
		if out.ValidateCLI == "" && defaults.ValidateCLI != "" {
			out.ValidateCLI = strings.ToUpper(strings.TrimSpace(defaults.ValidateCLI))
		}
		if out.RetryCLI == "" && defaults.RetryCLI != "" {
			out.RetryCLI = strings.ToUpper(strings.TrimSpace(defaults.RetryCLI))
		}
		if out.Timeout == 0 && defaults.Timeout > 0 {
			out.Timeout = defaults.Timeout
		}
		// Retries already set above from defaults.Retries when defaults != nil
	}

	// If NoValidation was set by tag, keep ValidatePrompt empty (already set above)
	if out.NoValidation {
		out.ValidatePrompt = ""
	}

	return out
}
