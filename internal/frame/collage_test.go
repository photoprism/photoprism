package frame

import (
	"image"
	"os"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/stretchr/testify/assert"
)

func TestCollage(t *testing.T) {
	t.Run("Polaroid", func(t *testing.T) {
		var images []image.Image

		img, err := imaging.Open("testdata/500x500.jpg")
		assert.NoError(t, err)

		for i := 0; i <= 5; i++ {
			images = append(images, img)
		}

		saveName := "testdata/test-polaroid-collage.jpg"
		preview, err := Collage(Polaroid, images)

		assert.NoError(t, err)

		err = imaging.Save(preview, saveName)

		assert.NoError(t, err)
		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypeJPEG, mimeType)

		_ = os.Remove(saveName)
	})

	t.Run("Two", func(t *testing.T) {
		var images []image.Image

		img, err := imaging.Open("testdata/500x500.jpg")
		assert.NoError(t, err)

		for i := 0; i <= 1; i++ {
			images = append(images, img)
		}

		saveName := "testdata/test-polaroid-collage-two.jpg"
		preview, err := Collage(Polaroid, images)

		assert.NoError(t, err)

		err = imaging.Save(preview, saveName)

		assert.NoError(t, err)
		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypeJPEG, mimeType)

		_ = os.Remove(saveName)
	})

	t.Run("No images", func(t *testing.T) {
		var images []image.Image

		saveName := "testdata/test-no-images-collage.jpg"
		preview, err := Collage(Polaroid, images)

		assert.NoError(t, err)

		err = imaging.Save(preview, saveName)

		assert.NoError(t, err)
		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypeJPEG, mimeType)

		_ = os.Remove(saveName)
	})

	t.Run("Unknown collage type", func(t *testing.T) {
		var images []image.Image

		img, err := imaging.Open("testdata/500x500.jpg")
		assert.NoError(t, err)

		for i := 0; i <= 5; i++ {
			images = append(images, img)
		}

		saveName := "testdata/test-unknown-type-collage.jpg"

		preview, err := Collage("Unknown", images)

		assert.Error(t, err)
		assert.Equal(t, "unknown collage type Unknown", err.Error())

		err = imaging.Save(preview, saveName)

		assert.NoError(t, err)

		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypeJPEG, mimeType)

		_ = os.Remove(saveName)

	})
}
