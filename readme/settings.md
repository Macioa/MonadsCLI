# Settings CLI

Manage settings with the `settings` subcommand.

## Read settings

Print settings as JSON:

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
