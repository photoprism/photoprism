package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestSubjectReports covers the people report, whose point is the two count columns.
func TestSubjectReports(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		people, err := SubjectReports("", 100, 0, true)
		require.NoError(t, err)
		require.NotEmpty(t, people)

		byUID := make(map[string]SubjectReport, len(people))
		for _, p := range people {
			byUID[p.SubjUID] = p
			assert.NotEmpty(t, p.SubjName, "a report row has to name its person")
		}

		// The persisted counts are whatever the last refresh left; the live ones are computed
		// here, so a subject with markers has to report them even when the stored count is stale.
		known := entity.SubjectFixtures.Get("actress-1")
		got, ok := byUID[known.SubjUID]
		require.True(t, ok, "a person with markers has to appear")
		assert.Positive(t, got.Markers, "the live marker count is computed, not read from the row")
	})
	t.Run("Stored", func(t *testing.T) {
		// The cheap variant skips the join over markers and files, which is half a second on a
		// large library, and reports whatever the last refresh left on the row instead.
		stored, err := SubjectReports("", 100, 0, false)
		require.NoError(t, err)
		require.NotEmpty(t, stored)

		live, err := SubjectReports("", 100, 0, true)
		require.NoError(t, err)
		require.Len(t, stored, len(live), "the two variants describe the same people")

		for i := range stored {
			assert.Equal(t, live[i].SubjUID, stored[i].SubjUID)
			assert.Equal(t, live[i].Markers, stored[i].Markers, "the marker count is live either way")
		}
	})
	t.Run("Paginates", func(t *testing.T) {
		first, err := SubjectReports("", 1, 0, true)
		require.NoError(t, err)
		require.Len(t, first, 1)

		second, err := SubjectReports("", 1, 1, true)
		require.NoError(t, err)
		require.Len(t, second, 1)

		assert.NotEqual(t, first[0].SubjUID, second[0].SubjUID, "an offset has to move the window")
	})
	t.Run("OffsetPastTheEnd", func(t *testing.T) {
		people, err := SubjectReports("", 10, 100000, true)
		require.NoError(t, err)
		assert.Empty(t, people)
	})
}

// TestFaceReports covers the cluster report.
func TestFaceReports(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		faces, err := FaceReports("", 100, 0)
		require.NoError(t, err)
		require.NotEmpty(t, faces)

		for _, f := range faces {
			assert.NotEmpty(t, f.ID)
		}

		// Ordered by the samples a cluster was built from, so the widest is first and two runs of
		// the report are comparable.
		for i := 1; i < len(faces); i++ {
			assert.GreaterOrEqual(t, faces[i-1].Samples, faces[i].Samples)
		}
	})
	t.Run("NamesTheSubject", func(t *testing.T) {
		faces, err := FaceReports("", 100, 0)
		require.NoError(t, err)

		var named int
		for _, f := range faces {
			if f.SubjName != "" {
				named++
			}
		}

		assert.Positive(t, named, "a cluster with a subject has to report the person's name")
	})
	t.Run("OffsetPastTheEnd", func(t *testing.T) {
		faces, err := FaceReports("", 10, 100000)
		require.NoError(t, err)
		assert.Empty(t, faces)
	})
}

// TestMarkerReports covers the marker report and each filter it offers.
func TestMarkerReports(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		markers, err := MarkerReports(MarkerReportFilter{Count: 10})
		require.NoError(t, err)
		require.NotEmpty(t, markers)

		for _, m := range markers {
			assert.NotEmpty(t, m.MarkerUID)
		}
	})
	t.Run("BySubject", func(t *testing.T) {
		subjUID := entity.SubjectFixtures.Get("actress-1").SubjUID

		markers, err := MarkerReports(MarkerReportFilter{Person: subjUID, Count: 100})
		require.NoError(t, err)
		require.NotEmpty(t, markers)

		for _, m := range markers {
			assert.Equal(t, subjUID, m.SubjUID)
		}
	})
	t.Run("ByFace", func(t *testing.T) {
		faceID := entity.FaceFixtures.Get("actress-1").ID

		markers, err := MarkerReports(MarkerReportFilter{FaceID: faceID, Count: 100})
		require.NoError(t, err)
		require.NotEmpty(t, markers)

		for _, m := range markers {
			assert.Equal(t, faceID, m.FaceID)
		}
	})
	t.Run("Unassigned", func(t *testing.T) {
		markers, err := MarkerReports(MarkerReportFilter{Unassigned: true, Count: 100})
		require.NoError(t, err)

		for _, m := range markers {
			assert.NotEmpty(t, m.SubjUID)
			assert.Empty(t, m.FaceID)
		}
	})
	t.Run("Dangling", func(t *testing.T) {
		markers, err := MarkerReports(MarkerReportFilter{Dangling: true, Count: 100})
		require.NoError(t, err)

		// Every one it returns must really point at a cluster that is gone.
		for _, m := range markers {
			require.NotEmpty(t, m.FaceID)
			assert.Nil(t, entity.FindFace(m.FaceID), "a dangling marker names a cluster that no longer exists")
		}
	})
	t.Run("ExcludesNonFaceMarkers", func(t *testing.T) {
		markers, err := MarkerReports(MarkerReportFilter{Count: 1000})
		require.NoError(t, err)

		for _, m := range markers {
			var stored entity.Marker
			require.NoError(t, UnscopedDb().Where("marker_uid = ?", m.MarkerUID).First(&stored).Error)
			assert.Equal(t, entity.MarkerFace, stored.MarkerType)
		}
	})
}

// TestEmbeddingDims covers the width a marker report shows, whose two zero-ish answers mean
// different things: a marker that was never embedded and one whose stored vector is broken.
func TestEmbeddingDims(t *testing.T) {
	t.Run("Absent", func(t *testing.T) {
		assert.Equal(t, 0, embeddingDims(nil))
		assert.Equal(t, 0, embeddingDims([]byte{}))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, InvalidJSON, embeddingDims([]byte("not json")))
		assert.Equal(t, InvalidJSON, embeddingDims([]byte(`{"a":1}`)))
	})
	t.Run("Dims", func(t *testing.T) {
		assert.Equal(t, 3, embeddingDims([]byte(`[[0.1,0.2,0.3]]`)))
		assert.Equal(t, 2, embeddingDims([]byte(`[[0.1,0.2],[0.3,0.4]]`)))
	})
	t.Run("EmptyArray", func(t *testing.T) {
		assert.Equal(t, 0, embeddingDims([]byte(`[]`)))
	})
}

// TestLandmarkCount covers the landmark column, which follows the same conventions.
func TestLandmarkCount(t *testing.T) {
	t.Run("Absent", func(t *testing.T) {
		assert.Equal(t, 0, landmarkCount(nil))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, InvalidJSON, landmarkCount([]byte("{")))
	})
	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, 2, landmarkCount([]byte(`[{"Name":"eye_l"},{"Name":"eye_r"}]`)))
		assert.Equal(t, 0, landmarkCount([]byte(`[]`)))
	})
}

// TestMarkerReports_Vectors covers that the report measures the stored vectors rather than
// returning them, since the vectors are most of the row and none of what a diagnosis reads.
func TestMarkerReports_Vectors(t *testing.T) {
	markers, err := MarkerReports(MarkerReportFilter{Count: 100})
	require.NoError(t, err)
	require.NotEmpty(t, markers)

	var embedded int

	for _, m := range markers {
		assert.NotEqual(t, InvalidJSON, m.EmbeddingDims, "a fixture must not hold unparsable embeddings")
		assert.NotEqual(t, InvalidJSON, m.Landmarks, "nor unparsable landmarks")

		if m.EmbeddingDims > 0 {
			embedded++
		}
	}

	assert.Positive(t, embedded, "the fixtures have to include an embedded marker for this to mean anything")
}

// TestPersonFilter covers how a report argument is read, which decides whether "js6sg6b..." selects
// one person or is searched for as a name.
func TestPersonFilter(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		uid, like := PersonFilter("")
		assert.Empty(t, uid)
		assert.Empty(t, like)

		uid, like = PersonFilter("   ")
		assert.Empty(t, uid)
		assert.Empty(t, like)
	})
	t.Run("SubjectUID", func(t *testing.T) {
		uid, like := PersonFilter(entity.SubjectFixtures.Get("actress-1").SubjUID)
		assert.Equal(t, entity.SubjectFixtures.Get("actress-1").SubjUID, uid)
		assert.Empty(t, like, "a uid selects one person rather than being searched for")
	})
	t.Run("Name", func(t *testing.T) {
		uid, like := PersonFilter("Actress")
		assert.Empty(t, uid)
		assert.Equal(t, "%Actress%", like)
	})
	// A name is matched literally: an operator typing a name that holds one of these is looking for
	// that person, not writing a pattern.
	t.Run("EscapesWildcards", func(t *testing.T) {
		_, like := PersonFilter("50%")
		assert.Equal(t, `%50\%%`, like)

		_, like = PersonFilter("a_b")
		assert.Equal(t, `%a\_b%`, like)

		_, like = PersonFilter(`back\slash`)
		assert.Equal(t, `%back\\slash%`, like)
	})
	t.Run("UIDOfAnotherType", func(t *testing.T) {
		// Only a subject uid selects by id; a marker uid is a name nobody has.
		uid, like := PersonFilter(entity.MarkerFixtures.Get("actress-a-1").MarkerUID)
		assert.Empty(t, uid)
		assert.NotEmpty(t, like)
	})
}

// TestFaceReports_Person covers narrowing the cluster report, so one person can be inspected
// without piping the output through grep.
func TestFaceReports_Person(t *testing.T) {
	name := entity.SubjectFixtures.Get("actress-1").SubjName
	subjUID := entity.SubjectFixtures.Get("actress-1").SubjUID

	t.Run("ByName", func(t *testing.T) {
		faces, err := FaceReports(name, 100, 0)
		require.NoError(t, err)
		require.NotEmpty(t, faces)

		for _, f := range faces {
			assert.Equal(t, name, f.SubjName)
		}
	})
	t.Run("BySubjectUID", func(t *testing.T) {
		faces, err := FaceReports(subjUID, 100, 0)
		require.NoError(t, err)
		require.NotEmpty(t, faces)

		for _, f := range faces {
			assert.Equal(t, subjUID, f.SubjUID)
		}
	})
	t.Run("NoMatch", func(t *testing.T) {
		faces, err := FaceReports("Nobody By That Name", 100, 0)
		require.NoError(t, err)
		assert.Empty(t, faces)
	})
}

// TestSubjectReports_Person covers narrowing the people report.
func TestSubjectReports_Person(t *testing.T) {
	known := entity.SubjectFixtures.Get("actress-1")

	t.Run("ByName", func(t *testing.T) {
		people, err := SubjectReports(known.SubjName, 100, 0, true)
		require.NoError(t, err)
		require.Len(t, people, 1)
		assert.Equal(t, known.SubjUID, people[0].SubjUID)
	})
	t.Run("BySubjectUID", func(t *testing.T) {
		people, err := SubjectReports(known.SubjUID, 100, 0, false)
		require.NoError(t, err)
		require.Len(t, people, 1)
		assert.Equal(t, known.SubjUID, people[0].SubjUID)
	})
	t.Run("PartialName", func(t *testing.T) {
		people, err := SubjectReports("ctress", 100, 0, true)
		require.NoError(t, err)
		assert.NotEmpty(t, people, "a fragment matches anywhere in the name")
	})
	t.Run("NoMatch", func(t *testing.T) {
		people, err := SubjectReports("Nobody By That Name", 100, 0, true)
		require.NoError(t, err)
		assert.Empty(t, people)
	})
}

// TestMarkerReports_Person covers narrowing the marker report by person.
func TestMarkerReports_Person(t *testing.T) {
	known := entity.SubjectFixtures.Get("actress-1")

	t.Run("ByName", func(t *testing.T) {
		markers, err := MarkerReports(MarkerReportFilter{Person: known.SubjName, Count: 100})
		require.NoError(t, err)
		require.NotEmpty(t, markers)

		for _, m := range markers {
			assert.Equal(t, known.SubjUID, m.SubjUID)
		}
	})
	t.Run("NoMatch", func(t *testing.T) {
		markers, err := MarkerReports(MarkerReportFilter{Person: "Nobody By That Name", Count: 100})
		require.NoError(t, err)
		assert.Empty(t, markers)
	})
}
