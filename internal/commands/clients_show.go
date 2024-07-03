package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClientsShowCommand configures the command name, flags, and action.
var ClientsShowCommand = cli.Command{
	Name:      "show",
	Usage:     "Shows client configuration details",
	ArgsUsage: "[client id]",
	Flags:     report.CliFlags,
	Action:    clientsShowAction,
}

// clientsShowAction displays the current client application settings
func clientsShowAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		id := clean.UID(ctx.Args().First())

		// Name or UID provided?
		if id == "" {
			return cli.ShowSubcommandHelp(ctx)
		}

		// Find client record.
		var m *entity.Client

		m = entity.FindClientByUID(id)

		if m == nil {
			return fmt.Errorf("client %s not found", clean.Log(id))
		}

		// Get client information.
		rows, cols := m.Report(true)

		// Sort values by name.
		report.Sort(rows)

		// Show client information.
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
