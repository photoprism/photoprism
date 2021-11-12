package commands

import (
	"context"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
)

// PlacesCommand registers the places subcommands.
var PlacesCommand = cli.Command{
	Name:  "places",
	Usage: "Location information subcommands",
	Subcommands: []cli.Command{
		{
			Name:   "update",
			Usage:  "Fetches updated location data",
			Action: placesUpdateAction,
		},
	},
}

// placesUpdateAction fetches updated location data.
func placesUpdateAction(ctx *cli.Context) error {
	start := time.Now()

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

		log.Infof("updated %s in %s", english.Plural(len(updated), "location", "locations"), elapsed)
	}

	conf.Shutdown()

	return nil
}
