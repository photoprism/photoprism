package photoprism

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/s2"
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

		if err = location.Find("places"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "姫路市", location.City())
		assert.Equal(t, "兵庫県", location.State())
		assert.Equal(t, "Japan", location.CountryName())
		assert.Equal(t, "", location.Category())
		assert.True(t, strings.HasPrefix(location.ID, s2.TokenPrefix+"3554df45"))
		location2, err := mediaFile.Location()

		if err != nil {
			t.Fatal(err)
		}

		if err = location.Find("places"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "姫路市", location2.City())
		assert.Equal(t, "兵庫県", location2.State())
	})
	t.Run("cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		f, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")

		if err != nil {
			t.Fatal(err)
		}

		loc, err := f.Location()

		if err != nil {
			t.Fatal(err)
		}

		if err = loc.Find("places"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Tübingen", loc.City())
		assert.Equal(t, "de", loc.CountryCode())
		assert.Equal(t, "Germany", loc.CountryName())
		assert.True(t, strings.HasPrefix(loc.ID, s2.TokenPrefix+"4799e4a5"))
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
			assert.Equal(t, "media: found no latitude and longitude", err.Error())
		}
	})
	t.Run("Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")

		if err != nil {
			t.Fatal(err)
		}

		location, err := mediaFile.Location()

		assert.Error(t, err, "metadata: found no exif header in Random.docx")
		assert.Nil(t, location)
	})
}
