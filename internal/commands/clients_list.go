package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/report"
)

// ClientsListCommand configures the command name, flags, and action.
var ClientsListCommand = cli.Command{
	Name:      "ls",
	Usage:     "Lists registered client applications",
	ArgsUsage: "[search]",
	Flags:     append(report.CliFlags, countFlag),
	Action:    clientsListAction,
}

// clientsListAction lists registered client applications
func clientsListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		cols := []string{"Client ID", "Name", "Authentication Method", "User", "Role", "Scope", "Enabled", "Access Token Lifetime", "Created At"}

		// Fetch clients from database.
		clients, err := query.Clients(ctx.Int("n"), 0, "", ctx.Args().First())

		if err != nil {
			return err
		}

		rows := make([][]string, len(clients))

		if len(clients) == 0 {
			log.Warnf("no clients registered")
			return nil
		}

		// Show log message.
		log.Infof("found %s", english.Plural(len(clients), "client", "clients"))

		// Display report.
		for i, client := range clients {
			var authExpires string

			if client.AuthExpires > 0 {
				authExpires = client.Expires().String()
			}

			if client.AuthTokens > 0 {
				if authExpires != "" {
					authExpires = fmt.Sprintf("%s; up to %s", authExpires, english.Plural(int(client.Tokens()), "token", "tokens"))
				} else {
					authExpires = fmt.Sprintf("up to %d tokens", client.AuthTokens)
				}
			}

			rows[i] = []string{
				client.GetUID(),
				client.Name(),
				client.AuthInfo(),
				client.UserInfo(),
				client.AclRole().String(),
				client.Scope(),
				report.Bool(client.AuthEnabled, report.Yes, report.No),
				authExpires,
				client.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
