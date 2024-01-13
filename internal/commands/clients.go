package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/authn"
)

// Usage hints for the client management subcommands.
const (
	ClientNameUsage        = "arbitrary name to help identify the `CLIENT` application"
	ClientUserName         = "`USERNAME` of the account the client application belongs to (leave empty for none)"
	ClientAuthScope        = "authorization `SCOPE` of the client e.g. \"metrics\" or \"photos albums\" (\"*\" to allow all scopes)"
	ClientAuthMethod       = "supported authentication `METHOD` for the client application"
	ClientAuthExpires      = "access token lifetime in `SECONDS`, after which a new token must be created by the client (-1 to disable)"
	ClientAuthTokens       = "maximum `NUMBER` of access tokens the client can create (-1 to disable)"
	ClientRegenerateSecret = "generate a new client secret and display it"
	ClientDisable          = "deactivate authentication with this client"
	ClientEnable           = "re-enable client authentication"
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
		Name:  "user, u",
		Usage: ClientUserName,
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
}

// ClientModFlags specifies the "photoprism client mod" command flags.
var ClientModFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "name, n",
		Usage: ClientNameUsage,
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
		Name:  "disable",
		Usage: ClientDisable,
	},
	cli.BoolFlag{
		Name:  "enable",
		Usage: ClientEnable,
	},
}
