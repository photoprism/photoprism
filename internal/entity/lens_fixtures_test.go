package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLensMap_Get(t *testing.T) {
	t.Run("GetExistingLens", func(t *testing.T) {
		r := LensFixtures.Get("lens-f-380")
		assert.Equal(t, uint(1000000), r.ID)
		assert.Equal(t, "lens-f-380", r.LensSlug)
		assert.IsType(t, Lens{}, r)
	})
	t.Run("GetNotExistingLens", func(t *testing.T) {
		r := LensFixtures.Get("Lens 123")
		assert.Equal(t, "lens-123", r.LensSlug)
		assert.IsType(t, Lens{}, r)
	})
}

func TestLensMap_Pointer(t *testing.T) {
	t.Run("GetExistingLensPointer", func(t *testing.T) {
		r := LensFixtures.Pointer("lens-f-380")
		assert.Equal(t, uint(1000000), r.ID)
		assert.Equal(t, "lens-f-380", r.LensSlug)
		assert.IsType(t, &Lens{}, r)
	})
	t.Run("GetNotExistingLensPointer", func(t *testing.T) {
		r := LensFixtures.Pointer("Lens new")
		assert.Equal(t, "lens-new", r.LensSlug)
		assert.IsType(t, &Lens{}, r)
	})
}
