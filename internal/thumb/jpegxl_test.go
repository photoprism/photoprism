package thumb

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJpegXLSupported(t *testing.T) {
	// The development and CI runtime ships libvips with the JXL module (libjxl).
	assert.True(t, JpegXLSupported())
}

func TestJpeg_JpegXL(t *testing.T) {
	if !JpegXLSupported() {
		t.Skip("libvips was built without JPEG XL support")
	}

	src := filepath.Join(SamplesPath, "dice.jxl")
	dst := filepath.Join(t.TempDir(), "dice.jxl.jpg")

	img, err := Jpeg(src, dst, OrientationNormal)
	require.NoError(t, err)
	require.NotNil(t, img)

	assert.FileExists(t, dst)
	assert.Greater(t, img.Bounds().Max.X, 0)
	assert.Greater(t, img.Bounds().Max.Y, 0)
}
