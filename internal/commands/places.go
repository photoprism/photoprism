package commands

import (
	"context"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
)

// PlacesCommand registers the places subcommands.
var PlacesCommand = cli.Command{
	Name:  "places",
	Usage: "Geodata management subcommands",
	Subcommands: []cli.Command{
		{
			Name:   "update",
			Usage:  "Downloads the latest location data and updates your places",
			Action: placesUpdateAction,
		},
	},
}

// placesUpdateAction fetches updated location data.
func placesUpdateAction(ctx *cli.Context) error {
	// Load config.
	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	if !conf.Sponsor() && !conf.Test() {
		log.Errorf(config.MsgSponsorCommand)
		return nil
	}

	confirmPrompt := promptui.Prompt{
		Label:     "Interrupting the update may result in inconsistent data. Proceed?",
		IsConfirm: true,
	}

	// Abort?
	if _, err := confirmPrompt.Run(); err != nil {
		return nil
	}

	start := time.Now()

	// Run places worker.
	if w := service.Places(); w != nil {
		_, err := w.Start()

		if err != nil {
			return err
		}
	}

	// Run moments worker.
	if w := service.Moments(); w != nil {
		err := w.Start()

		if err != nil {
			return err
		}
	}

	// Hide missing album contents.
	if err := query.UpdateMissingAlbumEntries(); err != nil {
		log.Errorf("index: %s (update album entries)", err)
	}

	// Update precalculated photo and file counts.
	if err := entity.UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	// Update album, subject, and label cover thumbs.
	if err := query.UpdateCovers(); err != nil {
		log.Warnf("index: %s (update covers)", err)
	}

	log.Infof("completed in %s", time.Since(start))

	conf.Shutdown()

	return nil
}
