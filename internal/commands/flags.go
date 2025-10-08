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
