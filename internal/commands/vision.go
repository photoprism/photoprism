package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VisionCommands configures the computer vision subcommands.
var VisionCommands = &cli.Command{
	Name:  "vision",
	Usage: "Computer vision subcommands",
	Subcommands: []*cli.Command{
		VisionListCommand,
		VisionRunCommand,
		VisionSourcesCommand,
		VisionSaveCommand,
	},
}

// VisionSourcesCommand configures the command name, flags, and action.
var VisionSourcesCommand = &cli.Command{
	Name:   "sources",
	Usage:  "Displays supported metadata sources and their priorities",
	Flags:  append(report.CliFlags),
	Action: showSourcesAction,
}
