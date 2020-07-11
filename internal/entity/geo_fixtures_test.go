package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocationMap_Get(t *testing.T) {
	t.Run("get existing location", func(t *testing.T) {
		r := GeoFixtures.Get("mexico")
		assert.Equal(t, "Adosada Platform", r.GeoName)
		assert.Equal(t, "s2:85d1ea7d382c", r.ID)
		assert.IsType(t, Geo{}, r)
	})
	t.Run("get not existing location", func(t *testing.T) {
		r := GeoFixtures.Get("Fusion 3333")
		assert.Equal(t, "zz", r.ID)
		assert.IsType(t, Geo{}, r)
	})
}

func TestLocationMap_Pointer(t *testing.T) {
	t.Run("get existing location pointer", func(t *testing.T) {
		r := GeoFixtures.Pointer("mexico")
		assert.Equal(t, "Adosada Platform", r.GeoName)
		assert.Equal(t, "s2:85d1ea7d382c", r.ID)
		assert.IsType(t, &Geo{}, r)
	})
	t.Run("get not existing location pointer", func(t *testing.T) {
		r := GeoFixtures.Pointer("Fusion 444")
		assert.Equal(t, "zz", r.ID)
		assert.IsType(t, &Geo{}, r)
	})
}
