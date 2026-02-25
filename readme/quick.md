# Quick start

**1. Install**

```bash
go install github.com/ryanmontgomery/MonadsCLI/cmd/monadscli@latest
```

See [Install](install.md) for PATH and supported platforms.

**2. Create a decision tree**

Design and export your tree in Lucidchart. See [Creating a Lucidchart decision tree](create-tree.md).

**3. Run the tree**

**From CSV** (no API key):

```bash
monadscli run-tree --csv path/to/your-tree.csv
```

**From Lucid cloud** (requires [Lucid developer API key](settings.md)):

```bash
monadscli lucid document --id <document-id>
```

---

## Docs

- [Quick](quick.md)
- [Install](install.md)
- [Creating a Lucidchart decision tree](create-tree.md)
- [Metadata in trees](metadata.md)
- [Settings and keys](settings.md)
