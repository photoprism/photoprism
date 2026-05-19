package thumb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestOpen(t *testing.T) {
	t.Run("JPEG", func(t *testing.T) {
		img, err := Open("testdata/example.jpg", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("BMP", func(t *testing.T) {
		img, err := Open("testdata/example.bmp", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("GIF", func(t *testing.T) {
		img, err := Open("testdata/example.gif", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("PNG", func(t *testing.T) {
		img, err := Open("testdata/example.png", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("TIFF", func(t *testing.T) {
		img, err := Open("testdata/example.tif", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("Layered16BitTiff", func(t *testing.T) {
		img, err := Open(fs.Abs("../../assets/samples/layered-16bit-small.tif"), 0)
		require.NoError(t, err)
		require.NotNil(t, img)
		assert.Equal(t, 236, img.Bounds().Dx())
		assert.Equal(t, 158, img.Bounds().Dy())
	})
	t.Run("TiffExplicitOrientation", func(t *testing.T) {
		img, err := Open(fs.Abs("../photoprism/testdata/rotate/6.tiff"), OrientationRotate90)
		require.NoError(t, err)
		require.NotNil(t, img)
		assert.Equal(t, 43, img.Bounds().Dx())
		assert.Equal(t, 65, img.Bounds().Dy())
	})
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil.tiff")
		payload := []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}
		require.NoError(t, os.WriteFile(fileName, payload, fs.ModeFile))

		img, err := Open(fileName, 0)

		require.Error(t, err)
		require.Nil(t, img)
	})
}
