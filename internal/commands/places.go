package commands

import (
	"context"
	"time"

	"github.com/manifoldco/promptui"

	"github.com/dustin/go-humanize/english"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
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
	start := time.Now()

	confirmPrompt := promptui.Prompt{
		Label:     "Interrupting the update may result in inconsistent data. Proceed?",
		IsConfirm: true,
	}

	if _, err := confirmPrompt.Run(); err != nil {
		// Abort.
		return nil
	}

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	w := service.Places()

	// Run places worker.
	if updated, err := w.Start(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("updated %s in %s", english.Plural(len(updated), "place", "places"), elapsed)
	}

	conf.Shutdown()

	return nil
}
