package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
)

// AuthShowCommand configures the command name, flags, and action.
var AuthShowCommand = cli.Command{
	Name:      "show",
	Usage:     "Shows detailed information about a session",
	ArgsUsage: "[identifier]",
	Flags:     report.CliFlags,
	Action:    authShowAction,
}

// authShowAction shows detailed session information.
func authShowAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		id := clean.ID(ctx.Args().First())

		// ID provided?
		if id == "" {
			return cli.ShowSubcommandHelp(ctx)
		}

		// Find session by name.
		sess, err := query.Session(id)

		if err != nil {
			return fmt.Errorf("session %s not found: %s", clean.LogQuote(id), err)
		}

		// Get session information.
		rows, cols := sess.Report(true)

		// Sort values by name.
		report.Sort(rows)

		// Show session information.
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
