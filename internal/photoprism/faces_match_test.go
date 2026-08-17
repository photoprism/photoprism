package photoprism

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestFaces_Match exercises the end-to-end matching flow with a loaded test configuration.
func TestFaces_Match(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	opt := FacesOptions{
		Force:     true,
		Threshold: 1,
	}

	r, err := m.Match(opt)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}

// TestBuildFaceCandidates validates that we drop non-matchable faces when building the index.
func TestBuildFaceCandidates(t *testing.T) {
	regular := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(3, face.RegularFace), face.EmbeddingModelName())
	require.NotNil(t, regular)

	// A cluster from another embedding space must never be compared with the current one.
	stale := *regular
	stale.ID = "stale-model"
	stale.EmbedModel = otherFaceModel(t, face.ConfiguredModel())

	// ResolveCollision raises the kind, which excludes a cluster from automatic matching.
	ambiguous := *regular
	ambiguous.ID = "ambiguous"
	ambiguous.FaceKind = int(face.AmbiguousFace)

	faces := entity.Faces{*regular, stale, ambiguous}

	index := buildFaceIndex(faces)

	require.Len(t, index.fallback, 1)
	require.Equal(t, regular.ID, index.fallback[0].ref.ID)

	// The candidate caches the clamped cutoff, not the raw column, so a stored radius
	// from an earlier calibration cannot widen the gate for a whole match run.
	require.InDelta(t, regular.AcceptDist(), index.fallback[0].acceptDist, 1e-9)
}

// TestFaceCandidateMatch covers the accept distance and collision gates in isolation.
func TestFaceCandidateMatch(t *testing.T) {
	embeddings := face.RandomEmbeddings(1, face.RegularFace)
	ref := entity.NewFace("", entity.SrcAuto, embeddings, face.EmbeddingModelName())
	require.NotNil(t, ref)

	t.Run("Match", func(t *testing.T) {
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: face.AcceptDist(0)}
		matched, dist := c.match(embeddings)
		require.True(t, matched)
		require.InDelta(t, 0.0, dist, 1e-9)
	})
	t.Run("TooFar", func(t *testing.T) {
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: -1}
		matched, _ := c.match(embeddings)
		require.False(t, matched)
	})
	t.Run("WithinCollisionRadius", func(t *testing.T) {
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: face.AcceptDist(0),
			collisionRadius: face.CollisionDist * 2}
		matched, _ := c.match(embeddings)
		require.True(t, matched)
	})
	t.Run("NoEmbeddings", func(t *testing.T) {
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: face.AcceptDist(0)}
		matched, dist := c.match(face.Embeddings{})
		require.False(t, matched)
		require.InDelta(t, -1.0, dist, 1e-9)
	})
}

// TestSelectBestFace ensures the best candidate is returned after indexing.
func TestSelectBestFace(t *testing.T) {
	markerEmb := face.RandomEmbeddings(1, face.RegularFace)

	matchFace := entity.NewFace("", entity.SrcAuto, markerEmb, face.EmbeddingModelName())
	require.NotNil(t, matchFace)

	// Force a different face that should not be a better match.
	otherEmb := face.RandomEmbeddings(4, face.RegularFace)
	otherFace := entity.NewFace("", entity.SrcAuto, otherEmb, face.EmbeddingModelName())
	require.NotNil(t, otherFace)

	faces := entity.Faces{*matchFace, *otherFace}

	index := buildFaceIndex(faces)
	require.Len(t, index.fallback, 2)

	best, dist := selectBestFace(markerEmb, index)
	require.NotNil(t, best)
	require.Equal(t, matchFace.ID, best.ID)
	require.InDelta(t, 0.0, dist, 1e-9)
}

func TestFacesMatchRespectsVeto(t *testing.T) {
	c := config.TestConfig()
	w := NewFaces(c)

	var marker entity.Marker
	require.NoError(t, entity.Db().Where("marker_type = ? AND marker_invalid = 0 AND face_id <> ''", entity.MarkerFace).Take(&marker).Error)

	origFaceID := marker.FaceID
	require.NotEqual(t, "", origFaceID)

	var f entity.Face
	require.NoError(t, entity.Db().Where("id = ?", origFaceID).Take(&f).Error)

	_, err := marker.ClearFace()
	require.NoError(t, err)

	stats := make(map[*entity.Face]*faceMatchStats)
	faces := entity.Faces{f}

	w.rememberVeto(marker.MarkerUID)
	_, err = w.MatchFaces(faces, false, nil, stats)
	require.NoError(t, err)

	require.NoError(t, entity.Db().Where("marker_uid = ?", marker.MarkerUID).Take(&marker).Error)
	require.Equal(t, "", marker.FaceID)

	// restore original assignment to keep fixtures consistent
	dist := minMarkerDistance(f.Embedding(), marker.Embeddings())
	_, err = marker.SetFace(&f, dist)
	require.NoError(t, err)
	w.clearVeto(marker.MarkerUID)
}
