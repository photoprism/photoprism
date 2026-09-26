package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var joinTokenSaveFlag = SaveFlag("write the generated join token to config/portal/secrets/join_token")

// ClusterJoinTokenCommand generates cluster join tokens for nodes.
var ClusterJoinTokenCommand = &cli.Command{
	Name:  "join-token",
	Usage: "Generates a portal join token for registering nodes",
	Flags: []cli.Flag{
		joinTokenSaveFlag,
		YesFlag(),
	},
	Action: clusterJoinTokenAction,
}

// clusterJoinTokenAction generates a portal join token for registering nodes.
func clusterJoinTokenAction(ctx *cli.Context) error {
	// Always print a freshly generated token; saving it is optional.
	token := rnd.JoinToken()
	fmt.Println(token)

	if !ctx.Bool("save") {
		return nil
	}

	return CallWithDependencies(ctx, func(conf *config.Config) error {
		tokenFile := conf.PortalJoinTokenFile()

		if fs.FileExistsNotEmpty(tokenFile) {
			if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), fmt.Sprintf("Replace existing join token in %s", clean.Log(tokenFile))); confirmErr != nil {
				return confirmErr
			} else if !proceed {
				log.Infof("cluster: join token was not updated")
				return nil
			}
		}

		_, savedFile, err := conf.SaveJoinToken(token)
		if err != nil {
			return cli.Exit(fmt.Errorf("failed to write join token: %w", err), 1)
		}

		log.Infof("cluster: new join token saved to %s", clean.Log(savedFile))
		return nil
	})
}
