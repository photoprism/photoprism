package commands

import (
	"context"
	"fmt"
	"time"

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
		&cli.BoolFlag{Name: "drop-db", Aliases: []string{"d"}, Usage: "also delete the node’s provisioned database and user, discarding the data they hold"},
		&cli.BoolFlag{Name: "all-ids", Usage: "delete all records that share the same UUID (admin cleanup)"},
		&cli.BoolFlag{Name: "purge", Usage: "remove the records permanently and release the node’s UUID for reuse"},
		YesFlag(),
	},
	Hidden: true, // Required for cluster-management only.
	Action: clusterNodesRemoveAction,
}

// nodeDatabase reports the database provisioned for a node, or an empty string when the
// registry records none. The name is derived from the node UUID, so the identifier is not
// released while the database it names still exists.
func nodeDatabase(node *reg.Node) string {
	if node == nil || node.Database == nil {
		return ""
	}

	if name := node.Database.Name; name != "" {
		return name
	}

	return node.Database.User
}

// releaseBlockedBy reports the database that keeps a node UUID from being released, or an empty
// string when none does. A release removes every record holding the identifier, so it is checked
// against the databases all of them name; cleaning is the one the same run deletes, which is
// therefore not an obstacle to it.
func releaseBlockedBy(databases []string, cleaning string) string {
	for _, name := range databases {
		if name != "" && name != cleaning {
			return name
		}
	}

	return ""
}

// dropNodeDatabase deletes a node's provisioned database and user and records the outcome.
// A failure names what is left behind, because the registry record that named it may already
// be gone and nothing else would say which database survived.
func dropNodeDatabase(who []string, uuid, dbName, dbUser string) error {
	dropCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provisioner.DropCredentials(dropCtx, dbName, dbUser); err != nil {
		event.AuditErr(append(who,
			string(acl.ResourceCluster),
			"drop database %s user %s",
			status.Error(err),
		), clean.Log(dbName), clean.Log(dbUser))

		return fmt.Errorf("failed to delete database %s and user %s of node %s: %w", clean.Log(dbName), clean.Log(dbUser), clean.Log(uuid), err)
	}

	log.Infof("database %s and user %s of node %s have been deleted", clean.Log(dbName), clean.Log(dbUser), clean.Log(uuid))

	event.AuditInfo(append(who,
		string(acl.ResourceCluster),
		"drop database %s user %s",
		status.Succeeded,
	), clean.Log(dbName), clean.Log(dbUser))

	return nil
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

		if node.Database != nil {
			dbName = node.Database.Name
			dbUser = node.Database.User
		}

		dropsDatabase := dropDB && (dbName != "" || dbUser != "")

		// Releasing an identifier is refused while a database named after it still exists.
		// A release removes every record that holds the uuid, not only the one resolved here,
		// so every database those records name has to be accounted for - the resolved one can
		// go with --drop-db, and any other has to be dealt with before the release.
		// A blocked release is still previewed, because a dry run is how an operator checks
		// what a destructive command would reach.
		blockedBy := ""

		if purge {
			cleaning := ""

			if dropsDatabase {
				cleaning = nodeDatabase(node)
			}

			blockedBy = releaseBlockedBy(r.NodeDatabases(uuid), cleaning)
		}

		if ctx.Bool("dry-run") {
			log.Infof("dry-run: would delete node %s (uuid=%s, clientId=%s)", clean.LogQuote(node.Name), clean.Log(uuid), clean.Log(node.ClientID))

			if purge && blockedBy == "" {
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

			if blockedBy != "" {
				log.Infof("dry-run: releasing uuid %s is refused while database %s exists", clean.Log(uuid), clean.Log(blockedBy))
			}

			return nil
		}

		// The identifier names the database, so it is not released while that database holds
		// data. What becomes of the data is the operator's decision, not a side effect.
		if blockedBy != "" {
			return cli.Exit(fmt.Errorf("node %s still owns database %s; back it up or move it before releasing the uuid, because --drop-db discards it permanently", clean.Log(uuid), clean.Log(blockedBy)), 2)
		}

		action := "Delete"

		if purge {
			action = "Permanently delete"
		}

		// Name the node the operator typed, not only the uuid it resolved to, so a wrong
		// target is visible. Name the database too, so nobody agrees to delete one without
		// being told which, or that it happens at all.
		label := fmt.Sprintf("%s node %s (uuid %s)?", action, clean.LogQuote(node.Name), clean.Log(uuid))

		if dropsDatabase {
			label = fmt.Sprintf("%s node %s (uuid %s) AND permanently delete database %s with user %s?",
				action, clean.LogQuote(node.Name), clean.Log(uuid), clean.Log(dbName), clean.Log(dbUser))
		}

		if proceed, confirmErr := ConfirmAction(ctx.Bool("yes"), label); confirmErr != nil {
			return confirmErr
		} else if !proceed {
			log.Infof("node %s was not deleted", clean.Log(uuid))
			return nil
		}

		who := clusterAuditWho(ctx, conf)

		// Name the target before anything irreversible runs, so a path that shows no prompt
		// still records what was reached, and a later failure can be traced to it.
		if dropsDatabase {
			log.Infof("removing node %s (uuid %s), deleting database %s and user %s", clean.LogQuote(node.Name), clean.Log(uuid), clean.Log(dbName), clean.Log(dbUser))
		} else {
			log.Infof("removing node %s (uuid %s)", clean.LogQuote(node.Name), clean.Log(uuid))
		}

		// D1: the identifier is released last. A drop that fails has to leave the uuid
		// reserved rather than free with the database still in place, so a purge deletes the
		// data it names before it removes the records that name it.
		droppedDatabase := false

		if purge && dropsDatabase {
			if err = dropNodeDatabase(who, uuid, dbName, dbUser); err != nil {
				return cli.Exit(err, 1)
			}

			droppedDatabase = true
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

		event.AuditInfo(append(who,
			string(acl.ResourceCluster),
			"node", "%s",
			status.Deleted,
		), clean.Log(uuid))

		deletion := "deleted"

		if purge {
			deletion = "permanently deleted"
		}

		// An ordinary removal retires the record first, so the node is out of service before
		// its database goes and the uuid stays reserved whatever happens next.
		if dropsDatabase && !droppedDatabase {
			if err = dropNodeDatabase(who, uuid, dbName, dbUser); err != nil {
				return cli.Exit(err, 1)
			}
		} else if dropDB && !dropsDatabase {
			log.Infof("node %s records no database credentials, so none were deleted", clean.Log(uuid))
		}

		log.Infof("node %s has been %s", clean.Log(uuid), deletion)

		if purge {
			log.Infof("uuid %s is available for reuse; its cluster grants remain and are yours to clean up", clean.Log(uuid))
		}

		return nil
	})
}
