package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenJpeg(t *testing.T) {
	t.Run("testdata/example.jpg", func(t *testing.T) {
		img, err := OpenJpeg("testdata/example.jpg", 0)

		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("testdata/broken.jpg", func(t *testing.T) {
		img, err := OpenJpeg("testdata/broken.jpg", 0)

		assert.Error(t, err)
		assert.Nil(t, img)
	})
	t.Run("testdata/fixed.jpg", func(t *testing.T) {
		img, err := OpenJpeg("testdata/fixed.jpg", 0)

		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("filename empty", func(t *testing.T) {
		img, err := OpenJpeg("", 0)

		assert.Error(t, err)
		assert.Equal(t, "filename missing", err.Error())
		assert.Nil(t, img)
	})
}
