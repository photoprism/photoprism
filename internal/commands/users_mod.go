package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// UsersModCommand configures the command name, flags, and action.
var UsersModCommand = cli.Command{
	Name:      "mod",
	Usage:     "Modifies an existing user account",
	ArgsUsage: "[username]",
	Flags:     UserFlags,
	Action:    usersModAction,
}

// usersModAction modifies an existing user account.
func usersModAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		username := clean.Username(ctx.Args().First())

		// Username provided?
		if username == "" {
			return cli.ShowSubcommandHelp(ctx)
		}

		// Find user by name.
		user := entity.FindUserByName(username)

		if user == nil {
			return fmt.Errorf("user %s not found", clean.LogQuote(username))
		}

		// Set values.
		if err := user.SetValuesFromCli(ctx); err != nil {
			return err
		}

		// Change password?
		if val := clean.Password(ctx.String("password")); ctx.IsSet("password") && val != "" {
			err := user.SetPassword(val)

			if err != nil {
				return err
			}

			log.Warnf("password has been changed")
		}

		// Save values.
		if err := user.Save(); err != nil {
			return err
		}

		log.Infof("user %s has been updated", clean.LogQuote(user.Name()))

		return nil
	})
}
