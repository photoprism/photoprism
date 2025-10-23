package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
)

// JsonFlag returns the shared CLI flag definition for JSON output across commands.
func JsonFlag() *cli.BoolFlag {
	return &cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "print machine-readable JSON"}
}

// DryRunFlag returns the shared CLI flag definition for dry runs across commands.
func DryRunFlag(usage string) *cli.BoolFlag {
	if usage == "" {
		usage = "performs a dry run without making any destructive changes"
	}
	return &cli.BoolFlag{Name: "dry-run", Aliases: []string{"dry"}, Usage: usage}
}

// YesFlag returns the shared CLI flag definition for non-interactive runs across commands.
func YesFlag() *cli.BoolFlag {
	return &cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "runs the command non-interactively"}
}

// SuperAdminFlag returns the shared super admin CLI flag definition.
func SuperAdminFlag(usage string) *cli.BoolFlag {
	if usage == "" {
		usage = "makes user super admin with full access"
	}

	return &cli.BoolFlag{Name: "superadmin", Aliases: []string{"super"}, Usage: usage}
}

// ScopeFlag returns the shared CLI flag definition for scopes.
func ScopeFlag(usage string) *cli.StringFlag {
	if usage == "" {
		usage = "authorization `SCOPE` as space-separated resources, or '*' for full access"
	}

	return &cli.StringFlag{
		Name:    "scope",
		Aliases: []string{"s"},
		Usage:   usage,
	}
}

// SaveFlag returns a reusable flag definition for commands that can persist generated values.
func SaveFlag(usage string) *cli.BoolFlag {
	if usage == "" {
		usage = "save the generated value to the configuration"
	}
	return &cli.BoolFlag{Name: "save", Aliases: []string{"s"}, Usage: usage}
}

// PicturesCountFlag returns a shared flag definition limiting how many pictures a batch operation processes.
// Usage: commands from the vision or import tooling that need to cap result size per invocation.
func PicturesCountFlag() *cli.IntFlag {
	return &cli.IntFlag{
		Name:    "count",
		Aliases: []string{"n"},
		Usage:   "maximum `NUMBER` of pictures to be processed",
		Value:   search.MaxResults,
	}
}

// VisionSourceFlag returns the CLI flag used to choose a metadata source for computer-vision commands.
// Allowing only whitelisted aliases keeps CLI input aligned with entity.SrcVisionCommands.
func VisionSourceFlag(src entity.Src) *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "source",
		Aliases: []string{"s"},
		Usage:   fmt.Sprintf("custom data source `TYPE` (%s)", visionSourceUsage()),
		Value:   src,
	}
}
