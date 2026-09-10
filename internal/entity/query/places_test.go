package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestCellIDs(t *testing.T) {
	entity.ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		result, err := CellIDs()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("cell count: %v", len(result))
	})
}
func TestPurgePlaces(t *testing.T) {
	entity.ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		if err := PurgePlaces(); err != nil {
			t.Fatal(err)
		}
	})
}
