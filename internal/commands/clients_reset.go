package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// ClientsResetCommand configures the command name, flags, and action.
var ClientsResetCommand = cli.Command{
	Name:  "reset",
	Usage: "Removes all registered client applications",
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
	Action: clientsResetAction,
}

// clientsResetAction removes all registered client applications.
func clientsResetAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		confirmed := ctx.Bool("yes")

		// Show prompt?
		if !confirmed {
			actionPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Reset the client database to a clean state?"),
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

		// Drop existing auth_clients table.
		if err := db.DropTableIfExists(entity.Client{}).Error; err != nil {
			return err
		}

		// Re-create auth_clients.
		if err := db.CreateTable(entity.Client{}).Error; err != nil {
			return err
		}

		log.Infof("the client database has been recreated and is now in a clean state")

		return nil
	})
}
