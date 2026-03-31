<div style="background-color: white;">

# Quick start

**1. Install**

```bash
go install github.com/ryanmontgomery/MonadsCLI/cmd/monadscli@latest
```

See [Install](install.md) for PATH and supported platforms.

**2. Install one or more agents and save your key** ([settings page](settings.md))

```bash
monadscli install cursor && monadscli settings set CURSOR_API_KEY=YOUR_KEY
```

**3. Create a decision tree**

Design and export your tree in Lucidchart. See [Creating a Lucidchart decision tree](create-tree.md).

**4. Run the tree**

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

</div>
