package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClientsModCommand configures the command name, flags, and action.
var ClientsModCommand = &cli.Command{
	Name:      "mod",
	Usage:     "Updates client application settings",
	ArgsUsage: "[client id | node uuid]",
	Flags:     ClientModFlags,
	Action:    clientsModAction,
}

// clientsModAction updates client application settings.
func clientsModAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		frm := form.ModClientFromCli(ctx)

		// Client UID or node UUID provided?
		id := clean.UID(ctx.Args().First())

		if id == "" {
			log.Infof("no valid client id specified")
			return ShowUsageError(ctx)
		}

		// Reject flags placed after the client id; the stdlib flag parser
		// would silently drop them and report "client updated" with no changes.
		if err := RejectTrailingFlags(ctx); err != nil {
			return err
		}

		// Find client record.
		client := entity.FindClient(id)

		if client == nil {
			return fmt.Errorf("client %s not found", clean.Log(id))
		}

		// Check if the client exists but has been deleted.
		if client.Deleted() {
			if !ctx.Bool("restore") && !RunNonInteractively(false) {
				prompt := promptui.Prompt{
					Label:     fmt.Sprintf("Restore client %s?", client.String()),
					IsConfirm: true,
				}

				if _, err := prompt.Run(); err != nil {
					return fmt.Errorf("client %s has been deleted", clean.Log(id))
				}
			} else if !ctx.Bool("restore") {
				return fmt.Errorf("client %s has been deleted, pass --restore to bring it back", clean.Log(id))
			}

			if err := client.Restore(); err != nil {
				return err
			}

			log.Infof("client %s has been restored", client.String())
		}

		// Update client from form values.
		client.SetFormValues(frm)

		if ctx.IsSet("enable") || ctx.IsSet("disable") {
			client.AuthEnabled = frm.AuthEnabled
			log.Infof("disabled client authentication")
		}

		if client.AuthEnabled {
			log.Infof("client authentication is enabled")
		} else {
			log.Warnf("client authentication is disabled")
		}

		// Update client record if valid.
		if err := client.Validate(); err != nil {
			return fmt.Errorf("invalid values: %s", err)
		} else if err = client.Save(); err != nil {
			return err
		} else {
			log.Infof("client %s has been updated", clean.LogQuote(client.ClientName))
		}

		// Change client secret if requested.
		var secret, message string
		var err error

		if ctx.IsSet("regenerate") && ctx.Bool("regenerate") {
			if secret, err = client.NewSecret(); err != nil {
				return fmt.Errorf("failed to regenerate client secret: %s", err)
			}

			message = fmt.Sprintf(ClientSecretInfo, "FOLLOWING RANDOMLY GENERATED")
		} else if secret = frm.Secret(); secret == "" {
			log.Debugf("client secret remains unchanged")
		} else if err = client.SetSecret(secret); err != nil {
			return fmt.Errorf("failed to set client secret: %s", err)
		} else {
			message = fmt.Sprintf(ClientSecretInfo, "NEW")
		}

		// Show new client secret.
		if secret != "" && err == nil {
			fmt.Println(message)
			result := report.Credentials("Client ID", client.ClientUID, "Client Secret", secret)
			fmt.Printf("\n%s\n", result)
		}

		return nil
	})
}
