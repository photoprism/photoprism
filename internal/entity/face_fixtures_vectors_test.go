package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

func TestGenerateFaceFixtureVectors(t *testing.T) {
	GenerateFaceFixtureVectors()

	model := face.EmbeddingModelName()

	t.Run("Clusters", func(t *testing.T) {
		for name := range faceFixtureSeeds {
			f := FaceFixtures.Get(name)
			embedding := f.Embedding()

			require.Len(t, embedding, face.ExpectedDims(), name)
			assert.Equal(t, model, f.EmbedModel, name)
			assert.True(t, f.SameEmbeddingModel(), name)
		}
	})
	t.Run("MarkersBelongToTheirCluster", func(t *testing.T) {
		// A marker generated inside its cluster has to be matchable by it, or a test that
		// looks like it exercises matching exercises the early exit instead.
		for name, spec := range markerFixtureVectors {
			m := MarkerFixtures.Get(name)

			require.Len(t, m.Embeddings(), 1, name)
			require.Len(t, m.Embeddings()[0], face.ExpectedDims(), name)
			assert.Equal(t, model, m.EmbedModel, name)

			f := FaceFixtures.Get(spec.face)
			match, dist := f.Match(m.Embeddings(), m.EmbedModel)

			assert.InDelta(t, spec.factor*face.AcceptDist(f.SampleRadius), dist, 1e-9, name)
			assert.Equal(t, spec.factor <= 1, match, name)
		}
	})
	t.Run("IdentitiesStayApart", func(t *testing.T) {
		// Two fixture people must never match each other, whatever the configured model
		// accepts, or a cross-subject assertion passes for the wrong reason.
		for name := range faceFixtureSeeds {
			f := FaceFixtures.Get(name)

			for other := range faceFixtureSeeds {
				if other == name {
					continue
				}

				o := FaceFixtures.Get(other)
				match, _ := f.Match(face.Embeddings{o.Embedding()}, model)
				assert.False(t, match, "%s must not match %s", name, other)
			}
		}
	})
	t.Run("EveryFaceMarkerHasAVector", func(t *testing.T) {
		// A marker fixture added without an entry in markerFixtureVectors still has to carry
		// a usable vector, or it is invisible to matching for a reason nothing reports.
		for name, m := range MarkerFixtures {
			if m.MarkerType != MarkerFace {
				continue
			}

			require.Len(t, m.Embeddings(), 1, name)
			assert.Len(t, m.Embeddings()[0], face.ExpectedDims(), name)
			assert.True(t, m.SameEmbeddingModel(), name)
		}
	})
	t.Run("Deterministic", func(t *testing.T) {
		before := FaceFixtures.Get("joe-biden").EmbeddingJSON

		GenerateFaceFixtureVectors()

		assert.Equal(t, before, FaceFixtures.Get("joe-biden").EmbeddingJSON)
	})
}

func TestFaceFixtureMarkerEmbedding(t *testing.T) {
	GenerateFaceFixtureVectors()

	biden := FaceFixtures.Get("joe-biden")
	centroid := biden.Embedding()
	centroids := map[string]face.Embedding{"joe-biden": centroid}

	t.Run("NearItsCluster", func(t *testing.T) {
		spec := markerFixtureVector{face: "joe-biden", factor: 0.5, seed: 900}
		embedding := faceFixtureMarkerEmbedding(spec, centroids)

		want := 0.5 * face.AcceptDist(biden.SampleRadius)
		assert.InDelta(t, want, centroid.Dist(embedding), 1e-9)
	})
	t.Run("PersonOfItsOwn", func(t *testing.T) {
		// A marker that names no cluster stands for somebody no fixture cluster holds, so it
		// has to land outside every accept distance rather than near one of them.
		spec := markerFixtureVector{seed: 901}
		embedding := faceFixtureMarkerEmbedding(spec, centroids)

		assert.Equal(t, face.FixtureEmbedding(901), embedding)
		assert.Greater(t, centroid.Dist(embedding), float64(face.ConfigDistMax))
	})
	t.Run("UnknownCluster", func(t *testing.T) {
		spec := markerFixtureVector{face: "nobody", factor: 0.5, seed: 902}
		assert.Equal(t, face.FixtureEmbedding(902), faceFixtureMarkerEmbedding(spec, centroids))
	})
}

func TestFixtureSeed(t *testing.T) {
	t.Run("Deterministic", func(t *testing.T) {
		assert.Equal(t, fixtureSeed("marker-a"), fixtureSeed("marker-a"))
	})
	t.Run("DiffersByName", func(t *testing.T) {
		assert.NotEqual(t, fixtureSeed("marker-a"), fixtureSeed("marker-b"))
	})
}
