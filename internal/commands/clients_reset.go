package commands

import (
	"github.com/dustin/go-humanize/english"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// ClientsResetDescription explains the effect of the clients reset command.
const ClientsResetDescription = "This command recreates the auth_clients database table so that it is compatible with the current version. As a result, all registered client applications are removed, including those registered to users, and the access tokens issued to them are deleted. App passwords are not affected."

// ClientsResetCommand configures the command name, flags, and action.
var ClientsResetCommand = &cli.Command{
	Name:        "reset",
	Usage:       "Removes all registered client applications",
	Description: ClientsResetDescription,
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
		if proceed, err := ConfirmAction(ctx.Bool("yes"), "Remove all registered client applications and the access tokens issued to them?"); err != nil {
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

		// Delete the sessions of the removed clients once the table has been recreated.
		if db.HasTable(entity.Session{}) {
			res := db.Where("client_uid <> ''").Delete(&entity.Session{})

			if res.Error != nil {
				return cli.Exit(res.Error, 1)
			}

			log.Infof("deleted %s", english.Plural(int(res.RowsAffected), "client session", "client sessions"))
		}

		log.Infof("the client database has been recreated and is now in a clean state")

		return nil
	})
}
