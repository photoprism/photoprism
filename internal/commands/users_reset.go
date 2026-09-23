package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// UsersResetDescription explains the effect of the users reset command.
const UsersResetDescription = "This command recreates the session and user management database tables so that they are compatible with the current version. Should you experience login problems, for example after an upgrade from an earlier version or a development preview, we recommend that you first try the \"photoprism auth reset --yes\" command to see if it solves the issue. Note that any client access tokens and app passwords that users may have created are also deleted and must be recreated."

// UsersResetCommand configures the command name, flags, and action.
var UsersResetCommand = &cli.Command{
	Name:        "reset",
	Usage:       "Removes all registered user accounts",
	Description: UsersResetDescription,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "shows trace logs for debugging",
		},
		YesFlag(),
	},
	Action: usersResetAction,
}

// usersResetAction drops and recreates the user management database tables.
func usersResetAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if proceed, err := ConfirmAction(ctx.Bool("yes"), "Remove all user accounts, sessions, access tokens, app passwords, 2FA passcodes, and user shares?"); err != nil {
			return err
		} else if !proceed {
			log.Infof("no user accounts were removed")
			return nil
		}

		if ctx.Bool("trace") {
			log.SetLevel(logrus.TraceLevel)
			log.Infoln("reset: enabled trace mode")
		}

		db := conf.Db()

		// Drop existing user management tables.
		if err := db.DropTableIfExists(entity.User{}, entity.UserDetails{}, entity.UserSettings{}, entity.UserShare{}, entity.Passcode{}, entity.Session{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create auth_users.
		if err := db.CreateTable(entity.User{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create auth_users_details.
		if err := db.CreateTable(entity.UserDetails{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create auth_users_settings.
		if err := db.CreateTable(entity.UserSettings{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create auth_users_shares.
		if err := db.CreateTable(entity.UserShare{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create passcodes.
		if err := db.CreateTable(entity.Passcode{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		// Re-create auth_sessions.
		if err := db.CreateTable(entity.Session{}).Error; err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("the user database has been recreated and is now in a clean state")

		return nil
	})
}
