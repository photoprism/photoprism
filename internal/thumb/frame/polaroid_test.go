package frame

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolaroid(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "testdata")
	require.NoError(t, os.MkdirAll(basePath, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(basePath) })

	t.Run("RandomAngle", func(t *testing.T) {
		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		saveName := filepath.Join(basePath, "test-polaroid.png")

		out, err := polaroid(img, RandomAngle(30))

		assert.NoError(t, err)

		err = thumb.Save(out, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypePng, mimeType)

		_ = os.Remove(saveName)
	})
}
