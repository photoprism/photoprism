package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UsersRemoveCommand configures the command name, flags, and action.
var UsersRemoveCommand = cli.Command{
	Name:      "rm",
	Usage:     "Deletes a registered user account",
	ArgsUsage: "[username]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "don't ask for confirmation",
		},
	},
	Action: usersRemoveAction,
}

// usersRemoveAction deletes a user account.
func usersRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		id := clean.Username(ctx.Args().First())

		// Name or UID provided?
		if id == "" {
			return cli.ShowSubcommandHelp(ctx)
		}

		// Find user record.
		var m *entity.User

		if rnd.IsUID(id, entity.UserUID) {
			m = entity.FindUserByUID(id)
		} else {
			m = entity.FindUserByName(id)
		}

		if m == nil {
			return fmt.Errorf("user %s not found", clean.LogQuote(id))
		} else if m.IsDeleted() {
			return fmt.Errorf("user %s has already been deleted", clean.LogQuote(id))
		}

		if !ctx.Bool("force") {
			actionPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Delete user %s?", m.String()),
				IsConfirm: true,
			}

			if _, err := actionPrompt.Run(); err != nil {
				log.Infof("user %s was not deleted", m.String())
				return nil
			}
		}

		if err := m.Delete(); err != nil {
			return err
		}

		log.Infof("user %s has been deleted", m.String())

		return nil
	})
}
