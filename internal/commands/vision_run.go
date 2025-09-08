package commands

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/workers"
)

// VisionRunCommand configures the command name, flags, and action.
var VisionRunCommand = &cli.Command{
	Name:      "run",
	Usage:     "Runs one or more computer vision models on a set of pictures that match the specified search filters",
	ArgsUsage: "[filter...]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "models",
			Aliases: []string{"m"},
			Usage:   "computer vision `MODELS` to run, e.g. caption, labels, or nsfw",
			Value:   "caption",
		},
		&cli.IntFlag{
			Name:    "count",
			Aliases: []string{"n"},
			Usage:   "maximum `NUMBER` of pictures to be processed",
			Value:   search.MaxResults,
		},
		&cli.StringFlag{
			Name:    "source",
			Aliases: []string{"s"},
			Value:   entity.SrcImage,
			Usage:   "custom data source `TYPE` e.g. default, image, meta, vision, or admin",
		},
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "replaces existing data if the model supports it and the source priority is equal or higher",
		},
	},
	Action: visionRunAction,
}

// visionListAction displays existing user accounts.
func visionRunAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		worker := workers.NewVision(conf)
		filter := strings.TrimSpace(strings.Join(ctx.Args().Slice(), " "))
		return worker.Start(
			filter,
			ctx.Int("count"),
			vision.ParseTypes(ctx.String("models")),
			ctx.String("source"),
			ctx.Bool("force"),
		)
	})
}
