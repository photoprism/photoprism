package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// AuthResetCommand configures the command name, flags, and action.
var AuthResetCommand = cli.Command{
	Name:  "reset",
	Usage: "Resets the authentication of all users and clients",
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
	Action: authResetAction,
}

// authResetAction removes all sessions and resets the related database table to a clean state.
func authResetAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		confirmed := ctx.Bool("yes")

		// Show prompt?
		if !confirmed {
			actionPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Remove all sessions and reset the database table to a clean state?"),
				IsConfirm: true,
			}

			if _, err := actionPrompt.Run(); err != nil {
				return nil
			}
		}

		if ctx.Bool("trace") {
			log.SetLevel(logrus.TraceLevel)
			log.Infoln("clear: enabled trace mode")
		}

		db := conf.Db()

		// Drop existing sessions table.
		if err := db.DropTableIfExists(entity.Session{}).Error; err != nil {
			return err
		}

		// Re-create auth_sessions.
		if err := db.CreateTable(entity.Session{}).Error; err != nil {
			return err
		}

		log.Infof("all sessions have been removed")

		return nil
	})
}
