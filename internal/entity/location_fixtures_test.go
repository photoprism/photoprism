package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocationMap_Get(t *testing.T) {
	t.Run("get existing location", func(t *testing.T) {
		r := LocationFixtures.Get("mexico")
		assert.Equal(t, "Adosada Platform", r.LocName)
		assert.Equal(t, "s2:85d1ea7d382c", r.ID)
		assert.IsType(t, Location{}, r)
	})
	t.Run("get not existing location", func(t *testing.T) {
		r := LocationFixtures.Get("Fusion 3333")
		assert.Equal(t, "zz", r.ID)
		assert.IsType(t, Location{}, r)
	})
}

func TestLocationMap_Pointer(t *testing.T) {
	t.Run("get existing location pointer", func(t *testing.T) {
		r := LocationFixtures.Pointer("mexico")
		assert.Equal(t, "Adosada Platform", r.LocName)
		assert.Equal(t, "s2:85d1ea7d382c", r.ID)
		assert.IsType(t, &Location{}, r)
	})
	t.Run("get not existing location pointer", func(t *testing.T) {
		r := LocationFixtures.Pointer("Fusion 444")
		assert.Equal(t, "zz", r.ID)
		assert.IsType(t, &Location{}, r)
	})
}
