# Node Tags and Metadata Variables

**Tags** and **metadata variables** are added to nodes in Lucidchart. The chart stores this configuration, so overrides for CLI, validation, retries, and timeouts live in the diagram rather than in separate config files.

---

## Adding Tags and Variables in Lucidchart

<p align="center"><strong>1. Open panel</strong> – Select a process node and open the context panel (right side of the canvas).</p>

<p align="center"><img src="../images/context_panel.png" alt="Open the context panel" /></p>

<p align="center"><strong>2. Add tags</strong> – Click <strong>Add tag</strong> to add a new tag to the node. Click <strong>Data</strong> to access metadata variables.</p>

<p align="center"><img src="../images/metadata.png" alt="Add tags" /></p>

<p align="center"><strong>3. Add metadata variables if not present</strong> – If no variables exist, click <strong>New data field</strong> to add the first one.</p>

<p align="center"><img src="../images/add_vars.png" alt="Add metadata variables if not present" /></p>

<p align="center"><strong>4. Add or edit metadata variables</strong> – Click to add new variables or edit existing variable names and values.</p>

<p align="center"><img src="../images/variables.png" alt="Add or edit metadata variables" /></p>

---

## Tags vs. Metadata Variables

| | **Tags** | **Metadata variables** |
|---|----------|-------------------------|
| **What they are** | Short labels on the node (e.g. `NoValidation`, `GEMINI`). | Key/value pairs in the node's **Data** section (e.g. `validate_prompt` = `Did the model follow the instructions?`). |
| **Where to set them** | In the **Tags** section of the node's context panel, with a process node selected. | In the **Data** section of the node's context panel, with a process node selected. |
| **Differences** | Visible from chart with color codes. No data values. | Visible only from data context panel. Allows custom data values. |

Both **tags** and **metadata variables** are **case-insensitive** and **snake/camel-safe**. For example, `NoValidation`, `novalidation`, and `no_validation` are equivalent tags; `validate_cli`, `validateCli`, and `Validate_CLI` are equivalent variable names. Variable values (CLI codenames) are normalized to uppercase; freeform values are trimmed of surrounding whitespace.

---

## Tags

| Tag | Effect |
|-----|--------|
| **NoValidation** | Validation is skipped for this node. Use for steps that don’t need a check. Accepts NoValidation, novalidation, no_validation, noValidation. |
| **&lt;CLI codename&gt;** | Use that CLI to run this node instead of the default. Any tag matching a known CLI codename (`GEMINI`, `CURSOR`, `CLAUDE`, `COPILOT`, `AIDER`, `QODO`) sets the node’s CLI. Case-insensitive (e.g. gemini, GEMINI) |



---

## Metadata Variables

Variables map to default settings. A node can override any default by setting the corresponding variable in its **Data** section.

| Variable | Effect | Example value |
|----------|--------|---------------|
| **validate_cli** | Which CLI validates this node’s response | Same as above |
| **retry_cli** | Which CLI retries after validation failure | Same as above |
| **retries** | Maximum retries when validation fails | `3`, `5` |
| **timeout** | Timeout in seconds for CLI operations (0 = use runner default) | `600`, `300` |
| **validate_prompt** | Custom validation prompt text; ignored if node has **NoValidation** tag | `Did the model follow the instructions exactly?` |

Unset variables fall back to your defaults; node values always override them. Variable names are case-insensitive and snake/camel-safe (e.g. `validate_cli`, `validateCli`, `Validate_CLI`). CLI codename values are normalized to uppercase.
