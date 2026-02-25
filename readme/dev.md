# MonadsCLI

Minimal Go CLI scaffold with subcommands, shell execution, and report output.

## Install

Install asdf:

```bash
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.14.0
```

Add to shell (bash/zsh):

```bash
. "$HOME/.asdf/asdf.sh"
```

Install Go via asdf:

```bash
asdf plugin add golang
asdf install golang 1.20.0
asdf global golang 1.20.0
```

## Usage

```bash
go run ./cmd/monadscli run --command "echo hello"
```

## Build and Run

Build a binary:

```bash
go build -o bin/monadscli ./cmd/monadscli
```

Run the binary:

```bash
./bin/monadscli run --command "echo hello"
```

### Examples

```bash
go run ./cmd/monadscli run --command "ls -la" --report ./reports/ls.json
go run ./cmd/monadscli run --command "Get-Process" --shell powershell --shell-arg -Command
```

Reports are written as JSON and include stdout/stderr, timing, and exit code.

---

## Docs

- [Quick](quick.md)
- [Install](install.md)
- [Creating a Lucidchart decision tree](create-tree.md)
- [Metadata in trees](metadata.md)
- [Settings and keys](settings.md)
