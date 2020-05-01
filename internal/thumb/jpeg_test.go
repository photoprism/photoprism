package thumb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJpeg(t *testing.T) {
	formats := []string{"bmp", "gif", "png", "tif"}

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + ".jpg"

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
	}

	t.Run("foo", func(t *testing.T) {

		src := "testdata/example.foo"
		dst := "testdata/example.foo.jpg"

		assert.NoFileExists(t, dst)

		img, err := Jpeg(src, dst)

		assert.NoFileExists(t, dst)

		if img != nil {
			t.Fatal("img should be nil")
		}

		assert.Error(t, err)
	})
}
