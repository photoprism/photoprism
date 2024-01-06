package commands

import (
	"github.com/urfave/cli"
)

// ShowCommands configures the show subcommands.
var ShowCommands = cli.Command{
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
