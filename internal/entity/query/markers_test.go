package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestMarkerByUID(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		if m, err := MarkerByUID("ms6sg6b1wowuy888"); err != nil {
			t.Fatal(err)
		} else if m == nil {
			t.Fatal("result is nil")
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		if _, err := MarkerByUID("mt9k3aa1wowuy888"); err == nil {
			t.Fatal("error expected")
		}
	})
}

func TestMarkers(t *testing.T) {
	t.Run("FindUmatched", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, false, entity.Now())

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("FindAll", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, false, time.Time{})

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("FindEmbeddings", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, true, false, time.Time{})

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("FindFalse", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, true, time.Time{})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(results))
	})
}

func TestUnmatchedFaceMarkers(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		results, err := UnmatchedFaceMarkers(3, 0, nil)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(results))
	})
	t.Run("Before", func(t *testing.T) {
		results, err := UnmatchedFaceMarkers(3, 0, entity.TimeStamp())

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(results))
	})
}

func TestFaceMarkers(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		results, err := FaceMarkers(3, 0)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(results))
	})
}

func TestEmbeddings(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		results, err := Embeddings(false, false, 0, 0)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, face.Embedding{}, val)
		}
	})
	t.Run("Size", func(t *testing.T) {
		results, err := Embeddings(false, false, 230, 0)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(results), 8)

		for _, val := range results {
			assert.IsType(t, face.Embedding{}, val)
		}
	})
	t.Run("Score", func(t *testing.T) {
		results, err := Embeddings(false, false, 0, 50)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, face.Embedding{}, val)
		}
	})
}

func TestMarkerCountsByFaceIDs(t *testing.T) {
	counts, err := MarkerCountsByFaceIDs(nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, counts)

	faces, err := Faces(false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(faces) == 0 {
		t.Skip("no faces available in test dataset")
	}

	ids := []string{faces[0].ID}

	counts, err = MarkerCountsByFaceIDs(ids)
	if err != nil {
		t.Fatal(err)
	}

	if len(counts) == 0 {
		t.Skip("no markers found for sampled face")
	}

	assert.GreaterOrEqual(t, counts[faces[0].ID], 0)
}

func TestRemoveInvalidMarkerReferences(t *testing.T) {
	affected, err := RemoveInvalidMarkerReferences()
	t.Cleanup(func() {
		for _, marker := range entity.MarkerFixtures {
			require.NoError(t, UnscopedDb().Model(&marker).UpdateColumns(entity.Values{"subj_uid": marker.SubjUID, "face_id": marker.FaceID, "face_dist": marker.FaceDist, "matched_at": marker.MatchedAt, "updated_at": marker.UpdatedAt}).Error)
		}
	})

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(0))
}

func TestRemoveNonExistentMarkerFaces(t *testing.T) {
	// Make sure that the data is valid for the test.
	_, err := RemoveAutoFaceClusters()
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, face := range entity.FaceFixtures {
			if face.FaceSrc == entity.SrcAuto {
				require.NoError(t, UnscopedDb().Create(&face).Error)
			}
		}
	})

	affected, err := RemoveNonExistentMarkerFaces()
	t.Cleanup(func() {
		for _, marker := range entity.MarkerFixtures {
			if marker.MarkerType == entity.MarkerFace {
				require.NoError(t, UnscopedDb().Model(&marker).UpdateColumns(entity.Values{"face_id": marker.FaceID, "face_dist": marker.FaceDist, "matched_at": marker.MatchedAt, "updated_at": marker.UpdatedAt}).Error)
			}
		}
	})

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
}

func TestRemoveNonExistentMarkerSubjects(t *testing.T) {
	affected, err := RemoveNonExistentMarkerSubjects()
	t.Cleanup(func() {
		for _, marker := range entity.MarkerFixtures {
			require.NoError(t, UnscopedDb().Model(&marker).UpdateColumns(entity.Values{"subj_uid": marker.SubjUID, "face_id": marker.FaceID, "face_dist": marker.FaceDist, "matched_at": marker.MatchedAt, "updated_at": marker.UpdatedAt}).Error)
		}
	})

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
}

func TestFixMarkerReferences(t *testing.T) {
	affected, err := FixMarkerReferences()
	t.Cleanup(func() {
		for _, marker := range entity.MarkerFixtures {
			require.NoError(t, UnscopedDb().Model(&marker).UpdateColumns(entity.Values{"subj_uid": marker.SubjUID, "face_id": marker.FaceID, "face_dist": marker.FaceDist, "matched_at": marker.MatchedAt, "updated_at": marker.UpdatedAt}).Error)
		}
	})

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(0))
}

func TestMarkersWithNonExistentReferences(t *testing.T) {
	f, s, err := MarkersWithNonExistentReferences()

	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(f), 0)
	assert.GreaterOrEqual(t, len(s), 0)
}

func TestMarkersWithSubjectConflict(t *testing.T) {
	m, err := MarkersWithSubjectConflict()

	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(m), 0)
}

func TestCountUnmatchedFaceMarkers(t *testing.T) {
	n := CountUnmatchedFaceMarkers()

	assert.GreaterOrEqual(t, n, 1)
}

func TestCountMarkers(t *testing.T) {
	n := CountMarkers(entity.MarkerFace)

	assert.GreaterOrEqual(t, n, 1)
}
