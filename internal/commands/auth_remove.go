package commands

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AuthRemoveCommand configures the command name, flags, and action.
var AuthRemoveCommand = cli.Command{
	Name:      "rm",
	Usage:     "Deletes a session by id or access token",
	ArgsUsage: "[identifier]",
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

		actionPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Remove session %s?", clean.LogQuote(id)),
			IsConfirm: true,
		}

		if _, err := actionPrompt.Run(); err == nil {
			if m, err := query.Session(id); err != nil {
				return errors.New("session not found")
			} else if err := m.Delete(); err != nil {
				return err
			} else {
				log.Infof("session %s has been removed", clean.LogQuote(id))
			}
		} else {
			log.Infof("session %s was not removed", clean.LogQuote(id))
		}

		return nil
	})
}
