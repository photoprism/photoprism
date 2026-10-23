package commands

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// Usage hints for the client management subcommands.
const (
	ClientIdUsage          = "static client `UID` for test purposes"
	ClientSecretUsage      = "static client `SECRET` for test purposes"
	ClientNameUsage        = "`CLIENT` name to help identify the application"
	ClientAuthScope        = "client authorization `SCOPE` as space-separated resources, or '*' for full access"
	ClientAuthProvider     = "client authentication `PROVIDER`"
	ClientRedirectURIUsage = "redirect `URI` of an app that users connect with OAuth sign-in and consent (repeat for each URI)"
	ClientAuthMethod       = "client authentication `METHOD`"
	ClientAuthExpires      = "access token `LIFETIME` in seconds, after which a new token must be requested"
	ClientAuthTokens       = "maximum `NUMBER` of access tokens that the client can request (-1 to disable the limit)"
	ClientRegenerateSecret = "set a new randomly generated client secret and revoke existing access tokens"
	ClientEnable           = "enable client authentication if disabled"
	ClientDisable          = "disable client authentication"
	ClientRestore          = "bring back a deleted client without asking for confirmation; it needs a new secret"
	ClientPurge            = "remove the record permanently and release the identifiers it reserves"
	ClientDeleted          = "list deleted clients instead of current ones"
	ClientSecretInfo       = "\nPLEASE WRITE DOWN THE %s CLIENT SECRET, AS YOU WILL NOT BE ABLE TO SEE IT AGAIN:" //nolint:gosec // informational message only
)

var (
	// ClientRoleUsage describes the assignable client roles for CLI help.
	ClientRoleUsage = fmt.Sprintf("client authorization `ROLE`, e.g. %s", acl.ClientRoles.CliUsageString())
)

// ClientsCommands configures the client application subcommands.
var ClientsCommands = &cli.Command{
	Name:    "clients",
	Aliases: []string{"client"},
	Usage:   "Client credentials subcommands",
	Subcommands: []*cli.Command{
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
	&cli.StringFlag{
		Name:   "id",
		Usage:  ClientIdUsage,
		Hidden: true,
	},
	&cli.StringFlag{
		Name:    "name",
		Aliases: []string{"n"},
		Usage:   ClientNameUsage,
	},
	&cli.StringFlag{
		Name:    "role",
		Aliases: []string{"r"},
		Usage:   ClientRoleUsage,
		Value:   acl.RoleClient.String(),
	},
	ScopeFlag(ClientAuthScope),
	&cli.StringSliceFlag{
		Name:  "redirect-uri",
		Usage: ClientRedirectURIUsage,
	},
	&cli.StringFlag{
		Name:    "provider",
		Aliases: []string{"p"},
		Usage:   ClientAuthProvider,
		Value:   authn.ProviderClient.String(),
		Hidden:  true,
	},
	&cli.StringFlag{
		Name:    "method",
		Aliases: []string{"m"},
		Usage:   ClientAuthMethod,
		Value:   authn.MethodOAuth2.String(),
		Hidden:  true,
	},
	&cli.Int64Flag{
		Name:    "expires",
		Aliases: []string{"e"},
		Usage:   ClientAuthExpires,
		Value:   unix.Day,
	},
	&cli.Int64Flag{
		Name:    "tokens",
		Aliases: []string{"t"},
		Usage:   ClientAuthTokens,
		Value:   10,
	},
	&cli.StringFlag{
		Name:   "secret",
		Usage:  ClientSecretUsage,
		Hidden: true,
	},
}

// ClientModFlags specifies the "photoprism client mod" command flags.
var ClientModFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "name",
		Aliases: []string{"n"},
		Usage:   ClientNameUsage,
	},
	&cli.StringFlag{
		Name:    "role",
		Aliases: []string{"r"},
		Usage:   ClientRoleUsage,
	},
	ScopeFlag(ClientAuthScope),
	&cli.StringSliceFlag{
		Name:  "redirect-uri",
		Usage: ClientRedirectURIUsage,
	},
	&cli.StringFlag{
		Name:    "provider",
		Aliases: []string{"p"},
		Usage:   ClientAuthProvider,
		Value:   authn.ProviderClient.String(),
		Hidden:  true,
	},
	&cli.StringFlag{
		Name:    "method",
		Aliases: []string{"m"},
		Usage:   ClientAuthMethod,
		Value:   authn.MethodOAuth2.String(),
		Hidden:  true,
	},
	&cli.Int64Flag{
		Name:    "expires",
		Aliases: []string{"e"},
		Usage:   ClientAuthExpires,
		Value:   unix.Day,
	},
	&cli.Int64Flag{
		Name:    "tokens",
		Aliases: []string{"t"},
		Usage:   ClientAuthTokens,
		Value:   10,
	},
	&cli.StringFlag{
		Name:   "secret",
		Usage:  ClientSecretUsage,
		Hidden: true,
	},
	&cli.BoolFlag{
		Name:  "regenerate",
		Usage: ClientRegenerateSecret,
	},
	&cli.BoolFlag{
		Name:  "enable",
		Usage: ClientEnable,
	},
	&cli.BoolFlag{
		Name:  "disable",
		Usage: ClientDisable,
	},
	&cli.BoolFlag{
		Name:  "restore",
		Usage: ClientRestore,
	},
}

// clientTokenInfo returns the access token lifetime and limit of an API client, and nothing for app clients,
// whose tokens are bounded by their grants.
func clientTokenInfo(client *entity.Client) string {
	if client == nil || client.IsApp() {
		return ""
	}

	var info string

	if client.AuthExpires > 0 {
		info = client.Expires().String()
	}

	if client.AuthTokens > 0 {
		if info != "" {
			info = fmt.Sprintf("%s; up to %s", info, english.Plural(int(client.Tokens()), "token", "tokens"))
		} else {
			info = fmt.Sprintf("up to %d tokens", client.AuthTokens)
		}
	}

	return info
}

// clientLastActive returns the time of the client's last activity in UTC, or nothing if it was never active.
func clientLastActive(client *entity.Client) string {
	if client == nil || client.LastActive <= 0 {
		return ""
	}

	return time.Unix(client.LastActive, 0).UTC().Format("2006-01-02 15:04:05")
}
