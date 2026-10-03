package photoprism

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestFaces_CorrectionCollision pins that correcting one face in a named cluster to another existing
// person records a collision that later runs keep, and moves the face to a face of that person.
func TestFaces_CorrectionCollision(t *testing.T) {
	w := isolatedTestFaces(t, "facescorrection")

	carol := consensusTestSubject(t, "Correction Carol")
	dave := consensusTestSubject(t, "Correction Dave")

	f := entity.NewFace(carol.SubjUID, entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(7301)}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())
	require.NoError(t, f.Updates(entity.Values{"samples": 5}))

	newMarker := func(t *testing.T, fraction float64, seed uint64) string {
		t.Helper()

		dist := fraction * f.AcceptDist()
		emb := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, seed)}
		m := entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        consensusTestFileUID,
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			SubjUID:        carol.SubjUID,
			SubjSrc:        entity.SrcAuto,
			FaceID:         f.ID,
			FaceDist:       dist,
			EmbeddingsJSON: emb.JSON(),
			EmbedModel:     f.EmbedModel,
			Size:           face.ClusterSizeThreshold,
			Score:          face.ClusterScore("") + 10,
			MatchedAt:      entity.TimeStamp(),
			W:              0.1,
			H:              0.1,
		}

		require.NoError(t, entity.UnscopedDb().Create(&m).Error)

		return m.MarkerUID
	}

	near := []string{newMarker(t, 0.2, 1), newMarker(t, 0.2, 2)}
	corrected := newMarker(t, 0.6, 3)

	m := entity.FindMarker(corrected)
	require.NotNil(t, m)
	changed, err := m.SetName(dave.SubjName, entity.SrcManual)
	require.NoError(t, err)
	require.True(t, changed)
	require.NoError(t, m.Save())

	check := func(t *testing.T) {
		t.Helper()

		cluster := entity.FindFace(f.ID)
		require.NotNil(t, cluster, "the cluster is kept")
		assert.Equal(t, carol.SubjUID, cluster.SubjUID)
		assert.Equal(t, 1, cluster.Collisions, "one correction is one collision")

		got := entity.FindMarker(corrected)
		require.NotNil(t, got)
		assert.Equal(t, dave.SubjUID, got.SubjUID)
		assert.NotEqual(t, f.ID, got.FaceID, "the corrected face does not return to the cluster")

		if own := entity.FindFace(got.FaceID); assert.NotNil(t, own, "it has a face of its own person") {
			assert.Equal(t, dave.SubjUID, own.SubjUID)
		}

		for _, uid := range near {
			assert.Equal(t, carol.SubjUID, entity.FindMarker(uid).SubjUID)
		}

		s := entity.FindSubject(carol.SubjUID)
		require.NotNil(t, s)
		assert.False(t, s.Deleted())
	}

	t.Run("Corrected", check)
	t.Run("Run", func(t *testing.T) {
		require.NoError(t, w.Start(FacesOptions{}))
		check(t)
	})
	t.Run("ForcedRun", func(t *testing.T) {
		require.NoError(t, w.Start(FacesOptions{Force: true}))
		check(t)
	})
}

// TestFaces_CorrectionInBand pins that a correction whose collision radius cannot narrow the cluster
// is recorded once, and that later runs neither count it again nor reopen the cluster.
func TestFaces_CorrectionInBand(t *testing.T) {
	w := isolatedTestFaces(t, "facescorrectionband")

	carol := consensusTestSubject(t, "Band Carol")
	dave := consensusTestSubject(t, "Band Dave")

	f := entity.NewFace(carol.SubjUID, entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(7311)}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())
	require.NoError(t, f.Updates(entity.Values{"samples": 5}))

	newMarker := func(t *testing.T, dist float64, seed uint64) string {
		t.Helper()

		emb := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, seed)}
		m := entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        consensusTestFileUID,
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			SubjUID:        carol.SubjUID,
			SubjSrc:        entity.SrcAuto,
			FaceID:         f.ID,
			FaceDist:       dist,
			EmbeddingsJSON: emb.JSON(),
			EmbedModel:     f.EmbedModel,
			Size:           face.ClusterSizeThreshold,
			Score:          face.ClusterScore("") + 10,
			MatchedAt:      entity.TimeStamp(),
			W:              0.1,
			H:              0.1,
		}

		require.NoError(t, entity.UnscopedDb().Create(&m).Error)

		return m.MarkerUID
	}

	members := []string{newMarker(t, 0.5*f.AcceptDist(), 1), newMarker(t, 0.5*f.AcceptDist(), 2)}
	dist := face.CollisionDist / 2
	require.Greater(t, dist, face.AmbiguityDist())
	corrected := newMarker(t, dist, 3)
	require.NoError(t, f.Matched())

	m := entity.FindMarker(corrected)
	require.NotNil(t, m)
	changed, err := m.SetName(dave.SubjName, entity.SrcManual)
	require.NoError(t, err)
	require.True(t, changed)
	require.NoError(t, m.Save())

	cluster := entity.FindFace(f.ID)
	require.NotNil(t, cluster)
	require.Equal(t, 1, cluster.Collisions)
	assert.NotNil(t, cluster.MatchedAt, "the correction does not reopen the cluster")

	stamps := func(t *testing.T) (face string, markers []string) {
		t.Helper()

		c := entity.FindFace(f.ID)
		require.NotNil(t, c)
		require.NotNil(t, c.MatchedAt, "the cluster is not left reopened")
		assert.Equal(t, 1, c.Collisions, "one correction is one collision")
		assert.Equal(t, carol.SubjUID, c.SubjUID)

		for _, uid := range members {
			got := entity.FindMarker(uid)
			require.NotNil(t, got)
			require.NotNil(t, got.MatchedAt)
			assert.Equal(t, carol.SubjUID, got.SubjUID)
			markers = append(markers, got.MatchedAt.String())
		}

		return c.MatchedAt.String(), markers
	}

	require.NoError(t, w.Start(FacesOptions{}))
	stamps(t)

	// Backdated, so a run that stamps them again within the same second still shows.
	past := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Face{}).Where("id = ?", f.ID).UpdateColumn("matched_at", past).Error)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid IN (?)", members).UpdateColumn("matched_at", past).Error)
	faceStamp, markerStamps := stamps(t)

	for range 2 {
		require.NoError(t, w.Start(FacesOptions{}))
		f2, m2 := stamps(t)
		assert.Equal(t, faceStamp, f2)
		assert.Equal(t, markerStamps, m2)
	}

	got := entity.FindMarker(corrected)
	require.NotNil(t, got)
	assert.Equal(t, dave.SubjUID, got.SubjUID)
	assert.NotEqual(t, f.ID, got.FaceID)
}

// TestFaces_SmallCorrectionInBand pins that a corrected face too small to seed a face of its own, whose
// collision cannot narrow the cluster, is matched once and not retried by every run.
func TestFaces_SmallCorrectionInBand(t *testing.T) {
	w := isolatedTestFaces(t, "facessmallcorrection")

	carol := consensusTestSubject(t, "Small Band Carol")
	dave := consensusTestSubject(t, "Small Band Dave")

	f := entity.NewFace(carol.SubjUID, entity.SrcAuto, face.Embeddings{face.FixtureEmbedding(7321)}, face.EmbeddingModelName())
	require.NotNil(t, f)
	require.NoError(t, f.Create())
	require.NoError(t, f.Updates(entity.Values{"samples": 5}))

	dist := face.CollisionDist / 2
	emb := face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), dist, 1)}
	m := entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        consensusTestFileUID,
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcImage,
		SubjUID:        carol.SubjUID,
		SubjSrc:        entity.SrcAuto,
		FaceID:         f.ID,
		FaceDist:       dist,
		EmbeddingsJSON: emb.JSON(),
		EmbedModel:     f.EmbedModel,
		Size:           face.ClusterSizeThreshold - 1,
		Score:          face.ClusterScore("") + 10,
		MatchedAt:      entity.TimeStamp(),
		W:              0.1,
		H:              0.1,
	}

	require.NoError(t, entity.UnscopedDb().Create(&m).Error)
	require.False(t, m.Clusterable())

	// Members keep the cluster from being removed once the corrected face leaves it.
	for i := range 2 {
		d := 0.5 * f.AcceptDist()
		member := m
		member.MarkerUID = rnd.GenerateUID('m')
		member.FaceDist = d
		member.EmbeddingsJSON = face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), d, uint64(10+i))}.JSON()
		member.Size = face.ClusterSizeThreshold
		require.NoError(t, entity.UnscopedDb().Create(&member).Error)
	}

	changed, err := m.SetName(dave.SubjName, entity.SrcManual)
	require.NoError(t, err)
	require.True(t, changed)
	require.NoError(t, m.Save())

	got := entity.FindMarker(m.MarkerUID)
	require.NotNil(t, got)
	require.Empty(t, got.FaceID, "left for matching")
	require.Nil(t, got.MatchedAt)

	for range 3 {
		require.NoError(t, w.Start(FacesOptions{}))

		got = entity.FindMarker(m.MarkerUID)
		require.NotNil(t, got)
		assert.NotNil(t, got.MatchedAt, "matched once, not retried by every run")
		assert.Equal(t, dave.SubjUID, got.SubjUID)
		assert.Equal(t, 1, entity.FindFace(f.ID).Collisions)
	}

	assert.Zero(t, query.CountUnmatchedFaceMarkers())

	// Naming it after the cluster's person again lets the next run place it there.
	got = entity.FindMarker(m.MarkerUID)
	require.NotNil(t, got)
	changed, err = got.SetName(carol.SubjName, entity.SrcManual)
	require.NoError(t, err)
	require.True(t, changed)
	require.NoError(t, got.Save())
	require.Nil(t, entity.FindMarker(m.MarkerUID).MatchedAt)

	require.NoError(t, w.Start(FacesOptions{}))

	got = entity.FindMarker(m.MarkerUID)
	require.NotNil(t, got)
	assert.Equal(t, carol.SubjUID, got.SubjUID)
	assert.Equal(t, f.ID, got.FaceID)
}

// TestFaces_AuditNotedCollision pins that the audit does not report a collision the cluster has
// recorded already and whose radius cannot narrow it.
func TestFaces_AuditNotedCollision(t *testing.T) {
	w := isolatedTestFaces(t, "facesauditnoted")

	carol := consensusTestSubject(t, "Audit Band Carol")
	dave := consensusTestSubject(t, "Audit Band Dave")

	f1 := entity.NewFace(carol.SubjUID, entity.SrcManual, face.Embeddings{face.FixtureEmbedding(7331)}, face.EmbeddingModelName())
	require.NoError(t, f1.Create())
	f2 := entity.NewFace(dave.SubjUID, entity.SrcManual, face.Embeddings{face.FixtureEmbeddingAt(f1.Embedding(), face.CollisionDist/2, 1)}, face.EmbeddingModelName())
	require.NoError(t, f2.Create())

	_, _, err := query.ResolveFaceCollisions()
	require.NoError(t, err)

	hook := captureLog(t)
	require.NoError(t, w.Audit(false, ""))

	for _, msg := range loggedMessages(hook, logrus.DebugLevel) {
		assert.False(t, strings.Contains(msg, f1.ID) && strings.Contains(msg, f2.ID), msg)
	}
}
