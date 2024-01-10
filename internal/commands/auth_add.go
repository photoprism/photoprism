package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// AuthAddFlags specifies the "photoprism auth add" command flags.
var AuthAddFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "name, n",
		Usage: "arbitrary `IDENTIFIER` for the new access token",
	},
	cli.StringFlag{
		Name:  "user, u",
		Usage: "provide a `USERNAME` if a personal access token for a specific user account should be created",
	},
	cli.StringFlag{
		Name:  "scope, s",
		Usage: "authorization `SCOPE` for the access token e.g. \"metrics\" or \"photos albums\" (\"*\" to allow all scopes)",
		Value: "*",
	},
	cli.Int64Flag{
		Name:  "expires, e",
		Usage: "access token lifetime in `SECONDS`, after which it expires and a new token must be created (-1 to disable)",
		Value: entity.UnixYear,
	},
}

// AuthAddCommand configures the command name, flags, and action.
var AuthAddCommand = cli.Command{
	Name:   "add",
	Usage:  "Creates a new client access token and shows it",
	Flags:  AuthAddFlags,
	Action: authAddAction,
}

// authAddAction shows detailed session information.
func authAddAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		name := ctx.String("name")

		if name == "" {
			prompt := promptui.Prompt{
				Label: "Token Name",
			}

			res, err := prompt.Run()

			if err != nil {
				return err
			}

			name = clean.Name(res)
		}

		// Set a default token name if no specific name has been provided.
		if name == "" {
			name = rnd.Name()
		}

		// Username provided?
		userName := ctx.String("user")
		user := entity.FindUserByName(userName)

		if user == nil && userName != "" {
			return fmt.Errorf("user %s not found", clean.LogQuote(userName))
		}

		// Create client session.
		sess, err := entity.CreateClientAccessToken(name, ctx.Int64("expires"), ctx.String("scope"), user)

		if err != nil {
			return fmt.Errorf("failed to create access token: %s", err)
		} else {
			// Show client authentication credentials.
			if sess.UserUID == "" {
				fmt.Printf("\nPLEASE WRITE DOWN THE FOLLOWING RANDOMLY GENERATED CLIENT ACCESS TOKEN, AS YOU WILL NOT BE ABLE TO SEE IT AGAIN:\n")
			} else {
				fmt.Printf("\nPLEASE WRITE DOWN THE FOLLOWING RANDOMLY GENERATED PERSONAL ACCESS TOKEN, AS YOU WILL NOT BE ABLE TO SEE IT AGAIN:\n")
			}

			result := report.Credentials("Access Token", sess.AuthToken(), "Authorization Scope", sess.Scope())

			fmt.Printf("\n%s\n", result)
		}

		return err
	})
}
