package commands

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowConfigYamlCommand configures the command name, flags, and action.
var ShowConfigYamlCommand = &cli.Command{
	Name:        "config-yaml",
	Usage:       "Displays supported YAML config options and CLI flags",
	Description: "For readability, standard and Markdown text output is divided into sections. The --json, --csv, and --tsv options return a flat list.",
	Flags:       report.CliFlags,
	Action:      showConfigYamlAction,
}

// showConfigYamlAction displays supported YAML config options and CLI flag.
func showConfigYamlAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.TraceLevel)

	rows, cols := conf.Options().Report()
	format, formatErr := report.CliFormatStrict(ctx)
	if formatErr != nil {
		return formatErr
	}

	// CSV/TSV/JSON exports use default single-table rendering.
	if format == report.CSV || format == report.TSV || format == report.JSON {
		result, err := report.RenderFormat(rows, cols, format)
		fmt.Println(result)
		return err
	}

	// JSON aggregation path (commented out because non-nested output is preferred for now).
	/* if format == report.JSON {
		type section struct {
			Title string              `json:"title"`
			Info  string              `json:"info,omitempty"`
			Items []map[string]string `json:"items"`
		}
		sectionsCfg := config.YamlReportSections
		agg := make([]section, 0, len(sectionsCfg))
		j := 0
		for i, sec := range sectionsCfg {
			secRows := make([][]string, 0)
			for {
				row := rows[j]
				if len(row) < 1 {
					continue
				}
				if i < len(sectionsCfg)-1 && sectionsCfg[i+1].Start == row[0] {
					break
				}
				secRows = append(secRows, row)
				j++
				if j >= len(rows) {
					break
				}
			}
			agg = append(agg, section{Title: sec.Title, Info: sec.Info, Items: report.RowsToObjects(secRows, cols)})
			if j >= len(rows) {
				break
			}
		}
		b, _ := json.Marshal(map[string]interface{}{"sections": agg})
		fmt.Println(string(b))
		return nil
	} */

	markDown := ctx.Bool("md")
	sections := config.YamlReportSections

	j := 0

	for i, section := range sections {
		if markDown {
			fmt.Printf("### %s\n\n", section.Title)
		} else {
			fmt.Printf("%s\n\n", strings.ToUpper(section.Title))
		}

		if section.Info != "" && markDown {
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

		// JSON handled earlier; Markdown and default render per section below.
		result, err := report.RenderFormat(secRows, cols, format)

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
