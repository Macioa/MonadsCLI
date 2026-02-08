package installer

import (
	"fmt"

	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
	"github.com/ryanmontgomery/MonadsCLI/types"
)

// InstallCLIs installs the requested CLIs. If names is empty, installs all.
func InstallCLIs(names []string) ([]runner.Result, error) {
	selected, err := types.SelectCLIs(names)
	if err != nil {
		return nil, err
	}

	shell, shellArgs := runner.DefaultShell()
	results := make([]runner.Result, 0, len(selected))

	for _, cli := range selected {
		result, runErr := runner.RunShellCommand(runner.CommandSpec{
			Shell:     shell,
			ShellArgs: shellArgs,
			Command:   cli.Install,
		})
		results = append(results, result)

		if runErr != nil {
			return results, fmt.Errorf("install %s: %w", cli.Command, runErr)
		}
	}

	return results, nil
}
