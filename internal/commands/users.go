package commands

import (
	"github.com/urfave/cli/v2"

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
var UsersCommands = &cli.Command{
	Name:    "users",
	Aliases: []string{"user"},
	Usage:   "User management subcommands",
	Subcommands: []*cli.Command{
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
	&cli.StringFlag{
		Name:    "name",
		Aliases: []string{"n"},
		Usage:   UserNameUsage,
	},
	&cli.StringFlag{
		Name:    "email",
		Aliases: []string{"m"},
		Usage:   UserEmailUsage,
	},
	&cli.StringFlag{
		Name:    "password",
		Aliases: []string{"p"},
		Usage:   UserPasswordUsage,
	},
	&cli.StringFlag{
		Name:    "role",
		Aliases: []string{"r"},
		Usage:   UserRoleUsage,
		Value:   acl.RoleAdmin.String(),
	},
	&cli.StringFlag{
		Name:    "auth",
		Aliases: []string{"A"},
		Usage:   UserAuthUsage,
		Value:   authn.ProviderDefault.String(),
	},
	&cli.StringFlag{
		Name:  "auth-id",
		Usage: UserAuthIDUsage,
		Value: "",
	},
	&cli.BoolFlag{
		Name:    "superadmin",
		Aliases: []string{"s"},
		Usage:   UserAdminUsage,
	},
	&cli.BoolFlag{
		Name:    "no-login",
		Aliases: []string{"l"},
		Usage:   UserNoLoginUsage,
	},
	&cli.BoolFlag{
		Name:    "webdav",
		Aliases: []string{"w"},
		Usage:   UserWebDAVUsage,
	},
}

// UserTokensFlag is a CLI flag for showing the security tokens in reports.
var UserTokensFlag = &cli.BoolFlag{
	Name:  "tokens",
	Usage: "show user preview and download tokens",
}

// UsersLoginFlag is a CLI flag for showing the last login timestamp in reports.
var UsersLoginFlag = &cli.BoolFlag{
	Name:    "login",
	Aliases: []string{"l"},
	Usage:   "show date and time of last login",
}

// UsersCreatedFlag is a CLI flag for showing the account creation timestamp in reports.
var UsersCreatedFlag = &cli.BoolFlag{
	Name:    "created",
	Aliases: []string{"a"},
	Usage:   "show account creation timestamp",
}

// UsersDeletedFlag is a CLI flag for showing deleted user accounts in reports.
var UsersDeletedFlag = &cli.BoolFlag{
	Name:    "deleted",
	Aliases: []string{"r"},
	Usage:   "show deleted user accounts",
}
