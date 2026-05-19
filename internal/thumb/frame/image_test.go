package frame

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"

	"github.com/stretchr/testify/assert"
)

func TestImage(t *testing.T) {
	t.Run("Polaroid", func(t *testing.T) {
		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := "testdata/test-image.png"

		out, err := Image(Polaroid, img, RandomAngle(30))

		assert.NoError(t, err)

		err = thumb.Save(out, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypePng, mimeType)

		_ = os.Remove(saveName)
	})
	t.Run("TypeUnknown", func(t *testing.T) {
		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := "testdata/test-image.png"

		out, err := Image("unknown", img, RandomAngle(30))

		assert.Error(t, err)
		assert.Equal(t, "unknown collage type unknown", err.Error())

		err = thumb.Save(out, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypePng, mimeType)

		_ = os.Remove(saveName)
	})
}
