package commands

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/workers"
)

// VisionRunCommand configures the command name, flags, and action.
var VisionRunCommand = &cli.Command{
	Name:      "run",
	Usage:     "Runs a computer vision model",
	ArgsUsage: "[filter]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "models",
			Aliases: []string{"m"},
			Usage:   "model types (labels, nsfw, caption)",
		},
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "force existing metadata to be updated",
		},
	},
	Action: visionRunAction,
	Hidden: true,
}

// visionListAction displays existing user accounts.
func visionRunAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		worker := workers.NewVision(conf)
		return worker.Start(strings.TrimSpace(ctx.Args().First()), vision.ParseTypes(ctx.String("models")), ctx.Bool("force"))
	})
}
