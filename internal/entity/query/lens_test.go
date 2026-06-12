package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestFindLensBySlug(t *testing.T) {
	t.Run("ExistingLens", func(t *testing.T) {
		lens := FindLensBySlug(entity.LensFixtures.Get("4.15mm-f/2.2").LensSlug)
		assert.NotNil(t, lens)
		if lens != nil {
			assert.Equal(t, entity.LensFixtures.Get("4.15mm-f/2.2").ID, lens.ID)
			assert.Equal(t, entity.LensFixtures.Get("4.15mm-f/2.2").LensModel, lens.LensModel)
		}
	})
	t.Run("NotExistingLens", func(t *testing.T) {
		lens := FindLensBySlug("IAmNotValid")
		assert.Nil(t, lens)
	})
}

func TestFindLensByID(t *testing.T) {
	t.Run("ExistingLens", func(t *testing.T) {
		lens := FindLensByID(entity.LensFixtures.Get("4.15mm-f/2.2").ID)
		assert.NotNil(t, lens)
		if lens != nil {
			assert.Equal(t, entity.LensFixtures.Get("4.15mm-f/2.2").ID, lens.ID)
			assert.Equal(t, entity.LensFixtures.Get("4.15mm-f/2.2").LensModel, lens.LensModel)
		}
	})
	t.Run("NotExistingLens", func(t *testing.T) {
		lens := FindLensByID(99885541348)
		assert.Nil(t, lens)
	})
}
