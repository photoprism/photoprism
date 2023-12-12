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
		cols := []string{"Client ID", "Client Name", "Authentication", "Scope", "User", "Enabled", "Access Token Expires", "Created At"}

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
			var userName string
			if client.UserUID == "" {
				userName = report.NotAssigned
			} else if client.UserName != "" {
				userName = client.UserName
			} else {
				userName = client.UserUID
			}

			var authExpires string
			if client.AuthExpires > 0 {
				authExpires = client.Expires().String()
			} else {
				authExpires = report.Never
			}

			if client.AuthTokens > 0 {
				authExpires = fmt.Sprintf("%s, max %d tokens", authExpires, client.AuthTokens)
			}

			rows[i] = []string{
				client.UID(),
				client.ClientName,
				client.AuthMethod,
				client.AuthScope,
				userName,
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
