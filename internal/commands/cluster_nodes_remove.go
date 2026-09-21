package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// ClusterNodesRemoveCommand deletes a node from the registry.
var ClusterNodesRemoveCommand = &cli.Command{
	Name:      "rm",
	Usage:     "Deletes a node from the registry",
	ArgsUsage: "<id|name>",
	Flags: []cli.Flag{
		DryRunFlag("preview deletion without modifying the registry or database"),
		&cli.BoolFlag{Name: "drop-db", Aliases: []string{"d"}, Usage: "drop the node’s provisioned database and user after registry deletion"},
		&cli.BoolFlag{Name: "all-ids", Usage: "delete all records that share the same UUID (admin cleanup)"},
		&cli.BoolFlag{Name: "purge", Usage: "remove the records permanently and release the node’s UUID for reuse"},
		YesFlag(),
	},
	Hidden: true, // Required for cluster-management only.
	Action: clusterNodesRemoveAction,
}

func clusterNodesRemoveAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.Portal() {
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
		// Resolve UUID to delete: accept uuid → clientId → name.
		var node *reg.Node

		purge := ctx.Bool("purge")

		if n, findErr := r.FindByNodeUUID(key); findErr == nil && n != nil {
			node = n
		} else if n, findErr = r.FindByClientID(key); findErr == nil && n != nil {
			node = n
		} else if name := clean.DNSLabel(key); name != "" {
			if n, findErr = r.FindByName(name); findErr == nil && n != nil {
				node = n
			}
		}

		// A node that was already deleted keeps its UUID reserved and resolves nowhere else,
		// so a purge is the one operation that still has to reach it.
		if node == nil && purge {
			if n, findErr := r.FindRetiredByNodeUUID(key); findErr == nil && n != nil {
				node = n
			}
		}

		if node == nil {
			return cli.Exit(fmt.Errorf("node not found"), 3)
		}

		uuid := node.UUID

		dropDB := ctx.Bool("drop-db")
		dbName, dbUser := "", ""

		if dropDB && node.Database != nil {
			dbName = node.Database.Name
			dbUser = node.Database.User
		}

		// Database and user names are derived from the UUID, so a release requires that the
		// schema go with it.
		if purge && !dropDB && node.Database != nil && (node.Database.Name != "" || node.Database.User != "") {
			return cli.Exit(fmt.Errorf("node %s has a provisioned database, so --purge requires --drop-db to release its uuid", clean.Log(uuid)), 2)
		}

		if ctx.Bool("dry-run") {
			log.Infof("dry-run: would delete node %s (uuid=%s, clientId=%s)", clean.LogQuote(node.Name), clean.Log(uuid), clean.Log(node.ClientID))

			if purge {
				log.Infof("dry-run: would permanently remove every entry for uuid %s and release it for reuse", clean.Log(uuid))
			} else if ctx.Bool("all-ids") {
				log.Infof("dry-run: would remove all registry entries that share uuid %s", clean.Log(uuid))
			}

			if dropDB {
				if dbName == "" && dbUser == "" {
					log.Infof("dry-run: --drop-db requested but no database credentials are recorded for node %s", clean.LogQuote(node.Name))
				} else {
					log.Infof("dry-run: would drop database %s and user %s", clean.Log(dbName), clean.Log(dbUser))
				}
			}

			return nil
		}

		if !RunNonInteractively(ctx.Bool("yes")) {
			action := "Delete"

			if purge {
				action = "Permanently delete"
			}

			prompt := promptui.Prompt{Label: fmt.Sprintf("%s node %s?", action, clean.Log(uuid)), IsConfirm: true}
			if _, err = prompt.Run(); err != nil {
				log.Infof("node %s was not deleted", clean.Log(uuid))
				return nil
			}
		}

		// A purge always covers every record for the UUID, because leaving one behind would
		// keep the identifier reserved and defeat the point of releasing it.
		if purge {
			if err = r.PurgeAllByUUID(uuid); err != nil {
				return cli.Exit(err, 1)
			}
		} else if ctx.Bool("all-ids") {
			if err = r.DeleteAllByUUID(uuid); err != nil {
				return cli.Exit(err, 1)
			}
		} else if err = r.Delete(uuid); err != nil {
			return cli.Exit(err, 1)
		}

		who := clusterAuditWho(ctx, conf)
		event.AuditInfo(append(who,
			string(acl.ResourceCluster),
			"node", "%s",
			status.Deleted,
		), clean.Log(uuid))

		deletion := "deleted"

		if purge {
			deletion = "permanently deleted"
		}

		loggedDeletion := false

		if dropDB {
			if dbName == "" && dbUser == "" {
				log.Infof("node %s has been %s (no database credentials recorded)", clean.Log(uuid), deletion)
				loggedDeletion = true
			} else {
				dropCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := provisioner.DropCredentials(dropCtx, dbName, dbUser); err != nil {
					return cli.Exit(fmt.Errorf("failed to drop database credentials for node %s: %w", clean.Log(uuid), err), 1)
				}
				log.Infof("node %s database %s and user %s have been dropped", clean.Log(uuid), clean.Log(dbName), clean.Log(dbUser))
				event.AuditInfo(append(who,
					string(acl.ResourceCluster),
					"drop database %s user %s",
					status.Succeeded,
				), clean.Log(dbName), clean.Log(dbUser))
			}
		}

		if !loggedDeletion {
			log.Infof("node %s has been %s", clean.Log(uuid), deletion)
		}

		if purge {
			log.Infof("uuid %s is available for reuse; drop its database and cluster grants first if it had any", clean.Log(uuid))
		}

		return nil
	})
}
