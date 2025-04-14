package commands

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VisionListCommand configures the command name, flags, and action.
var VisionListCommand = &cli.Command{
	Name:   "ls",
	Usage:  "Lists configured computer vision models",
	Flags:  append(report.CliFlags),
	Action: visionListAction,
}

// visionListAction displays existing user accounts.
func visionListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		var rows [][]string

		cols := []string{"Type", "Name", "Version", "Resolution", "Uri", "Tags", "Disabled"}

		// Show log message.
		log.Infof("found %s", english.Plural(len(vision.Config.Models), "model", "models"))

		if n := len(vision.Config.Models); n == 0 {
			return nil
		} else {
			rows = make([][]string, n)
		}

		// Display report.
		for i, model := range vision.Config.Models {
			modelUri, _ := model.Endpoint()
			rows[i] = []string{
				model.Type,
				model.Name,
				model.Version,
				fmt.Sprintf("%d", model.Resolution),
				modelUri,
				strings.Join(model.Tags, ", "),
				report.Bool(model.Disabled, report.Yes, report.No),
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
