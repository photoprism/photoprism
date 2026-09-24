package commands

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
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
			return cli.ShowSubcommandHelp(ctx)
		}

		m, err := query.Session(id)

		switch {
		case errors.Is(err, query.ErrInvalidSessionID):
			return cli.Exit(err, 2)
		case gorm.IsRecordNotFoundError(err):
			return cli.Exit(errors.New("session not found"), 3)
		case err != nil:
			return cli.Exit(err, 1)
		}

		if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), fmt.Sprintf("Remove session %s?", clean.LogQuote(id))); confirmErr != nil {
			return confirmErr
		} else if !proceed {
			log.Infof("session %s was not removed", clean.LogQuote(id))
			return nil
		}

		if err = m.Delete(); err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("session %s has been removed", clean.LogQuote(id))

		return nil
	})
}
