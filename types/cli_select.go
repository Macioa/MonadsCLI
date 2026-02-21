package types

import (
	"fmt"
	"strings"
)

// AvailableCLIs returns a copy of all known CLI definitions.
func AvailableCLIs() []CLI {
	return append([]CLI{}, AllCLIs...)
}

// SelectCLIs returns specific CLIs by name or command.
// If names is empty, it returns all available CLIs.
func SelectCLIs(names []string) ([]CLI, error) {
	if len(names) == 0 {
		return AvailableCLIs(), nil
	}

	index := map[string]CLI{}
	for _, cli := range AllCLIs {
		index[normalizeCLIKey(cli.Name)] = cli
		index[normalizeCLIKey(cli.Command)] = cli
		if cli.Codename != "" {
			index[normalizeCLIKey(cli.Codename)] = cli
		}
	}

	selected := make([]CLI, 0, len(names))
	for _, name := range names {
		key := normalizeCLIKey(name)
		cli, ok := index[key]
		if !ok {
			return nil, fmt.Errorf("unknown cli: %s", name)
		}
		selected = append(selected, cli)
	}

	return selected, nil
}

func normalizeCLIKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
