package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
)

// VisionSaveCommand writes the model configuration to vision.yml.
var VisionSaveCommand = &cli.Command{
	Name:   "save",
	Usage:  "Writes the model configuration to vision.yml",
	Action: visionSaveAction,
}

// visionListAction displays existing user accounts.
func visionSaveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		return vision.Config.Save(conf.VisionYaml())
	})
}
