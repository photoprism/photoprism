package thumb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestJpeg(t *testing.T) {
	formats := []string{"bmp", "gif", "png", "tif"}

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationRotate270)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 67, bounds.Max.X)
			assert.Equal(t, 100, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationFlipH", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationFlipH)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationFlipV", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationFlipV)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationRotate90", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationRotate90)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 67, bounds.Max.X)
			assert.Equal(t, 100, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationRotate180", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationRotate180)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationTranspose", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationTranspose)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 67, bounds.Max.X)
			assert.Equal(t, 100, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationTransverse", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationTransverse)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 67, bounds.Max.X)
			assert.Equal(t, 100, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationUnspecified", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationUnspecified)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("OrientationNormal", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, OrientationNormal)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("InvalidOrientation", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJpeg

			assert.NoFileExists(t, dst)

			img, err := Jpeg(src, dst, 500)

			if err != nil {
				t.Fatal(err)
			}

			assert.FileExists(t, dst)

			bounds := img.Bounds()
			assert.Equal(t, 100, bounds.Max.X)
			assert.Equal(t, 67, bounds.Max.Y)

			if err := os.Remove(dst); err != nil {
				t.Fatal(err)
			}
		})
	}

	t.Run("Foo", func(t *testing.T) {
		src := "testdata/example.foo"
		dst := "testdata/example.foo.jpg"

		assert.NoFileExists(t, dst)

		img, err := Jpeg(src, dst, OrientationFlipV)

		assert.NoFileExists(t, dst)

		if img != nil {
			t.Fatal("img should be nil")
		}

		assert.Error(t, err)
	})
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "evil.tiff")
		dst := filepath.Join(t.TempDir(), "evil.jpg")
		payload := []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}
		require.NoError(t, os.WriteFile(src, payload, fs.ModeFile))

		img, err := Jpeg(src, dst, OrientationNormal)

		assert.NoFileExists(t, dst)
		assert.Nil(t, img)
		assert.Error(t, err)
	})
	t.Run("MalformedBigEndianTiffIfdOffset", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "evil-mm.tiff")
		dst := filepath.Join(t.TempDir(), "evil-mm.jpg")
		payload := []byte{0x4d, 0x4d, 0x00, 0x2a, 0xff, 0xff, 0xff, 0xff}
		require.NoError(t, os.WriteFile(src, payload, fs.ModeFile))

		img, err := Jpeg(src, dst, OrientationNormal)

		assert.NoFileExists(t, dst)
		assert.Nil(t, img)
		assert.Error(t, err)
	})
	t.Run("TruncatedTiffBody", func(t *testing.T) {
		src := filepath.Join(t.TempDir(), "truncated.tiff")
		dst := filepath.Join(t.TempDir(), "truncated.jpg")
		payload := []byte{0x49, 0x49, 0x2a, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00}
		require.NoError(t, os.WriteFile(src, payload, fs.ModeFile))

		img, err := Jpeg(src, dst, OrientationNormal)

		assert.NoFileExists(t, dst)
		assert.Nil(t, img)
		require.Error(t, err)
	})
}
