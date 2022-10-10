package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
)

// UsersShowCommand configures the command name, flags, and action.
var UsersShowCommand = cli.Command{
	Name:      "show",
	Usage:     "Shows user account information",
	ArgsUsage: "[username]",
	Flags:     report.CliFlags,
	Action:    usersShowAction,
}

// usersShowAction Shows user account details.
func usersShowAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
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

		// Get user information.
		rows, cols := user.Report(true)

		// Sort values by name.
		report.Sort(rows)

		// Show user information.
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
