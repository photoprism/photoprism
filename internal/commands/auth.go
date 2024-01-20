package commands

import (
	"github.com/urfave/cli"
)

// AuthCommands registers the API authentication subcommands.
var AuthCommands = cli.Command{
	Name:    "auth",
	Aliases: []string{"sess"},
	Usage:   "API authentication subcommands",
	Subcommands: []cli.Command{
		AuthListCommand,
		AuthAddCommand,
		AuthShowCommand,
		AuthRemoveCommand,
		AuthResetCommand,
	},
}

// tokensFlag represents a CLI flag to include tokens in a report.
var tokensFlag = cli.BoolFlag{
	Name:  "tokens",
	Usage: "show preview and download tokens",
}
