package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/report"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClientsAddCommand configures the command name, flags, and action.
var ClientsAddCommand = cli.Command{
	Name:        "add",
	Usage:       "Registers a new client application",
	Description: "If you specify a username as argument, the new client will belong to this user and inherit its privileges.",
	ArgsUsage:   "[username]",
	Flags:       ClientAddFlags,
	Action:      clientsAddAction,
}

// clientsAddAction registers a new client application.
func clientsAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		conf.MigrateDb(false, nil)

		frm := form.AddClientFromCli(ctx)

		interactive := true

		if frm.ClientName != "" && frm.AuthScope != "" {
			log.Debugf("client will be added in non-interactive mode")
			interactive = false
		}

		if interactive && frm.ClientName == "" {
			prompt := promptui.Prompt{
				Label:   "Client Name",
				Default: rnd.Name(),
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			frm.ClientName = clean.Name(res)
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
			return addErr
		} else {
			log.Infof("successfully registered new client %s", clean.LogQuote(client.ClientName))

			// Display client details.
			cols := []string{"Client ID", "Name", "Authentication Method", "User", "Role", "Scope", "Enabled", "Access Token Lifetime", "Created At"}
			rows := make([][]string, 1)

			var authExpires string

			if client.AuthExpires > 0 {
				authExpires = client.Expires().String()
			}

			if client.AuthTokens > 0 {
				if authExpires != "" {
					authExpires = fmt.Sprintf("%s; up to %s", authExpires, english.Plural(int(client.Tokens()), "token", "tokens"))
				} else {
					authExpires = fmt.Sprintf("up to %d tokens", client.AuthTokens)
				}
			}

			rows[0] = []string{
				client.GetUID(),
				client.Name(),
				client.AuthInfo(),
				client.UserInfo(),
				client.AclRole().String(),
				client.Scope(),
				report.Bool(client.AuthEnabled, report.Yes, report.No),
				authExpires,
				client.CreatedAt.Format("2006-01-02 15:04:05"),
			}

			if result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx)); err == nil {
				fmt.Printf("\n%s", result)
			}
		}

		// Se a random secret or the secret specified in the command flags, if any.
		var secret, message string
		var err error

		if secret = frm.Secret(); secret == "" {
			secret, err = client.NewSecret()
			message = fmt.Sprintf(ClientSecretInfo, "FOLLOWING RANDOMLY GENERATED")
		} else {
			err = client.SetSecret(secret)
			message = fmt.Sprintf(ClientSecretInfo, "SPECIFIED")
		}

		// Check if the secret has been saved successfully or return an error otherwise.
		if err != nil {
			return fmt.Errorf("failed to set client secret: %s", err)
		}

		// Show client authentication credentials.
		fmt.Printf(message)
		result := report.Credentials("Client ID", client.ClientUID, "Client Secret", secret)
		fmt.Printf("\n%s\n", result)

		return nil
	})
}
