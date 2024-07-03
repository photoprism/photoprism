package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt/report"
	"github.com/photoprism/photoprism/pkg/unix"
)

// AuthAddFlags specifies the "photoprism auth add" command flags.
var AuthAddFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "name, n",
		Usage: "`CLIENT` name to help identify the application",
	},
	cli.StringFlag{
		Name:  "scope, s",
		Usage: "authorization `SCOPES` e.g. \"metrics\" or \"photos albums\" (\"*\" to allow all)",
	},
	cli.Int64Flag{
		Name:  "expires, e",
		Usage: "authentication `LIFETIME` in seconds, after which access expires (-1 to disable the limit)",
		Value: unix.Year,
	},
}

// AuthAddCommand configures the command name, flags, and action.
var AuthAddCommand = cli.Command{
	Name:  "add",
	Usage: "Adds a new authentication secret for client applications",
	Description: "If you specify a username as argument, an app password will be created for this user account." +
		" It can be used as a password replacement to grant limited access to client applications.",
	ArgsUsage: "[username]",
	Flags:     AuthAddFlags,
	Action:    authAddAction,
}

// authAddAction shows detailed session information.
func authAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		// Get username from command flag.
		userName := clean.Username(ctx.Args().First())

		// Find user account.
		user := entity.FindUserByName(userName)

		if user == nil && userName != "" {
			return fmt.Errorf("user %s not found", clean.LogQuote(userName))
		}

		// Get client name from command flag or ask for it.
		clientName := ctx.String("name")

		if clientName == "" {
			prompt := promptui.Prompt{
				Label:   "Client Name",
				Default: rnd.Name(),
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			clientName = clean.Name(res)
		}

		// Get auth scope from command flag or ask for it.
		authScope := ctx.String("scope")

		if authScope == "" {
			prompt := promptui.Prompt{
				Label:   "Authorization Scope",
				Default: "*",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			authScope = clean.Scope(res)
		}

		// Create session and show the authentication secret.
		sess, err := entity.AddClientSession(clientName, ctx.Int64("expires"), authScope, authn.GrantCLI, user)

		if err != nil {
			return fmt.Errorf("failed to create authentication secret: %s", err)
		} else {
			// Show client authentication credentials.
			if sess.UserUID == "" {
				fmt.Printf("\nPLEASE COPY THE FOLLOWING RANDOMLY GENERATED ACCESS TOKEN AND KEEP IT IN A SAFE PLACE, AS YOU WILL NOT BE ABLE TO SEE IT AGAIN:\n")
				fmt.Printf("\n%s\n", report.Credentials("Access Token", sess.AuthToken(), "Authorization Scope", sess.Scope()))
			} else {
				fmt.Printf("\nPLEASE COPY THE FOLLOWING RANDOMLY GENERATED APP PASSWORD AND KEEP IT IN A SAFE PLACE, AS YOU WILL NOT BE ABLE TO SEE IT AGAIN:\n")
				fmt.Printf("\n%s\n", report.Credentials("App Password", sess.AuthToken(), "Authorization Scope", sess.Scope()))
			}

		}

		return err
	})
}
