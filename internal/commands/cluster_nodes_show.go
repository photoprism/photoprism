package commands

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClusterNodesShowCommand shows node details.
var ClusterNodesShowCommand = &cli.Command{
	Name:      "show",
	Usage:     "Shows node details",
	ArgsUsage: "<id|name>",
	Flags:     report.CliFlags,
	Hidden:    true, // Required for cluster-management only.
	Action:    clusterNodesShowAction,
}

func clusterNodesShowAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.Portal() {
			return cli.Exit(fmt.Errorf("node show is only available on a Portal node"), 2)
		}

		key := ctx.Args().First()
		if key == "" {
			return cli.Exit(fmt.Errorf("node id or name is required"), 2)
		}

		r, err := reg.NewClientRegistryWithConfig(conf)
		if err != nil {
			return cli.Exit(err, 1)
		}

		// Resolve by id first, then by normalized name.
		n, getErr := r.FindByNodeUUID(key)
		if getErr != nil || n == nil {
			n, getErr = r.FindByClientID(key)
		}
		if getErr != nil || n == nil {
			name := clean.DNSLabel(key)
			if name == "" {
				return cli.Exit(fmt.Errorf("invalid node identifier"), 2)
			}
			n, getErr = r.FindByName(name)
		}
		if getErr != nil || n == nil {
			return cli.Exit(fmt.Errorf("node not found"), 3)
		}

		opts := reg.NodeOpts{IncludeAdvertiseUrl: true, IncludeDatabase: true}
		dto := reg.BuildClusterNode(*n, opts)

		who := clusterAuditWho(ctx, conf)
		event.AuditInfo(append(who,
			string(acl.ResourceCluster),
			"show node", "%s",
			status.Succeeded,
		), clean.Log(dto.UUID))

		if ctx.Bool("json") {
			b, _ := json.Marshal(dto)
			fmt.Println(string(b))
			return nil
		}

		cols := []string{"UUID", "ClientID", "Name", "Role", "Internal URL", "DB Driver", "DB Name", "DB User", "DB Last Rotated", "Created At", "Updated At"}
		var dbName, dbUser, dbRot, dbDriver string
		if dto.Database != nil {
			dbName, dbUser, dbRot, dbDriver = dto.Database.Name, dto.Database.User, dto.Database.RotatedAt, dto.Database.Driver
		}
		rows := [][]string{{dto.UUID, dto.ClientID, dto.Name, dto.Role, dto.AdvertiseUrl, dbDriver, dbName, dbUser, dbRot, dto.CreatedAt, dto.UpdatedAt}}
		out, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))
		fmt.Printf("\n%s\n", out)
		if err != nil {
			return cli.Exit(err, 1)
		}
		return nil
	})
}
