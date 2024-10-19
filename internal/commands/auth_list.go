package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// AuthListCommand configures the command name, flags, and action.
var AuthListCommand = cli.Command{
	Name:      "ls",
	Usage:     "Lists currently authenticated users and clients",
	ArgsUsage: "[search]",
	Flags:     append(report.CliFlags, CountFlag, tokensFlag),
	Action:    authListAction,
}

// authListAction finds and displays sessions.
func authListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		var rows [][]string

		cols := []string{"Session ID", "User", "Authentication Method", "Client", "Scope", "Login IP", "Current IP", "Last Active", "Created At", "Expires At"}

		if ctx.Bool("tokens") {
			cols = append(cols, "Preview Token", "Download Token")
		}

		// Fetch sessions from database.
		results, err := query.Sessions(ctx.Int("n"), 0, "", ctx.Args().First())

		if err != nil {
			return err
		}

		// Show log message.
		log.Infof("found %s", english.Plural(len(results), "session", "sessions"))

		if len(results) == 0 {
			return nil
		}

		rows = make([][]string, len(results))

		// Display report.
		for i, res := range results {
			rows[i] = []string{
				res.RefID,
				res.UserInfo(),
				res.AuthInfo(),
				res.ClientInfo(),
				res.AuthScope,
				res.LoginIP,
				res.ClientIP,
				report.UnixTime(res.LastActive),
				report.DateTime(&res.CreatedAt),
				report.UnixTime(res.SessExpires),
			}

			if ctx.Bool("tokens") {
				rows[i] = append(rows[i], res.PreviewToken, res.DownloadToken)
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
