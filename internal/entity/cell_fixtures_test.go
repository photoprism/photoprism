package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocationMap_Get(t *testing.T) {
	t.Run("get existing location", func(t *testing.T) {
		r := CellFixtures.Get("mexico")
		assert.Equal(t, "Adosada Platform", r.CellName)
		assert.Equal(t, "s2:85d1ea7d382c", r.ID)
		assert.IsType(t, Cell{}, r)
	})
	t.Run("get not existing location", func(t *testing.T) {
		r := CellFixtures.Get("Fusion 3333")
		assert.Equal(t, "zz", r.ID)
		assert.IsType(t, Cell{}, r)
	})
}

func TestLocationMap_Pointer(t *testing.T) {
	t.Run("get existing location pointer", func(t *testing.T) {
		r := CellFixtures.Pointer("mexico")
		assert.Equal(t, "Adosada Platform", r.CellName)
		assert.Equal(t, "s2:85d1ea7d382c", r.ID)
		assert.IsType(t, &Cell{}, r)
	})
	t.Run("get not existing location pointer", func(t *testing.T) {
		r := CellFixtures.Pointer("Fusion 444")
		assert.Equal(t, "zz", r.ID)
		assert.IsType(t, &Cell{}, r)
	})
}
