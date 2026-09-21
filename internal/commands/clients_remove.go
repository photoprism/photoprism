package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClientsRemoveCommand configures the command name, flags, and action.
var ClientsRemoveCommand = &cli.Command{
	Name:      "rm",
	Usage:     "Deletes the specified client application",
	ArgsUsage: "[client id | node uuid]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "skips asking for confirmation",
		},
		&cli.BoolFlag{
			Name:  "purge",
			Usage: ClientPurge,
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
		m := entity.FindClient(id)

		purge := ctx.Bool("purge")

		if m == nil {
			return fmt.Errorf("client %s not found", clean.Log(id))
		} else if m.Deleted() && !purge {
			return fmt.Errorf("client %s has already been deleted", clean.Log(id))
		} else if purge && m.NodeUUID != "" {
			// A node UUID also names a provisioned database and cluster grants, and may be
			// held by more than one record, so releasing it goes through the cluster command.
			return fmt.Errorf("client %s registers cluster node %s, use \"cluster nodes rm --purge\" to remove it permanently", clean.Log(id), clean.Log(m.NodeUUID))
		}

		action := "Delete"

		if purge {
			action = "Permanently delete"
		}

		if !ctx.Bool("force") && !RunNonInteractively(false) {
			actionPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("%s client %s?", action, m.GetUID()),
				IsConfirm: true,
			}

			if _, err := actionPrompt.Run(); err != nil {
				log.Infof("client %s was not deleted", m.GetUID())
				return nil
			}
		}

		// A purge releases the identifiers the record reserves, so a later client can take
		// them; an ordinary delete keeps them reserved and can be undone with "clients mod".
		if purge {
			if err := m.Purge(); err != nil {
				return err
			}

			log.Infof("client %s has been permanently deleted", m.GetUID())

			return nil
		}

		if err := m.Delete(); err != nil {
			return err
		}

		log.Infof("client %s has been deleted", m.GetUID())

		return nil
	})
}
