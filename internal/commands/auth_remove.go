package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AuthRemoveCommand configures the command name, flags, and action.
var AuthRemoveCommand = &cli.Command{
	Name:      "rm",
	Usage:     "Deletes a session by id or access token",
	ArgsUsage: "[identifier]",
	Flags:     []cli.Flag{YesFlag()},
	Action:    authRemoveAction,
}

// authRemoveAction deletes the specified session.
func authRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		id := clean.ID(ctx.Args().First())

		// ID provided?
		if id == "" {
			return ShowUsageError(ctx)
		}

		m, err := authFindSession(id)

		if err != nil {
			return err
		}

		label := authSessionLabel(m)

		if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), fmt.Sprintf("Remove %s?", label)); confirmErr != nil {
			return confirmErr
		} else if !proceed {
			log.Infof("session %s was not removed", clean.Log(m.RefID))
			return nil
		}

		if err = m.Delete(); err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("%s has been removed", label)

		return nil
	})
}
