package commands

import "github.com/urfave/cli/v2"

// AuthJWTCommands groups JWT-related auth helpers under photoprism auth jwt.
var AuthJWTCommands = &cli.Command{
	Name:   "jwt",
	Usage:  "JWT issuance and diagnostics",
	Hidden: true, // Required for cluster-management only.
	Subcommands: []*cli.Command{
		AuthJWTIssueCommand,
		AuthJWTInspectCommand,
		AuthJWTKeysCommand,
		AuthJWTStatusCommand,
	},
}
