package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_Location(t *testing.T) {
	t.Run("iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		location, err := mediaFile.Location()

		if err != nil {
			t.Fatal(err)
		}

		if err = location.Find(conf.Db()); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Himeji", location.City())
		assert.Equal(t, "Kinki Region", location.State())
		assert.Equal(t, "Japan", location.CountryName())
		assert.Equal(t, "", location.Category())
		assert.Equal(t, 34.79745, location.Latitude())
		assert.Equal(t, uint64(0x1554df45c0000000), location.ID)
		location2, err := mediaFile.Location()

		if err != nil {
			t.Fatal(err)
		}

		if err = location.Find(conf.Db()); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Himeji", location2.LocCity)
		assert.Equal(t, "Kinki Region", location2.LocState)
	})
	t.Run("cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		if err != nil {
			t.Fatal(err)
		}

		location, err := mediaFile.Location()

		if err != nil {
			t.Fatal(err)
		}

		if err = location.Find(conf.Db()); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Tübingen", location.City())
		assert.Equal(t, "de", location.CountryCode())
		assert.Equal(t, "Germany", location.CountryName())
		assert.Equal(t, 48.53870833333333, location.Latitude())
		assert.Equal(t, uint64(0x799e4a5c0000000), location.ID)
	})
	t.Run("dog_orange.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/dog_orange.jpg")

		if err != nil {
			t.Fatal(err)
		}

		if _, err := mediaFile.Location(); err == nil {
			t.Fatal("mediaFile.Location() should return error")
		} else {
			assert.Equal(t, "mediafile: no latitude and longitude in image metadata", err.Error())
		}
	})
}
