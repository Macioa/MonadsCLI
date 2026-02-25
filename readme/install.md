# Install

<p align="center">Install the CLI with Go.</p>

```bash
go install github.com/ryanmontgomery/MonadsCLI/cmd/monadscli@latest
```

<p align="center">Ensure <code>$(go env GOPATH)/bin</code> is on your <code>PATH</code> so <code>monadscli</code> runs from anywhere.</p>

---

## Supported operating systems

MonadsCLI is a single binary. The following are **tested in CI** (GitHub Actions):

| OS    | Architecture | Notes                                                                                          |
|-------|--------------|------------------------------------------------------------------------------------------------|
| Linux | amd64        | **Ubuntu**, **Debian**, **Fedora**, **Alpine**. Use Bash. PowerShell is also tested on Ubuntu. |

<p align="center">macOS (arm64, amd64) and Windows (arm64, amd64) are supported by Go and should work but are <strong>not tested</strong> in this repo’s CI. On macOS use Bash or zsh in Terminal; on Windows use <code>monadscli.exe</code> in <strong>PowerShell</strong> or Windows Terminal.</p>

Requires **Go 1.24 or later** for the install command above. If you prefer not to install Go, you can build from source (see [dev](dev.md)) or use a pre-built binary from the project’s releases when available.

---

## Adding API keys

Configure your agent and API keys after install. For supported CLIs, key names, and how to set them (environment, <code>monadscli settings set</code>, or <code>.env</code>), see [Settings and keys](settings.md).

---

## Docs

- [Quick](quick.md)
- [Install](install.md)
- [Creating a Lucidchart decision tree](create-tree.md)
- [Metadata in trees](metadata.md)
- [Settings and keys](settings.md)
