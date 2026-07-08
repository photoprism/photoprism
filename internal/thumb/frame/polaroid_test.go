package frame

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"

	"github.com/stretchr/testify/assert"
)

func TestPolaroid(t *testing.T) {
	t.Run("RandomAngle", func(t *testing.T) {
		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := "testdata/test-polaroid.png"

		out, err := polaroid(img, RandomAngle(30))

		assert.NoError(t, err)

		err = thumb.Save(out, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypePng, mimeType)

		_ = os.Remove(saveName)
	})
}
