package commands

import (
	"context"
	"time"

	"github.com/manifoldco/promptui"

	"github.com/photoprism/photoprism/internal/photoprism"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/urfave/cli"
)

// FacesCommand registers the faces cli command.
var FacesCommand = cli.Command{
	Name:  "faces",
	Usage: "Facial recognition sub-commands",
	Subcommands: []cli.Command{
		{
			Name:   "stats",
			Usage:  "Shows stats on face samples",
			Action: facesStatsAction,
		},
		{
			Name:   "reset",
			Usage:  "Resets recognized faces",
			Action: facesResetAction,
		},
		{
			Name:  "index",
			Usage: "Performs facial recognition",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "reindex existing faces",
				},
			},
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

// facesResetAction resets face clusters and matches.
func facesResetAction(ctx *cli.Context) error {
	actionPrompt := promptui.Prompt{
		Label:     "Remove automatically recognized faces, matches, and dangling subjects?",
		IsConfirm: true,
	}

	if _, err := actionPrompt.Run(); err != nil {
		return nil
	}

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

	if err := w.Reset(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}

// facesIndexAction performs face clustering and matching.
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

	opt := photoprism.FacesOptions{
		Force: ctx.Bool("force"),
	}

	w := service.Faces()

	if err := w.Start(opt); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
