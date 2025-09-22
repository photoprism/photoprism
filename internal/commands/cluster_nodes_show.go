package commands

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClusterNodesShowCommand shows node details.
var ClusterNodesShowCommand = &cli.Command{
	Name:      "show",
	Usage:     "Shows node details (Portal-only)",
	ArgsUsage: "<id|name>",
	Flags:     report.CliFlags,
	Action:    clusterNodesShowAction,
}

func clusterNodesShowAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.IsPortal() {
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
		n, getErr := r.Get(key)
		if getErr != nil {
			name := clean.TypeLowerDash(key)
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

		if ctx.Bool("json") {
			b, _ := json.Marshal(dto)
			fmt.Println(string(b))
			return nil
		}

		cols := []string{"ID", "Name", "Role", "Internal URL", "DB Name", "DB User", "DB Last Rotated", "Created At", "Updated At"}
		var dbName, dbUser, dbRot string
		if dto.Database != nil {
			dbName, dbUser, dbRot = dto.Database.Name, dto.Database.User, dto.Database.RotatedAt
		}
		rows := [][]string{{dto.ID, dto.Name, dto.Role, dto.AdvertiseUrl, dbName, dbUser, dbRot, dto.CreatedAt, dto.UpdatedAt}}
		out, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))
		fmt.Printf("\n%s\n", out)
		if err != nil {
			return cli.Exit(err, 1)
		}
		return nil
	})
}
