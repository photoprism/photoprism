package commands

import (
	"github.com/urfave/cli/v2"
)

// VisionCommands configures the computer vision subcommands.
var VisionCommands = &cli.Command{
	Name:  "vision",
	Usage: "Computer vision subcommands",
	Subcommands: []*cli.Command{
		VisionListCommand,
		VisionRunCommand,
		VisionSaveCommand,
	},
}
