package commands

import "github.com/urfave/cli"

// MigrateCommand configures the command name, flags, and action.
var MigrateCommand = cli.Command{
	Name:      "migrate",
	Usage:     MigrationsRunCommand.Usage,
	ArgsUsage: MigrationsRunCommand.ArgsUsage,
	Flags:     MigrationsRunCommand.Flags,
	Action:    migrationsRunAction,
}
