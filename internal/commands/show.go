package commands

import (
	"github.com/urfave/cli"
)

// ShowCommand registers the show subcommands.
var ShowCommand = cli.Command{
	Name:  "show",
	Usage: "Shows supported formats, standards, and features",
	Subcommands: []cli.Command{
		ShowConfigCommand,
		ShowOptionsCommand,
		ShowFiltersCommand,
		ShowFormatsCommand,
		ShowTagsCommand,
	},
}
