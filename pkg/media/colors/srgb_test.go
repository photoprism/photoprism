package colors

import (
	"errors"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeImage(path string, img image.Image) (err error) {
	imgFile, err := os.Create(path) //nolint:gosec // test temp file

	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, imgFile.Close())
	}()

	opt := jpeg.Options{
		Quality: 95,
	}

	err = jpeg.Encode(imgFile, img, &opt)
	return err
}

func TestToSRGB(t *testing.T) {
	t.Run("DisplayP3", func(t *testing.T) {
		testFile, _ := filepath.Abs("./testdata/DisplayP3.jpg")

		t.Logf("testfile: %s", testFile)

		imgFile, err := os.Open(testFile) //nolint:gosec // test temp file

		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			assert.NoError(t, imgFile.Close())
		})

		img, _, err := image.Decode(imgFile)

		if err != nil {
			t.Fatal(err)
		}

		imgSRGB := ToSRGB(img, ProfileDisplayP3)

		srgbFile := "./testdata/SRGB.jpg"

		if err := writeImage(srgbFile, imgSRGB); err != nil {
			t.Error(err)
		}

		assert.FileExists(t, srgbFile)

		_ = os.Remove(srgbFile)
	})
}
