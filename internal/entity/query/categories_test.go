package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
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
		t.Cleanup(func() {
			entity.Entities.Truncate(entity.Db())
			entity.CreateDefaultFixtures()
			entity.CreateTestFixtures()
			entity.File{}.RegenerateIndex()
		})
		// Clean the database as if it's brand new
		entity.Entities.Truncate(entity.Db())
		entity.CreateDefaultFixtures()
		entity.File{}.RegenerateIndex()

		categories := CategoryLabels(1000, 0)
		assert.NotNil(t, categories)
		assert.Len(t, categories, 0)
	})
}
