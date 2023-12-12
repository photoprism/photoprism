package commands

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/urfave/cli"
)

// Usage hints for the client management subcommands.
const (
	ClientNameUsage        = "arbitrary name to help identify the `CLIENT` application"
	ClientUserName         = "a `USERNAME` is only required if the client belongs to a specific account"
	ClientAuthMethod       = "supported authentication `METHOD` for the client application"
	ClientAuthScope        = "authorization `SCOPE` of the client e.g. \"metrics\" (\"*\" to allow all scopes)"
	ClientAuthExpires      = "access token expiration time in `SECONDS`, after which a new token must be created"
	ClientAuthTokens       = "maximum `NUMBER` of access tokens the client can create (-1 to disable the limit)"
	ClientRegenerateSecret = "generate a new client secret and display it"
	ClientDisable          = "deactivate authentication with this client"
	ClientEnable           = "re-enable client authentication"
)

// ClientsCommand configures the client application subcommands.
var ClientsCommand = cli.Command{
	Name:  "clients",
	Usage: "API authentication subcommands",
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
		Name:  "method, m",
		Usage: ClientAuthMethod,
		Value: authn.MethodOAuth2.String(),
	},
	cli.StringFlag{
		Name:  "scope, s",
		Usage: ClientAuthScope,
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
		Name:  "method, m",
		Usage: ClientAuthMethod,
		Value: authn.MethodOAuth2.String(),
	},
	cli.StringFlag{
		Name:  "scope, s",
		Usage: ClientAuthScope,
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
