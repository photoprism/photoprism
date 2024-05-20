package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/report"
)

// ShowConfigYamlCommand configures the command name, flags, and action.
var ShowConfigYamlCommand = cli.Command{
	Name:   "config-yaml",
	Usage:  "Displays supported YAML config options and CLI flags",
	Flags:  report.CliFlags,
	Action: showConfigYamlAction,
}

// showConfigYamlAction displays supported YAML config options and CLI flag.
func showConfigYamlAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.TraceLevel)

	rows, cols := conf.Options().Report()

	// CSV Export?
	if ctx.Bool("csv") || ctx.Bool("tsv") {
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Println(result)

		return err
	}

	sections := config.YamlReportSections

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
