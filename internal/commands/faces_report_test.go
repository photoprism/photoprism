package commands

import (
	"encoding/json"
	"flag"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// newCommandContext parses args against a subcommand's own flags, which the shared test context
// does not do: it applies the app's flags, not a subcommand's.
func newCommandContext(t *testing.T, cmd *cli.Command, args ...string) *cli.Context {
	t.Helper()

	flagSet := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	for _, f := range cmd.Flags {
		require.NoError(t, f.Apply(flagSet))
	}

	require.NoError(t, flagSet.Parse(args))

	return cli.NewContext(cli.NewApp(), flagSet, nil)
}

// TestReportPerson covers how a report argument is sanitized before it becomes a filter.
func TestReportPerson(t *testing.T) {
	t.Run("Joined", func(t *testing.T) {
		assert.Equal(t, "Jane Roe", reportPerson(newCommandContext(t, FacesSubjectsCommand, "Jane", "Roe")))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, reportPerson(newCommandContext(t, FacesSubjectsCommand)))
	})
	t.Run("KeepsUnderscore", func(t *testing.T) {
		// A stored name may hold one, so the filter has to keep it and escape it downstream.
		assert.Equal(t, "Ann_Marie", reportPerson(newCommandContext(t, FacesSubjectsCommand, "Ann_Marie")))
	})
	t.Run("SanitizedLikeAStoredName", func(t *testing.T) {
		// clean.Name is what Subject.SetName applies, so the argument and the column it is matched
		// against have been through the same filter. Not clean.SearchString, which turns "%" into
		// the wildcard "*" and would search for a character no stored name can contain.
		assert.Equal(t, "Bob 100", reportPerson(newCommandContext(t, FacesSubjectsCommand, "Bob 100%")))
	})
	t.Run("SubjectUID", func(t *testing.T) {
		assert.Equal(t, "js6sg6b1h1njaaac", reportPerson(newCommandContext(t, FacesSubjectsCommand, "js6sg6b1h1njaaac")))
	})
}

// TestFacesSubjectsCommand covers the people report, whose reason for existing is the pair of
// count columns: the stored ones and the ones the markers currently support drift, and the drift is
// what keeps a person off People > Recognized after a CLI-only reset and re-cluster.
func TestFacesSubjectsCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesSubjectsCommand, []string{"subjects"})
		require.NoError(t, err)

		for _, col := range []string{"Subject", "Name", "Src", "Favorite", "Verified", "Hidden", "Markers", "Clusters", "Files", "Photos", "Created At"} {
			assert.Contains(t, output, col)
		}

		// A cluster holds markers, so the count of the parts comes before the count of the wholes.
		assert.Less(t, strings.Index(output, "Markers"), strings.Index(output, "Clusters"))
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
		assert.Contains(t, rows[0], "favorite")
		assert.Contains(t, rows[0], "verified")
		assert.Contains(t, rows[0], "hidden")
		assert.Contains(t, rows[0], "files")
		assert.Contains(t, rows[0], "clusters")
		assert.Contains(t, rows[0], "created_at")
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
		// Asserted on reportPaging rather than on the row count: the fixtures hold fewer than 100
		// people, so every bound produces the same output and the command cannot tell them apart.
		for _, c := range []struct {
			args          []string
			count, offset int
		}{
			{nil, 100, 0},
			{[]string{"--count", "0"}, 100, 0},
			{[]string{"--count", "2000"}, 100, 0},
			{[]string{"--count", "250"}, 250, 0},
			{[]string{"--offset", "-5"}, 100, 0},
			{[]string{"--count", "10", "--offset", "20"}, 10, 20},
		} {
			count, offset := reportPaging(newCommandContext(t, FacesSubjectsCommand, c.args...))

			assert.Equal(t, c.count, count, "count for %v", c.args)
			assert.Equal(t, c.offset, offset, "offset for %v", c.args)
		}
	})
}

// TestFacesListCommand covers the cluster report.
func TestFacesListCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesListCommand, []string{"ls"})
		require.NoError(t, err)

		// Samples against the live marker count is the pair worth reading: the first is what the
		// cluster was built from, the second is what points at it now.
		for _, col := range []string{"Face", "Name", "Subject", "Src", "Kind", "Markers", "Samples", "Radius", "Collisions", "Collision Radius", "Matched At"} {
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

func TestFacesConflictsCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(FacesConflictsCommand, []string{"conflicts"})
		require.NoError(t, err)

		for _, col := range []string{"Face", "Name", "Subject", "Face 2", "Name 2", "Subject 2", "Resolution", "Dist", "Accept", "Accept 2", "Samples", "Samples 2"} {
			assert.Contains(t, output, col)
		}
	})
	t.Run("NotesBelowTheTable", func(t *testing.T) {
		output, err := RunWithTestContext(FacesConflictsCommand, []string{"conflicts"})
		require.NoError(t, err)

		// What the table cannot say has to be on the same page as the table, or an empty one
		// reads as "no conflicts" when it may mean "none of them was compared".
		assert.Contains(t, output, "Compared")
		assert.Contains(t, output, "retires a cluster as ambiguous")
	})
	t.Run("JSONCarriesTheNotes", func(t *testing.T) {
		// A saved report has to keep them: dropping the notes for machine output would strip
		// exactly the part that says what the rows do not cover.
		output, err := RunWithTestContext(FacesConflictsCommand, []string{"conflicts", "--json"})
		require.NoError(t, err)

		var result struct {
			Conflicts []map[string]any `json:"conflicts"`
			Scan      map[string]int   `json:"scan"`
			Notes     []string         `json:"notes"`
		}

		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.NotEmpty(t, result.Notes)
		assert.Contains(t, result.Scan, "clusters")
		assert.Contains(t, result.Scan, "compared")
	})
	t.Run("UnknownPerson", func(t *testing.T) {
		output, err := RunWithTestContext(FacesConflictsCommand, []string{"conflicts", "--json", "Nobody By That Name"})
		require.NoError(t, err)

		var result struct {
			Conflicts []map[string]any `json:"conflicts"`
			Scan      map[string]int   `json:"scan"`
		}

		require.NoError(t, json.Unmarshal([]byte(output), &result))
		assert.Empty(t, result.Conflicts)
		assert.Zero(t, result.Scan["clusters"])
	})
	t.Run("InvalidFormat", func(t *testing.T) {
		_, err := RunWithTestContext(FacesConflictsCommand, []string{"conflicts", "--format=nonsense"})
		assert.Error(t, err)
	})
}

func TestReportResolution(t *testing.T) {
	named := "js6sg6b1qekk9jx8"
	t.Run("Ambiguous", func(t *testing.T) {
		assert.Equal(t, "ambiguous", reportResolution(query.FaceConflict{SubjUID: named, Dist: face.AmbiguityDist() / 2}))
	})
	t.Run("Narrow", func(t *testing.T) {
		assert.Equal(t, "narrow", reportResolution(query.FaceConflict{SubjUID: named, Dist: face.AmbiguityDist() * 10}))
	})
	t.Run("Unmeasured", func(t *testing.T) {
		// Match never reports a negative distance alongside a match, so this is a guard: what
		// matters is that an unmeasured pair does not claim a cluster would be narrowed.
		assert.NotEqual(t, "narrow", reportResolution(query.FaceConflict{SubjUID: named, Dist: -1}))
	})
	t.Run("AnonymousFirstCluster", func(t *testing.T) {
		// ResolveCollision returns early on a cluster that names nobody, so neither branch runs
		// however close the pair is. Naming one here would assert an outcome that cannot happen.
		assert.Equal(t, "none", reportResolution(query.FaceConflict{Dist: face.AmbiguityDist() / 2}))
		assert.Equal(t, "none", reportResolution(query.FaceConflict{Dist: face.AmbiguityDist() * 10}))
	})
	t.Run("AnonymousSecondCluster", func(t *testing.T) {
		// The other orientation does resolve: the named cluster is the receiver, and it narrows
		// against the anonymous one's vector.
		assert.Equal(t, "narrow", reportResolution(query.FaceConflict{
			SubjUID: "js6sg6b1qekk9jx8", OtherSubjUID: "", Dist: face.AmbiguityDist() * 10,
		}))
	})
}

func TestFaceConflictRows(t *testing.T) {
	c := query.FaceConflict{
		ID: "FACE1", SubjName: "Alice", SubjUID: "js6sg6b1qekk9jx8", Samples: 7, Accept: 0.95,
		OtherID: "FACE2", OtherSubjName: "Bob", OtherSubjUID: "js6sg6b1h1njaaab", OtherSamples: 3, OtherAccept: 0.42,
		Dist: face.CollisionDist + face.Epsilon + 0.5,
	}
	t.Run("EveryRowFillsEveryColumn", func(t *testing.T) {
		// Nothing else pins the row against its headers, so a transposed pair of cells - Accept
		// against Accept 2, Samples against Samples 2 - would ship silently.
		rows := faceConflictRows([]query.FaceConflict{c, {}})
		require.Len(t, rows, 2)
		for i, row := range rows {
			assert.Len(t, row, len(faceConflictCols()), "row %d", i)
		}
	})
	t.Run("ValuesLandInTheNamedColumn", func(t *testing.T) {
		row := faceConflictRows([]query.FaceConflict{c})[0]
		cols := faceConflictCols()
		got := make(map[string]string, len(cols))
		for i, col := range cols {
			got[col] = row[i]
		}
		assert.Equal(t, "FACE1", got["Face"])
		assert.Equal(t, "Alice", got["Name"])
		assert.Equal(t, "js6sg6b1qekk9jx8", got["Subject"])
		assert.Equal(t, "FACE2", got["Face 2"])
		assert.Equal(t, "Bob", got["Name 2"])
		assert.Equal(t, "js6sg6b1h1njaaab", got["Subject 2"])
		assert.Equal(t, "narrow", got["Resolution"])
		assert.Equal(t, report.Distance(c.Accept), got["Accept"])
		assert.Equal(t, report.Distance(c.OtherAccept), got["Accept 2"])
		assert.Equal(t, "7", got["Samples"])
		assert.Equal(t, "3", got["Samples 2"])
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, faceConflictRows(nil))
	})
}

func TestPrintFaceConflictsJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cols := faceConflictCols()
		rows := faceConflictRows([]query.FaceConflict{{ID: "FACE1", OtherID: "FACE2"}})
		var printErr error
		output := capture.Output(func() {
			printErr = printFaceConflictsJSON(rows, cols,
				query.FaceConflictScan{Clusters: 2, Compared: 1}, []string{"a note"})
		})
		require.NoError(t, printErr)

		var result struct {
			Conflicts []map[string]any `json:"conflicts"`
			Scan      map[string]int   `json:"scan"`
			Notes     []string         `json:"notes"`
		}

		require.NoError(t, json.Unmarshal([]byte(output), &result))
		require.Len(t, result.Conflicts, 1)
		assert.Equal(t, "FACE1", result.Conflicts[0]["face"])
		assert.Equal(t, 1, result.Scan["compared"])
		assert.Equal(t, []string{"a note"}, result.Notes)
	})
}

func TestFaceConflictNotes(t *testing.T) {
	t.Run("AlwaysStatesWhatWasCompared", func(t *testing.T) {
		lines := faceConflictNotes(query.FaceConflictScan{Clusters: 3, Compared: 7}, query.FaceConflictNotes{})
		require.Len(t, lines, 2)
		assert.Contains(t, lines[0], "7 pairs")
		assert.Contains(t, lines[0], "3 clusters")
		assert.Contains(t, lines[1], "ambiguous")
	})
	t.Run("EveryNote", func(t *testing.T) {
		lines := faceConflictNotes(query.FaceConflictScan{},
			query.FaceConflictNotes{Ambiguous: 1, Hidden: 2, InertRadius: 3, BelowOwnSpread: 4})
		require.Len(t, lines, 6)
		joined := strings.Join(lines, "\n")
		assert.Contains(t, joined, "1 cluster in the index is already retired")
		assert.Contains(t, joined, "2 clusters in the index are hidden")
		assert.Contains(t, joined, "3 compared clusters record")
		assert.Contains(t, joined, "4 compared clusters record")
	})
}
