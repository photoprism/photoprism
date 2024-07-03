package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// FindCommand configures the command name, flags, and action.
var FindCommand = cli.Command{
	Name:      "find",
	Usage:     "Searches the index for specific files",
	ArgsUsage: "filter",
	Flags: append(report.CliFlags, cli.UintFlag{
		Name:  "n",
		Usage: "maximum number of search `RESULTS`",
		Value: 10000,
	}),
	Action: findAction,
}

// findAction searches the index for specific files.
func findAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	conf.RegisterDb()
	defer conf.Shutdown()

	frm := form.SearchPhotos{
		Query:   strings.TrimSpace(ctx.Args().First()),
		Primary: false,
		Merged:  false,
		Count:   ctx.Int("n"),
		Offset:  0,
		Order:   sortby.Name,
	}

	results, _, err := search.Photos(frm)

	if err != nil {
		return err
	}

	format := report.CliFormat(ctx)

	// Display just the filename?
	if format == report.Default {
		for _, found := range results {
			fmt.Println(found.FileName)
		}

		return nil
	}

	cols := []string{"File Name", "Mime Type", "Size", "SHA1 Hash"}
	rows := make([][]string, 0, len(results))

	for _, found := range results {
		v := []string{found.FileName, found.FileMime, humanize.Bytes(uint64(found.FileSize)), found.FileHash}
		rows = append(rows, v)
	}

	result, err := report.RenderFormat(rows, cols, format)

	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
