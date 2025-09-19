package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClusterNodesRemoveCommand deletes a node from the registry.
var ClusterNodesRemoveCommand = &cli.Command{
	Name:      "rm",
	Usage:     "Deletes a node from the registry (Portal-only)",
	ArgsUsage: "<id|name>",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "runs the command non-interactively"},
	},
	Action: clusterNodesRemoveAction,
}

func clusterNodesRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.IsPortal() {
			return cli.Exit(fmt.Errorf("node delete is only available on a Portal node"), 2)
		}

		key := ctx.Args().First()
		if key == "" {
			return cli.Exit(fmt.Errorf("node id or name is required"), 2)
		}

		r, err := reg.NewClientRegistryWithConfig(conf)
		if err != nil {
			return cli.Exit(err, 1)
		}

		// Resolve to id for deletion, but also support name.
		id := key
		if _, getErr := r.Get(id); getErr != nil {
			if n, err2 := r.FindByName(clean.TypeLowerDash(key)); err2 == nil && n != nil {
				id = n.ID
			} else {
				return cli.Exit(fmt.Errorf("node not found"), 3)
			}
		}

		confirmed := RunNonInteractively(ctx.Bool("yes"))
		if !confirmed {
			prompt := promptui.Prompt{Label: fmt.Sprintf("Delete node %s?", clean.Log(id)), IsConfirm: true}
			if _, err := prompt.Run(); err != nil {
				log.Infof("node %s was not deleted", clean.Log(id))
				return nil
			}
		}

		if err := r.Delete(id); err != nil {
			return cli.Exit(err, 1)
		}

		log.Infof("node %s has been deleted", clean.Log(id))
		return nil
	})
}
