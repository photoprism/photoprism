package photoprism

import (
	"strings"
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

		if err = location.Find(conf.Db(), "places"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Takasago", location.City())
		assert.Equal(t, "Hyogo Prefecture", location.State())
		assert.Equal(t, "Japan", location.CountryName())
		assert.Equal(t, "", location.Category())
		assert.True(t, strings.HasPrefix(location.ID, "3554df45"))
		location2, err := mediaFile.Location()

		if err != nil {
			t.Fatal(err)
		}

		if err = location.Find(conf.Db(), "places"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Takasago", location2.City())
		assert.Equal(t, "Hyogo Prefecture", location2.State())
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

		if err = location.Find(conf.Db(), "places"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Tübingen", location.City())
		assert.Equal(t, "de", location.CountryCode())
		assert.Equal(t, "Germany", location.CountryName())
		assert.True(t, strings.HasPrefix(location.ID, "4799e4a5"))
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
			assert.Equal(t, "file: no latitude and longitude in metadata", err.Error())
		}
	})
	t.Run("Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")

		if err != nil {
			t.Fatal(err)
		}

		location, err := mediaFile.Location()

		assert.Error(t, err, "meta: no exif data")
		assert.Nil(t, location)

	})
}
