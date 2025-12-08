package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ShowConfigCommand configures the command name, flags, and action.
var ShowConfigCommand = &cli.Command{
	Name:   "config",
	Usage:  "Displays global config options and their current values",
	Flags:  report.CliFlags,
	Action: showConfigAction,
}

// ConfigReports specifies which configuration reports to display.
var ConfigReports = []Report{
	{Title: "Global Config Options", NoWrap: true, Report: func(conf *config.Config) ([][]string, []string) {
		return conf.Report()
	}},
	{Title: "OpenID Connect (OIDC)", NoWrap: true, Report: func(conf *config.Config) ([][]string, []string) {
		return conf.OIDCReport()
	}},
}

// showConfigAction displays global config option names and values.
func showConfigAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)
	get.SetConfig(conf)

	if err := conf.Init(); err != nil {
		log.Debug(err)
	}

	format, formatErr := report.CliFormatStrict(ctx)
	if formatErr != nil {
		return formatErr
	}

	if format == report.JSON {
		type section struct {
			Title string              `json:"title"`
			Items []map[string]string `json:"items"`
		}
		sections := make([]section, 0, len(ConfigReports))
		for _, rep := range ConfigReports {
			rows, cols := rep.Report(conf)
			sections = append(sections, section{Title: rep.Title, Items: report.RowsToObjects(rows, cols)})
		}
		b, _ := json.Marshal(map[string]interface{}{"sections": sections})
		fmt.Println(string(b))
		return nil
	}

	for _, rep := range ConfigReports {
		rows, cols := rep.Report(conf)
		opt := report.Options{Format: format, NoWrap: rep.NoWrap}
		result, _ := report.Render(rows, cols, opt)
		switch opt.Format {
		case report.Markdown:
			fmt.Printf("### %s\n\n", rep.Title)
		case report.Default:
			fmt.Printf("%s\n\n", strings.ToUpper(rep.Title))
		}
		fmt.Println(result)
	}
	return nil
}
