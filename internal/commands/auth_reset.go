package commands

import (
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// AuthResetDescription explains the effect of the auth reset command.
const AuthResetDescription = "This command recreates the auth_sessions, auth_grants, and auth_codes database tables so that they are compatible with the current version. As a result, all users and clients must re-authenticate. Note that any client access tokens and app passwords that users may have created are also deleted and must be recreated, and connected apps must be connected again."

// AuthResetCommand configures the command name, flags, and action.
var AuthResetCommand = &cli.Command{
	Name:        "reset",
	Usage:       "Resets the authentication of all users and clients",
	Description: AuthResetDescription,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "shows trace logs for debugging",
		},
		YesFlag(),
	},
	Action: authResetAction,
}

// authResetAction removes all sessions and resets the related database table to a clean state.
func authResetAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if proceed, err := ConfirmAction(ctx.Bool("yes"), "Remove all sessions, access tokens, and app passwords?"); err != nil {
			return err
		} else if !proceed {
			log.Infof("no sessions were removed")
			return nil
		}

		if ctx.Bool("trace") {
			log.SetLevel(logrus.TraceLevel)
			log.Infoln("reset: enabled trace mode")
		}

		db := conf.Db()

		// Drop and re-create the tables with sessions, app grants, and pending authorization codes.
		if err := db.Migrator().DropTable(&entity.Session{}); err != nil {
			return cli.Exit(err, 1)
		} else if err = entity.CreateTable(db, &entity.Session{}); err != nil {
			return cli.Exit(err, 1)
		} else if err = ResetAppGrants(db); err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("all sessions have been removed")

		return nil
	})
}
