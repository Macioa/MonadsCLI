package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ryanmontgomery/MonadsCLI/internal/cli"
)

func loginCommand() cli.Command {
	return cli.Command{
		Name:        "login",
		Description: "Login to supported services (lucid)",
		Run: func(_ *flag.FlagSet) error {
			args := os.Args[2:]
			if len(args) != 1 {
				return fmt.Errorf("usage: monadscli login <lucid>")
			}

			switch args[0] {
			case "lucid":
				return lucidLogin(nil)
			default:
				return fmt.Errorf("unknown login target: %s", args[0])
			}
		},
	}
}
