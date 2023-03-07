package thumb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestJpeg(t *testing.T) {
	formats := []string{"bmp", "gif", "png", "tif"}

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
			dst := "testdata/example." + ext + fs.ExtJPEG

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
		t.Run("invalid orientation", func(t *testing.T) {
			src := "testdata/example." + ext
			dst := "testdata/example." + ext + fs.ExtJPEG

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

	t.Run("foo", func(t *testing.T) {
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
}
