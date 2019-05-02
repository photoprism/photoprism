package commands

import (
	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/photoprism"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Converts RAW files to JPEG images, if no JPEG already exists
var ConvertCommand = cli.Command{
	Name:   "convert",
	Usage:  "Converts RAW originals to JPEG",
	Action: convertAction,
}

func convertAction(ctx *cli.Context) error {
	conf := context.NewConfig(ctx)

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	log.Infof("converting RAW images in %s to JPEG", conf.OriginalsPath())

	converter := photoprism.NewConverter(conf.DarktableCli())

	converter.ConvertAll(conf.OriginalsPath())

	log.Infof("image conversion complete")

	return nil
}
