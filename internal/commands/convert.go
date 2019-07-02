package commands

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

// Converts RAW files to JPEG images, if no JPEG already exists
var ConvertCommand = cli.Command{
	Name:   "convert",
	Usage:  "Converts RAW originals to JPEG",
	Action: convertAction,
}

func convertAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)

	if conf.ReadOnly() {
		return config.ErrReadOnly
	}

	if err := conf.CreateDirectories(); err != nil {
		return err
	}

	log.Infof("converting RAW images in %s to JPEG", conf.OriginalsPath())

	converter := photoprism.NewConverter(conf)

	converter.ConvertAll(conf.OriginalsPath())

	elapsed := time.Since(start)

	log.Infof("image conversion completed in %s", elapsed)

	return nil
}
