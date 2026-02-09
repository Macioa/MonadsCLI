# Settings CLI

Manage settings with the `settings` subcommand.

Available settings:

### Agentic CLIs

- ANTHROPIC_API_KEY — [Get key](https://console.anthropic.com/settings/keys)
- CURSOR_API_KEY — [Get key](https://cursor.com/dashboard)
- GEMINI_API_KEY — [Get key](https://aistudio.google.com/app/apikey)
- GH_TOKEN — [Get key](https://github.com/settings/personal-access-tokens/new)
- OPENAI_API_KEY — [Get key](https://platform.openai.com/api-keys)
- QODO_API_KEY — [Get key](https://app.qodo.ai/)

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
