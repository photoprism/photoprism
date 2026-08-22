package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestCategoryLabels(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		categories := CategoryLabels(1000, 0)

		assert.GreaterOrEqual(t, 1, len(categories))
		assert.LessOrEqual(t, 1, len(categories))

		for _, r := range categories {
			assert.IsType(t, CategoryLabel{}, r)
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}
		t.Cleanup(func() {
			entity.ResetTestFixtures()
		})
		// Clean the database as if it's brand new
		entity.Entities.Truncate(entity.Db())
		entity.CreateDefaultFixtures()
		entity.FlushCaches()
		entity.File{}.RegenerateIndex()

		categories := CategoryLabels(1000, 0)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 0)
	})
}
