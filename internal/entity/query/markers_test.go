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
		results, err := UnmatchedFaceMarkers(3, "", nil)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(results))
	})
	t.Run("Before", func(t *testing.T) {
		results, err := UnmatchedFaceMarkers(3, "", entity.TimeStamp())

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(results))
	})
	t.Run("Cursor", func(t *testing.T) {
		// Paging by cursor is what keeps a run from re-reading markers it visited without
		// stamping, so a page must start after the uid it is given and never repeat one.
		first, err := UnmatchedFaceMarkers(2, "", nil)
		require.NoError(t, err)
		require.Len(t, first, 2)
		require.Less(t, first[0].MarkerUID, first[1].MarkerUID, "a page is ordered by the cursor column")

		next, err := UnmatchedFaceMarkers(2, first[1].MarkerUID, nil)
		require.NoError(t, err)

		for _, m := range next {
			assert.Greater(t, m.MarkerUID, first[1].MarkerUID)
		}
	})
	t.Run("CursorPastTheEnd", func(t *testing.T) {
		results, err := UnmatchedFaceMarkers(2, "zzzzzzzzzzzzzzzz", nil)
		require.NoError(t, err)
		assert.Empty(t, results)
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

	unmatched, err := UnmatchedFaceMarkers(1000, "", nil)
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
		markers, err := UnmatchedFaceMarkers(1000, "", nil)
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
			Where("matched_at IS NULL AND marker_invalid = 0 AND LENGTH(embeddings_json) > 0").
			Where("marker_type = ?", entity.MarkerFace).
			Count(&expected).Error)
		assert.Equal(t, expected, CountUnmatchedFaceMarkers())
	})
	t.Run("CountNewFaceMarkers", func(t *testing.T) {
		var expected int
		require.NoError(t, entity.Db().Model(&entity.Markers{}).
			Where("marker_type = ?", entity.MarkerFace).
			Where("face_id = '' AND marker_invalid = 0 AND LENGTH(embeddings_json) > 0").
			Count(&expected).Error)
		assert.Equal(t, expected, CountNewFaceMarkers(0, 0))
	})
}

func TestFaceMarkersWithEmptyEmbeddings(t *testing.T) {
	restore := face.ConfiguredModel()
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))

	beforeUnmatched := CountUnmatchedFaceMarkers()
	beforeNew := CountNewFaceMarkers(0, 0)

	// Embeddings.JSON returns an empty non-nil slice when there is nothing to store, and
	// the migration writes one to clear a vector. SQLite stores that as a zero-length blob
	// rather than NULL, which an "embeddings_json <> ''" comparison does not exclude.
	empty := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		MarkerType:     entity.MarkerFace,
		EmbedModel:     face.ModelFaceNet,
		EmbeddingsJSON: face.Embeddings{}.JSON(),
	}
	require.NoError(t, entity.Db().Create(empty).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(empty) })

	t.Run("UnmatchedFaceMarkers", func(t *testing.T) {
		markers, err := UnmatchedFaceMarkers(1000, "", nil)
		require.NoError(t, err)

		for _, marker := range markers {
			assert.NotEqual(t, empty.MarkerUID, marker.MarkerUID)
		}
	})
	t.Run("CountUnmatchedFaceMarkers", func(t *testing.T) {
		assert.Equal(t, beforeUnmatched, CountUnmatchedFaceMarkers())
	})
	t.Run("CountNewFaceMarkers", func(t *testing.T) {
		assert.Equal(t, beforeNew, CountNewFaceMarkers(0, 0))
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
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Make sure that the data is valid for the test.
	_, err := RemoveAutoFaceClusters()
	require.NoError(t, err)

	affected, err := RemoveNonExistentMarkerFaces()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
	// Post test cleanup
	entity.ResetTestFixtures()
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

// TestWhereClusterScore pins the bar that decides which markers clustering may use. It is per
// marker, from the detector that produced it, so an upgrade cannot exclude a marker for a
// calibration it was never scored against.
func TestWhereClusterScore(t *testing.T) {
	newMarker := func(t *testing.T, detector string, score int) *entity.Marker {
		t.Helper()

		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			Score:          score,
			Size:           160,
			DetectModel:    detector,
			EmbedModel:     face.EmbeddingModelName(),
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		}

		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

		return m
	}
	matched := func(t *testing.T, uid string, floor int) bool {
		t.Helper()

		var n int
		require.NoError(t, whereClusterScore(entity.Db().Model(&entity.Marker{}).Where("marker_uid = ?", uid), floor).
			Count(&n).Error)

		return n == 1
	}
	yunet := face.ClusterScore(face.DetectorYuNet)
	scrfd := face.ClusterScore(face.DetectorSCRFD)

	// Which detector sits higher is a calibration that moves, so the case is stated by the
	// relationship rather than by naming one of them.
	lower, higher := face.DetectorYuNet, face.DetectorSCRFD
	if scrfd < yunet {
		lower, higher = face.DetectorSCRFD, face.DetectorYuNet
	}

	t.Run("EachDetectorItsOwnBar", func(t *testing.T) {
		// A score between the two bars is admitted for the detector with the lower one and
		// refused for the other, which is the whole point of registering them separately.
		between := max(face.ClusterScore(lower), face.ClusterScore(higher)-1)

		require.Less(t, face.ClusterScore(lower), face.ClusterScore(higher))
		assert.True(t, matched(t, newMarker(t, lower, between).MarkerUID, face.ClusterScoreAuto))
		assert.False(t, matched(t, newMarker(t, higher, between).MarkerUID, face.ClusterScoreAuto))
	})
	t.Run("ConfiguredFloorOverridesEveryDetector", func(t *testing.T) {
		// FACE_CLUSTER_SCORE is a choice rather than a calibration a marker was never scored
		// against, so unlike a detector's bar it applies to every marker. Without this the
		// option reached no query at all.
		restore := face.ClusterScoreThreshold
		t.Cleanup(func() { face.ClusterScoreThreshold = restore })
		face.ClusterScoreThreshold = 90

		assert.False(t, matched(t, newMarker(t, face.DetectorYuNet, 89).MarkerUID, face.ClusterScoreAuto))
		assert.True(t, matched(t, newMarker(t, face.DetectorSCRFD, 90).MarkerUID, face.ClusterScoreAuto))
	})
	t.Run("ConfiguredFloorRemovesTheBar", func(t *testing.T) {
		restore := face.ClusterScoreThreshold
		t.Cleanup(func() { face.ClusterScoreThreshold = restore })
		face.ClusterScoreThreshold = -1

		assert.True(t, matched(t, newMarker(t, face.DetectorSCRFD, 1).MarkerUID, face.ClusterScoreAuto))
	})
	t.Run("UnrecordedDetectorKeepsTheDefault", func(t *testing.T) {
		// Every row written before the provenance column existed, which must not be stranded.
		legacy := newMarker(t, "", face.ClusterScoreThresholdDefault)

		assert.True(t, matched(t, legacy.MarkerUID, face.ClusterScoreAuto))
		assert.False(t, matched(t, newMarker(t, "", face.ClusterScoreThresholdDefault-1).MarkerUID, face.ClusterScoreAuto))
	})
	t.Run("ExplicitFloorOverrides", func(t *testing.T) {
		m := newMarker(t, face.DetectorYuNet, yunet)

		assert.True(t, matched(t, m.MarkerUID, yunet))
		assert.False(t, matched(t, m.MarkerUID, yunet+1))
	})
	t.Run("ZeroDoesNotFilter", func(t *testing.T) {
		assert.True(t, matched(t, newMarker(t, "", 1).MarkerUID, 0))
	})
}

// TestResetAllFaceMarkerMatches covers the difference between the two reset scopes, which is
// whether a marker a person named keeps its identity.
func TestResetAllFaceMarkerMatches(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Cleanup(entity.ResetTestFixtures)

	named := entity.MarkerFixtures.Get("actress-a-1")
	require.Equal(t, entity.SrcManual, named.SubjSrc, "the fixture this pins must be hand-named")

	t.Run("KeepsManual", func(t *testing.T) {
		_, err := ResetFaceMarkerMatches()
		require.NoError(t, err)

		m, err := MarkerByUID(named.MarkerUID)
		require.NoError(t, err)
		assert.Equal(t, named.SubjUID, m.SubjUID, "an automatic reset must not touch a hand-named marker")
		assert.Equal(t, named.MarkerName, m.MarkerName)
		assert.Equal(t, named.FaceID, m.FaceID)
	})
	t.Run("ClearsManual", func(t *testing.T) {
		removed, err := ResetAllFaceMarkerMatches()
		require.NoError(t, err)
		assert.Positive(t, removed)

		m, err := MarkerByUID(named.MarkerUID)
		require.NoError(t, err)
		assert.Empty(t, m.SubjUID)
		assert.Empty(t, m.SubjSrc)
		assert.Empty(t, m.MarkerName)
		assert.Empty(t, m.FaceID)
		assert.Nil(t, m.MatchedAt)
	})
	t.Run("KeepsEmbeddings", func(t *testing.T) {
		// The whole point of the scope: detection survives, so a later run re-clusters from the
		// stored vectors instead of decoding every file again.
		m, err := MarkerByUID(named.MarkerUID)
		require.NoError(t, err)
		assert.Equal(t, named.Size, m.Size)
		assert.Equal(t, named.Score, m.Score)
		assert.NotEmpty(t, m.Thumb)
	})
}
