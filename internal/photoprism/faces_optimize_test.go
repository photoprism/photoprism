package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

func TestFaces_Optimize(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	r, err := m.Optimize()

	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}

// namedFace saves a manual cluster of one sample, which is what naming a face in the app creates.
func namedFace(t *testing.T, subjUID string, e face.Embedding) *entity.Face {
	t.Helper()

	m := entity.NewFace(subjUID, entity.SrcManual, face.Embeddings{e}, face.EmbeddingModelName())

	require.NotNil(t, m)
	require.NoError(t, m.Create())

	return m
}

// TestFaces_OptimizeSingletons covers the clusters naming a face by hand produces. A one-sample
// cluster stores no measurable extent, so what it accepts is decided entirely by the default
// SetEmbeddings applies - and at a radius of zero it accepts nothing and merges with nothing.
func TestFaces_OptimizeSingletons(t *testing.T) {
	t.Run("SameSubjectMerges", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-same-subj")

		base := face.FixtureEmbedding(8101)
		near := face.AcceptDist(face.ClusterRadius) * 0.9

		for i, e := range []face.Embedding{
			base,
			face.FixtureEmbeddingAt(base, near, 8102),
			face.FixtureEmbeddingAt(base, near, 8103),
		} {
			f := namedFace(t, "js6sg6b1qekk9jx8", e)
			require.InDelta(t, face.ClusterRadius, f.SampleRadius, 1e-9, "cluster %d", i)
		}

		before, err := query.ManuallyAddedFaces(false, false, "js6sg6b1qekk9jx8")
		require.NoError(t, err)
		require.Len(t, before, 3)

		r, err := w.OptimizeFor("js6sg6b1qekk9jx8")
		require.NoError(t, err)
		assert.Positive(t, r.Merged, "singletons of one subject have to be able to merge")

		after, err := query.ManuallyAddedFaces(false, false, "js6sg6b1qekk9jx8")
		require.NoError(t, err)
		assert.Less(t, len(after), len(before), "the subject ends up with fewer clusters")
	})
	t.Run("SameSubjectPairMerges", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-same-subj-pair")

		base := face.FixtureEmbedding(8301)
		near := face.AcceptDist(face.ClusterRadius) * 0.9

		namedFace(t, "js6sg6b1qekk9jx7", base)
		namedFace(t, "js6sg6b1qekk9jx7", face.FixtureEmbeddingAt(base, near, 8302))

		r, err := w.OptimizeFor("js6sg6b1qekk9jx7")
		require.NoError(t, err)
		assert.Positive(t, r.Merged, "naming a second face of one person has to be able to merge")

		after, err := query.ManuallyAddedFaces(false, false, "js6sg6b1qekk9jx7")
		require.NoError(t, err)
		assert.Len(t, after, 1)
	})
	// The clusters arrive ordered by subject, and a subject that follows another is the case a
	// boundary can lose: the cluster the run crosses into has to anchor the next group rather
	// than be passed over, or every subject but the first merges nothing.
	t.Run("EachSubjectsRunMergesIndependently", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-two-subjects")

		near := face.AcceptDist(face.ClusterRadius) * 0.9
		subjects := []string{"js6sg6b1qekk9ja1", "js6sg6b1qekk9ja2"}

		for i, subjUID := range subjects {
			base := face.FixtureEmbedding(uint64(8501 + i*10))
			namedFace(t, subjUID, base)
			namedFace(t, subjUID, face.FixtureEmbeddingAt(base, near, uint64(8502+i*10)))
		}

		r, err := w.Optimize()
		require.NoError(t, err)
		assert.Positive(t, r.Merged)

		for _, subjUID := range subjects {
			after, err := query.ManuallyAddedFaces(false, false, subjUID)
			require.NoError(t, err)
			assert.Len(t, after, 1, "subject %s merges its own pair", subjUID)
		}
	})
	// A merge builds its midpoint from the source centroids alone, so it can be narrower than
	// either source and refuse a marker both of them held. That candidate is then kept and the
	// midpoint created anyway, leaving the same cluster count - which has to end the pass.
	t.Run("ConvergesWhenAMergeCannotAbsorbItsMarkers", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-retained")

		const subjUID = "js6sg6b1qekk9jb1"

		base := face.FixtureEmbedding(8601)
		apart := face.AcceptDist(face.ClusterRadius) * 0.75

		c1 := namedFace(t, subjUID, base)
		c2 := namedFace(t, subjUID, face.FixtureEmbeddingAt(base, apart, 8602))

		// The marker its own cluster holds, placed where the midpoint will not reach it.
		stray := face.FixtureEmbeddingAt(c1.Embedding(), apart, 8603)
		_, radius, _ := face.EmbeddingsMidpoint(face.Embeddings{c1.Embedding(), c2.Embedding()})
		midpoint, _, _ := face.EmbeddingsMidpoint(face.Embeddings{c1.Embedding(), c2.Embedding()})

		require.Greater(t, face.Embeddings{stray}.Dist(midpoint), face.AcceptDist(radius),
			"the fixture has to reproduce a merge that cannot absorb its own marker")

		for _, m := range []struct {
			face *entity.Face
			emb  face.Embedding
		}{{c1, stray}, {c2, c2.Embedding()}} {
			marker := &entity.Marker{
				FileUID:    "fs6sg6bw45bnlqdw",
				MarkerType: entity.MarkerFace,
				MarkerSrc:  entity.SrcImage,
				FaceID:     m.face.ID,
				SubjUID:    subjUID,
				SubjSrc:    entity.SrcManual,
				Size:       face.SizeThreshold,
				Score:      50,
				X:          0.2,
				Y:          0.2,
				W:          0.1,
				H:          0.1,
			}

			marker.SetEmbeddings(face.Embeddings{m.emb}, m.face.EmbedModel, face.DetectorYuNet)
			require.NoError(t, entity.Db().Create(marker).Error)
		}

		if _, err := w.OptimizeFor(subjUID); err != nil {
			t.Fatal(err)
		}

		second, err := w.OptimizeFor(subjUID)
		require.NoError(t, err)
		assert.Zero(t, second.Merged, "a settled subject reports no further merges")
	})
	t.Run("DifferentSubjectsDoNotMerge", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-other-subj")

		base := face.FixtureEmbedding(8201)
		near := face.AcceptDist(face.ClusterRadius) * 0.9

		// Same geometry as above, so only the subject decides the outcome. ManuallyAddedFaces
		// groups by subject, which is why it holds however far the clusters reach.
		namedFace(t, "js6sg6b1qekk9jx1", base)
		namedFace(t, "js6sg6b1qekk9jx2", face.FixtureEmbeddingAt(base, near, 8202))
		namedFace(t, "js6sg6b1qekk9jx3", face.FixtureEmbeddingAt(base, near, 8203))

		r, err := w.Optimize()
		require.NoError(t, err)
		assert.Zero(t, r.Merged, "clusters of different people must never be combined")

		for _, subjUID := range []string{"js6sg6b1qekk9jx1", "js6sg6b1qekk9jx2", "js6sg6b1qekk9jx3"} {
			remaining, err := query.ManuallyAddedFaces(false, false, subjUID)
			require.NoError(t, err)
			assert.Len(t, remaining, 1, "subject %s keeps its own cluster", subjUID)
		}
	})
}
