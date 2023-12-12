package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/report"
)

// UsersLegacyCommand configures the command name, flags, and action.
var UsersLegacyCommand = cli.Command{
	Name:      "legacy",
	Usage:     "Lists legacy user accounts",
	ArgsUsage: "[search]",
	Flags:     report.CliFlags,
	Action:    usersLegacyAction,
}

// usersLegacyAction displays legacy user accounts.
func usersLegacyAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		cols := []string{"ID", "UID", "Name", "User", "Email", "Admin", "Created At"}

		// Fetch users from database.
		users := entity.FindLegacyUsers(ctx.Args().First())
		rows := make([][]string, len(users))

		// Show log message.
		log.Infof("found %s", english.Plural(len(users), "legacy user", "legacy users"))

		// Display report.
		for i, user := range users {
			rows[i] = []string{
				fmt.Sprintf("%d", user.ID),
				user.UserUID,
				user.FullName,
				user.UserName,
				user.PrimaryEmail,
				report.Bool(user.Admin(), report.Yes, report.No),
				user.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
