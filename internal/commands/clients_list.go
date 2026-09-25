package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClientsListCommand configures the command name, flags, and action.
var ClientsListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists registered client applications",
	ArgsUsage: "[search]",
	Flags: append(report.CliFlags, CountFlag, &cli.BoolFlag{
		Name:  "deleted",
		Usage: ClientDeleted,
	}),
	Action: clientsListAction,
}

// clientsListAction lists registered client applications
func clientsListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		deleted := ctx.Bool("deleted")

		cols := []string{"Client ID", "Name", "Authentication Method", "User", "Role", "Scope", "Enabled", "Access Token Lifetime", "Created At"}

		if deleted {
			cols = append(cols, "Deleted At")
		}

		// Fetch clients from database.
		clients, err := query.Clients(ctx.Int("count"), 0, "", ctx.Args().First(), deleted)

		if err != nil {
			return err
		}

		rows := make([][]string, len(clients))

		if len(clients) == 0 {
			if deleted {
				log.Warnf("no deleted clients found")
			} else {
				log.Warnf("no clients registered")
			}

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

			// A deleted client resolves as role none for access checks, so the listing
			// reports the role it was configured with rather than that outcome.
			role := client.AclRole().String()

			if deleted {
				role = acl.ClientRoles[clean.Role(client.ClientRole)].String()
			}

			rows[i] = []string{
				client.GetUID(),
				client.Name(),
				client.AuthInfo(),
				client.UserInfo(),
				role,
				client.Scope(),
				report.Bool(client.AuthEnabled, report.Yes, report.No),
				authExpires,
				client.CreatedAt.Format("2006-01-02 15:04:05"),
			}

			if deleted {
				deletedAt := ""

				if client.DeletedAt != nil {
					deletedAt = client.DeletedAt.Format("2006-01-02 15:04:05")
				}

				rows[i] = append(rows[i], deletedAt)
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
