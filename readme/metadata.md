# Node tags and metadata variables

You can customize how each node in your decision tree behaves by adding **tags** and **metadata variables** directly on the node in Lucidchart. This lets you override defaults (which CLI runs the node, validation, retries, timeouts) without changing your application code—the chart itself carries the configuration.

---

## Tags vs. metadata variables

| | **Tags** | **Metadata variables** |
|---|----------|-------------------------|
| **What they are** | Short labels on the node (e.g. `NoValidation`, `GEMINI`). | Key/value pairs in the node’s **Data** section (e.g. `validate_prompt` = `Did the model follow the instructions?`). |
| **Where to set them** | In the **Tags** section of the node’s context panel, with a process node selected. | In the **Data** section of the node’s context panel, with a process node selected. |
| **Differences** | Visible from chart with color codes. No data values. | Visible only from data context panel. Allows custom data values. |

Both **tags** and **metadata variables** are **case-insensitive** and **snake/camel-safe**. For example, `NoValidation`, `novalidation`, and `no_validation` are equivalent tags; `validate_cli`, `validateCli`, and `Validate_CLI` are equivalent variable names. Variable values (CLI codenames) are normalized to uppercase; freeform values are trimmed of surrounding whitespace.

---

## Adding tags and variables in Lucidchart

**1. Open panel** – Select a process node and open the context panel (right side of the canvas).

<p align="center"><img src="https://raw.githubusercontent.com/Macioa/MonadsCLI/main/images/context_panel.png" alt="Open the context panel" /></p>

**2. Add tags** – Click **Add tag** to add a new tag to the node. Click **Data** to access metadata variables.

<p align="center"><img src="https://raw.githubusercontent.com/Macioa/MonadsCLI/main/images/metadata.png" alt="Add tags" /></p>

**3. Add metadata variables if not present** – If no variables exist, click **New data field** to add the first one.

<p align="center"><img src="https://raw.githubusercontent.com/Macioa/MonadsCLI/main/images/add_vars.png" alt="Add metadata variables if not present" /></p>

**4. Add or edit metadata variables** – Click to add new variables or edit existing variable names and values.

<p align="center"><img src="https://raw.githubusercontent.com/Macioa/MonadsCLI/main/images/variables.png" alt="Add or edit metadata variables" /></p>

---

## Tags

| Tag | Effect |
|-----|--------|
| **NoValidation** | Validation is skipped for this node. Use for steps that don’t need a check. Accepts NoValidation, novalidation, no_validation, noValidation. |
| **&lt;CLI codename&gt;** | Use that CLI to run this node instead of the default. Any tag matching a known CLI codename (`GEMINI`, `CURSOR`, `CLAUDE`, `COPILOT`, `AIDER`, `QODO`) sets the node’s CLI. Case-insensitive (e.g. gemini, GEMINI). The first matching tag wins. |

If a node has no CLI tag and no metadata variable for CLI, the runner uses the default from settings.

---

## Metadata variables

Variables map to default settings. A node can override any default by setting the corresponding variable in its **Data** section.

| Variable | Effect | Example value |
|----------|--------|---------------|
| **cli** / **codename** | Which CLI runs this node | `GEMINI`, `CURSOR`, `CLAUDE`, `COPILOT`, `AIDER`, `QODO` |
| **validate_cli** | Which CLI validates this node’s response | Same as above |
| **retry_cli** | Which CLI retries after validation failure | Same as above |
| **retries** | Maximum retries when validation fails | `3`, `5` |
| **timeout** | Timeout in seconds for CLI operations (0 = use runner default) | `600`, `300` |
| **validate_prompt** | Custom validation prompt text; ignored if node has **NoValidation** tag | `Did the model follow the instructions exactly?` |

Unset variables fall back to your defaults; node values always override them. Variable names are case-insensitive and snake/camel-safe (e.g. `validate_cli`, `validateCli`, `Validate_CLI`). CLI codename values are normalized to uppercase.
