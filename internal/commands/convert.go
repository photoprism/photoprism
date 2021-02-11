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
	Usage:  "Converts originals in other formats to JPEG and AVC sidecar files",
	Action: convertAction,
}

// convertAction converts originals in other formats to JPEG and AVC sidecar files.
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

	log.Infof("converting originals in %s", conf.OriginalsPath())

	convert := service.Convert()

	if err := convert.Start(conf.OriginalsPath()); err != nil {
		log.Error(err)
	}

	elapsed := time.Since(start)

	log.Infof("converting completed in %s", elapsed)

	return nil
}
