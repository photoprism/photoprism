package commands

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/urfave/cli"
)

// ConvertCommand is used to register the convert cli command
var ConvertCommand = cli.Command{
	Name:   "convert",
	Usage:  "Converts originals in other formats to JPEG",
	Action: convertAction,
}

// convertAction converts RAW files to JPEG images, if no JPEG already exists
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

	convert := photoprism.NewConvert(conf)

	convert.Start(conf.OriginalsPath())

	elapsed := time.Since(start)

	log.Infof("image conversion completed in %s", elapsed)

	return nil
}
