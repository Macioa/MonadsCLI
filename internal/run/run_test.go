package run

import (
	"strings"
	"testing"

	"github.com/ryanmontgomery/MonadsCLI/types"
)

func TestResponseKind(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := ResponseKind(nil); got != ResponseKindProcess {
			t.Errorf("ResponseKind(nil) = %q, want process", got)
		}
	})
	t.Run("childless", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P"}
		if got := ResponseKind(n); got != ResponseKindProcess {
			t.Errorf("ResponseKind(childless) = %q, want process", got)
		}
	})
	t.Run("has_children", func(t *testing.T) {
		n := &types.ProcessedNode{Children: map[string]*types.ProcessedNode{"Yes": {}}}
		if got := ResponseKind(n); got != ResponseKindDecision {
			t.Errorf("ResponseKind(has children) = %q, want decision", got)
		}
	})
}

func TestResolveCLI(t *testing.T) {
	t.Run("node_codename", func(t *testing.T) {
		n := &types.ProcessedNode{CLI: "GEMINI"}
		cli, err := ResolveCLI(n, "CURSOR")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "GEMINI" {
			t.Errorf("ResolveCLI = codename %q, want GEMINI", cli.Codename)
		}
	})
	t.Run("default_codename", func(t *testing.T) {
		n := &types.ProcessedNode{}
		cli, err := ResolveCLI(n, "CURSOR")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "CURSOR" {
			t.Errorf("ResolveCLI = codename %q, want CURSOR", cli.Codename)
		}
	})
	t.Run("no_cli", func(t *testing.T) {
		n := &types.ProcessedNode{}
		_, err := ResolveCLI(n, "")
		if err != ErrNoCLI {
			t.Errorf("ResolveCLI(no default) err = %v, want ErrNoCLI", err)
		}
	})
	t.Run("nil_node_uses_default", func(t *testing.T) {
		cli, err := ResolveCLI(nil, "CLAUDE")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "CLAUDE" {
			t.Errorf("ResolveCLI(nil, CLAUDE) = codename %q", cli.Codename)
		}
	})
	t.Run("unknown_codename", func(t *testing.T) {
		n := &types.ProcessedNode{CLI: "UNKNOWN"}
		_, err := ResolveCLI(n, "CURSOR")
		if err == nil {
			t.Error("ResolveCLI(unknown) want error")
		}
	})
}

func TestBuildRunPrompt(t *testing.T) {
	t.Run("childless", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "Do the task"}
		got := BuildRunPrompt(n)
		if !strings.Contains(got, "Do the task") {
			t.Errorf("BuildRunPrompt missing base prompt: %q", got)
		}
		if !strings.Contains(got, "completed") || !strings.Contains(got, "secs_taken") {
			t.Errorf("BuildRunPrompt should include process response instruction: %q", got)
		}
	})
	t.Run("decision", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "Choose one", Children: map[string]*types.ProcessedNode{"A": {}}}
		got := BuildRunPrompt(n)
		if !strings.Contains(got, "Choose one") {
			t.Errorf("BuildRunPrompt missing base prompt: %q", got)
		}
		if !strings.Contains(got, "choices") || !strings.Contains(got, "answer") {
			t.Errorf("BuildRunPrompt should include decision response instruction: %q", got)
		}
	})
	t.Run("nil", func(t *testing.T) {
		got := BuildRunPrompt(nil)
		if got != "" {
			t.Errorf("BuildRunPrompt(nil) = %q, want empty", got)
		}
	})
}

func TestBuildCommand(t *testing.T) {
	cli := types.CLI{Prompt: "agent \"<prompt>\""}
	got := BuildCommand(cli, "hello world")
	want := `agent "hello world"`
	if got != want {
		t.Errorf("BuildCommand = %q, want %q", got, want)
	}
	// Escaping
	cli2 := types.CLI{Prompt: "cmd \"<prompt>\""}
	got2 := BuildCommand(cli2, `say "hi"`)
	want2 := `cmd "say \"hi\""`
	if got2 != want2 {
		t.Errorf("BuildCommand(quoted) = %q, want %q", got2, want2)
	}
	// Gemini CLI template includes --yolo and prompt
	list, err := types.SelectCLIs([]string{"GEMINI"})
	if err != nil {
		t.Fatal(err)
	}
	geminiCmd := BuildCommand(list[0], "test prompt")
	if !strings.Contains(geminiCmd, "--yolo") {
		t.Errorf("Gemini BuildCommand should contain --yolo: %q", geminiCmd)
	}
	if !strings.Contains(geminiCmd, "test prompt") {
		t.Errorf("Gemini BuildCommand should contain prompt: %q", geminiCmd)
	}
}

func TestProcessResponseInstructionForKind(t *testing.T) {
	p := ProcessResponseInstructionForKind(ResponseKindProcess)
	if !strings.Contains(p, "completed") {
		t.Errorf("process instruction: %q", p)
	}
	d := ProcessResponseInstructionForKind(ResponseKindDecision)
	if !strings.Contains(d, "choices") {
		t.Errorf("decision instruction: %q", d)
	}
	// unknown kind defaults to process
	u := ProcessResponseInstructionForKind("unknown")
	if !strings.Contains(u, "completed") {
		t.Errorf("unknown kind should default to process: %q", u)
	}
}

func TestRunNode_ErrorPaths(t *testing.T) {
	t.Run("no_cli_returns_ErrNoCLI", func(t *testing.T) {
		node := &types.ProcessedNode{Prompt: "x"}
		_, err := RunNode(node, RunOptions{DefaultCLI: ""})
		if err != ErrNoCLI {
			t.Errorf("RunNode(no default CLI) err = %v, want ErrNoCLI", err)
		}
	})
	t.Run("unknown_codename_returns_error", func(t *testing.T) {
		node := &types.ProcessedNode{Prompt: "x", CLI: "UNKNOWN_CODENAME"}
		_, err := RunNode(node, RunOptions{DefaultCLI: "CURSOR"})
		if err == nil {
			t.Error("RunNode(unknown codename) want error")
		}
	})
}

func TestShouldValidate(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if ShouldValidate(nil) {
			t.Error("ShouldValidate(nil) want false")
		}
	})
	t.Run("no_validation_prompt_skipped", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P", ValidatePrompt: ""}
		if ShouldValidate(n) {
			t.Error("ShouldValidate(empty ValidatePrompt) want false")
		}
	})
	t.Run("has_children_skipped", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P", ValidatePrompt: "Check it", Children: map[string]*types.ProcessedNode{"A": {}}}
		if ShouldValidate(n) {
			t.Error("ShouldValidate(node with children) want false")
		}
	})
	t.Run("childless_with_validate_prompt", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P", ValidatePrompt: "Did it follow instructions?"}
		if !ShouldValidate(n) {
			t.Error("ShouldValidate(childless with ValidatePrompt) want true")
		}
	})
}

func TestResolveValidateCLI(t *testing.T) {
	t.Run("node_validate_cli", func(t *testing.T) {
		n := &types.ProcessedNode{ValidateCLI: "GEMINI"}
		cli, err := ResolveValidateCLI(n, "CURSOR")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "GEMINI" {
			t.Errorf("ResolveValidateCLI = codename %q, want GEMINI", cli.Codename)
		}
	})
	t.Run("default_validate_cli", func(t *testing.T) {
		n := &types.ProcessedNode{}
		cli, err := ResolveValidateCLI(n, "CURSOR")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "CURSOR" {
			t.Errorf("ResolveValidateCLI = codename %q, want CURSOR", cli.Codename)
		}
	})
	t.Run("no_validate_cli", func(t *testing.T) {
		n := &types.ProcessedNode{}
		_, err := ResolveValidateCLI(n, "")
		if err != ErrNoValidateCLI {
			t.Errorf("ResolveValidateCLI(no default) err = %v, want ErrNoValidateCLI", err)
		}
	})
	t.Run("unknown_codename", func(t *testing.T) {
		n := &types.ProcessedNode{ValidateCLI: "UNKNOWN"}
		_, err := ResolveValidateCLI(n, "CURSOR")
		if err == nil {
			t.Error("ResolveValidateCLI(unknown) want error")
		}
	})
}

func TestBuildValidatePrompt(t *testing.T) {
	t.Run("includes_validate_prompt_and_context", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "Do X", ValidatePrompt: "Check that X was done."}
		got := BuildValidatePrompt(n, "output from node")
		if !strings.Contains(got, "Check that X was done.") {
			t.Errorf("BuildValidatePrompt missing validate text: %q", got)
		}
		if !strings.Contains(got, "Do X") {
			t.Errorf("BuildValidatePrompt missing original prompt: %q", got)
		}
		if !strings.Contains(got, "output from node") {
			t.Errorf("BuildValidatePrompt missing node output: %q", got)
		}
		if !strings.Contains(got, "fully_completed") || !strings.Contains(got, "should_retry") {
			t.Errorf("BuildValidatePrompt should include validation response instruction: %q", got)
		}
	})
	t.Run("nil", func(t *testing.T) {
		got := BuildValidatePrompt(nil, "out")
		if got != "" {
			t.Errorf("BuildValidatePrompt(nil) = %q, want empty", got)
		}
	})
	t.Run("empty_validate_prompt", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P", ValidatePrompt: ""}
		got := BuildValidatePrompt(n, "out")
		if got != "" {
			t.Errorf("BuildValidatePrompt(empty ValidatePrompt) = %q, want empty", got)
		}
	})
}
