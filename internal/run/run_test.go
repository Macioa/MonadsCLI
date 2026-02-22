package run

import (
	"strings"
	"testing"

	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
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

func TestResolveRetryCLI(t *testing.T) {
	t.Run("node_retry_cli", func(t *testing.T) {
		n := &types.ProcessedNode{RetryCLI: "GEMINI"}
		cli, err := ResolveRetryCLI(n, "CURSOR")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "GEMINI" {
			t.Errorf("ResolveRetryCLI = codename %q, want GEMINI", cli.Codename)
		}
	})
	t.Run("default_retry_cli", func(t *testing.T) {
		n := &types.ProcessedNode{}
		cli, err := ResolveRetryCLI(n, "CURSOR")
		if err != nil {
			t.Fatal(err)
		}
		if cli.Codename != "CURSOR" {
			t.Errorf("ResolveRetryCLI = codename %q, want CURSOR", cli.Codename)
		}
	})
	t.Run("no_retry_cli", func(t *testing.T) {
		n := &types.ProcessedNode{}
		_, err := ResolveRetryCLI(n, "")
		if err != ErrNoRetryCLI {
			t.Errorf("ResolveRetryCLI(no default) err = %v, want ErrNoRetryCLI", err)
		}
	})
	t.Run("unknown_codename", func(t *testing.T) {
		n := &types.ProcessedNode{RetryCLI: "UNKNOWN"}
		_, err := ResolveRetryCLI(n, "CURSOR")
		if err == nil {
			t.Error("ResolveRetryCLI(unknown) want error")
		}
	})
}

func TestEffectiveRetryLimit(t *testing.T) {
	t.Run("nil_uses_default_3", func(t *testing.T) {
		if got := EffectiveRetryLimit(nil); got != 3 {
			t.Errorf("EffectiveRetryLimit(nil) = %d, want 3", got)
		}
	})
	t.Run("zero_uses_3", func(t *testing.T) {
		n := &types.ProcessedNode{Retries: 0}
		if got := EffectiveRetryLimit(n); got != 3 {
			t.Errorf("EffectiveRetryLimit(Retries=0) = %d, want 3", got)
		}
	})
	t.Run("positive_uses_node", func(t *testing.T) {
		n := &types.ProcessedNode{Retries: 5}
		if got := EffectiveRetryLimit(n); got != 5 {
			t.Errorf("EffectiveRetryLimit(Retries=5) = %d, want 5", got)
		}
	})
}

func TestBuildRetryPrompt(t *testing.T) {
	t.Run("original_plus_critiques_plus_instruction", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "Do X"}
		critiques := []string{"Warning: incomplete.", "Validation did not pass."}
		got := BuildRetryPrompt(n, critiques)
		if got == "" {
			t.Fatal("BuildRetryPrompt want non-empty")
		}
		if !strings.Contains(got, "Do X") {
			t.Errorf("BuildRetryPrompt missing original prompt: %q", got)
		}
		if !strings.Contains(got, "Previous validation feedback") {
			t.Errorf("BuildRetryPrompt missing critique section: %q", got)
		}
		if !strings.Contains(got, "Warning: incomplete") {
			t.Errorf("BuildRetryPrompt missing critique text: %q", got)
		}
		if !strings.Contains(got, "completed") || !strings.Contains(got, "secs_taken") {
			t.Errorf("BuildRetryPrompt should include process response instruction: %q", got)
		}
	})
	t.Run("decision_includes_decision_instruction", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "Choose", Children: map[string]*types.ProcessedNode{"A": {}}}
		got := BuildRetryPrompt(n, []string{"Fix it"})
		if !strings.Contains(got, "choices") || !strings.Contains(got, "answer") {
			t.Errorf("BuildRetryPrompt(decision) should include decision instruction: %q", got)
		}
	})
	t.Run("nil_node_empty", func(t *testing.T) {
		got := BuildRetryPrompt(nil, []string{"x"})
		if got != "" {
			t.Errorf("BuildRetryPrompt(nil) = %q, want empty", got)
		}
	})
	t.Run("no_critiques_just_prompt_and_instruction", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "Task"}
		got := BuildRetryPrompt(n, nil)
		if !strings.Contains(got, "Task") {
			t.Errorf("BuildRetryPrompt(no critiques) missing prompt: %q", got)
		}
	})
}

func TestFormatValidationCritique(t *testing.T) {
	t.Run("with_warnings", func(t *testing.T) {
		v := types.ValidationResponse{Warnings: []string{"w1", "w2"}, FullyCompleted: false}
		got := FormatValidationCritique(v)
		if !strings.Contains(got, "w1") || !strings.Contains(got, "w2") {
			t.Errorf("FormatValidationCritique = %q", got)
		}
		if !strings.Contains(got, "Validation did not pass") {
			t.Errorf("FormatValidationCritique = %q", got)
		}
	})
	t.Run("no_warnings", func(t *testing.T) {
		v := types.ValidationResponse{FullyCompleted: false}
		got := FormatValidationCritique(v)
		if !strings.Contains(got, "Validation did not pass") {
			t.Errorf("FormatValidationCritique = %q", got)
		}
	})
}

func TestVerifyRunOutput(t *testing.T) {
	t.Run("process_valid", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P"}
		stdout := `{"completed": true, "secs_taken": 0, "tokens_used": 0, "comments": []}`
		if err := VerifyRunOutput(n, stdout); err != nil {
			t.Errorf("VerifyRunOutput(process) = %v", err)
		}
	})
	t.Run("process_invalid", func(t *testing.T) {
		n := &types.ProcessedNode{Prompt: "P"}
		if err := VerifyRunOutput(n, "not json"); err == nil {
			t.Error("VerifyRunOutput(process, invalid) want error")
		}
	})
	t.Run("decision_valid", func(t *testing.T) {
		n := &types.ProcessedNode{Children: map[string]*types.ProcessedNode{"A": {}}}
		stdout := `{"choices": ["A","B"], "answer": "A", "reasons": []}`
		if err := VerifyRunOutput(n, stdout); err != nil {
			t.Errorf("VerifyRunOutput(decision) = %v", err)
		}
	})
	t.Run("nil_node", func(t *testing.T) {
		if err := VerifyRunOutput(nil, "{}"); err == nil {
			t.Error("VerifyRunOutput(nil) want error")
		}
	})
}

// TestRunNodeThenValidate_RetryOnExplicitFailPrompt uses an explicit validation-fail response
// (fully_completed: false + warnings) to verify that validate and retry work: first validation
// fails, retry runs with the fail feedback in the prompt, then validation passes.
// It explicitly records and asserts each of the 4 shell invocations in order.
func TestRunNodeThenValidate_RetryOnExplicitFailPrompt(t *testing.T) {
	processOut := `{"completed": true, "secs_taken": 0, "tokens_used": 0, "comments": []}`
	validateFail := `{"fully_completed": false, "partially_completed": true, "should_retry": true, "warnings": ["Output did not meet the criteria. Please fix and try again."]}`
	validatePass := `{"fully_completed": true, "partially_completed": false, "should_retry": false, "warnings": []}`

	type callKind int
	const (
		callNodeRun callKind = iota
		callValidateFail
		callRetryRun
		callValidatePass
	)
	var sequence []callKind
	SetShellRunner(func(spec runner.CommandSpec) (runner.Result, error) {
		cmd := spec.Command
		// Distinguish: node run has "Complete the task" + process response instruction (completed, secs_taken).
		// Validation has "Check that the task" + "fully_completed" + "Output to validate".
		// Retry has "Complete the task" + "Previous validation feedback" + "Output did not meet the criteria".
		isNodeRun := strings.Contains(cmd, "Complete the task") && strings.Contains(cmd, "completed") && strings.Contains(cmd, "secs_taken") && !strings.Contains(cmd, "Previous validation feedback") && !strings.Contains(cmd, "Output to validate")
		isValidate := strings.Contains(cmd, "fully_completed") && strings.Contains(cmd, "Output to validate")
		isRetry := strings.Contains(cmd, "Complete the task") && strings.Contains(cmd, "Previous validation feedback") && strings.Contains(cmd, "Output did not meet the criteria")

		var kind callKind
		var stdout string
		switch {
		case isNodeRun && !isRetry:
			kind = callNodeRun
			stdout = processOut
		case isValidate && len(sequence) == 1:
			kind = callValidateFail
			stdout = validateFail
		case isRetry:
			kind = callRetryRun
			stdout = processOut
		case isValidate && len(sequence) == 3:
			kind = callValidatePass
			stdout = validatePass
		default:
			t.Fatalf("unexpected shell call #%d: command snippet %q", len(sequence)+1, cmd)
		}
		sequence = append(sequence, kind)
		return runner.Result{Stdout: stdout, Success: true}, nil
	})
	defer SetShellRunner(nil)

	node := &types.ProcessedNode{
		Prompt:         "Complete the task.",
		ValidatePrompt: "Check that the task was completed correctly.",
		Retries:        2,
	}
	opts := RunOptions{
		DefaultCLI:         "CURSOR",
		DefaultValidateCLI: "CURSOR",
		DefaultRetryCLI:    "CURSOR",
	}

	result, err := RunNodeThenValidate(node, opts)
	if err != nil {
		t.Fatalf("RunNodeThenValidate: %v", err)
	}

	// Explicit verification: exactly 4 calls in order.
	wantSequence := []callKind{callNodeRun, callValidateFail, callRetryRun, callValidatePass}
	if len(sequence) != len(wantSequence) {
		t.Fatalf("shell call count: got %d, want %d (sequence: %v)", len(sequence), len(wantSequence), sequence)
	}
	for i := range wantSequence {
		if sequence[i] != wantSequence[i] {
			t.Fatalf("call #%d: got %v, want %v (full sequence: %v)", i+1, sequence[i], wantSequence[i], sequence)
		}
	}
	t.Logf("verified sequence: 1=RunNode, 2=Validate(fail), 3=RunRetry, 4=Validate(pass)")

	if !result.Valid {
		t.Fatalf("Valid = false; retry must have passed validation")
	}
	if result.ValidationError != nil {
		t.Fatalf("ValidationError = %v; want nil after successful retry", result.ValidationError)
	}
	if node.Retried != 1 {
		t.Fatalf("node.Retried = %d; want 1 (exactly one retry attempt)", node.Retried)
	}
	if result.Validation == nil {
		t.Fatalf("Validation = nil; validation ran and should be set")
	}
	if !result.Validation.Valid {
		t.Fatalf("Validation.Valid = false; final validation must have passed")
	}
	// Retry prompt must include the explicit fail feedback.
	retryPrompt := BuildRetryPrompt(node, []string{FormatValidationCritique(types.ValidationResponse{
		FullyCompleted: false, Warnings: []string{"Output did not meet the criteria. Please fix and try again."},
	})})
	if !strings.Contains(retryPrompt, "Output did not meet the criteria") {
		t.Fatalf("retry prompt must include fail feedback; got: %q", retryPrompt)
	}
	t.Logf("verified: retry prompt contained validation fail feedback and response-type instruction")
}
