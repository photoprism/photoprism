package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

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
		assert.Equal(t, "zz", country.CountrySlug)
	})
}
func TestCountry_FirstOrCreate(t *testing.T) {
	t.Run("country already existing", func(t *testing.T) {
		country := entity.NewCountry("de", "Germany")
		country.FirstOrCreate()
		assert.Equal(t, "de", country.Code())
		assert.Equal(t, "Germany", country.Name())
		assert.Equal(t, "Country description", country.CountryDescription)
		assert.Equal(t, "Country Notes", country.CountryNotes)
		assert.Equal(t, uint(0), country.CountryPhotoID)
	})
	t.Run("country not yet existing", func(t *testing.T) {
		country := entity.NewCountry("wl", "Wonder Land")
		country.FirstOrCreate()
		assert.Equal(t, "wl", country.Code())
		assert.Equal(t, "Wonder Land", country.Name())
	})
}
