package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// visionSaveFlags specifies the available command flags.
var visionSaveFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "replaces an existing vision.yml file",
	},
}

// VisionSaveCommand writes the model configuration to vision.yml.
var VisionSaveCommand = &cli.Command{
	Name:   "save",
	Usage:  "Saves the current model configuration to the vision.yml file",
	Flags:  visionSaveFlags,
	Action: visionSaveAction,
}

// visionListAction displays existing user accounts.
func visionSaveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		force := ctx.Bool("force")

		fileName := conf.VisionYaml()

		if !force && fs.FileExistsNotEmpty(fileName) {
			return fmt.Errorf("%s already exists", clean.Log(fileName))
		}

		return vision.Config.Save(fileName)
	})
}
