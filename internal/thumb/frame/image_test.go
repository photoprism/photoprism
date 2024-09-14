package frame

import (
	"os"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/stretchr/testify/assert"
)

func TestImage(t *testing.T) {
	t.Run("Polaroid", func(t *testing.T) {
		img, err := imaging.Open("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := "testdata/test-image.png"

		out, err := Image(Polaroid, img, RandomAngle(30))

		assert.NoError(t, err)

		err = imaging.Save(out, saveName)

		assert.NoError(t, err)
		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypePNG, mimeType)

		_ = os.Remove(saveName)
	})

	t.Run("TypeUnknown", func(t *testing.T) {
		img, err := imaging.Open("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := "testdata/test-image.png"

		out, err := Image("unknown", img, RandomAngle(30))

		assert.Error(t, err)
		assert.Equal(t, "unknown collage type unknown", err.Error())

		err = imaging.Save(out, saveName)

		assert.NoError(t, err)
		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypePNG, mimeType)

		_ = os.Remove(saveName)
	})
}
