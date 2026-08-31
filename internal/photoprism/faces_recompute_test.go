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

	// Seeded with a full core, or the cluster is a labeled example rather than a grouping and
	// matching never offers it - which is the state this measurement exists to correct.
	embeddings := make(face.Embeddings, face.ManualClusterCore)
	for i := range embeddings {
		embeddings[i] = base
	}

	f := entity.NewFace(subjUID, entity.SrcManual, embeddings, face.EmbeddingModelName())

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

		// Four markers against a three-embedding centroid: the numbers have to differ, or "samples was
		// left alone" cannot be told from "samples was set to the member count".
		f, _ := recomputeTestCluster(t, subjUID, 9101, 0.05, 0.08, 0.11, 0.09)

		// The ratchet widens to the clamp, which is what a measurement over the members replaces.
		require.NoError(t, f.UpdateMatchStats(1, face.ClusterRadius))
		require.InDelta(t, face.ClusterRadius, f.SampleRadius, 1e-9)

		samples := f.Samples

		measured, err := recomputeFaceStats(f)
		require.NoError(t, err)
		require.True(t, measured)

		stored := entity.FindFace(f.ID)
		require.NotNil(t, stored)

		assert.Less(t, stored.SampleRadius, face.ClusterRadius, "the radius stops being the clamp")
		assert.GreaterOrEqual(t, stored.SampleRadius, 0.11, "while still covering the furthest member")
		require.Equal(t, face.ManualClusterCore, samples)
		assert.Equal(t, samples, stored.Samples, "and the centroid's inputs are not what was measured")
		assert.NotEqual(t, 4, stored.Samples, "which is not the number of markers it holds")
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
	t.Run("MeasuresAcrossTheBlankModelAlias", func(t *testing.T) {
		// A cluster predating the provenance column holds markers stamped facenet, and the two are
		// one space in both directions - so this must be measured, not declined. String equality
		// would decline it, and it is the common shape on any library from before that column.
		const subjUID = "js6sg6b1qekk9je6"
		isolatedTestFaces(t, "faces-recompute-blank-model")

		f, _ := recomputeTestCluster(t, subjUID, 9351, 0.05, 0.09)

		require.NoError(t, entity.Db().Model(&entity.Marker{}).
			Where("face_id = ?", f.ID).UpdateColumn("embed_model", face.ModelFaceNet).Error)
		require.NoError(t, f.Update("EmbedModel", ""))
		f.EmbedModel = ""

		measured, err := recomputeFaceStats(f)

		require.NoError(t, err)
		assert.True(t, measured, "a blank model is FaceNet's space, not a foreign one")
	})
	t.Run("DeclinesUnmeasurableMembers", func(t *testing.T) {
		// Same space, but nothing comparable in it: the widest radius in the schema is what this
		// replaces, so it must not be the answer when the vectors cannot be measured.
		const subjUID = "js6sg6b1qekk9je7"
		isolatedTestFaces(t, "faces-recompute-unusable")

		f, _ := recomputeTestCluster(t, subjUID, 9361, 0.05)
		require.NoError(t, f.SetSampleRadius(0.05))

		members, err := query.FaceMembers(f.ID)
		require.NoError(t, err)
		require.Len(t, members, 1)

		require.NoError(t, entity.Db().Model(&entity.Marker{}).
			Where("marker_uid = ?", members[0].MarkerUID).
			UpdateColumn("embeddings_json", []byte("[[0.1,0.2]]")).Error)

		measured, err := recomputeFaceStats(f)
		require.NoError(t, err)
		assert.False(t, measured)

		stored := entity.FindFace(f.ID)
		require.NotNil(t, stored)
		assert.InDelta(t, 0.05, stored.SampleRadius, 1e-9, "a declined cluster keeps what it stored")
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

// TestFacesMatchRecomputeStatsFlag pins the gate rather than the measurement: the flag's whole value
// is that a run with it off is an uncontaminated baseline, so both arms have to be asserted.
func TestFacesMatchRecomputeStatsFlag(t *testing.T) {
	// One cluster stored at the clamp with members far inside it, which is the state the ratchet
	// produces and a measurement corrects - so the two arms cannot agree.
	arm := func(t *testing.T, name string, recompute bool) float64 {
		t.Helper()

		w := isolatedTestFaces(t, name)
		w.conf.Options().FaceRecomputeStats = recompute
		require.Equal(t, recompute, w.conf.FaceRecomputeStats())

		f, _ := recomputeTestCluster(t, "js6sg6b1qekk9jf1", 9701, 0.04, 0.06, 0.08)
		require.NoError(t, f.UpdateMatchStats(1, face.ClusterRadius))
		require.InDelta(t, face.ClusterRadius, f.SampleRadius, 1e-9)

		_, err := w.Match(FacesOptions{Force: true})
		require.NoError(t, err)

		stored := entity.FindFace(f.ID)
		require.NotNil(t, stored)

		return stored.SampleRadius
	}

	t.Run("Off", func(t *testing.T) {
		assert.InDelta(t, face.ClusterRadius, arm(t, "faces-flag-off", false), 1e-9,
			"the ratchet is untouched, which is what makes a flag-off run a baseline")
	})
	t.Run("On", func(t *testing.T) {
		assert.Less(t, arm(t, "faces-flag-on", true), face.ClusterRadius,
			"the same run measures the members instead")
	})
}

// TestFaceMembers covers the set the radius is measured over, which has to exclude a marker that
// carries no vector: a member the measurement cannot reach would describe a different cluster than
// the one that is stored.
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
