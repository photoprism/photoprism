package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

const UsersResetDescription = "This command recreates the session and user management database tables so that they are compatible with the current version. Should you experience login problems, for example after an upgrade from an earlier version or a development preview, we recommend that you first try the \"photoprism auth reset --yes\" command to see if it solves the issue. Note that any client access tokens and app passwords that users may have created are also deleted and must be recreated."

// UsersResetCommand configures the command name, flags, and action.
var UsersResetCommand = cli.Command{
	Name:        "reset",
	Usage:       "Removes all registered user accounts",
	Description: UsersResetDescription,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "trace, t",
			Usage: "show trace logs for debugging",
		},
		cli.BoolFlag{
			Name:  "yes, y",
			Usage: "assume \"yes\" and run non-interactively",
		},
	},
	Action: usersResetAction,
}

// usersResetAction deletes recreates the user management database tables.
func usersResetAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		confirmed := ctx.Bool("yes")

		// Show prompt?
		if !confirmed {
			actionPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Reset the user database to a clean state?"),
				IsConfirm: true,
			}

			if _, err := actionPrompt.Run(); err != nil {
				return nil
			}
		}

		if ctx.Bool("trace") {
			log.SetLevel(logrus.TraceLevel)
			log.Infoln("reset: enabled trace mode")
		}

		db := conf.Db()

		// Drop existing user management tables.
		if err := db.DropTableIfExists(entity.User{}, entity.UserDetails{}, entity.UserSettings{}, entity.UserShare{}, entity.Passcode{}, entity.Session{}).Error; err != nil {
			return err
		}

		// Re-create auth_users.
		if err := db.CreateTable(entity.User{}).Error; err != nil {
			return err
		}

		// Re-create auth_users_details.
		if err := db.CreateTable(entity.UserDetails{}).Error; err != nil {
			return err
		}

		// Re-create auth_users_settings.
		if err := db.CreateTable(entity.UserSettings{}).Error; err != nil {
			return err
		}

		// Re-create auth_users_shares.
		if err := db.CreateTable(entity.UserShare{}).Error; err != nil {
			return err
		}

		// Re-create passcodes.
		if err := db.CreateTable(entity.Passcode{}).Error; err != nil {
			return err
		}

		// Re-create auth_sessions.
		if err := db.CreateTable(entity.Session{}).Error; err != nil {
			return err
		}

		log.Infof("the user database has been recreated and is now in a clean state")

		return nil
	})
}
