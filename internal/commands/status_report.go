package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// printStatusReport prints the status lines of a status command above one table per section.
func printStatusReport(format report.Format, status []string, sections []config.StatusSection) error {
	if format == report.JSON {
		return printStatusJSON(status, sections)
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

// printStatusJSON writes the status report as a single JSON object, so that a caller parsing
// it gets the notes and the status lines too rather than only the tables.
func printStatusJSON(status []string, sections []config.StatusSection) error {
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
