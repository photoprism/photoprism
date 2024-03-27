package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/acl"
)

// Usage hints for the user management subcommands.
const (
	UserNameUsage     = "full `NAME` for display in the interface"
	UserEmailUsage    = "unique `EMAIL` address of the user"
	UserPasswordUsage = "`PASSWORD` for local authentication"
	UserRoleUsage     = "user role `NAME` (leave blank for default)"
	UserAdminUsage    = "make user super admin with full access"
	UserNoLoginUsage  = "disable login on the web interface"
	UserWebDAVUsage   = "allow to sync files via WebDAV"
	UserDisable2FA    = "deactivate two-factor authentication"
)

// UsersCommands configures the user management subcommands.
var UsersCommands = cli.Command{
	Name:    "users",
	Aliases: []string{"user"},
	Usage:   "User management subcommands",
	Subcommands: []cli.Command{
		UsersListCommand,
		UsersLegacyCommand,
		UsersAddCommand,
		UsersShowCommand,
		UsersModCommand,
		UsersRemoveCommand,
		UsersResetCommand,
	},
}

// UserFlags specifies the add and modify user command flags.
var UserFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "name, n",
		Usage: UserNameUsage,
	},
	cli.StringFlag{
		Name:  "email, m",
		Usage: UserEmailUsage,
	},
	cli.StringFlag{
		Name:  "password, p",
		Usage: UserPasswordUsage,
	},
	cli.StringFlag{
		Name:  "role, r",
		Usage: UserRoleUsage,
		Value: acl.RoleAdmin.String(),
	},
	cli.BoolFlag{
		Name:  "superadmin, s",
		Usage: UserAdminUsage,
	},
	cli.BoolFlag{
		Name:  "no-login, l",
		Usage: UserNoLoginUsage,
	},
	cli.BoolFlag{
		Name:  "webdav, w",
		Usage: UserWebDAVUsage,
	},
}
