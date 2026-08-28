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
// state a database connection reveals: the model the library holds, whether embeddings are paused,
// and which threshold is holding automatic clustering back. It is named for the status rather than
// the configuration, because `photoprism config` reports the same options and cannot state any of it.
var FacesStatusCommand = &cli.Command{
	Name:    "status",
	Aliases: []string{"config"},
	Usage:   "Reports the face detection and recognition status",
	Flags:   report.CliFlags,
	Action:  facesStatusAction,
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

	status := conf.FaceStatus()
	sections := conf.FaceReportSections()

	if format == report.JSON {
		return printFacesStatusJSON(status, sections)
	}

	fmt.Printf("\n%s\n\n", strings.Join(status, " "))

	for _, section := range sections {
		result, renderErr := report.Render(section.Rows, section.Cols, report.Options{Format: format, NoWrap: true})

		if renderErr != nil {
			return renderErr
		}

		switch format {
		case report.Markdown:
			fmt.Printf("### %s\n\n", section.Title)
		default:
			fmt.Printf("%s\n\n", strings.ToUpper(section.Title))
		}

		fmt.Println(result)

		if section.Note != "" {
			fmt.Printf("%s\n\n", section.Note)
		}
	}

	return nil
}

// printFacesStatusJSON writes the status report as a single JSON object, so that a caller parsing
// it gets the notes and the status lines too rather than only the tables.
func printFacesStatusJSON(status []string, sections []config.FaceReportSection) error {
	type jsonSection struct {
		Title string              `json:"title"`
		Items []map[string]string `json:"items"`
		Note  string              `json:"note,omitempty"`
	}

	out := make([]jsonSection, 0, len(sections))

	for _, section := range sections {
		out = append(out, jsonSection{
			Title: section.Title,
			Items: report.RowsToObjects(section.Rows, section.Cols),
			Note:  section.Note,
		})
	}

	b, err := json.Marshal(map[string]any{"status": status, "sections": out})

	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}
