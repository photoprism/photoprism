package commands

import (
	"github.com/urfave/cli/v2"
)

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
		ClusterJoinTokenCommand,
		ClusterThemePullCommand,
	},
}
