package commands

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
)

// ConnectCommand configures the command name, flags, and action.
var ConnectCommand = cli.Command{
	Name:      "connect",
	Usage:     "Connects your membership account",
	ArgsUsage: "[activation code]",
	Action:    connectAction,
}

// connectAction connects your membership account.
func connectAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		token := ctx.Args().First()

		// Fail if no code was provided.
		if token == "" {
			return cli.ShowSubcommandHelp(ctx)
		}

		// Connect to hub.
		if err := conf.ResyncHub(token); err != nil {
			return err
		}

		log.Infof("successfully connected your account")

		return nil
	})
}
