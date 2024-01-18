package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// Usage hints for the client management subcommands.
const (
	ClientNameUsage        = "`CLIENT` name to help identify the application"
	ClientRoleUsage        = "client authorization `ROLE`"
	ClientAuthScope        = "client authorization `SCOPES` e.g. \"metrics\" or \"photos albums\" (\"*\" to allow all)"
	ClientAuthMethod       = "client authentication `METHOD`"
	ClientAuthExpires      = "authentication `LIFETIME` in seconds, after which a new access token must be requested (-1 to disable the limit)"
	ClientAuthTokens       = "maximum 'NUMBER' of access tokens that the client can request (-1 to disable the limit)"
	ClientRegenerateSecret = "generate a new client secret and display it"
	ClientEnable           = "enable client authentication if disabled"
	ClientDisable          = "disable client authentication"
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
		Name:   "method, m",
		Usage:  ClientAuthMethod,
		Value:  authn.MethodOAuth2.String(),
		Hidden: true,
	},
	cli.Int64Flag{
		Name:  "expires, e",
		Usage: ClientAuthExpires,
		Value: entity.UnixDay,
	},
	cli.Int64Flag{
		Name:  "tokens, t",
		Usage: ClientAuthTokens,
		Value: 10,
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
		Name:   "method, m",
		Usage:  ClientAuthMethod,
		Value:  authn.MethodOAuth2.String(),
		Hidden: true,
	},
	cli.Int64Flag{
		Name:  "expires, e",
		Usage: ClientAuthExpires,
	},
	cli.Int64Flag{
		Name:  "tokens, t",
		Usage: ClientAuthTokens,
	},
	cli.BoolFlag{
		Name:  "regenerate-secret, r",
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
