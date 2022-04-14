package commands

import (
	"github.com/urfave/cli"
)

// ShowCommand registers the show subcommands.
var ShowCommand = cli.Command{
	Name:  "show",
	Usage: "Configuration and system report subcommands",
	Subcommands: []cli.Command{
		ShowConfigCommand,
		ShowTagsCommand,
		ShowFiltersCommand,
		ShowFormatsCommand,
	},
}
