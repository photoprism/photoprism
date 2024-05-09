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
	Usage: "Maps and location information subcommands",
	Subcommands: []cli.Command{
		{
			Name:        "update",
			Usage:       "Updates location information",
			Description: "Updates missing location information only if used without the --force flag.",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "forces the location of all pictures to be updated",
				},
				cli.BoolFlag{
					Name:  "yes, y",
					Usage: "assume \"yes\" and run non-interactively",
				},
			},
			Action: placesUpdateAction,
		},
	},
}

// placesUpdateAction updates the location information.
func placesUpdateAction(ctx *cli.Context) error {
	// Load config.
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	// Force update of all locations?
	force := ctx.Bool("force")

	// Show info in case the force option is used without support.
	if force && !conf.Sponsor() && !conf.Test() {
		log.Errorf("Since updating the location details of all pictures puts a high load on our infrastructure, this option cannot be used with our Community Edition.")
		return nil
	}

	// Initialize database connection.
	conf.InitDb()
	defer conf.Shutdown()

	if !ctx.Bool("yes") {
		confirmPrompt := promptui.Prompt{
			Label:     "Interrupting the update may lead to inconsistent location information. Continue?",
			IsConfirm: true,
		}

		// Abort?
		if _, confirmErr := confirmPrompt.Run(); confirmErr != nil {
			return nil
		}
	}

	start := time.Now()

	// Run places worker.
	if w := get.Places(); w != nil {
		_, err = w.Start(force)

		if err != nil {
			return err
		}
	}

	// Run moments worker.
	if w := get.Moments(); w != nil {
		err = w.Start()

		if err != nil {
			return err
		}
	}

	// Hide missing album contents.
	if err = query.UpdateMissingAlbumEntries(); err != nil {
		log.Errorf("index: %s (update album entries)", err)
	}

	// Update precalculated photo and file counts.
	if err = entity.UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	// Update album, subject, and label cover thumbs.
	if err = query.UpdateCovers(); err != nil {
		log.Warnf("index: %s (update covers)", err)
	}

	log.Infof("completed in %s", time.Since(start))

	return nil
}
