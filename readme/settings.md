<div style="background-color: white;">

# Settings CLI

Manage settings with the `settings` subcommand.

## Supported CLIs (codename)

Use these **codename** values for `DEFAULT_CLI`, `DEFAULT_RETRY_CLI`, and `DEFAULT_VALIDATE_CLI`:

| Codename | CLI            | API key setting(s)                    |
|----------|----------------|----------------------------------------|
| GEMINI   | Gemini CLI     | GEMINI_API_KEY                         |
| CURSOR   | Cursor CLI     | CURSOR_API_KEY                         |
| CLAUDE   | Claude CLI     | ANTHROPIC_API_KEY                      |
| COPILOT  | GitHub Copilot CLI | GH_TOKEN                           |
| AIDER    | Aider          | OPENAI_API_KEY, ANTHROPIC_API_KEY      |
| QODO     | Qodo Gen CLI   | QODO_API_KEY                           |

## Available settings

### Defaults / behavior

| Key | Description | Default |
|-----|-------------|---------|
| DEFAULT_CLI | Codename of CLI to use by default | CURSOR |
| DEFAULT_TIMEOUT | Default timeout in seconds for CLI operations | 600 (10 min) |
| DEFAULT_RETRY_CLI | Codename of CLI to use for retries | CURSOR |
| DEFAULT_RETRY_COUNT | Maximum number of retries | 3 |
| DEFAULT_VALIDATE_CLI | Codename of CLI to use for validation | CURSOR |
| LOG_DIR | Relative path for run logs (from CLI cwd) | ./_monad_logs/ |
| WRITE_LOG_SHORT | Write short log (response JSONs per node + validations/retries) | true |
| WRITE_LOG_LONG | Write long log (full LLM output per run) | true |

### Agentic CLI API keys

- ANTHROPIC_API_KEY — [Get key](https://console.anthropic.com/settings/keys)
- CURSOR_API_KEY — [Get key](https://cursor.com/dashboard?tab=integrations)
- GEMINI_API_KEY — [Get key](https://aistudio.google.com/app/apikey)
- GH_TOKEN — [Get key](https://github.com/settings/personal-access-tokens/new)
- OPENAI_API_KEY — [Get key](https://platform.openai.com/api-keys)
- QODO_API_KEY — https://app.qodo.ai/ — **⚠️ Warning:** Must install on a browser-enabled device and use `qodo login` to obtain a key. 

### Lucidchart

- LUCIDCHART_API_KEY — [Get key](https://lucid.app/developer#/apikeys)

## Read settings

Print settings as `.env` format:

```bash
monadscli settings get
```

## Write settings

Set key/value pairs:

```bash
monadscli settings set GEMINI_API_KEY=... CURSOR_API_KEY=...
```

Load from current environment:

```bash
monadscli settings from-env
```

Load from a `.env` file:

```bash
monadscli settings from-file --path ./secrets.env
```

Load from JSON:

```bash
monadscli settings from-json --json '{"GEMINI_API_KEY":"..."}'
```

## Export settings

Apply settings to the current process environment:

```bash
monadscli settings to-env
```

Write settings to a `.env` file:

```bash
monadscli settings to-file --path ./out.env
```

---

## Docs

- [Quick](quick.md)
- [Install](install.md)
- [Creating a Lucidchart decision tree](create-tree.md)
- [Metadata in trees](metadata.md)
- [Settings and keys](settings.md)

</div>
