package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

func TestPrintStatusJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var err error
		output := capture.Output(func() {
			err = printStatusJSON([]string{"Ready."}, []config.StatusSection{{Title: "Empty", Cols: []string{"Name", "Value"}}})
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"status":["Ready."],"sections":[{"title":"Empty","items":[]}]}`, output)
	})
	t.Run("NoSections", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printStatusJSON(nil, nil) })
		require.NoError(t, err)
		assert.JSONEq(t, `{"status":null,"sections":[]}`, output)
	})
}

func TestPrintStatusReport(t *testing.T) {
	status := []string{"First line.", "Second line."}
	sections := []config.StatusSection{{
		Title: "Some Options",
		Cols:  []string{"Name", "Value"},
		Rows:  [][]string{{"some-option", "on"}},
		Note:  "A note.",
	}}

	t.Run("Table", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printStatusReport(report.Default, status, sections) })
		require.NoError(t, err)
		assert.Contains(t, output, "First line. Second line.")
		assert.Contains(t, output, "SOME OPTIONS")
		assert.Contains(t, output, "some-option")
		assert.Contains(t, output, "A note.")
	})
	t.Run("Markdown", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printStatusReport(report.Markdown, status, sections) })
		require.NoError(t, err)
		assert.Contains(t, output, "### Some Options")
	})
	t.Run("JSON", func(t *testing.T) {
		var err error
		output := capture.Output(func() { err = printStatusReport(report.JSON, status, sections) })
		require.NoError(t, err)

		var payload struct {
			Status   []string `json:"status"`
			Sections []struct {
				Title string              `json:"title"`
				Items []map[string]string `json:"items"`
				Note  string              `json:"note"`
			} `json:"sections"`
		}

		require.NoError(t, json.Unmarshal([]byte(output), &payload))
		assert.Equal(t, status, payload.Status)
		require.Len(t, payload.Sections, 1)
		assert.Equal(t, "Some Options", payload.Sections[0].Title)
		assert.Equal(t, "on", payload.Sections[0].Items[0]["value"])
		assert.Equal(t, "A note.", payload.Sections[0].Note)
	})
}
