package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/unix"
)

// Usage hints for the client management subcommands.
const (
	ClientIdUsage          = "static client `UID` for test purposes"
	ClientSecretUsage      = "static client `SECRET` for test purposes"
	ClientNameUsage        = "`CLIENT` name to help identify the application"
	ClientRoleUsage        = "client authorization `ROLE`"
	ClientAuthScope        = "client authorization `SCOPES` e.g. \"metrics\" or \"photos albums\" (\"*\" to allow all)"
	ClientAuthProvider     = "client authentication `PROVIDER`"
	ClientAuthMethod       = "client authentication `METHOD`"
	ClientAuthExpires      = "access token `LIFETIME` in seconds, after which a new token must be requested"
	ClientAuthTokens       = "maximum `NUMBER` of access tokens that the client can request (-1 to disable the limit)"
	ClientRegenerateSecret = "set a new randomly generated client secret"
	ClientEnable           = "enable client authentication if disabled"
	ClientDisable          = "disable client authentication"
	ClientSecretInfo       = "\nPLEASE WRITE DOWN THE %s CLIENT SECRET, AS YOU WILL NOT BE ABLE TO SEE IT AGAIN:\n"
)

// ClientsCommands configures the client application subcommands.
var ClientsCommands = cli.Command{
	Name:    "clients",
	Aliases: []string{"client"},
	Usage:   "Client credentials subcommands",
	Subcommands: []cli.Command{
		ClientsListCommand,
		ClientsAddCommand,
		ClientsShowCommand,
		ClientsModCommand,
		ClientsRemoveCommand,
		ClientsResetCommand,
	},
}

// ClientAddFlags specifies the "photoprism client add" command flags.
var ClientAddFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "id",
		Usage:  ClientIdUsage,
		Hidden: true,
	},
	cli.StringFlag{
		Name:  "name, n",
		Usage: ClientNameUsage,
	},
	cli.StringFlag{
		Name:  "role, r",
		Usage: ClientRoleUsage,
		Value: acl.RoleClient.String(),
	},
	cli.StringFlag{
		Name:  "scope, s",
		Usage: ClientAuthScope,
	},
	cli.StringFlag{
		Name:   "provider, p",
		Usage:  ClientAuthProvider,
		Value:  authn.ProviderClient.String(),
		Hidden: true,
	},
	cli.StringFlag{
		Name:   "method, m",
		Usage:  ClientAuthMethod,
		Value:  authn.MethodOAuth2.String(),
		Hidden: true,
	},
	cli.Int64Flag{
		Name:  "expires, e",
		Usage: ClientAuthExpires,
		Value: unix.Day,
	},
	cli.Int64Flag{
		Name:  "tokens, t",
		Usage: ClientAuthTokens,
		Value: 10,
	},
	cli.StringFlag{
		Name:   "secret",
		Usage:  ClientSecretUsage,
		Hidden: true,
	},
}

// ClientModFlags specifies the "photoprism client mod" command flags.
var ClientModFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "name, n",
		Usage: ClientNameUsage,
	},
	cli.StringFlag{
		Name:  "role, r",
		Usage: ClientRoleUsage,
		Value: acl.RoleClient.String(),
	},
	cli.StringFlag{
		Name:  "scope, s",
		Usage: ClientAuthScope,
	},
	cli.StringFlag{
		Name:   "provider, p",
		Usage:  ClientAuthProvider,
		Value:  authn.ProviderClient.String(),
		Hidden: true,
	},
	cli.StringFlag{
		Name:   "method, m",
		Usage:  ClientAuthMethod,
		Value:  authn.MethodOAuth2.String(),
		Hidden: true,
	},
	cli.Int64Flag{
		Name:  "expires, e",
		Usage: ClientAuthExpires,
		Value: unix.Day,
	},
	cli.Int64Flag{
		Name:  "tokens, t",
		Usage: ClientAuthTokens,
		Value: 10,
	},
	cli.StringFlag{
		Name:   "secret",
		Usage:  ClientSecretUsage,
		Hidden: true,
	},
	cli.BoolFlag{
		Name:  "regenerate",
		Usage: ClientRegenerateSecret,
	},
	cli.BoolFlag{
		Name:  "enable",
		Usage: ClientEnable,
	},
	cli.BoolFlag{
		Name:  "disable",
		Usage: ClientDisable,
	},
}
