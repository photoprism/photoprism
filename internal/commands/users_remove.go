package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// UsersRemoveCommand configures the command name, flags, and action.
var UsersRemoveCommand = cli.Command{
	Name:      "rm",
	Usage:     "Removes a user account",
	ArgsUsage: "[username]",
	Action:    usersRemoveAction,
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
		}

		actionPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Remove user %s?", m.String()),
			IsConfirm: true,
		}

		if _, err := actionPrompt.Run(); err == nil {
			if err = m.Delete(); err != nil {
				return err
			} else {
				log.Infof("user %s has been removed", m.String())
			}
		} else {
			log.Infof("user %s was not removed", m.String())
		}

		return nil
	})
}
