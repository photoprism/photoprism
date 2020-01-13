package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXMP(t *testing.T) {
	t.Run("photoshop", func(t *testing.T) {
		data, err := XMP("testdata/photoshop.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Night Shift / Berlin / 2020", data.Title)
		assert.Equal(t, "Michael Mayer", data.Artist)
		assert.Equal(t, "Example file for development", data.Description)
		assert.Equal(t, "This is an (edited) legal notice", data.Copyright)
		assert.Equal(t, "HUAWEI", data.CameraMake)
		assert.Equal(t, "ELE-L29", data.CameraModel)
		assert.Equal(t, "HUAWEI P30 Rear Main Camera", data.LensModel)
	})

	t.Run("canon_eos_6d", func(t *testing.T) {
		data, err := XMP("testdata/canon_eos_6d.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Canon", data.CameraMake)
		assert.Equal(t, "Canon EOS 6D", data.CameraModel)
		assert.Equal(t, "EF24-105mm f/4L IS USM", data.LensModel)
	})

	t.Run("iphone_7", func(t *testing.T) {
		data, err := XMP("testdata/iphone_7.xmp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "iPhone 7 / September 2018", data.Title)
		assert.Equal(t, "", data.Artist)
		assert.Equal(t, "", data.Description)
		assert.Equal(t, "", data.Copyright)
		assert.Equal(t, "Apple", data.CameraMake)
		assert.Equal(t, "iPhone 7", data.CameraModel)
		assert.Equal(t, "iPhone 7 back camera 3.99mm f/1.8", data.LensModel)
	})

}
