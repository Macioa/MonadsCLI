package types

import (
	"reflect"
	"testing"
)

func TestNodeToProcessedNode(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := NodeToProcessedNode(nil); got != nil {
			t.Errorf("NodeToProcessedNode(nil) = %v, want nil", got)
		}
	})

	// When metadata is not present and no defaults are passed: CLI and ValidateCLI
	// stay empty (runner applies settings); Retries is conventional 3; ValidatePrompt
	// is the default prompt. Conversion does not read settings; callers use
	// NodeToProcessedNodeWithDefaults(n, settingsDefaults) to apply settings.
	t.Run("no_metadata_no_defaults_CLI_and_ValidateCLI_empty_Retries_3", func(t *testing.T) {
		n := &Node{Text: "Do the thing"}
		p := NodeToProcessedNode(n)
		if p == nil {
			t.Fatal("ProcessedNode is nil")
		}
		if p.Prompt != "Do the thing" {
			t.Errorf("Prompt = %q, want Do the thing", p.Prompt)
		}
		if p.ValidatePrompt == "" {
			t.Error("ValidatePrompt should be default when no NoValidation tag")
		}
		if p.CLI != "" {
			t.Errorf("CLI = %q, want empty (runner uses settings DEFAULT_CLI)", p.CLI)
		}
		if p.ValidateCLI != "" {
			t.Errorf("ValidateCLI = %q, want empty (runner uses settings DEFAULT_VALIDATE_CLI)", p.ValidateCLI)
		}
		if p.Retries != 3 {
			t.Errorf("Retries = %d, want 3 (conventional default)", p.Retries)
		}
		if p.RetryCLI != "" {
			t.Errorf("RetryCLI = %q, want empty (runner uses settings DEFAULT_RETRY_CLI)", p.RetryCLI)
		}
		if p.Timeout != 0 {
			t.Errorf("Timeout = %d, want 0 (runner uses settings DEFAULT_TIMEOUT)", p.Timeout)
		}
	})

	t.Run("no_metadata_with_defaults_uses_settings_defaults", func(t *testing.T) {
		n := &Node{Text: "Step"}
		defaults := &ProcessedNodeDefaults{
			CLI:         "CURSOR",
			ValidateCLI: "GEMINI",
			RetryCLI:    "AIDER",
			Retries:     5,
			Timeout:     600,
		}
		p := NodeToProcessedNodeWithDefaults(n, defaults)
		if p.CLI != "CURSOR" {
			t.Errorf("CLI = %q, want CURSOR (from defaults when metadata not present)", p.CLI)
		}
		if p.ValidateCLI != "GEMINI" {
			t.Errorf("ValidateCLI = %q, want GEMINI (from defaults)", p.ValidateCLI)
		}
		if p.RetryCLI != "AIDER" {
			t.Errorf("RetryCLI = %q, want AIDER (from defaults)", p.RetryCLI)
		}
		if p.Retries != 5 {
			t.Errorf("Retries = %d, want 5 (from defaults)", p.Retries)
		}
		if p.Timeout != 600 {
			t.Errorf("Timeout = %d, want 600 (from defaults)", p.Timeout)
		}
	})

	t.Run("defaults_Retries_zero_uses_conventional_3", func(t *testing.T) {
		n := &Node{Text: "Step"}
		defaults := &ProcessedNodeDefaults{CLI: "CURSOR", Retries: 0}
		p := NodeToProcessedNodeWithDefaults(n, defaults)
		if p.Retries != 3 {
			t.Errorf("Retries = %d, want 3 when defaults.Retries is 0", p.Retries)
		}
	})

	t.Run("node_metadata_overrides_defaults", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"cli": "CLAUDE", "validate_cli": "AIDER", "retry_cli": "QODO", "retries": "2", "timeout": "120"},
		}
		defaults := &ProcessedNodeDefaults{CLI: "CURSOR", ValidateCLI: "CURSOR", RetryCLI: "CURSOR", Retries: 3, Timeout: 600}
		p := NodeToProcessedNodeWithDefaults(n, defaults)
		if p.CLI != "CLAUDE" {
			t.Errorf("CLI = %q, want CLAUDE (metadata overrides default)", p.CLI)
		}
		if p.ValidateCLI != "AIDER" {
			t.Errorf("ValidateCLI = %q, want AIDER (metadata overrides default)", p.ValidateCLI)
		}
		if p.RetryCLI != "QODO" {
			t.Errorf("RetryCLI = %q, want QODO (metadata overrides default)", p.RetryCLI)
		}
		if p.Retries != 2 {
			t.Errorf("Retries = %d, want 2 (metadata overrides default)", p.Retries)
		}
		if p.Timeout != 120 {
			t.Errorf("Timeout = %d, want 120 (metadata overrides default)", p.Timeout)
		}
	})

	t.Run("NoValidation tag", func(t *testing.T) {
		n := &Node{Text: "Skip", Tags: []string{TagNoValidation}}
		p := NodeToProcessedNode(n)
		if p.ValidatePrompt != "" {
			t.Errorf("ValidatePrompt = %q, want empty when NoValidation tag", p.ValidatePrompt)
		}
	})

	t.Run("CLI from tag", func(t *testing.T) {
		n := &Node{Text: "Use Gemini", Tags: []string{"GEMINI"}}
		p := NodeToProcessedNode(n)
		if p.CLI != "GEMINI" {
			t.Errorf("CLI = %q, want GEMINI", p.CLI)
		}
	})

	t.Run("CLI from metadata", func(t *testing.T) {
		n := &Node{
			Text:     "Use Claude",
			Metadata: map[string]string{"cli": "CLAUDE"},
		}
		p := NodeToProcessedNode(n)
		if p.CLI != "CLAUDE" {
			t.Errorf("CLI = %q, want CLAUDE", p.CLI)
		}
	})

	t.Run("codename metadata alias", func(t *testing.T) {
		n := &Node{
			Text:     "Codename",
			Metadata: map[string]string{"codename": "cursor"},
		}
		p := NodeToProcessedNode(n)
		if p.CLI != "CURSOR" {
			t.Errorf("CLI = %q, want CURSOR", p.CLI)
		}
	})

	t.Run("validate_prompt from metadata", func(t *testing.T) {
		custom := "Did the model follow instructions?"
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"validate_prompt": custom},
		}
		p := NodeToProcessedNode(n)
		if p.ValidatePrompt != custom {
			t.Errorf("ValidatePrompt = %q, want %q", p.ValidatePrompt, custom)
		}
	})

	t.Run("validate_cli from metadata", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"validate_cli": "GEMINI"},
		}
		p := NodeToProcessedNode(n)
		if p.ValidateCLI != "GEMINI" {
			t.Errorf("ValidateCLI = %q, want GEMINI", p.ValidateCLI)
		}
	})

	t.Run("retries from metadata", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"retries": "5"},
		}
		p := NodeToProcessedNode(n)
		if p.Retries != 5 {
			t.Errorf("Retries = %d, want 5", p.Retries)
		}
	})

	t.Run("retry_cli from metadata", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"retry_cli": "GEMINI"},
		}
		p := NodeToProcessedNode(n)
		if p.RetryCLI != "GEMINI" {
			t.Errorf("RetryCLI = %q, want GEMINI", p.RetryCLI)
		}
	})

	t.Run("timeout from metadata", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"timeout": "300"},
		}
		p := NodeToProcessedNode(n)
		if p.Timeout != 300 {
			t.Errorf("Timeout = %d, want 300", p.Timeout)
		}
	})

	t.Run("children converted recursively", func(t *testing.T) {
		n := &Node{
			Text: "Root",
			Children: map[string]*Node{
				"Yes": {Text: "Yes branch", Tags: []string{"NoValidation"}},
				"No":  {Text: "No branch", Metadata: map[string]string{"cli": "GEMINI"}},
			},
		}
		p := NodeToProcessedNode(n)
		if p.Prompt != "Root" {
			t.Errorf("root Prompt = %q", p.Prompt)
		}
		yes, ok := p.Children["Yes"]
		if !ok || yes == nil {
			t.Fatal("missing Children[Yes]")
		}
		if yes.ValidatePrompt != "" {
			t.Errorf("Yes.ValidatePrompt = %q, want empty", yes.ValidatePrompt)
		}
		no, ok := p.Children["No"]
		if !ok || no == nil {
			t.Fatal("missing Children[No]")
		}
		if no.CLI != "GEMINI" {
			t.Errorf("No.CLI = %q, want GEMINI", no.CLI)
		}
	})

	t.Run("metadata case-insensitive", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"Validate_Prompt": "Custom?"},
		}
		p := NodeToProcessedNode(n)
		if p.ValidatePrompt != "Custom?" {
			t.Errorf("ValidatePrompt = %q, want Custom? (case-insensitive key)", p.ValidatePrompt)
		}
	})

	t.Run("metadata snake/camel key", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"validateCli": "GEMINI"},
		}
		p := NodeToProcessedNode(n)
		if p.ValidateCLI != "GEMINI" {
			t.Errorf("ValidateCLI = %q, want GEMINI (validateCli camelCase key)", p.ValidateCLI)
		}
	})

	t.Run("metadata value case-insensitive", func(t *testing.T) {
		n := &Node{
			Text:     "Step",
			Metadata: map[string]string{"cli": "cursor"},
		}
		p := NodeToProcessedNode(n)
		if p.CLI != "CURSOR" {
			t.Errorf("CLI = %q, want CURSOR (value normalized to uppercase)", p.CLI)
		}
	})

	t.Run("NoValidation tag case-insensitive", func(t *testing.T) {
		n := &Node{Text: "Skip", Tags: []string{"novalidation"}}
		p := NodeToProcessedNode(n)
		if p.ValidatePrompt != "" {
			t.Errorf("ValidatePrompt = %q, want empty when novalidation tag", p.ValidatePrompt)
		}
	})

	t.Run("NoValidation tag snake/camel", func(t *testing.T) {
		n := &Node{Text: "Skip", Tags: []string{"no_validation"}}
		p := NodeToProcessedNode(n)
		if p.ValidatePrompt != "" {
			t.Errorf("ValidatePrompt = %q, want empty when no_validation tag", p.ValidatePrompt)
		}
		n2 := &Node{Text: "Skip", Tags: []string{"noValidation"}}
		p2 := NodeToProcessedNode(n2)
		if p2.ValidatePrompt != "" {
			t.Errorf("ValidatePrompt = %q, want empty when noValidation tag", p2.ValidatePrompt)
		}
	})

	t.Run("CLI tag case-insensitive", func(t *testing.T) {
		n := &Node{Text: "Use Gemini", Tags: []string{"gemini"}}
		p := NodeToProcessedNode(n)
		if p.CLI != "GEMINI" {
			t.Errorf("CLI = %q, want GEMINI from lowercase tag", p.CLI)
		}
	})
}

func TestNodeVariableRegistry_completeness(t *testing.T) {
	// One metadata variable per default setting in readme/settings.md, plus validate_prompt and codename alias.
	wantKeys := []string{"cli", "codename", "validate_prompt", "validate_cli", "retries", "retry_cli", "timeout"}
	for _, k := range wantKeys {
		if _, ok := NodeVariableRegistry[k]; !ok {
			t.Errorf("NodeVariableRegistry missing key %q", k)
		}
	}
	if len(NodeVariableRegistry) != len(wantKeys) {
		t.Errorf("NodeVariableRegistry has %d entries, want %d (one per default setting + validate_prompt + codename)", len(NodeVariableRegistry), len(wantKeys))
	}
}

// defaultSettingToVariable maps each "default" setting key (readme/settings.md) to its metadata variable name.
// Every default setting must have a corresponding metadata variable so nodes can override it.
var defaultSettingToVariable = map[string]string{
	"DEFAULT_CLI":          "cli",           // codename is alias
	"DEFAULT_TIMEOUT":      "timeout",
	"DEFAULT_RETRY_CLI":    "retry_cli",
	"DEFAULT_RETRY_COUNT":  "retries",
	"DEFAULT_VALIDATE_CLI": "validate_cli",
}

func TestEveryDefaultSettingHasMetadataVariable(t *testing.T) {
	for setting, variable := range defaultSettingToVariable {
		if _, ok := NodeVariableRegistry[variable]; !ok {
			t.Errorf("setting %q maps to variable %q but NodeVariableRegistry has no key %q", setting, variable, variable)
		}
	}
	// All five default settings must be covered
	wantSettings := []string{"DEFAULT_CLI", "DEFAULT_TIMEOUT", "DEFAULT_RETRY_CLI", "DEFAULT_RETRY_COUNT", "DEFAULT_VALIDATE_CLI"}
	for _, s := range wantSettings {
		if _, ok := defaultSettingToVariable[s]; !ok {
			t.Errorf("default setting %q has no metadata variable in defaultSettingToVariable", s)
		}
	}
}

func TestKnownCLICodenames(t *testing.T) {
	codenames := KnownCLICodenames()
	for _, cli := range AllCLIs {
		if cli.Codename == "" {
			continue
		}
		if _, ok := codenames[cli.Codename]; !ok {
			t.Errorf("KnownCLICodenames missing %q", cli.Codename)
		}
	}
	// Tag that matches codename sets CLI
	n := &Node{Text: "X", Tags: []string{"CURSOR"}}
	p := NodeToProcessedNode(n)
	if p.CLI != "CURSOR" {
		t.Errorf("CLI from CURSOR tag = %q, want CURSOR", p.CLI)
	}
}

func TestResolveNodeValues_NoValidation_overrides_metadata_validate_prompt(t *testing.T) {
	// When NoValidation tag is set, ValidatePrompt must stay empty even if metadata has validate_prompt.
	n := &Node{
		Text:     "Skip",
		Tags:     []string{TagNoValidation},
		Metadata: map[string]string{"validate_prompt": "Custom"},
	}
	res := resolveNodeValues(n, "default", KnownCLICodenames(), nil)
	if res.ValidatePrompt != "" {
		t.Errorf("with NoValidation tag, ValidatePrompt = %q, want empty", res.ValidatePrompt)
	}
}

func TestDeepEqualProcessedNode(t *testing.T) {
	// Sanity: Document from CSV round-trip then convert root to processed; structure preserved.
	// This test lives in document package; here we only ensure ProcessedNode is comparable for tests.
	a := &ProcessedNode{Prompt: "A", CLI: "CURSOR"}
	b := &ProcessedNode{Prompt: "A", CLI: "CURSOR"}
	if !reflect.DeepEqual(a, b) {
		t.Error("DeepEqual expected for same ProcessedNode values")
	}
	b.CLI = "GEMINI"
	if reflect.DeepEqual(a, b) {
		t.Error("DeepEqual should differ when CLI differs")
	}
}
