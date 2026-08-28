package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/txt/report"
)

// TestFacesSubjectsCommand covers the people report, whose reason for existing is the pair of
// count columns: the stored ones and the ones the markers currently support drift, and the drift is
// what keeps a person off People > Recognized after a CLI-only reset and re-cluster.
func TestFacesSubjectsCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects"})
		require.NoError(t, err)

		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "Verified")
		assert.Contains(t, output, "Files")
		assert.Contains(t, output, "Live Files")
		assert.Contains(t, output, "Markers")
	})
	t.Run("JSON", func(t *testing.T) {
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects", "--json"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))

		if len(rows) > 0 {
			assert.Contains(t, rows[0], "verified")
			assert.Contains(t, rows[0], "live_files")
		}
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
		assert.Contains(t, output, "Samples")
		assert.Contains(t, output, "Markers")
		assert.Contains(t, output, "Collision Radius")
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

		assert.Contains(t, output, "Marker")
		assert.Contains(t, output, "Cluster")
		// Embeddings are most of the row and none of what a diagnosis reads.
		assert.NotContains(t, output, "embeddings")
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
	t.Run("UnknownSubject", func(t *testing.T) {
		output, err := RunWithTestContext(FacesMarkersCommand, []string{"markers", "--subject", "js6sg6b1qekk0000", "--json"})
		require.NoError(t, err)

		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		assert.Empty(t, rows)
	})
}

// TestReportBool covers the flag rendering, which uses the shared labels so a column of them reads
// the same as in every other report.
func TestReportBool(t *testing.T) {
	assert.Equal(t, report.Yes, reportBool(true))
	assert.Equal(t, report.No, reportBool(false))
}
