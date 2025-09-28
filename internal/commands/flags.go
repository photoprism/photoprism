package commands

import "github.com/urfave/cli/v2"

// JsonFlag returns the shared CLI flag definition for JSON output across commands.
func JsonFlag() *cli.BoolFlag {
	return &cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "print machine-readable JSON"}
}
