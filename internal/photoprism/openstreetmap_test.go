package photoprism

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMediaFile_Location(t *testing.T) {
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		location, err := mediaFile.Location()
		assert.Nil(t, err)
		assert.Equal(t, "Himeji", location.LocCity)
		assert.Equal(t, "Kinki Region", location.LocState)
		assert.Equal(t, "Japan", location.LocCountry)
		assert.Equal(t, "highway", location.LocCategory)
		assert.Equal(t, 34.7974872, location.LocLat)
		location2, err := mediaFile.Location()
		assert.Nil(t, err)
		assert.Equal(t, "Himeji", location2.LocCity)
		assert.Equal(t, "Kinki Region", location2.LocState)
	})
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		location, err := mediaFile.Location()
		assert.Nil(t, err)
		assert.Equal(t, "Geißwiesenstraße", location.LocStreet)
		assert.Equal(t, "14", location.LocHouseNr)
		assert.Equal(t, "72070", location.LocPostcode)
		assert.Equal(t, "Tübingen", location.LocCity)
		assert.Equal(t, "Landkreis Tübingen", location.LocCounty)
		assert.Equal(t, "Germany", location.LocCountry)
		assert.Equal(t, "building", location.LocCategory)
		assert.Equal(t, 48.53870475, location.LocLat)
	})
	t.Run("/dog_orange.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/dog_orange.jpg")
		assert.Nil(t, err)
		location, err := mediaFile.Location()
		assert.Nil(t, location)
		assert.Equal(t, "no latitude and longitude in image metadata", err.Error())
	})
}
