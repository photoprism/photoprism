package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// ClientsResetCommand configures the command name, flags, and action.
var ClientsResetCommand = &cli.Command{
	Name:  "reset",
	Usage: "Removes all registered client applications",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "shows trace logs for debugging",
		},
		YesFlag(),
	},
	Action: clientsResetAction,
}

// clientsResetAction removes all registered client applications.
func clientsResetAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if proceed, err := ConfirmAction(ctx.Bool("yes"), "Remove all registered client applications?"); err != nil {
			return err
		} else if !proceed {
			log.Infof("no client applications were removed")
			return nil
		}

		if ctx.Bool("trace") {
			log.SetLevel(logrus.TraceLevel)
			log.Infoln("reset: enabled trace mode")
		}

		db := conf.Db()

		// Drop existing auth_clients table.
		if err := db.DropTableIfExists(entity.Client{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create auth_clients.
		if err := db.CreateTable(entity.Client{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("the client database has been recreated and is now in a clean state")

		return nil
	})
}
