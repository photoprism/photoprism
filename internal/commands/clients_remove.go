package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClientsRemoveCommand configures the command name, flags, and action.
var ClientsRemoveCommand = cli.Command{
	Name:      "rm",
	Usage:     "Deletes the specified client application",
	ArgsUsage: "[client id]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "don't ask for confirmation",
		},
	},
	Action: clientsRemoveAction,
}

// clientsRemoveAction deletes a registered client application
func clientsRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		id := clean.UID(ctx.Args().First())

		// Name or UID provided?
		if id == "" {
			log.Infof("no valid client id specified")
			return cli.ShowSubcommandHelp(ctx)
		}

		// Find client record.
		var m *entity.Client

		m = entity.FindClientByUID(id)

		if m == nil {
			return fmt.Errorf("client %s not found", clean.Log(id))
		} else if m.Deleted() {
			return fmt.Errorf("client %s has already been deleted", clean.Log(id))
		}

		if !ctx.Bool("force") {
			actionPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Delete client %s?", m.UID()),
				IsConfirm: true,
			}

			if _, err := actionPrompt.Run(); err != nil {
				log.Infof("client %s was not deleted", m.UID())
				return nil
			}
		}

		if err := m.Delete(); err != nil {
			return err
		}

		log.Infof("client %s has been deleted", m.UID())

		return nil
	})
}
