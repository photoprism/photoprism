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
		// What every cluster is created with since the classifiers were removed, so a column of
		// them would say nothing. Blank rather than named, so a value that is there to be read is
		// one that differs from it.
		assert.Equal(t, "", UnclassifiedFace.String())
		assert.Equal(t, "", Kind(0).String())
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
