package commands

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// TestFacesSubjectsCommand covers the people report, whose reason for existing is the pair of
// count columns: the stored ones and the ones the markers currently support drift, and the drift is
// what keeps a person off People > Recognized after a CLI-only reset and re-cluster.
func TestFacesSubjectsCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects"})
		require.NoError(t, err)

		assert.Contains(t, output, "UID")
		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "Markers")
		assert.Contains(t, output, "Verified")
		assert.Contains(t, output, "Hidden")
		assert.Contains(t, output, "Files")
		assert.Contains(t, output, "Photos")
	})
	t.Run("JSON", func(t *testing.T) {
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects", "--json"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))

		require.NotEmpty(t, rows)
		assert.Contains(t, rows[0], "verified")
		assert.Contains(t, rows[0], "hidden")
		assert.Contains(t, rows[0], "files")
	})
	t.Run("ByPerson", func(t *testing.T) {
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects", "--json", "Actress A"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		require.Len(t, rows, 1)
		assert.Equal(t, "Actress A", rows[0]["name"])
	})
	t.Run("Stored", func(t *testing.T) {
		// Same shape, without the pass over markers and files that counting live costs.
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects", "--stored", "--json"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		require.NotEmpty(t, rows)
		assert.Contains(t, rows[0], "files")
		assert.Contains(t, rows[0], "photos")
	})
	t.Run("CountIsBounded", func(t *testing.T) {
		// An out-of-range count falls back to the API default rather than fetching everything.
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects", "--count", "0", "--json"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		assert.LessOrEqual(t, len(rows), 100)
	})
}

// TestFacesListCommand covers the cluster report.
func TestFacesListCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesListCommand, []string{"ls"})
		require.NoError(t, err)

		// Samples against the live marker count is the pair worth reading: the first is what the
		// cluster was built from, the second is what points at it now.
		for _, col := range []string{"ID", "Name", "Subject", "Src", "Kind", "Markers", "Samples", "Radius", "Collisions", "Collision Radius", "Matched At"} {
			assert.Contains(t, output, col)
		}

		// The kind is named rather than numbered, so it reads without a lookup table. Kind sits
		// left of the counts so everything to its right is a number.
		assert.Less(t, strings.Index(output, "Kind"), strings.Index(output, "Markers"))
		assert.Contains(t, output, "regular", "a cluster states its kind rather than leaving the column unset")
	})
	t.Run("ByPerson", func(t *testing.T) {
		output, err := RunWithTestContext(FacesListCommand, []string{"ls", "--json", "Actress A"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		require.NotEmpty(t, rows)

		for _, row := range rows {
			assert.Equal(t, "Actress A", row["name"])
		}
	})
	t.Run("Markdown", func(t *testing.T) {
		output, err := RunWithTestContext(FacesListCommand, []string{"ls", "--md"})
		require.NoError(t, err)

		assert.Contains(t, output, "|")
		assert.Contains(t, output, "Samples")
	})
}

// TestFacesMarkersCommand covers the marker report and its filters.
func TestFacesMarkersCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesMarkersCommand, []string{"markers"})
		require.NoError(t, err)

		for _, col := range []string{"Marker", "Name", "Size", "Score", "Subject", "Src", "Face", "Dist", "Embedding", "Landmarks", "Invalid", "File", "Matched At"} {
			assert.Contains(t, output, col)
		}

		// The vectors are measured, not printed: they are most of the row and none of what a
		// diagnosis reads.
		assert.NotContains(t, output, "embeddings_json")
		assert.NotContains(t, output, "0.1234")
	})
	t.Run("Dangling", func(t *testing.T) {
		// The fixtures are consistent, so this selects nothing - which is the point: the filter
		// exists to answer "is anything pointing at a cluster that is gone" without hand-written SQL.
		output, err := RunWithTestContext(FacesMarkersCommand, []string{"markers", "--dangling", "--json"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		assert.Empty(t, rows)
	})
	t.Run("UnknownPerson", func(t *testing.T) {
		output, err := RunWithTestContext(FacesMarkersCommand, []string{"markers", "--json", "Nobody By That Name"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		assert.Empty(t, rows)
	})
	t.Run("ByPerson", func(t *testing.T) {
		// The argument is what makes inspecting one person possible without piping through grep.
		output, err := RunWithTestContext(FacesMarkersCommand, []string{"markers", "--json", "Actress A"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		assert.NotEmpty(t, rows)
	})
}

// TestReportVectors covers the vector columns, whose empty and invalid readings mean different
// things: a marker that was never embedded, and one whose stored vector cannot be parsed.
func TestReportVectors(t *testing.T) {
	assert.Equal(t, "128", reportVectors(128))
	assert.Equal(t, "5", reportVectors(5))
	assert.Equal(t, "", reportVectors(0), "an absent vector reads as blank, not as a zero width")
	assert.Equal(t, "invalid", reportVectors(query.InvalidJSON))
}

// TestReportBool covers the flag rendering, which uses the shared labels so a column of them reads
// the same as in every other report.
func TestReportBool(t *testing.T) {
	assert.Equal(t, report.Yes, reportBool(true))
	assert.Equal(t, report.No, reportBool(false))
}
