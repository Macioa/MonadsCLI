package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
)

type Command struct {
	Name        string
	Description string
	Flags       func(*flag.FlagSet)
	Run         func(*flag.FlagSet) error
}

type ExitError struct {
	Code int
}

func (e ExitError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

type StringList []string

func (s *StringList) String() string {
	return fmt.Sprint([]string(*s))
}

func (s *StringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Execute(commands []Command) {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	if len(os.Args) < 2 {
		printUsage(commands)
		os.Exit(2)
	}

	name := os.Args[1]
	if name == "-h" || name == "--help" {
		printUsage(commands)
		return
	}

	cmd, ok := findCommand(commands, name)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
		printUsage(commands)
		os.Exit(2)
	}

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		printCommandUsage(cmd, fs)
	}

	if cmd.Flags != nil {
		cmd.Flags(fs)
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if cmd.Run == nil {
		return
	}

	if err := cmd.Run(fs); err != nil {
		var exitErr ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func findCommand(commands []Command, name string) (Command, bool) {
	for _, cmd := range commands {
		if cmd.Name == name {
			return cmd, true
		}
	}
	return Command{}, false
}

func printUsage(commands []Command) {
	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "  monadscli <command> [options]")
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Commands:")

	for _, cmd := range commands {
		if cmd.Description == "" {
			fmt.Fprintf(os.Stdout, "  %s\n", cmd.Name)
			continue
		}
		fmt.Fprintf(os.Stdout, "  %-12s %s\n", cmd.Name, cmd.Description)
	}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Run 'monadscli <command> --help' for command usage.")
}

func printCommandUsage(cmd Command, fs *flag.FlagSet) {
	fmt.Fprintf(os.Stdout, "Usage:\n  monadscli %s [options]\n\n", cmd.Name)
	if cmd.Description != "" {
		fmt.Fprintln(os.Stdout, cmd.Description)
		fmt.Fprintln(os.Stdout)
	}
	fmt.Fprintln(os.Stdout, "Options:")
	fs.PrintDefaults()
}
