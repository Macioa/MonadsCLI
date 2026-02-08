package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ryanmontgomery/MonadsCLI/internal/cli"
	"github.com/ryanmontgomery/MonadsCLI/internal/installer"
	"github.com/ryanmontgomery/MonadsCLI/internal/report"
	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
)

func main() {
	cli.Execute([]cli.Command{
		installCommand(),
		runCommand(),
	})
}

func installCommand() cli.Command {
	return cli.Command{
		Name:        "install",
		Description: "Install one or more supported CLIs",
		Run: func(fs *flag.FlagSet) error {
			names := fs.Args()
			_, err := installer.InstallCLIs(names)
			return err
		},
	}
}

func runCommand() cli.Command {
	var command string
	var reportPath string
	var shell string
	var workDir string
	var shellArgs cli.StringList

	return cli.Command{
		Name:        "run",
		Description: "Run a shell command and write a report",
		Flags: func(fs *flag.FlagSet) {
			fs.StringVar(&command, "command", "", "Shell command to run")
			fs.StringVar(&reportPath, "report", "", "Path to write report JSON")
			fs.StringVar(&shell, "shell", "", "Shell executable to use")
			fs.Var(&shellArgs, "shell-arg", "Shell arg (repeatable)")
			fs.StringVar(&workDir, "workdir", "", "Working directory for the command")
		},
		Run: func(fs *flag.FlagSet) error {
			if command == "" {
				return fmt.Errorf("missing --command")
			}

			shellCmd, defaultArgs := runner.DefaultShell()
			if shell != "" {
				shellCmd = shell
			}

			var args []string
			if len(shellArgs) > 0 {
				args = append(args, shellArgs...)
			} else {
				args = append(args, defaultArgs...)
			}

			result, err := runner.RunShellCommand(runner.CommandSpec{
				Shell:     shellCmd,
				ShellArgs: args,
				Command:   command,
				WorkDir:   workDir,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "command failed: %v\n", err)
			}

			if reportPath == "" {
				reportPath = filepath.Join("reports", fmt.Sprintf("report-%s.json", time.Now().Format("20060102-150405")))
			}

			if err := report.WriteJSON(reportPath, result); err != nil {
				return fmt.Errorf("write report: %w", err)
			}

			if result.ExitCode != 0 {
				return cli.ExitError{Code: result.ExitCode}
			}
			return nil
		},
	}
}
