package commands

import (
	"github.com/urfave/cli"
)

// ShowCommand configures the show subcommands.
var ShowCommand = cli.Command{
	Name:  "show",
	Usage: "Shows supported formats, features, and config options",
	Subcommands: []cli.Command{
		ShowConfigCommand,
		ShowConfigOptionsCommand,
		ShowConfigYamlCommand,
		ShowSearchFiltersCommand,
		ShowFileFormatsCommand,
		ShowThumbSizesCommand,
		ShowVideoSizesCommand,
		ShowMetadataCommand,
	},
}
