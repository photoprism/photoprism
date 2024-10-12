package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// UsersListCommand configures the command name, flags, and action.
var UsersListCommand = cli.Command{
	Name:   "ls",
	Usage:  "Lists registered user accounts",
	Flags:  append(report.CliFlags, CountFlag, UsersLoginFlag, UsersCreatedFlag, UsersDeletedFlag),
	Action: usersListAction,
}

// usersListAction displays existing user accounts.
func usersListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		var rows [][]string

		cols := []string{"UID", "Username", "Role", "Authentication", "Super Admin", "Web Login", "WebDAV"}

		if ctx.Bool("login") {
			cols = append(cols, "Last Login")
		}

		if ctx.Bool("created") {
			cols = append(cols, "Created At")
		}

		if ctx.Bool("deleted") {
			cols = append(cols, "Deleted At")
		}

		// Fetch users from database.
		users, err := query.Users(ctx.Int("n"), 0, "", ctx.Args().First(), ctx.Bool("deleted"))

		if err != nil {
			return err
		}

		// Show log message.
		log.Infof("found %s", english.Plural(len(users), "user", "users"))

		if len(users) == 0 {
			return nil
		}

		rows = make([][]string, len(users))

		// Display report.
		for i, user := range users {
			rows[i] = []string{
				user.GetUID(),
				user.Username(),
				user.AclRole().Pretty(),
				user.AuthInfo(),
				report.Bool(user.SuperAdmin, report.Yes, report.No),
				report.Bool(user.CanLogIn(), report.Enabled, report.Disabled),
				report.Bool(user.CanUseWebDAV(), report.Enabled, report.Disabled),
			}

			if ctx.Bool("login") {
				rows[i] = append(rows[i], report.DateTime(user.LoginAt))
			}

			if ctx.Bool("created") {
				rows[i] = append(rows[i], report.DateTime(&user.CreatedAt))
			}

			if ctx.Bool("deleted") {
				rows[i] = append(rows[i], report.DateTime(user.DeletedAt))
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
