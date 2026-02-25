# Decision tree process

This document describes the technical pipeline for running a decision tree: import, run, validate, and retry. It is intended for maintainers and integrators.

---

## Overview

A decision tree is a rooted tree of nodes. Each node is run by invoking a configured CLI with a prompt; the CLI’s stdout is parsed as a structured response. Childless nodes use a process response type; nodes with children use a decision response type so the runner can select the next branch. Optionally, a node’s output is validated by a second CLI call; if validation fails, the node is retried with a combined prompt until validation passes or a retry limit is reached.

---

## Pipeline stages

### 1. Import

Import converts an external representation into an internal node tree, then into a processed node tree used by the runner.

**Sources**

- **CSV:** Lucid CSV export. Parsed by `internal/document.TransformFromCSV` into a `types.Document` whose `Root` is a `*types.Node` tree.
- **Lucid JSON:** Lucid API document contents. Parsed by `internal/document.TransformFromLucidJSON` into the same `Document` / `Node` shape.

**Internal node type (`types.Node`)**

- Shape metadata: `ID`, `Label`, `Text`, `TextAreas`, `Tags`, `Metadata`, `Children` (route → child).
- Tags and metadata are preserved for the next stage.

**Processed node type (`types.ProcessedNode`)**

- Conversion: `types.NodeToProcessedNodeWithDefaults(node, defaults)`.
- Defaults (e.g. from settings) supply `DEFAULT_CLI`, `DEFAULT_VALIDATE_CLI`, `DEFAULT_RETRY_CLI`, `DEFAULT_RETRY_COUNT`, `DEFAULT_TIMEOUT` when the node does not override them via tag or metadata.
- Result: each node has `Prompt`, `ValidatePrompt`, `CLI`, `ValidateCLI`, `RetryCLI`, `Retries`, and `Children` (route → `*ProcessedNode`). Tags like `NoValidation` and metadata (e.g. `validate_prompt`, `validate_cli`) are applied during this conversion; see `readme/metadata.md`.

The runner operates only on the processed tree; it does not use raw `Node` or CSV/JSON.

---

### 2. Run

Running a node executes the node’s CLI with a constructed prompt and (in the retry path) verifies that the CLI output matches the expected response type.

**CLI selection**

- If the node has a CLI codename set (e.g. via tag or metadata), that codename is used.
- Otherwise the default CLI from options (e.g. settings `DEFAULT_CLI`) is used.
- Implementation: `internal/run.ResolveCLI(node, opts.DefaultCLI)`.

**Prompt construction**

- Base: the node’s `Prompt`.
- The runner appends a response-type instruction so the CLI returns parseable JSON:
  - **Childless node:** process response type (`completed`, `secs_taken`, `tokens_used`, `comments`). Instruction from `prompts.ProcessResponseInstruction()`.
  - **Node with children:** decision response type (`choices`, `answer`, `reasons`). Instruction from `prompts.DecisionResponseInstruction()`.
- Implementation: `internal/run.BuildRunPrompt(node)` uses `ResponseKind(node)` to choose the instruction.

**Execution**

- The chosen CLI’s command template (e.g. `cursor ask "<prompt>"`) is filled with the full prompt; the command is run in the configured work directory.
- Implementation: `internal/run.RunNode(node, opts)` → shell execution via `internal/runner`.

**Response type verification**

- The runner expects stdout to parse as either `ProcessResponse` or `DecisionResponse` depending on node kind. Verification is performed in the retry loop after each retry run; see §4. Implementation: `internal/run.VerifyRunOutput(node, stdout)` → `types.ParseProcessResponse` or `types.ParseDecisionResponse`.

---

### 3. Validate

After a successful run, the node may be validated. Validation is a second CLI call that judges whether the run output is acceptable; its stdout must parse as `ValidationResponse`.

**When validation runs**

- Validation runs only if:
  - The node has no children (decision nodes are not validated), and
  - The node does not have the `NoValidation` tag (in that case `ValidatePrompt` is empty and validation is skipped).
- Implementation: `internal/run.ShouldValidate(node)` is true when `node.ValidatePrompt` is non-empty and `node.Children` is empty.

**Validation prompt**

- Default: embedded prompt from `prompts.DefaultValidatePrompt()` (see `prompts/validate.txt`).
- Override: node metadata `validate_prompt` sets a custom validation prompt.
- The runner builds a full validation prompt that includes: the validation prompt text, the original node prompt, the run output to validate, and the validation response-type instruction (`prompts.ValidationResponseInstruction()`: `fully_completed`, `partially_completed`, `should_retry`, `warnings`).
- Implementation: `internal/run.BuildValidatePrompt(node, nodeOutput)`.

**Validation CLI**

- Node’s `ValidateCLI` if set; otherwise options’ default validate CLI (e.g. `DEFAULT_VALIDATE_CLI`).
- Implementation: `internal/run.ResolveValidateCLI(node, opts.DefaultValidateCLI)`.

**Execution and parsing**

- The validation CLI is invoked with the validation prompt. Stdout is parsed as `ValidationResponse` via `types.ParseValidationResponse`. Success is defined as `fully_completed == true`.
- Implementation: `internal/run.RunValidation(node, opts, nodeOutput)`.

**Outcome**

- If validation is not run: the runner proceeds to the next step (e.g. run next node or finish).
- If validation runs and `fully_completed` is true: proceed.
- If validation runs and `fully_completed` is false: trigger retry; see §4.

---

### 4. Retry

When validation fails (or validation was run and did not pass), the node is retried: the same node is run again with a retry prompt, then validated again, until validation passes or a retry limit is reached.

**Retry prompt**

- Original node prompt, plus one section per prior attempt: “Previous validation feedback” and the critique from that attempt (e.g. warnings and “Validation did not pass (fully_completed: false)”).
- The same response-type instruction as for the initial run (process or decision) is appended.
- Implementation: `internal/run.BuildRetryPrompt(node, priorCritiques)`, `internal/run.FormatValidationCritique(validationResponse)`.

**Retry CLI**

- Node’s `RetryCLI` if set; otherwise options’ default retry CLI (e.g. `DEFAULT_RETRY_CLI`).
- Implementation: `internal/run.ResolveRetryCLI(node, opts.DefaultRetryCLI)`.

**Retry limit**

- Node’s `Retries` if &gt; 0; otherwise conventional default 3.
- Implementation: `internal/run.EffectiveRetryLimit(node)`.

**Per-retry steps**

1. Build retry prompt (original + prior critiques + response-type instruction).
2. Run the retry CLI with that prompt.
3. Verify stdout parses as the correct response type (`VerifyRunOutput`).
4. Run validation on the new stdout.
5. If `fully_completed` is true: success; stop retries and proceed.
6. Otherwise append this attempt’s critique to the list and repeat until the retry limit. If the limit is reached without success, the runner reports validation did not pass (max retries reached).

Implementation: `internal/run.runRetryLoop` (called from `RunNodeThenValidate` when the first validation fails).

---

## Execution flow

- Entry: `internal/runlog.ExecuteTree(root, opts, workDir, logDir, chartName, writeShort, writeLong)`.
- For each node, in tree order (depth-first from root):
  1. Call `internal/run.RunNodeThenValidate(node, opts)`:
     - Run node (RunNode).
     - If `ShouldValidate(node)`: run validation (RunValidation). If not valid, run retry loop until valid or limit.
  2. Log the node result (run output, validation if any, retry count).
  3. If the node has children: parse run stdout as `DecisionResponse`, select child by `d.Answer`, recurse on that child. If parsing fails, the tree run returns the parse error.
  4. If the node has no children: continue to the next sibling or end of tree.
- When the tree walk completes, logs are written (short JSON and/or long log) to the configured log directory.

---

## Response types (summary)

| Kind        | Type                 | Use                    | Keys (JSON)                                                                 |
|------------|----------------------|------------------------|-----------------------------------------------------------------------------|
| Process    | `ProcessResponse`    | Childless node output  | `completed`, `secs_taken`, `tokens_used`, `comments`                        |
| Decision   | `DecisionResponse`   | Node with children     | `choices`, `answer`, `reasons`                                             |
| Validation | `ValidationResponse`| Validation step output | `fully_completed`, `partially_completed`, `should_retry`, `warnings`         |

Defined in `types/responses.go`; parsed via `ParseProcessResponse`, `ParseDecisionResponse`, `ParseValidationResponse` (with handling for markdown fences and trailing non-JSON).

---

## Related docs

- **Tags and metadata:** `readme/metadata.md` (NoValidation, CLI codename, validate_prompt, validate_cli, retry_cli, retries, timeout).
- **Settings and defaults:** `readme/settings.md` (DEFAULT_CLI, DEFAULT_VALIDATE_CLI, DEFAULT_RETRY_CLI, DEFAULT_RETRY_COUNT, LOG_DIR, etc.).
