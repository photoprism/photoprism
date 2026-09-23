package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UsersRemoveCommand configures the command name, flags, and action.
var UsersRemoveCommand = &cli.Command{
	Name:      "rm",
	Usage:     "Deletes a registered user account",
	ArgsUsage: "[username]",
	Flags: []cli.Flag{
		YesFlag(),
		DeprecatedForceFlag(),
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
			return cli.Exit(fmt.Errorf("user %s not found", clean.LogQuote(id)), 3)
		} else if m.IsDeleted() {
			return cli.Exit(fmt.Errorf("user %s has already been deleted", clean.LogQuote(id)), 3)
		}

		if proceed, err := ConfirmAction(ctx.Bool("yes") || ctx.Bool("force"), fmt.Sprintf("Delete user %s?", m.String())); err != nil {
			return err
		} else if !proceed {
			log.Infof("user %s was not deleted", m.String())
			return nil
		}

		if err := m.Delete(); err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("user %s has been deleted", m.String())

		return nil
	})
}
