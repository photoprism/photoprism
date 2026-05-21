package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// FacesConfigCommand displays the config options relevant for face detection
// and recognition. It mirrors `photoprism config` but restricts the report to
// the face-related rows produced by Config.FaceReport.
var FacesConfigCommand = &cli.Command{
	Name:   "config",
	Usage:  "Displays the config options relevant for face detection and recognition",
	Flags:  report.CliFlags,
	Action: facesConfigAction,
}

// FacesConfigReports specifies which face-related reports to display.
var FacesConfigReports = []Report{
	{Title: "Face Detection & Recognition", NoWrap: true, Report: func(conf *config.Config) ([][]string, []string) {
		return conf.FaceReport()
	}},
}

// facesConfigAction prints face-related config option names and values.
func facesConfigAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)
	hub.Disable()

	if err := conf.InitCore(); err != nil {
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
		sections := make([]section, 0, len(FacesConfigReports))
		for _, rep := range FacesConfigReports {
			rows, cols := rep.Report(conf)
			sections = append(sections, section{Title: rep.Title, Items: report.RowsToObjects(rows, cols)})
		}
		b, _ := json.Marshal(map[string]any{"sections": sections})
		fmt.Println(string(b))
		return nil
	}

	for _, rep := range FacesConfigReports {
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
