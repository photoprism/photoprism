package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
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

func TestFaceMarkerModelBoundaries(t *testing.T) {
	restore := face.ConfiguredModel()
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace}))
	beforeUnmatched := CountUnmatchedFaceMarkers()

	compatible := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		MarkerType:     entity.MarkerFace,
		EmbedModel:     face.ModelSFace,
		EmbeddingsJSON: face.Embeddings{{0.1, 0.2}}.JSON(),
	}
	legacy := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		MarkerType:     entity.MarkerFace,
		EmbeddingsJSON: face.Embeddings{{0.1, 0.2, 0.3}}.JSON(),
	}
	require.NoError(t, entity.Db().Create(compatible).Error)
	require.NoError(t, entity.Db().Create(legacy).Error)
	t.Cleanup(func() {
		entity.UnscopedDb().Delete(compatible)
		entity.UnscopedDb().Delete(legacy)
	})

	unmatched, err := UnmatchedFaceMarkers(1000, 0, nil)
	require.NoError(t, err)
	foundCompatible, foundLegacy := false, false
	for _, marker := range unmatched {
		foundCompatible = foundCompatible || marker.MarkerUID == compatible.MarkerUID
		foundLegacy = foundLegacy || marker.MarkerUID == legacy.MarkerUID
	}
	assert.True(t, foundCompatible)
	assert.False(t, foundLegacy)
	assert.Equal(t, beforeUnmatched+1, CountUnmatchedFaceMarkers())

	all, err := FaceMarkers(1000, 0)
	require.NoError(t, err)
	foundCompatible, foundLegacy = false, false
	for _, marker := range all {
		foundCompatible = foundCompatible || marker.MarkerUID == compatible.MarkerUID
		foundLegacy = foundLegacy || marker.MarkerUID == legacy.MarkerUID
	}
	assert.True(t, foundCompatible)
	assert.False(t, foundLegacy)

	embeddings, err := Embeddings(false, false, 0, 0, face.ModelSFace)
	require.NoError(t, err)
	foundCompatible, foundLegacy = false, false
	for _, embedding := range embeddings {
		foundCompatible = foundCompatible || len(embedding) == 2
		foundLegacy = foundLegacy || len(embedding) == 3
	}
	assert.True(t, foundCompatible)
	assert.False(t, foundLegacy)
}

func TestFaceMarkersWithoutConfiguredModel(t *testing.T) {
	restore := face.ConfiguredModel()
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	recorded := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		MarkerType:     entity.MarkerFace,
		EmbedModel:     face.ModelSFace,
		EmbeddingsJSON: face.Embeddings{{0.1, 0.2}}.JSON(),
	}
	require.NoError(t, entity.Db().Create(recorded).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(recorded) })

	// An embedder that fails to initialize leaves no model name behind, which must not
	// narrow matching to the legacy rows that predate the provenance column.
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))
	require.Equal(t, "", face.EmbeddingModelName())

	found := func(markers entity.Markers) bool {
		for _, marker := range markers {
			if marker.MarkerUID == recorded.MarkerUID {
				return true
			}
		}

		return false
	}

	t.Run("UnmatchedFaceMarkers", func(t *testing.T) {
		markers, err := UnmatchedFaceMarkers(1000, 0, nil)
		require.NoError(t, err)
		assert.True(t, found(markers))
	})
	t.Run("FaceMarkers", func(t *testing.T) {
		markers, err := FaceMarkers(1000, 0)
		require.NoError(t, err)
		assert.True(t, found(markers))
	})
	t.Run("CountUnmatchedFaceMarkers", func(t *testing.T) {
		var expected int
		require.NoError(t, entity.Db().Model(&entity.Markers{}).
			Where("matched_at IS NULL AND marker_invalid = 0 AND embeddings_json <> ''").
			Where("marker_type = ?", entity.MarkerFace).
			Count(&expected).Error)
		assert.Equal(t, expected, CountUnmatchedFaceMarkers())
	})
	t.Run("CountNewFaceMarkers", func(t *testing.T) {
		var expected int
		require.NoError(t, entity.Db().Model(&entity.Markers{}).
			Where("marker_type = ?", entity.MarkerFace).
			Where("face_id = '' AND marker_invalid = 0 AND embeddings_json <> ''").
			Count(&expected).Error)
		assert.Equal(t, expected, CountNewFaceMarkers(0, 0))
	})
}

func TestEmbeddings(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		results, err := Embeddings(false, false, 0, 0, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, face.Embedding{}, val)
		}
	})
	t.Run("Size", func(t *testing.T) {
		results, err := Embeddings(false, false, 230, 0, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(results), 8)

		for _, val := range results {
			assert.IsType(t, face.Embedding{}, val)
		}
	})
	t.Run("Score", func(t *testing.T) {
		results, err := Embeddings(false, false, 0, 50, "")

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

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(0))
}

func TestRemoveNonExistentMarkerFaces(t *testing.T) {
	affected, err := RemoveNonExistentMarkerFaces()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
}

func TestRemoveNonExistentMarkerSubjects(t *testing.T) {
	affected, err := RemoveNonExistentMarkerSubjects()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
}

func TestFixMarkerReferences(t *testing.T) {
	affected, err := FixMarkerReferences()

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
