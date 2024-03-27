package commands

import (
	"context"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/query"
)

// PlacesCommands configures the command name, flags, and action.
var PlacesCommands = cli.Command{
	Name:  "places",
	Usage: "Maps and location details subcommands",
	Subcommands: []cli.Command{
		{
			Name:  "update",
			Usage: "Retrieves updated location details",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "yes, y",
					Hidden: true,
					Usage:  "assume \"yes\" as answer to all prompts and run non-interactively",
				},
			},
			Action: placesUpdateAction,
		},
	},
}

// placesUpdateAction fetches updated location data.
func placesUpdateAction(ctx *cli.Context) error {
	// Load config.
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	// Show info to non-members.
	if !conf.Sponsor() && !conf.Test() {
		log.Errorf("Since running this command puts additional load on our infrastructure, we can unfortunately only offer it to members at this time.")
		return nil
	}

	// Initialize database connection.
	conf.InitDb()
	defer conf.Shutdown()

	if !ctx.Bool("yes") {
		confirmPrompt := promptui.Prompt{
			Label:     "Interrupting the update may result in inconsistent location details. Proceed?",
			IsConfirm: true,
		}

		// Abort?
		if _, err := confirmPrompt.Run(); err != nil {
			return nil
		}
	}

	start := time.Now()

	// Run places worker.
	if w := get.Places(); w != nil {
		_, err := w.Start()

		if err != nil {
			return err
		}
	}

	// Run moments worker.
	if w := get.Moments(); w != nil {
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

	return nil
}
