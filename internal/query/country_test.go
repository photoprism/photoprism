package query

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCountry(t *testing.T) {
	t.Run("name Fantasy code fy", func(t *testing.T) {
		country := entity.NewCountry("fy", "Fantasy")
		assert.Equal(t, "fy", country.ID)
		assert.Equal(t, "Fantasy", country.CountryName)
		assert.Equal(t, "fantasy", country.CountrySlug)
	})
	t.Run("name Unknown code Unknown", func(t *testing.T) {
		country := entity.NewCountry("", "")
		assert.Equal(t, "zz", country.ID)
		assert.Equal(t, "Unknown", country.CountryName)
		assert.Equal(t, "unknown", country.CountrySlug)
	})
}
func TestCountry_FirstOrCreate(t *testing.T) {
	t.Run("country already existing", func(t *testing.T) {
		country := entity.NewCountry("pp", "PhotoPrism")
		c := config.TestConfig()
		country.FirstOrCreate(c.Db())
		assert.Equal(t, "pp", country.Code())
		assert.Equal(t, "PhotoPrism", country.Name())

	})
	t.Run("country not yet existing", func(t *testing.T) {
		country := entity.NewCountry("wl", "Wonder Land")
		c := config.TestConfig()
		country.FirstOrCreate(c.Db())
		assert.Equal(t, "wl", country.Code())
		assert.Equal(t, "Wonder Land", country.Name())
	})
}
