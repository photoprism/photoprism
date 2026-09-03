package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestCellIDs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := CellIDs()

		if assert.Nil(t, err) {
			t.Logf("cell count: %v", len(result))
			assert.GreaterOrEqual(t, len(result), 8)
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

		result, err := CellIDs()

		if assert.Nil(t, err) {
			assert.NotNil(t, result)
			assert.Len(t, result, 0)
		}
	})
}
func TestPurgePlaces(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := PurgePlaces(); err != nil {
			t.Fatal(err)
		}
	})
}
