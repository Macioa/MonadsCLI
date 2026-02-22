package run

import (
	"errors"
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

// RunOptions configures RunNode.
type RunOptions struct {
	DefaultCLI string // Codename when node.CLI is empty (e.g. from settings DEFAULT_CLI).
	WorkDir   string // Working directory for the shell command; empty means current dir.
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
	return runner.RunShellCommand(runner.CommandSpec{
		Shell:     shell,
		ShellArgs: shellArgs,
		Command:   command,
		WorkDir:   opts.WorkDir,
	})
}
