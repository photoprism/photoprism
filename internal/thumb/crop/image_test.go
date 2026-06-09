package crop

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestThumbFileName(t *testing.T) {
	t.Run("InvalidHash", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("xxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid file hash")
	})
	t.Run("PathMissing", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "path missing")
	})
	t.Run("InvalidWidth", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.000, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid area width")
	})
	t.Run("InvalidCropSize", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 0, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid crop size")
	})
	t.Run("FileNotFound", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "not found")
	})
	t.Run("FileExists", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile500, "", "FaceNet", 500, 500, DefaultOptions}
		r, err := ThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", a, s, "./testdata")
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
}

func TestFileWidth(t *testing.T) {
	t.Run("Tile50", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		assert.Equal(t, 49999, FileWidth(a, Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}))
	})
	t.Run("Tile500", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		assert.Equal(t, 499999, FileWidth(a, Size{Tile500, "", "FaceNet", 500, 500, DefaultOptions}))
	})
}

func TestThumbHash(t *testing.T) {
	t.Run("ValidFilename", func(t *testing.T) {
		assert.Equal(t, "23b05bc917a5aa61382210cedafc162dd3517dc0", thumbHash("23b05bc917a5aa61382210cedafc162dd3517dc0_2048x2048_fit.jpg"))
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		assert.Equal(t, "", thumbHash(""))
	})
}

func TestFindIdealThumbFileName(t *testing.T) {
	t.Run("HashEmpty", func(t *testing.T) {
		r := findIdealThumbFileName("", 500, "path/b")
		assert.Equal(t, "", r)
	})
	t.Run("PathEmpty", func(t *testing.T) {
		r := findIdealThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", 500, "")
		assert.Equal(t, "", r)
	})
	t.Run("FileDoesNotExist", func(t *testing.T) {
		r := findIdealThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", 500, "path/b")
		assert.Equal(t, "", r)
	})
	t.Run("WidthNum500", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 500, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
	t.Run("WidthNum720", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 720, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
	t.Run("WidthNum800", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 800, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
	t.Run("WidthNum60", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 60, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
}

func TestImageFromThumb(t *testing.T) {
	t.Run("Layered16BitTiffThumbnail", func(t *testing.T) {
		prevLibrary := thumb.Library
		thumb.Library = thumb.LibVips
		t.Cleanup(func() {
			thumb.Library = prevLibrary
		})

		cachePath := t.TempDir()
		src := fs.Abs("../../../assets/samples/layered-16bit-small.tif")
		fit720 := thumb.Sizes[thumb.Fit720]

		thumbName, err := thumb.FromFile(src, "bccfeaa526a36e19b555fd4ca5e8f767d5604289", cachePath, fit720.Width, fit720.Height, thumb.OrientationNormal, fit720.Options...)
		if err != nil {
			t.Fatal(err)
		}

		img, cropName, err := ImageFromThumb(thumbName, NewArea("crop", 0, 0, 1, 1), Sizes[Tile50], false)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, cropName)
		assert.Equal(t, filepath.Join(filepath.Dir(thumbName), "bccfeaa526a36e19b555fd4ca5e8f767d5604289_50x50_crop_0000003e83e8.jpg"), cropName)
		assert.Equal(t, 50, img.Bounds().Dx())
		assert.Equal(t, 50, img.Bounds().Dy())
	})
}
