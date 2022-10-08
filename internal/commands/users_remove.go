package commands

import (
	"errors"
	"fmt"

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

		username := clean.Username(ctx.Args().First())

		// Username provided?
		if username == "" {
			return cli.ShowSubcommandHelp(ctx)
		}

		actionPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Remove user %s?", clean.LogQuote(username)),
			IsConfirm: true,
		}

		if _, err := actionPrompt.Run(); err == nil {
			if m := entity.FindUserByName(username); m == nil {
				return errors.New("user not found")
			} else if err := m.Delete(); err != nil {
				return err
			} else {
				log.Infof("user %s has been removed", clean.LogQuote(username))
			}
		} else {
			log.Infof("user %s was not removed", clean.LogQuote(username))
		}

		return nil
	})
}
