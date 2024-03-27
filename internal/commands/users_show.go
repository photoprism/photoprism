package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UsersShowCommand configures the command name, flags, and action.
var UsersShowCommand = cli.Command{
	Name:      "show",
	Usage:     "Shows detailed account information",
	ArgsUsage: "[username]",
	Flags:     report.CliFlags,
	Action:    usersShowAction,
}

// usersShowAction Shows user account details.
func usersShowAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
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

		// Get user information.
		rows, cols := m.Report(true)

		// Sort values by name.
		report.Sort(rows)

		// Show user information.
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
