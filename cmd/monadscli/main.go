package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryanmontgomery/MonadsCLI/internal/cli"
	"github.com/ryanmontgomery/MonadsCLI/internal/installer"
	"github.com/ryanmontgomery/MonadsCLI/internal/report"
	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
	"github.com/ryanmontgomery/MonadsCLI/internal/settings"
)

func main() {
	cli.Execute([]cli.Command{
		installCommand(),
		runCommand(),
		settingsCommand(),
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

func settingsCommand() cli.Command {
	return cli.Command{
		Name:        "settings",
		Description: "Manage encrypted settings",
		Run: func(fs *flag.FlagSet) error {
			args := fs.Args()
			if len(args) == 0 {
				return fmt.Errorf("missing settings subcommand")
			}

			switch args[0] {
			case "get":
				payload, err := settings.Get()
				if err != nil {
					return err
				}
				fmt.Fprintln(os.Stdout, string(payload))
				return nil
			case "set":
				return settingsSet(args[1:])
			case "from-env":
				_, err := settings.FromEnv()
				return err
			case "from-file":
				return settingsFromFile(args[1:])
			case "from-json":
				return settingsFromJSON(args[1:])
			case "to-env":
				_, err := settings.ToEnv()
				return err
			case "to-file":
				return settingsToFile(args[1:])
			default:
				return fmt.Errorf("unknown settings subcommand: %s", args[0])
			}
		},
	}
}

func settingsSet(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing KEY=VALUE pairs")
	}

	values := settings.Settings{}
	for _, pair := range args {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid setting: %s", pair)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return fmt.Errorf("invalid setting: %s", pair)
		}
		values[key] = parts[1]
	}

	return settings.Set(values)
}

func settingsFromFile(args []string) error {
	fs := flag.NewFlagSet("settings from-file", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var path string
	fs.StringVar(&path, "path", "", "Path to .env file")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if path == "" {
		return fmt.Errorf("missing --path")
	}

	_, err := settings.FromFile(path)
	return err
}

func settingsFromJSON(args []string) error {
	fs := flag.NewFlagSet("settings from-json", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var payload string
	fs.StringVar(&payload, "json", "", "Stringified JSON settings payload")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if payload == "" {
		return fmt.Errorf("missing --json")
	}

	_, err := settings.FromJSON(payload)
	return err
}

func settingsToFile(args []string) error {
	fs := flag.NewFlagSet("settings to-file", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var path string
	fs.StringVar(&path, "path", "", "Path to write .env file")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if path == "" {
		return fmt.Errorf("missing --path")
	}

	_, err := settings.ToFile(path)
	return err
}
