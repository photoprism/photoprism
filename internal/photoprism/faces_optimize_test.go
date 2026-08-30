package photoprism

import (
	"strings"
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

// TestMergeGroups covers the partition Optimize merges, which has to follow the distances rather
// than the order the clusters were fetched in.
func TestMergeGroups(t *testing.T) {
	const subjUID = "js6sg6b1qekk9jc1"

	// A chain: a-b and b-c are within the clustering distance, a-c is not. Anchored on whichever
	// cluster comes first, c joins when b anchors and is dropped when a does.
	base := face.FixtureEmbedding(8701)
	near := face.ClusterDist * 0.9

	a := &entity.Face{ID: "A", SubjUID: subjUID, EmbedModel: face.EmbeddingModelName()}
	b := &entity.Face{ID: "B", SubjUID: subjUID, EmbedModel: face.EmbeddingModelName()}
	c := &entity.Face{ID: "C", SubjUID: subjUID, EmbedModel: face.EmbeddingModelName()}

	a.EmbeddingJSON = base.JSON()
	b.EmbeddingJSON = face.FixtureEmbeddingAt(base, near, 8702).JSON()
	c.EmbeddingJSON = face.FixtureEmbeddingAt(b.Embedding(), near, 8703).JSON()

	require.Greater(t, a.Embedding().Dist(c.Embedding()), face.ClusterDist,
		"the fixture has to reproduce a chain the ends of which do not merge directly")

	t.Run("LinksTransitively", func(t *testing.T) {
		groups := mergeGroups(entity.Faces{*a, *b, *c})

		require.Len(t, groups, 1)
		assert.Len(t, groups[0], 3)
	})
	t.Run("SameForEveryOrder", func(t *testing.T) {
		for _, order := range []entity.Faces{
			{*a, *b, *c}, {*a, *c, *b}, {*b, *a, *c},
			{*b, *c, *a}, {*c, *a, *b}, {*c, *b, *a},
		} {
			groups := mergeGroups(order)

			require.Len(t, groups, 1, "order %s", strings.Join(order.IDs(), ""))
			assert.Len(t, groups[0], 3, "order %s", strings.Join(order.IDs(), ""))
		}
	})
	t.Run("SplitsBySubject", func(t *testing.T) {
		// Same geometry, so only the subject keeps them apart.
		other := *c
		other.SubjUID = "js6sg6b1qekk9jc2"

		groups := mergeGroups(entity.Faces{*a, *b, other})

		require.Len(t, groups, 2)
		assert.Len(t, groups[0], 2)
		assert.Len(t, groups[1], 1)
	})
	t.Run("NoneClose", func(t *testing.T) {
		far := *c
		far.EmbeddingJSON = face.FixtureEmbeddingAt(base, face.ClusterDist*2.5, 8704).JSON()

		groups := mergeGroups(entity.Faces{*a, far})

		require.Len(t, groups, 2)
		assert.Len(t, groups[0], 1)
		assert.Len(t, groups[1], 1)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, mergeGroups(nil))
	})
}

// namedFace saves a manual cluster of one sample, which is what naming a face in the app creates.
func namedFace(t *testing.T, subjUID string, e face.Embedding) *entity.Face {
	t.Helper()

	m := entity.NewFace(subjUID, entity.SrcManual, face.Embeddings{e}, face.EmbeddingModelName())

	require.NotNil(t, m)
	require.NoError(t, m.Create())

	return m
}

// TestFaces_OptimizeSingletons covers the clusters naming a face by hand produces. Each stores no
// measurable extent, so the distances are expressed against the clustering distance the merge is
// bounded by rather than against a radius none of them measured.
func TestFaces_OptimizeSingletons(t *testing.T) {
	t.Run("SameSubjectMerges", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-same-subj")

		base := face.FixtureEmbedding(8101)
		near := face.ClusterDist * 0.9

		for i, e := range []face.Embedding{
			base,
			face.FixtureEmbeddingAt(base, near, 8102),
			face.FixtureEmbeddingAt(base, near, 8103),
		} {
			f := namedFace(t, "js6sg6b1qekk9jx8", e)
			require.InDelta(t, face.Epsilon, f.SampleRadius, 1e-9, "cluster %d", i)
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
	// Two labels of one person are close enough to link, and still must not merge: their midpoint
	// would be a pair's average, and a centroid needs a core to be one.
	t.Run("SameSubjectPairWaits", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-same-subj-pair")

		base := face.FixtureEmbedding(8301)
		near := face.ClusterDist * 0.9

		a := namedFace(t, "js6sg6b1qekk9jx7", base)
		b := namedFace(t, "js6sg6b1qekk9jx7", face.FixtureEmbeddingAt(base, near, 8302))

		mergeable, _ := a.Mergeable(b)
		require.True(t, mergeable, "the pair links, so only the core decides")

		r, err := w.OptimizeFor("js6sg6b1qekk9jx7")
		require.NoError(t, err)
		assert.Zero(t, r.Merged)

		after, err := query.ManuallyAddedFaces(false, false, "js6sg6b1qekk9jx7")
		require.NoError(t, err)
		assert.Len(t, after, 2, "both wait for a third")

		// And the third completes the group in one pass, since the link is transitive.
		namedFace(t, "js6sg6b1qekk9jx7", face.FixtureEmbeddingAt(base, near, 8303))

		r, err = w.OptimizeFor("js6sg6b1qekk9jx7")
		require.NoError(t, err)
		assert.Positive(t, r.Merged)

		after, err = query.ManuallyAddedFaces(false, false, "js6sg6b1qekk9jx7")
		require.NoError(t, err)
		require.Len(t, after, 1)
		assert.Equal(t, face.ManualClusterCore, after[0].Samples, "the centroid records its inputs")
		assert.Greater(t, after[0].SampleRadius, face.Epsilon, "and now has a measured extent")
	})
	// The clusters arrive ordered by subject, so a subject that follows another is the case a
	// boundary can lose: one run must not consume the clusters the next subject needs.
	t.Run("EachSubjectsRunMergesIndependently", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-two-subjects")

		near := face.ClusterDist * 0.9
		subjects := []string{"js6sg6b1qekk9ja1", "js6sg6b1qekk9ja2"}

		for i, subjUID := range subjects {
			base := face.FixtureEmbedding(uint64(8501 + i*10))
			namedFace(t, subjUID, base)

			for j := 1; j < face.ManualClusterCore; j++ {
				namedFace(t, subjUID, face.FixtureEmbeddingAt(base, near, uint64(8502+i*10+j)))
			}
		}

		r, err := w.Optimize()
		require.NoError(t, err)
		assert.Positive(t, r.Merged)

		for _, subjUID := range subjects {
			after, err := query.ManuallyAddedFaces(false, false, subjUID)
			require.NoError(t, err)
			assert.Len(t, after, 1, "subject %s merges its own core", subjUID)
		}
	})
	// A merge builds its midpoint from the source centroids alone, so it can be narrower than
	// either source and refuse a marker both of them held. That candidate is then kept and the
	// midpoint created anyway, leaving the same cluster count - which has to end the pass.
	t.Run("ConvergesWhenAMergeCannotAbsorbItsMarkers", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-retained")

		const subjUID = "js6sg6b1qekk9jb1"

		base := face.FixtureEmbedding(8601)
		apart := face.ClusterDist * 0.95

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
	// The case the symmetric criterion exists for, and the one every other subtest here misses:
	// a tight multi-sample cluster accepts almost nothing, so under a predicate reading the anchor's
	// own extent the pair is decided by which of the two the fetch order puts first.
	t.Run("TightAnchorStillMergesItsFarSingleton", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-tight-anchor")

		const subjUID = "js6sg6b1qekk9jd1"

		base := face.FixtureEmbedding(8801)
		apart := face.ClusterDist * 0.85

		// Five near-identical samples, so the measured extent is small and sorts first on samples.
		tight := face.Embeddings{base}
		for i := range 4 {
			tight = append(tight, face.FixtureEmbeddingAt(base, 0.02, uint64(8802+i)))
		}

		anchor := entity.NewFace(subjUID, entity.SrcManual, tight, face.EmbeddingModelName())
		require.NotNil(t, anchor)
		require.NoError(t, anchor.Create())

		// Beyond what the tight cluster accepts, inside what one midpoint may stand for.
		require.Greater(t, apart, anchor.AcceptDist())
		require.LessOrEqual(t, apart, face.ClusterDist)

		// Enough singletons to complete a core with the anchor, all beyond what it would accept.
		for i := 1; i < face.ManualClusterCore; i++ {
			namedFace(t, subjUID, face.FixtureEmbeddingAt(anchor.Embedding(), apart, uint64(8806+i)))
		}

		r, err := w.OptimizeFor(subjUID)
		require.NoError(t, err)
		assert.Positive(t, r.Merged, "the bound is the clustering distance, not the anchor's extent")

		after, err := query.ManuallyAddedFaces(false, false, subjUID)
		require.NoError(t, err)
		assert.Len(t, after, 1)
	})
	t.Run("DifferentSubjectsDoNotMerge", func(t *testing.T) {
		w := isolatedTestFaces(t, "faces-optimize-other-subj")

		base := face.FixtureEmbedding(8201)
		near := face.ClusterDist * 0.9

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
