package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClusterNodesCommands groups node subcommands.
var ClusterNodesCommands = &cli.Command{
	Name:  "nodes",
	Usage: "Node registry subcommands",
	Subcommands: []*cli.Command{
		ClusterNodesListCommand,
		ClusterNodesShowCommand,
		ClusterNodesModCommand,
		ClusterNodesRemoveCommand,
		ClusterNodesRotateCommand,
	},
}

// ClusterNodesListCommand lists registered nodes.
var ClusterNodesListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists registered cluster nodes (Portal-only)",
	Flags:     append(append(report.CliFlags, JsonFlag), CountFlag, OffsetFlag),
	ArgsUsage: "",
	Action:    clusterNodesListAction,
}

func clusterNodesListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		if !conf.IsPortal() {
			return cli.Exit(fmt.Errorf("node listing is only available on a Portal node"), 2)
		}

		r, err := reg.NewClientRegistryWithConfig(conf)
		if err != nil {
			return cli.Exit(err, 1)
		}

		items, err := r.List()
		if err != nil {
			return cli.Exit(err, 1)
		}

		// Pagination identical to API defaults.
		count := int(ctx.Uint("count"))
		if count <= 0 || count > 1000 {
			count = 100
		}
		offset := ctx.Int("offset")
		if offset < 0 {
			offset = 0
		}
		if offset > len(items) {
			offset = len(items)
		}
		end := offset + count
		if end > len(items) {
			end = len(items)
		}
		page := items[offset:end]

		// Build admin view (include internal URL and DB meta).
		opts := reg.NodeOpts{IncludeAdvertiseUrl: true, IncludeDatabase: true}
		out := reg.BuildClusterNodes(page, opts)

		if ctx.Bool("json") {
			b, _ := json.Marshal(out)
			fmt.Println(string(b))
			return nil
		}

		cols := []string{"ID", "Name", "Role", "Labels", "Internal URL", "DB Name", "DB User", "DB Last Rotated", "Created At", "Updated At"}
		rows := make([][]string, 0, len(out))
		for _, n := range out {
			var dbName, dbUser, dbRot string
			if n.Database != nil {
				dbName, dbUser, dbRot = n.Database.Name, n.Database.User, n.Database.RotatedAt
			}
			rows = append(rows, []string{
				n.ID, n.Name, n.Role, formatLabels(n.Labels), n.AdvertiseUrl, dbName, dbUser, dbRot, n.CreatedAt, n.UpdatedAt,
			})
		}

		if len(rows) == 0 {
			log.Warnf("no nodes registered")
			return nil
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))
		fmt.Printf("\n%s\n", result)
		if err != nil {
			return cli.Exit(err, 1)
		}
		return nil
	})
}

func formatLabels(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ", ")
}
