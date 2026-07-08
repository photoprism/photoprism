package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VideoListCommand configures the command name, flags, and action.
var VideoListCommand = &cli.Command{
	Name:      "ls",
	Usage:     "Lists indexed video files matching the specified filters",
	ArgsUsage: "[filter]...",
	Flags: append(append([]cli.Flag{}, report.CliFlags...),
		videoCountFlag,
		OffsetFlag,
	),
	Action: videoListAction,
}

// videoListAction renders a filtered list of indexed video files.
func videoListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		// Ensure config is initialized before querying the index.
		if conf == nil {
			return fmt.Errorf("config is not available")
		}

		format, err := report.CliFormatStrict(ctx)

		if err != nil {
			return err
		}

		filter := videoNormalizeFilter(ctx.Args().Slice())
		results, err := videoSearchResults(filter, ctx.Int(videoCountFlag.Name), ctx.Int(OffsetFlag.Name))

		if err != nil {
			return err
		}

		cols := videoListColumns()

		if format == report.JSON {
			rows := make([]map[string]any, 0, len(results))
			for _, found := range results {
				rows = append(rows, videoListJSONRow(found))
			}

			payload, jsonErr := videoListJSON(rows, cols)
			if jsonErr != nil {
				return jsonErr
			}

			fmt.Println(payload)
			return nil
		}

		rows := make([][]string, 0, len(results))

		for _, found := range results {
			rows = append(rows, videoListRow(found))
		}

		output, err := report.RenderFormat(rows, cols, format)

		if err != nil {
			return err
		}

		fmt.Println(output)

		return nil
	})
}
