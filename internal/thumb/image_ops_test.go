package thumb

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestSave(t *testing.T) {
	t.Run("JPEG", func(t *testing.T) {
		dst := filepath.Join(t.TempDir(), "test.jpg")
		img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

		err := Save(img, dst)
		require.NoError(t, err)
		assert.FileExists(t, dst)

		// Verify it's a valid JPEG.
		mime := fs.MimeType(dst)
		assert.Equal(t, "image/jpeg", mime)
	})
	t.Run("PNG", func(t *testing.T) {
		dst := filepath.Join(t.TempDir(), "test.png")
		img := image.NewNRGBA(image.Rect(0, 0, 10, 10))

		err := Save(img, dst)
		require.NoError(t, err)
		assert.FileExists(t, dst)

		mime := fs.MimeType(dst)
		assert.Equal(t, "image/png", mime)
	})
	t.Run("NilImage", func(t *testing.T) {
		dst := filepath.Join(t.TempDir(), "test.jpg")
		err := Save(nil, dst)
		assert.Error(t, err)
		assert.NoFileExists(t, dst)
	})
	t.Run("InvalidDir", func(t *testing.T) {
		err := Save(image.NewNRGBA(image.Rect(0, 0, 1, 1)), "/nonexistent/dir/test.jpg")
		assert.Error(t, err)
	})
	t.Run("NoPartialFileOnEncodeSuccess", func(t *testing.T) {
		dir := t.TempDir()
		dst := filepath.Join(dir, "output.jpg")
		img := image.NewNRGBA(image.Rect(0, 0, 2, 2))

		require.NoError(t, Save(img, dst))
		assert.FileExists(t, dst)

		// No temp files should remain.
		entries, _ := os.ReadDir(dir)
		assert.Equal(t, 1, len(entries), "only the final file should remain")
	})
}

func TestCloneImage(t *testing.T) {
	t.Run("NonOriginBounds", func(t *testing.T) {
		// SubImage returns an image with non-zero Min.
		src := image.NewNRGBA(image.Rect(0, 0, 10, 10))
		src.SetNRGBA(5, 5, color.NRGBA{R: 200, A: 255})
		sub := src.SubImage(image.Rect(5, 5, 10, 10))

		clone := cloneImage(sub)
		assert.Equal(t, 0, clone.Bounds().Min.X)
		assert.Equal(t, 0, clone.Bounds().Min.Y)
		assert.Equal(t, 5, clone.Bounds().Dx())
		assert.Equal(t, 5, clone.Bounds().Dy())

		// Verify pixel was preserved.
		r, _, _, _ := clone.At(0, 0).RGBA()
		assert.Equal(t, uint32(200<<8|200), r)
	})
}

func TestResizeImage(t *testing.T) {
	t.Run("ZeroSource", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 0, 0))
		result := resizeImage(src, 100, 100, ResampleLanczos)
		assert.Equal(t, 0, result.Bounds().Dx())
		assert.Equal(t, 0, result.Bounds().Dy())
	})
	t.Run("IdentitySize", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 50, 50))
		result := resizeImage(src, 50, 50, ResampleLanczos)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
	t.Run("Downscale", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := resizeImage(src, 50, 100, ResampleLanczos)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 100, result.Bounds().Dy())
	})
	t.Run("WidthZeroPreservesRatio", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := resizeImage(src, 0, 100, ResampleLanczos)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 100, result.Bounds().Dy())
	})
	t.Run("HeightZeroPreservesRatio", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := resizeImage(src, 50, 0, ResampleLanczos)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 100, result.Bounds().Dy())
	})
	t.Run("BothZeroReturnsClone", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := resizeImage(src, 0, 0, ResampleLanczos)
		assert.Equal(t, 100, result.Bounds().Dx())
		assert.Equal(t, 200, result.Bounds().Dy())
	})
}

func TestFitImage(t *testing.T) {
	t.Run("FitLarger", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := fitImage(src, 50, 200, ResampleLanczos)
		// Scale limited by width: 50/100 = 0.5, so 100*0.5 = 50, 200*0.5 = 100.
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 100, result.Bounds().Dy())
	})
	t.Run("NoUpscale", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 50, 50))
		result := fitImage(src, 200, 200, ResampleLanczos)
		// fitImage caps scale at 1.0, so image stays at 50x50.
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
	t.Run("ZeroDimension", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 50, 50))
		result := fitImage(src, 0, 100, ResampleLanczos)
		// Returns clone when width is 0.
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
}

func TestFillImage(t *testing.T) {
	t.Run("Center", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := fillImage(src, 50, 50, ResampleLanczos, cropAnchorCenter)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
	t.Run("TopLeft", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := fillImage(src, 50, 50, ResampleLanczos, cropAnchorTopLeft)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
	t.Run("BottomRight", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 200))
		result := fillImage(src, 50, 50, ResampleLanczos, cropAnchorBottomRight)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
	t.Run("ZeroDimension", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 50, 50))
		result := fillImage(src, 0, 50, ResampleLanczos, cropAnchorCenter)
		assert.Equal(t, 50, result.Bounds().Dx())
		assert.Equal(t, 50, result.Bounds().Dy())
	})
}

func TestCropImage(t *testing.T) {
	t.Run("ValidRegion", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 100, 100))
		result := cropImage(src, image.Rect(10, 10, 50, 50))
		assert.Equal(t, 40, result.Bounds().Dx())
		assert.Equal(t, 40, result.Bounds().Dy())
	})
	t.Run("OutOfBounds", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 10, 10))
		result := cropImage(src, image.Rect(20, 20, 30, 30))
		// No overlap — returns empty image.
		assert.Equal(t, 0, result.Bounds().Dx())
		assert.Equal(t, 0, result.Bounds().Dy())
	})
	t.Run("PartialOverlap", func(t *testing.T) {
		src := image.NewNRGBA(image.Rect(0, 0, 10, 10))
		result := cropImage(src, image.Rect(5, 5, 20, 20))
		assert.Equal(t, 5, result.Bounds().Dx())
		assert.Equal(t, 5, result.Bounds().Dy())
	})
}

func TestResampleScaler(t *testing.T) {
	t.Run("Linear", func(t *testing.T) {
		assert.NotNil(t, resampleScaler(ResampleLinear))
	})
	t.Run("Nearest", func(t *testing.T) {
		assert.NotNil(t, resampleScaler(ResampleNearest))
	})
	t.Run("DefaultCatmullRom", func(t *testing.T) {
		assert.NotNil(t, resampleScaler(ResampleLanczos))
	})
}
