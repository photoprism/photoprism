package commands

import (
	"context"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// FacesCommand registers the faces cli command.
var FacesCommand = cli.Command{
	Name:  "faces",
	Usage: "Runs facial recognition sub-commands",
	Subcommands: []cli.Command{
		{
			Name:   "stats",
			Usage:  "Shows stats on face embeddings",
			Action: facesStatsAction,
		},
		{
			Name:   "index",
			Usage:  "Performs face clustering and recognition",
			Action: facesIndexAction,
		},
	},
}

// facesStatsAction shows stats on face embeddings.
func facesStatsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Faces()

	if err := w.Analyze(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesIndexAction performs face clustering and recognition.
func facesIndexAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Faces()

	if err := w.Start(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
