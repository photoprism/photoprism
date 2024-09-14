package frame

import (
	"os"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/stretchr/testify/assert"
)

func TestPolaroid(t *testing.T) {
	t.Run("RandomAngle", func(t *testing.T) {
		img, err := imaging.Open("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := "testdata/test-polaroid.png"

		out, err := polaroid(img, RandomAngle(30))

		assert.NoError(t, err)

		err = imaging.Save(out, saveName)

		assert.NoError(t, err)
		mimeType := fs.MimeType(saveName)
		assert.Equal(t, fs.MimeTypePNG, mimeType)

		_ = os.Remove(saveName)
	})
}
