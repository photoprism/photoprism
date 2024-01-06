package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
)

// ClientsModCommand configures the command name, flags, and action.
var ClientsModCommand = cli.Command{
	Name:      "mod",
	Usage:     "Updates client application settings",
	ArgsUsage: "[id]",
	Flags:     ClientModFlags,
	Action:    clientsModAction,
}

// clientsModAction updates client application settings.
func clientsModAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		id := clean.UID(ctx.Args().First())

		// Name or UID provided?
		if id == "" {
			log.Infof("no valid client id specified")
			return cli.ShowSubcommandHelp(ctx)
		}

		// Find client record.
		var client *entity.Client

		client = entity.FindClient(id)

		if client == nil {
			return fmt.Errorf("client %s not found", clean.LogQuote(id))
		}

		if name := clean.Name(ctx.String("name")); name != "" {
			client.ClientName = name
		}

		if ctx.IsSet("method") {
			client.AuthMethod = authn.Method(ctx.String("method")).String()
		}

		if ctx.IsSet("scope") {
			client.SetScope(ctx.String("scope"))
		}

		if expires := ctx.Int64("expires"); expires != 0 {
			if expires > entity.UnixMonth {
				client.AuthExpires = entity.UnixMonth
			} else if expires > 0 {
				client.AuthExpires = expires
			} else if expires <= 0 {
				client.AuthExpires = entity.UnixHour
			}
		}

		if tokens := ctx.Int64("tokens"); tokens != 0 {
			if tokens > 2147483647 {
				client.AuthTokens = 2147483647
			} else if tokens > 0 {
				client.AuthTokens = tokens
			} else if tokens < 0 {
				client.AuthTokens = -1
			}
		}

		if ctx.IsSet("disable") && ctx.Bool("disable") {
			client.AuthEnabled = false
			log.Infof("disabled client authentication")
		} else if ctx.IsSet("enable") && ctx.Bool("enable") {
			client.AuthEnabled = true
			log.Warnf("enabled client authentication")
		}

		// Save changes.
		if err := client.Validate(); err != nil {
			return fmt.Errorf("invalid client settings: %s", err)
		} else if err = client.Save(); err != nil {
			return fmt.Errorf("failed to update client settings: %s", err)
		} else {
			log.Infof("client %s has been updated", clean.LogQuote(client.ClientName))
		}

		// Regenerate and display secret, if requested.
		if ctx.IsSet("regenerate-secret") && ctx.Bool("regenerate-secret") {
			if secret, err := client.NewSecret(); err != nil {
				// Failed to create client secret.
				return fmt.Errorf("failed to create client secret: %s", err)
			} else {
				// Show client authentication credentials.
				fmt.Printf("\nTHE FOLLOWING RANDOMLY GENERATED CLIENT ID AND SECRET ARE REQUIRED FOR AUTHENTICATION:\n")
				result := report.Credentials("Client ID", client.ClientUID, "Client Secret", secret)
				fmt.Printf("\n%s", result)
				fmt.Printf("\nPLEASE WRITE THE CREDENTIALS DOWN AND KEEP THEM IN A SAFE PLACE, AS THE SECRET CANNOT BE DISPLAYED AGAIN.\n\n")
			}
		}

		return nil
	})
}
