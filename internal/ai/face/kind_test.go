package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKind_String(t *testing.T) {
	t.Run("Known", func(t *testing.T) {
		assert.Equal(t, "regular", RegularFace.String())
		assert.Equal(t, "ambiguous", AmbiguousFace.String())
	})
	t.Run("Unclassified", func(t *testing.T) {
		// The zero value, which nothing stores deliberately, so seeing it in a report means the
		// cluster was never classified rather than that it is ordinary.
		assert.Equal(t, "unset", UnclassifiedFace.String())
	})
	t.Run("Retired", func(t *testing.T) {
		// Named rather than reported as unknown: a library indexed before they were removed
		// still holds them, and the numbers stay reserved so they are never reused.
		assert.Equal(t, "children", Kind(2).String())
		assert.Equal(t, "background", Kind(3).String())
	})
	t.Run("Unknown", func(t *testing.T) {
		// Shown rather than hidden behind a placeholder, so an unexpected value is visible.
		assert.Equal(t, "9", Kind(9).String())
	})
	t.Run("MatchesSkipMatching", func(t *testing.T) {
		// Everything above RegularFace is excluded from matching, which is what the name has to
		// let a reader see without a lookup table. The two that are not are the two that read as
		// ordinary: the unclassified state and the legacy regular one.
		for _, k := range []Kind{2, 3, AmbiguousFace} {
			assert.Greater(t, int(k), int(RegularFace), "kind %s", k)
		}

		assert.LessOrEqual(t, int(UnclassifiedFace), int(RegularFace))
	})
}

// TestEmbedding_Kind covers the classification a cluster is created with.
//
// Nothing classifies a face any further since the child and background filters were measured to be
// inert, so the only thing it separates is a vector that describes a face from one that does not.
func TestEmbedding_Kind(t *testing.T) {
	t.Run("Regular", func(t *testing.T) {
		assert.Equal(t, RegularFace, Embedding{0.1, 0.2, 0.3}.Kind())
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, UnclassifiedFace, Embedding{}.Kind())
		assert.Equal(t, UnclassifiedFace, Embedding(nil).Kind())
	})
	t.Run("Zero", func(t *testing.T) {
		// A vector with no magnitude sits one unit from every unit vector, so a cluster built from
		// it would accept whatever a model reaching past 1 compares with it. Not a face.
		assert.Equal(t, UnclassifiedFace, Embedding{0, 0, 0}.Kind())
	})
}

// TestEmbeddings_Kind covers the classification of a set, which is the highest any member carries.
func TestEmbeddings_Kind(t *testing.T) {
	t.Run("Regular", func(t *testing.T) {
		assert.Equal(t, RegularFace, Embeddings{{0.1, 0.2}, {0.3, 0.4}}.Kind())
	})
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, UnclassifiedFace, Embeddings{}.Kind())
	})
	t.Run("HighestWins", func(t *testing.T) {
		// One usable member is enough, so a set is not demoted by an empty neighbor.
		assert.Equal(t, RegularFace, Embeddings{{}, {0.1, 0.2}}.Kind())
	})
}
