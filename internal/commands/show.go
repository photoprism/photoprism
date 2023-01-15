package commands

import (
	"github.com/urfave/cli"
)

// ShowCommand registers the show subcommands.
var ShowCommand = cli.Command{
	Name:  "show",
	Usage: "Shows supported formats, features, and config options",
	Subcommands: []cli.Command{
		ShowConfigCommand,
		ShowFlagsCommand,
		ShowOptionsCommand,
		ShowFiltersCommand,
		ShowFormatsCommand,
		ShowTagsCommand,
	},
}
