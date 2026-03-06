package runlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ryanmontgomery/MonadsCLI/internal/run"
	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
	"github.com/ryanmontgomery/MonadsCLI/types"
)

// TestExecuteTree_singleChildFollowedRegardlessOfAnswer verifies that when a node has
// exactly one child, the runner follows it regardless of the parsed answer (e.g. Process
// with unlabeled edge: agent returns "animal3.txt" but the only child is keyed "").
func TestExecuteTree_singleChildFollowedRegardlessOfAnswer(t *testing.T) {
	callCount := 0
	run.SetShellRunner(func(spec runner.CommandSpec) (runner.Result, error) {
		callCount++
		var stdout string
		switch callCount {
		case 1:
			// Root (Start): return answer that does NOT match the single child key ""
			stdout = `{"choices":[""],"answer":"wrong_key","reasons":["Proceeding."]}`
		case 2:
			// Child (Next): leaf response so execution stops
			stdout = `{"completed": true, "secs_taken": 0, "tokens_used": 0, "comments": []}`
		default:
			t.Fatalf("unexpected shell call #%d", callCount)
		}
		return runner.Result{Stdout: stdout, Success: true}, nil
	})
	defer run.SetShellRunner(nil)

	// Root has one child keyed "" (unlabeled edge). Without single-child fallback,
	// we would look up Children["wrong_key"] = nil and stop after the first node.
	child := &types.ProcessedNode{Name: "Next", Prompt: "Do next"}
	root := &types.ProcessedNode{
		Name:     "Start",
		Prompt:   "Start",
		Children: map[string]*types.ProcessedNode{"": child},
	}

	workDir := t.TempDir()
	logDir := "_monad_logs"
	opts := run.RunOptions{
		DefaultCLI:         "CURSOR",
		DefaultValidateCLI: "CURSOR",
		DefaultRetryCLI:    "CURSOR",
	}

	err := ExecuteTree(root, opts, workDir, logDir, "TestChart", true, false)
	if err != nil {
		t.Fatalf("ExecuteTree: %v", err)
	}

	if callCount != 2 {
		t.Errorf("shell calls: got %d, want 2 (root and single child must both run)", callCount)
	}

	absLogDir := filepath.Join(workDir, logDir)
	entries, err := os.ReadDir(absLogDir)
	if err != nil {
		t.Fatalf("ReadDir log: %v", err)
	}
	var shortPath string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			shortPath = filepath.Join(absLogDir, e.Name())
			break
		}
	}
	if shortPath == "" {
		t.Fatal("no short log JSON found")
	}
	data, err := os.ReadFile(shortPath)
	if err != nil {
		t.Fatalf("ReadFile short log: %v", err)
	}
	var body shortLogBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("Unmarshal short log: %v", err)
	}
	if len(body.Nodes) != 2 {
		t.Errorf("short log nodes: got %d, want 2 (Start + Next)", len(body.Nodes))
	}
	if len(body.Nodes) >= 1 && body.Nodes[0].NodeName != "Start" {
		t.Errorf("first node name: got %q, want Start", body.Nodes[0].NodeName)
	}
	if len(body.Nodes) >= 2 && body.Nodes[1].NodeName != "Next" {
		t.Errorf("second node name: got %q, want Next", body.Nodes[1].NodeName)
	}
}

// TestExecuteTree_decisionUsesAnswer verifies that when a node has multiple children,
// the runner uses the parsed answer to select the child (existing behavior).
func TestExecuteTree_decisionUsesAnswer(t *testing.T) {
	callCount := 0
	run.SetShellRunner(func(spec runner.CommandSpec) (runner.Result, error) {
		callCount++
		var stdout string
		switch callCount {
		case 1:
			stdout = `{"choices":["Yes","No"],"answer":"No","reasons":["Not yes."]}`
		case 2:
			stdout = `{"completed": true, "secs_taken": 0, "tokens_used": 0, "comments": []}`
		default:
			t.Fatalf("unexpected shell call #%d", callCount)
		}
		return runner.Result{Stdout: stdout, Success: true}, nil
	})
	defer run.SetShellRunner(nil)

	yesChild := &types.ProcessedNode{Name: "YesBranch", Prompt: "Yes"}
	noChild := &types.ProcessedNode{Name: "NoBranch", Prompt: "No"}
	root := &types.ProcessedNode{
		Name:     "Decision",
		Prompt:   "Yes or No?",
		Children: map[string]*types.ProcessedNode{"Yes": yesChild, "No": noChild},
	}

	workDir := t.TempDir()
	logDir := "_monad_logs"
	opts := run.RunOptions{
		DefaultCLI:         "CURSOR",
		DefaultValidateCLI: "CURSOR",
		DefaultRetryCLI:    "CURSOR",
	}

	err := ExecuteTree(root, opts, workDir, logDir, "TestChart", true, false)
	if err != nil {
		t.Fatalf("ExecuteTree: %v", err)
	}

	if callCount != 2 {
		t.Errorf("shell calls: got %d, want 2", callCount)
	}

	absLogDir := filepath.Join(workDir, logDir)
	entries, _ := os.ReadDir(absLogDir)
	var shortPath string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			shortPath = filepath.Join(absLogDir, e.Name())
			break
		}
	}
	if shortPath == "" {
		t.Fatal("no short log JSON found")
	}
	data, _ := os.ReadFile(shortPath)
	var body shortLogBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("Unmarshal short log: %v", err)
	}
	if len(body.Nodes) != 2 {
		t.Errorf("short log nodes: got %d, want 2", len(body.Nodes))
	}
	if len(body.Nodes) >= 2 && body.Nodes[1].NodeName != "NoBranch" {
		t.Errorf("second node (branch taken): got %q, want NoBranch (answer was No)", body.Nodes[1].NodeName)
	}
}

// TestExecuteTree_decisionCaseInsensitive verifies that when the agent's answer differs
// only by case from the child key (e.g. "no" vs "No"), the runner still selects the correct branch.
func TestExecuteTree_decisionCaseInsensitive(t *testing.T) {
	callCount := 0
	run.SetShellRunner(func(spec runner.CommandSpec) (runner.Result, error) {
		callCount++
		var stdout string
		switch callCount {
		case 1:
			// Answer is lowercase "no"; tree has keys "Yes" and "No"
			stdout = `{"choices":["yes","no"],"answer":"no","reasons":["No."]}`
		case 2:
			stdout = `{"completed": true, "secs_taken": 0, "tokens_used": 0, "comments": []}`
		default:
			t.Fatalf("unexpected shell call #%d", callCount)
		}
		return runner.Result{Stdout: stdout, Success: true}, nil
	})
	defer run.SetShellRunner(nil)

	yesChild := &types.ProcessedNode{Name: "YesBranch", Prompt: "Yes"}
	noChild := &types.ProcessedNode{Name: "NoBranch", Prompt: "No"}
	root := &types.ProcessedNode{
		Name:     "Decision",
		Prompt:   "Yes or No?",
		Children: map[string]*types.ProcessedNode{"Yes": yesChild, "No": noChild},
	}

	workDir := t.TempDir()
	logDir := "_monad_logs"
	opts := run.RunOptions{
		DefaultCLI:         "CURSOR",
		DefaultValidateCLI: "CURSOR",
		DefaultRetryCLI:     "CURSOR",
	}

	err := ExecuteTree(root, opts, workDir, logDir, "TestChart", true, false)
	if err != nil {
		t.Fatalf("ExecuteTree: %v", err)
	}

	if callCount != 2 {
		t.Errorf("shell calls: got %d, want 2", callCount)
	}

	absLogDir := filepath.Join(workDir, logDir)
	entries, _ := os.ReadDir(absLogDir)
	var shortPath string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			shortPath = filepath.Join(absLogDir, e.Name())
			break
		}
	}
	if shortPath == "" {
		t.Fatal("no short log JSON found")
	}
	data, _ := os.ReadFile(shortPath)
	var body shortLogBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("Unmarshal short log: %v", err)
	}
	if len(body.Nodes) != 2 {
		t.Errorf("short log nodes: got %d, want 2", len(body.Nodes))
	}
	if len(body.Nodes) >= 2 && body.Nodes[1].NodeName != "NoBranch" {
		t.Errorf("second node (case-insensitive no): got %q, want NoBranch", body.Nodes[1].NodeName)
	}
}
