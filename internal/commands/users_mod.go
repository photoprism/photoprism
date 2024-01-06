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

// UsersModCommand configures the command name, flags, and action.
var UsersModCommand = cli.Command{
	Name:      "mod",
	Usage:     "Changes user account settings",
	ArgsUsage: "[username]",
	Flags:     UserFlags,
	Action:    usersModAction,
}

// usersModAction modifies an existing user account.
func usersModAction(ctx *cli.Context) error {
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

		// Check if account exists but is deleted.
		if m.IsDeleted() {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Restore user %s?", m.String()),
				IsConfirm: true,
			}

			if _, err := prompt.Run(); err != nil {
				return fmt.Errorf("user already exists")
			}

			m.DeletedAt = nil
			log.Infof("user %s will be restored", m.String())
		}

		// Set values.
		if err := m.SetValuesFromCli(ctx); err != nil {
			return err
		}

		// Change password?
		if val := clean.Password(ctx.String("password")); ctx.IsSet("password") && val != "" {
			err := m.SetPassword(val)

			if err != nil {
				return err
			}

			log.Warnf("password has been changed")
		}

		// Save values.
		if err := m.Save(); err != nil {
			return err
		}

		log.Infof("user %s has been updated", m.String())

		return nil
	})
}
