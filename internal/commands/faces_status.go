package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// FacesStatusCommand reports the face detection and recognition configuration together with the
// state a database connection reveals: the model the library holds, and whether embeddings are
// paused. It is named for the status rather than the configuration, because `photoprism config`
// reports the same options and cannot state either.
var FacesStatusCommand = &cli.Command{
	Name:    "status",
	Aliases: []string{"config"},
	Usage:   "Reports the face detection and recognition status",
	Flags:   report.CliFlags,
	Action:  facesStatusAction,
}

// FacesStatusReports specifies which face-related reports to display.
var FacesStatusReports = []Report{
	{Title: "Face Detection & Recognition", NoWrap: true, Report: func(conf *config.Config) ([][]string, []string) {
		return conf.FaceReport()
	}},
}

// facesStatusAction prints the face detection and recognition status.
func facesStatusAction(ctx *cli.Context) error {
	conf, err := InitCoreConfig(ctx, true)

	if err != nil {
		log.Debug(err)
	}

	// Detecting the model asks the library which one produced its vectors, so without a
	// connection this report could not name the model that is in force. Connecting is
	// idempotent and fails over to the old behavior.
	conf.RegisterDb()

	// A report states what is configured and must not change it, so this resolves the model
	// for display without writing the result to "options.yml".
	conf.ResolveFaceModel()

	format, formatErr := report.CliFormatStrict(ctx)
	if formatErr != nil {
		return formatErr
	}

	if format == report.JSON {
		type section struct {
			Title string              `json:"title"`
			Items []map[string]string `json:"items"`
		}
		sections := make([]section, 0, len(FacesStatusReports))
		for _, rep := range FacesStatusReports {
			rows, cols := rep.Report(conf)
			sections = append(sections, section{Title: rep.Title, Items: report.RowsToObjects(rows, cols)})
		}
		b, _ := json.Marshal(map[string]any{"sections": sections})
		fmt.Println(string(b))
		return nil
	}

	for _, rep := range FacesStatusReports {
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
