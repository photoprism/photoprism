package commands

import (
	"github.com/urfave/cli/v2"
)

// JsonFlag enables machine-readable JSON output for cluster commands.
var JsonFlag = &cli.BoolFlag{
	Name:  "json",
	Usage: "print machine-readable JSON",
}

// OffsetFlag for pagination offset (>= 0).
var OffsetFlag = &cli.IntFlag{
	Name:  "offset",
	Usage: "result `OFFSET` (>= 0)",
	Value: 0,
}

// ClusterCommands configures the cluster command group and subcommands.
var ClusterCommands = &cli.Command{
	Name:  "cluster",
	Usage: "Cluster operations and management (portal, nodes)",
	Subcommands: []*cli.Command{
		ClusterSummaryCommand,
		ClusterHealthCommand,
		ClusterNodesCommands,
		ClusterRegisterCommand,
		ClusterThemePullCommand,
	},
}
