package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

// Usage hints for the user management subcommands.
const (
	UserNameUsage     = "full `NAME` for display in the interface"
	UserEmailUsage    = "unique `EMAIL` address of the user"
	UserPasswordUsage = "`PASSWORD` for local authentication (8-72 characters)"
	UserRoleUsage     = "user account `ROLE` (admin or guest)"
	UserAuthUsage     = "authentication `PROVIDER` (default, local, oidc or none)"
	UserAuthIDUsage   = "authentication `ID` e.g. Subject ID or Distinguished Name (DN)"
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
	cli.StringFlag{
		Name:  "auth, A",
		Usage: UserAuthUsage,
		Value: authn.ProviderDefault.String(),
	},
	cli.StringFlag{
		Name:  "auth-id",
		Usage: UserAuthIDUsage,
		Value: "",
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
