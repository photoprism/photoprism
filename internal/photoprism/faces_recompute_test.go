package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

// recomputeTestCluster saves a cluster of one tight sample plus markers at the given distances from
// it, which is the shape the ratchet and a measurement disagree about.
func recomputeTestCluster(t *testing.T, subjUID string, seed uint64, dists ...float64) (*entity.Face, face.Embedding) {
	t.Helper()

	base := face.FixtureEmbedding(seed)
	f := entity.NewFace(subjUID, entity.SrcManual, face.Embeddings{base}, face.EmbeddingModelName())

	require.NotNil(t, f)
	require.NoError(t, f.Create())

	for i, d := range dists {
		m := &entity.Marker{
			FileUID:    "fs6sg6bw45bnlqdw",
			MarkerType: entity.MarkerFace,
			MarkerSrc:  entity.SrcImage,
			FaceID:     f.ID,
			SubjUID:    subjUID,
			SubjSrc:    entity.SrcManual,
			Size:       face.SizeThreshold,
			Score:      50,
			X:          0.2,
			Y:          0.2,
			W:          0.1,
			H:          0.1,
		}

		m.SetEmbeddings(face.Embeddings{face.FixtureEmbeddingAt(f.Embedding(), d, seed+uint64(i)+1)},
			f.EmbedModel, face.DetectorYuNet)
		require.NoError(t, entity.Db().Create(m).Error)
	}

	return f, base
}

// TestRecomputeFaceStats covers the measurement that replaces the ratchet, through the stored row
// rather than through the helper, so it fails against the shipped default and passes with the flag.
func TestRecomputeFaceStats(t *testing.T) {
	t.Run("MeasuresMembers", func(t *testing.T) {
		const subjUID = "js6sg6b1qekk9je1"
		isolatedTestFaces(t, "faces-recompute-members")

		f, _ := recomputeTestCluster(t, subjUID, 9101, 0.05, 0.08, 0.11)

		// The ratchet would leave it here, at the widest a pass may accept rather than at a spread.
		require.NoError(t, f.UpdateMatchStats(1, f.AcceptDist()))
		require.InDelta(t, face.ClusterRadius, f.SampleRadius, 1e-9)

		measured, err := recomputeFaceStats(f)
		require.NoError(t, err)
		require.True(t, measured)

		stored := entity.FindFace(f.ID)
		require.NotNil(t, stored)

		assert.Equal(t, 3, stored.Samples, "samples counts the markers it holds")
		assert.Less(t, stored.SampleRadius, face.ClusterRadius, "and the radius stops being the clamp")
		assert.GreaterOrEqual(t, stored.SampleRadius, 0.11, "while still covering the furthest member")
	})
	t.Run("KeepsClusterIdentity", func(t *testing.T) {
		// The centroid is read, never re-derived: faces.id is its hash, so a new one would orphan
		// every marker pointing at the cluster.
		const subjUID = "js6sg6b1qekk9je2"
		isolatedTestFaces(t, "faces-recompute-identity")

		f, _ := recomputeTestCluster(t, subjUID, 9201, 0.05, 0.09)
		before := f.ID

		measured, err := recomputeFaceStats(f)
		require.NoError(t, err)
		require.True(t, measured)

		assert.Equal(t, before, f.ID)
		assert.NotNil(t, entity.FindFace(before))
	})
	t.Run("DeclinesMixedEmbeddingSpace", func(t *testing.T) {
		// No library holds this shape, so it ships unrun without a synthetic case.
		const subjUID = "js6sg6b1qekk9je3"
		isolatedTestFaces(t, "faces-recompute-mixed")

		f, _ := recomputeTestCluster(t, subjUID, 9301, 0.05, 0.09)

		members, err := query.FaceMembers(f.ID)
		require.NoError(t, err)
		require.Len(t, members, 2)

		foreign := face.ModelSFace
		if f.EmbedModel == face.ModelSFace {
			foreign = face.ModelFaceNet
		}

		require.NoError(t, entity.Db().Model(&entity.Marker{}).
			Where("marker_uid = ?", members[0].MarkerUID).
			UpdateColumn("embed_model", foreign).Error)

		measured, err := recomputeFaceStats(f)

		require.NoError(t, err)
		assert.False(t, measured, "a distance across two embedding spaces measures nothing")
	})
	t.Run("DeclinesWithoutMembers", func(t *testing.T) {
		const subjUID = "js6sg6b1qekk9je4"
		isolatedTestFaces(t, "faces-recompute-empty")

		f, _ := recomputeTestCluster(t, subjUID, 9401)

		measured, err := recomputeFaceStats(f)

		require.NoError(t, err)
		assert.False(t, measured)
	})
}

// TestFaceMembers covers the set both columns are derived from, which has to exclude a marker that
// carries no vector: counting one the radius cannot include would leave the two describing
// different clusters.
func TestFaceMembers(t *testing.T) {
	const subjUID = "js6sg6b1qekk9je5"
	isolatedTestFaces(t, "faces-members")

	f, _ := recomputeTestCluster(t, subjUID, 9501, 0.05, 0.09)

	vectorless := &entity.Marker{
		FileUID:    "fs6sg6bw45bnlqdw",
		MarkerType: entity.MarkerFace,
		MarkerSrc:  entity.SrcXmp,
		FaceID:     f.ID,
		SubjUID:    subjUID,
		SubjSrc:    entity.SrcManual,
		Size:       face.SizeThreshold,
		Score:      50,
		X:          0.4,
		Y:          0.4,
		W:          0.1,
		H:          0.1,
	}

	require.NoError(t, entity.Db().Create(vectorless).Error)

	members, err := query.FaceMembers(f.ID)
	require.NoError(t, err)

	assert.Len(t, members, 2, "a marker holding a cluster id but no vector is not a member")

	for _, m := range members {
		assert.NotEqual(t, vectorless.MarkerUID, m.MarkerUID)
	}
}
