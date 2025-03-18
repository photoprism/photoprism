package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowConfigOptionsCommand configures the command name, flags, and action.
var ShowConfigOptionsCommand = &cli.Command{
	Name:   "config-options",
	Usage:  "Displays supported environment variables and CLI flags",
	Flags:  report.CliFlags,
	Action: showConfigOptionsAction,
}

// showConfigOptionsAction displays supported environment variables and CLI flags.
func showConfigOptionsAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)

	rows, cols := config.Flags.Report()

	// CSV Export?
	if ctx.Bool("csv") || ctx.Bool("tsv") {
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Println(result)

		return err
	}

	sections := config.OptionsReportSections

	j := 0

	for i, section := range sections {
		fmt.Printf("### %s ###\n\n", section.Title)
		if section.Info != "" && ctx.Bool("md") {
			fmt.Printf("%s\n\n", section.Info)
		}

		secRows := make([][]string, 0, len(rows))

		for {
			row := rows[j]

			if len(row) < 1 {
				continue
			}

			if i < len(sections)-1 {
				if sections[i+1].Start == row[0] {
					break
				}
			}

			secRows = append(secRows, row)
			j++

			if j >= len(rows) {
				break
			}
		}

		result, err := report.RenderFormat(secRows, cols, report.CliFormat(ctx))

		if err != nil {
			return err
		}

		fmt.Println(result)

		if j >= len(rows) {
			break
		}
	}

	return nil
}
