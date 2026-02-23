package run

import (
	"errors"
	"io"
	"strings"

	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
	"github.com/ryanmontgomery/MonadsCLI/prompts"
	"github.com/ryanmontgomery/MonadsCLI/types"
)

var (
	// ErrNoCLI is returned when the node has no CLI set and no default codename is provided.
	ErrNoCLI = errors.New("no CLI codename for node and no default")
)

const (
	// ResponseKindProcess is used for childless (leaf) nodes.
	ResponseKindProcess = "process"
	// ResponseKindDecision is used for nodes that have children.
	ResponseKindDecision = "decision"
)

// ResponseKind returns "process" for childless nodes and "decision" for nodes with children.
func ResponseKind(node *types.ProcessedNode) string {
	if node == nil {
		return ResponseKindProcess
	}
	if len(node.Children) > 0 {
		return ResponseKindDecision
	}
	return ResponseKindProcess
}

// ResolveCLI returns the CLI to use for the node: node.CLI if set, otherwise defaultCodename.
func ResolveCLI(node *types.ProcessedNode, defaultCodename string) (types.CLI, error) {
	codename := ""
	if node != nil {
		codename = strings.TrimSpace(node.CLI)
	}
	if codename == "" {
		codename = strings.TrimSpace(defaultCodename)
	}
	if codename == "" {
		return types.CLI{}, ErrNoCLI
	}
	list, err := types.SelectCLIs([]string{codename})
	if err != nil {
		return types.CLI{}, err
	}
	return list[0], nil
}

// BuildRunPrompt returns the node's prompt plus the appropriate response-type instruction
// (process or decision) so the CLI output matches ProcessResponse or DecisionResponse.
func BuildRunPrompt(node *types.ProcessedNode) string {
	if node == nil {
		return ""
	}
	base := strings.TrimSpace(node.Prompt)
	instruction := ProcessResponseInstructionForKind(ResponseKind(node))
	if base == "" {
		return instruction
	}
	if instruction == "" {
		return base
	}
	return base + "\n\n" + instruction
}

// ProcessResponseInstructionForKind returns the prompt instruction for the given kind.
func ProcessResponseInstructionForKind(kind string) string {
	switch kind {
	case ResponseKindDecision:
		return prompts.DecisionResponseInstruction()
	case ResponseKindProcess:
		return prompts.ProcessResponseInstruction()
	default:
		return prompts.ProcessResponseInstruction()
	}
}

// BuildCommand substitutes the full prompt into the CLI's Prompt template.
// The template uses "<prompt>" as the placeholder. The prompt is escaped for safe use inside double quotes in a shell command.
func BuildCommand(cli types.CLI, fullPrompt string) string {
	escaped := escapeForShellPrompt(fullPrompt)
	return strings.Replace(cli.Prompt, "<prompt>", escaped, 1)
}

// escapeForShellPrompt escapes fullPrompt for embedding inside double-quoted string in a shell -c command.
func escapeForShellPrompt(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}

// RunOptions configures RunNode, validation, and retries.
type RunOptions struct {
	DefaultCLI         string // Codename when node.CLI is empty (e.g. from settings DEFAULT_CLI).
	DefaultValidateCLI string // Codename when node.ValidateCLI is empty (e.g. from settings DEFAULT_VALIDATE_CLI).
	DefaultRetryCLI    string // Codename when node.RetryCLI is empty (e.g. from settings DEFAULT_RETRY_CLI).
	WorkDir            string // Working directory for the shell command; empty means current dir.
	// LogLongWriter, when set, receives each LLM stdout (run, validate, retry) for long log.
	LogLongWriter io.Writer
}

// shellRunner is set by tests to fake shell execution; when nil, the real runner is used.
var shellRunner func(runner.CommandSpec) (runner.Result, error)

// SetShellRunner sets the shell runner used by RunNode, RunValidation, and RunRetry (for tests). Pass nil to restore default.
func SetShellRunner(f func(runner.CommandSpec) (runner.Result, error)) {
	shellRunner = f
}

func runShell(spec runner.CommandSpec) (runner.Result, error) {
	if shellRunner != nil {
		return shellRunner(spec)
	}
	return runner.RunShellCommand(spec)
}

func appendLongLog(w io.Writer, stdout string) {
	if w == nil || stdout == "" {
		return
	}
	_, _ = io.WriteString(w, stdout)
	if len(stdout) > 0 && stdout[len(stdout)-1] != '\n' {
		_, _ = io.WriteString(w, "\n")
	}
	_, _ = io.WriteString(w, "\n---\n")
}

// RunNode runs the processed node: resolves CLI, builds prompt with response type, invokes CLI, returns result.
// It does not verify the response type or retry.
func RunNode(node *types.ProcessedNode, opts RunOptions) (runner.Result, error) {
	cli, err := ResolveCLI(node, opts.DefaultCLI)
	if err != nil {
		return runner.Result{}, err
	}
	fullPrompt := BuildRunPrompt(node)
	command := BuildCommand(cli, fullPrompt)
	shell, shellArgs := runner.DefaultShell()
	res, err := runShell(runner.CommandSpec{
		Shell:     shell,
		ShellArgs: shellArgs,
		Command:   command,
		WorkDir:   opts.WorkDir,
	})
	if err == nil {
		appendLongLog(opts.LogLongWriter, res.Stdout)
	}
	return res, err
}

// ShouldValidate reports whether the node should be validated after it runs.
// Validation is skipped when the node has the NoValidation tag (ValidatePrompt empty) or has children.
func ShouldValidate(node *types.ProcessedNode) bool {
	if node == nil {
		return false
	}
	if len(node.Children) > 0 {
		return false
	}
	return strings.TrimSpace(node.ValidatePrompt) != ""
}

// ResolveValidateCLI returns the CLI to use for validation: node.ValidateCLI if set, otherwise defaultValidateCodename.
func ResolveValidateCLI(node *types.ProcessedNode, defaultValidateCodename string) (types.CLI, error) {
	codename := ""
	if node != nil {
		codename = strings.TrimSpace(node.ValidateCLI)
	}
	if codename == "" {
		codename = strings.TrimSpace(defaultValidateCodename)
	}
	if codename == "" {
		return types.CLI{}, ErrNoValidateCLI
	}
	list, err := types.SelectCLIs([]string{codename})
	if err != nil {
		return types.CLI{}, err
	}
	return list[0], nil
}

// ErrNoValidateCLI is returned when the node has no ValidateCLI set and no default is provided.
var ErrNoValidateCLI = errors.New("no validate CLI codename for node and no default")

// ResolveRetryCLI returns the CLI to use for retries: node.RetryCLI if set, otherwise defaultRetryCodename.
func ResolveRetryCLI(node *types.ProcessedNode, defaultRetryCodename string) (types.CLI, error) {
	codename := ""
	if node != nil {
		codename = strings.TrimSpace(node.RetryCLI)
	}
	if codename == "" {
		codename = strings.TrimSpace(defaultRetryCodename)
	}
	if codename == "" {
		return types.CLI{}, ErrNoRetryCLI
	}
	list, err := types.SelectCLIs([]string{codename})
	if err != nil {
		return types.CLI{}, err
	}
	return list[0], nil
}

// ErrNoRetryCLI is returned when the node has no RetryCLI set and no default is provided.
var ErrNoRetryCLI = errors.New("no retry CLI codename for node and no default")

const defaultRetryCount = 3

// EffectiveRetryLimit returns the maximum retry count for the node (conventional default 3 when node.Retries <= 0).
func EffectiveRetryLimit(node *types.ProcessedNode) int {
	if node == nil || node.Retries <= 0 {
		return defaultRetryCount
	}
	return node.Retries
}

// BuildRetryPrompt returns a retry prompt: the original prompt, prior validation critiques (one section per attempt),
// and the appropriate response-type instruction (process or decision). Used when validation fails and the node is retried.
func BuildRetryPrompt(node *types.ProcessedNode, priorCritiques []string) string {
	if node == nil {
		return ""
	}
	base := strings.TrimSpace(node.Prompt)
	for _, c := range priorCritiques {
		if strings.TrimSpace(c) == "" {
			continue
		}
		base += "\n\n---\nPrevious validation feedback:\n"
		base += strings.TrimSpace(c)
	}
	instruction := ProcessResponseInstructionForKind(ResponseKind(node))
	if instruction == "" {
		return base
	}
	return base + "\n\n" + instruction
}

// FormatValidationCritique turns a failed ValidationResponse into a single critique string for the retry prompt.
func FormatValidationCritique(v types.ValidationResponse) string {
	var b strings.Builder
	if len(v.Warnings) > 0 {
		b.WriteString(strings.Join(v.Warnings, "\n"))
		b.WriteString("\n")
	}
	b.WriteString("Validation did not pass (fully_completed: false).")
	return b.String()
}

// VerifyRunOutput parses the CLI stdout as ProcessResponse or DecisionResponse based on node kind and returns an error if parsing fails.
func VerifyRunOutput(node *types.ProcessedNode, stdout string) error {
	if node == nil {
		return errors.New("cannot verify run output: nil node")
	}
	kind := ResponseKind(node)
	switch kind {
	case ResponseKindDecision:
		_, err := types.ParseDecisionResponse(stdout)
		return err
	case ResponseKindProcess:
		_, err := types.ParseProcessResponse(stdout)
		return err
	default:
		_, err := types.ParseProcessResponse(stdout)
		return err
	}
}

// BuildValidatePrompt returns the full validation prompt: the node's validation prompt text (default or custom),
// context (original prompt and output to validate), and the validation response type instruction.
func BuildValidatePrompt(node *types.ProcessedNode, nodeOutput string) string {
	if node == nil {
		return ""
	}
	validateText := strings.TrimSpace(node.ValidatePrompt)
	if validateText == "" {
		return ""
	}
	instruction := prompts.ValidationResponseInstruction()
	var b strings.Builder
	b.WriteString(validateText)
	b.WriteString("\n\n---\nOriginal task:\n")
	b.WriteString(strings.TrimSpace(node.Prompt))
	b.WriteString("\n\nOutput to validate:\n")
	b.WriteString(nodeOutput)
	b.WriteString("\n\n---\n")
	b.WriteString(instruction)
	return b.String()
}

// ValidationResult holds the result of running the validation prompt and verifying the response type.
type ValidationResult struct {
	RunnerResult runner.Result
	Response     types.ValidationResponse
	Valid        bool // true when response was parsed and fully_completed is true
}

// RunValidation runs the validation prompt for the node using the validate CLI, then parses and verifies
// the response matches ValidationResponse. Valid is true only when fully_completed is true.
func RunValidation(node *types.ProcessedNode, opts RunOptions, nodeOutput string) (ValidationResult, error) {
	var out ValidationResult
	cli, err := ResolveValidateCLI(node, opts.DefaultValidateCLI)
	if err != nil {
		return out, err
	}
	fullPrompt := BuildValidatePrompt(node, nodeOutput)
	if fullPrompt == "" {
		return out, errors.New("validation prompt is empty")
	}
	command := BuildCommand(cli, fullPrompt)
	shell, shellArgs := runner.DefaultShell()
	res, err := runShell(runner.CommandSpec{
		Shell:     shell,
		ShellArgs: shellArgs,
		Command:   command,
		WorkDir:   opts.WorkDir,
	})
	out.RunnerResult = res
	if err == nil {
		appendLongLog(opts.LogLongWriter, res.Stdout)
	}
	if err != nil {
		return out, err
	}
	parsed, parseErr := types.ParseValidationResponse(res.Stdout)
	out.Response = parsed
	if parseErr != nil {
		return out, parseErr
	}
	out.Valid = parsed.FullyCompleted
	return out, nil
}

// NodeResult holds the result of running a node and optionally validating it.
type NodeResult struct {
	RunResult      runner.Result
	Validation      *ValidationResult // nil if validation was skipped
	ValidationRan   bool              // true when validation was run (even if parse failed)
	Valid          bool              // true when no validation or validation passed (fully_completed)
	ValidationError error             // set when validation was run but failed (parse or not fully_completed)
}

// RunNodeThenValidate runs the node, then automatically runs validation when ShouldValidate(node) is true.
// On success (node run succeeds and, if validation ran, validation passed), Valid is true and the caller can run the next node.
// On validation failure, retries run with a custom retry prompt (original + prior validation critiques + response type) until validation passes or EffectiveRetryLimit is reached.
// Returns a non-nil error only for run or validation CLI/shell/parse failures; when validation ran and fully_completed is false and retries exhausted, error is nil and Valid is false.
func RunNodeThenValidate(node *types.ProcessedNode, opts RunOptions) (NodeResult, error) {
	var out NodeResult
	runRes, err := RunNode(node, opts)
	out.RunResult = runRes
	if err != nil {
		return out, err
	}
	if !ShouldValidate(node) {
		out.Valid = true
		return out, nil
	}
	out.ValidationRan = true
	valRes, err := RunValidation(node, opts, runRes.Stdout)
	out.Validation = &valRes
	if err != nil {
		out.ValidationError = err
		out.Valid = false
		return out, err
	}
	out.Valid = valRes.Valid
	if !out.Valid {
		out.ValidationError = errors.New("validation did not pass: fully_completed is false")
		runOut, retryErr := runRetryLoop(node, opts, &out, []string{FormatValidationCritique(valRes.Response)})
		if retryErr != nil {
			return runOut, retryErr
		}
		return runOut, nil
	}
	return out, nil
}

// RunRetry runs the node with a custom retry prompt using the retry CLI. Caller must verify response type and run validation.
func RunRetry(node *types.ProcessedNode, opts RunOptions, retryPrompt string) (runner.Result, error) {
	cli, err := ResolveRetryCLI(node, opts.DefaultRetryCLI)
	if err != nil {
		return runner.Result{}, err
	}
	command := BuildCommand(cli, retryPrompt)
	shell, shellArgs := runner.DefaultShell()
	res, err := runShell(runner.CommandSpec{
		Shell:     shell,
		ShellArgs: shellArgs,
		Command:   command,
		WorkDir:   opts.WorkDir,
	})
	if err == nil {
		appendLongLog(opts.LogLongWriter, res.Stdout)
	}
	return res, err
}

// runRetryLoop runs retries until validation passes or EffectiveRetryLimit is reached. Mutates node.Retried and out.
func runRetryLoop(node *types.ProcessedNode, opts RunOptions, out *NodeResult, critiques []string) (NodeResult, error) {
	limit := EffectiveRetryLimit(node)
	for node.Retried < limit {
		node.Retried++
		retryPrompt := BuildRetryPrompt(node, critiques)
		if retryPrompt == "" {
			return *out, errors.New("retry prompt is empty")
		}
		runRes, err := RunRetry(node, opts, retryPrompt)
		if err != nil {
			out.ValidationError = err
			return *out, err
		}
		if err := VerifyRunOutput(node, runRes.Stdout); err != nil {
			out.ValidationError = err
			return *out, err
		}
		valRes, err := RunValidation(node, opts, runRes.Stdout)
		if err != nil {
			out.ValidationError = err
			return *out, err
		}
		out.RunResult = runRes
		out.Validation = &valRes
		if valRes.Valid {
			out.Valid = true
			out.ValidationError = nil
			return *out, nil
		}
		critiques = append(critiques, FormatValidationCritique(valRes.Response))
	}
	out.ValidationError = errors.New("validation did not pass: max retries reached")
	return *out, nil
}
