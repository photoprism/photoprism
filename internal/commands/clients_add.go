package commands

import (
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/report"
)

// ClientsAddCommand configures the command name, flags, and action.
var ClientsAddCommand = cli.Command{
	Name:   "add",
	Usage:  "Registers a new client application",
	Flags:  ClientAddFlags,
	Action: clientsAddAction,
}

// clientsAddAction registers a new client application.
func clientsAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		frm := form.NewClientFromCli(ctx)

		interactive := true

		if frm.ClientName != "" && frm.AuthScope != "" {
			log.Debugf("client will be added in non-interactive mode")
			interactive = false
		}

		if interactive && frm.ClientName == "" {
			prompt := promptui.Prompt{
				Label: "Client Name",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			frm.ClientName = clean.Name(res)
		}

		// Set a default client name if no specific name has been provided.
		if frm.ClientName == "" {
			frm.ClientName = time.Now().UTC().Format(time.DateTime)
		}

		if interactive && frm.AuthScope == "" {
			prompt := promptui.Prompt{
				Label: "Authorization Scope",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			frm.AuthScope = clean.Scope(res)
		}

		// Set a default client name if no specific name has been provided.
		if frm.AuthScope == "" {
			frm.AuthScope = list.All
		}

		client, addErr := entity.AddClient(frm)

		if addErr != nil {
			return fmt.Errorf("failed to add client: %s", addErr)
		} else {
			log.Infof("successfully registered new client %s", clean.LogQuote(client.ClientName))

			// Display client details.
			cols := []string{"Client ID", "Client Name", "Authentication", "Scope", "User", "Enabled", "Access Token Expires", "Created At"}
			rows := make([][]string, 1)

			var userName string
			if client.UserUID == "" {
				userName = report.NotAssigned
			} else if client.UserName != "" {
				userName = client.UserName
			} else {
				userName = client.UserUID
			}

			var authExpires string
			if client.AuthExpires > 0 {
				authExpires = client.Expires().String()
			} else {
				authExpires = report.Never
			}

			if client.AuthTokens > 0 {
				authExpires = fmt.Sprintf("%s, max %d tokens", authExpires, client.AuthTokens)
			}

			rows[0] = []string{
				client.UID(),
				client.ClientName,
				client.AuthMethod,
				client.AuthScope,
				userName,
				report.Bool(client.AuthEnabled, report.Yes, report.No),
				authExpires,
				client.CreatedAt.Format("2006-01-02 15:04:05"),
			}

			if result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx)); err == nil {
				fmt.Printf("\n%s", result)
			}
		}

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

		return nil
	})
}
