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
		people, err := SubjectReports(100, 0)
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
	t.Run("Paginates", func(t *testing.T) {
		first, err := SubjectReports(1, 0)
		require.NoError(t, err)
		require.Len(t, first, 1)

		second, err := SubjectReports(1, 1)
		require.NoError(t, err)
		require.Len(t, second, 1)

		assert.NotEqual(t, first[0].SubjUID, second[0].SubjUID, "an offset has to move the window")
	})
	t.Run("OffsetPastTheEnd", func(t *testing.T) {
		people, err := SubjectReports(10, 100000)
		require.NoError(t, err)
		assert.Empty(t, people)
	})
}

// TestFaceReports covers the cluster report.
func TestFaceReports(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		faces, err := FaceReports(100, 0)
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
		faces, err := FaceReports(100, 0)
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
		faces, err := FaceReports(10, 100000)
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

		markers, err := MarkerReports(MarkerReportFilter{SubjUID: subjUID, Count: 100})
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
