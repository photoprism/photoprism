package thumb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFileInfo(t *testing.T) {
	t.Run("Jpeg", func(t *testing.T) {
		fileName := fs.Abs("./testdata/example.jpg")

		if fileInfo, err := FileInfo(fileName); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 750, fileInfo.Width)
			assert.Equal(t, 500, fileInfo.Height)
		}
	})
	t.Run("BrokenJpeg", func(t *testing.T) {
		fileName := fs.Abs("./testdata/broken.jpg")

		if fileInfo, err := FileInfo(fileName); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 705, fileInfo.Width)
			assert.Equal(t, 725, fileInfo.Height)
		}
	})
	t.Run("Png", func(t *testing.T) {
		fileName := fs.Abs("./testdata/example.png")

		if fileInfo, err := FileInfo(fileName); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 100, fileInfo.Width)
			assert.Equal(t, 67, fileInfo.Height)
		}
	})
	t.Run("Bmp", func(t *testing.T) {
		fileName := fs.Abs("./testdata/example.bmp")

		if fileInfo, err := FileInfo(fileName); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 100, fileInfo.Width)
			assert.Equal(t, 67, fileInfo.Height)
		}
	})
	t.Run("Gif", func(t *testing.T) {
		fileName := fs.Abs("./testdata/example.bmp")

		if fileInfo, err := FileInfo(fileName); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 100, fileInfo.Width)
			assert.Equal(t, 67, fileInfo.Height)
		}
	})
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil.tiff")
		payload := []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}
		require.NoError(t, os.WriteFile(fileName, payload, fs.ModeFile))

		_, err := FileInfo(fileName)

		require.Error(t, err)
	})
}
