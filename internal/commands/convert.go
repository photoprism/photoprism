package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// ConvertCommand registers the convert cli command.
var ConvertCommand = cli.Command{
	Name:   "convert",
	Usage:  "Converts originals in other formats to JPEG",
	Action: convertAction,
}

// convertAction converts RAW files to JPEG images, if no JPEG already exists
func convertAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if !conf.SidecarWritable() {
		return config.ErrReadOnly
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	log.Infof("creating JPEGs for other file types in %s", conf.OriginalsPath())

	convert := service.Convert()

	if err := convert.Start(conf.OriginalsPath()); err != nil {
		log.Error(err)
	}

	elapsed := time.Since(start)

	log.Infof("converting to JPEG completed in %s", elapsed)

	return nil
}
