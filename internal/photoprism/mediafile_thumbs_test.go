package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/thumb"
)

func TestMediaFile_Thumbnail(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	thumbsPath := conf.CachePath() + "/.test_mediafile_thumbnail"

	defer os.RemoveAll(thumbsPath)

	t.Run("elephants.jpg", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, "tile_500")

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbnail)
	})
	t.Run("invalid image format", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.xmp")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, "tile_500")

		assert.EqualError(t, err, "media: failed to create thumbnail for canon_eos_6d.xmp (image: unknown format)")

		t.Log(thumbnail)
	})
	t.Run("invalid thumbnail type", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, "invalid_500")

		assert.EqualError(t, err, "media: invalid type invalid_500")

		t.Log(thumbnail)
	})
}

func TestMediaFile_Resample(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	thumbsPath := conf.CachePath() + "/.test_mediafile_resample"

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(thumbsPath)

	t.Run("elephants.jpg", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Resample(thumbsPath, thumb.Tile500)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, thumbnail)

	})
	t.Run("invalid type", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Resample(thumbsPath, "xxx_500")

		if err == nil {
			t.Fatal("err should not be nil")
		}

		assert.Equal(t, "media: invalid type xxx_500", err.Error())
		assert.Empty(t, thumbnail)
	})

}

func TestMediaFile_CreateThumbnailSize(t *testing.T) {
	t.Run("elephants.jpg", func(t *testing.T) {
		m, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "elephants.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, m.CreateThumbnailSize(thumb.SizeColors))
		assert.True(t, m.CreateThumbnailSize(thumb.SizeTile100))
		assert.True(t, m.CreateThumbnailSize(thumb.SizeTile224))
		assert.True(t, m.CreateThumbnailSize(thumb.SizeTile500))
		assert.True(t, m.CreateThumbnailSize(thumb.SizeFit720))
		assert.False(t, m.CreateThumbnailSize(thumb.SizeFit1280))
		assert.False(t, m.CreateThumbnailSize(thumb.SizeFit1920))
	})
}

func TestMediaFile_CreateThumbnails(t *testing.T) {
	c := config.TestConfig()

	thumbsPath := "./.test_mediafile_createthumbnails"

	if p, err := filepath.Abs(thumbsPath); err != nil {
		t.Fatal(err)
	} else {
		thumbsPath = p
	}

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(thumbsPath)

	if err := c.CreateDirectories(); err != nil {
		t.Fatal(err)
	}

	t.Run("elephants.jpg", func(t *testing.T) {
		m, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "elephants.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		err = m.CreateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.CreateThumbnails(thumbsPath, false))
	})

	t.Run("animated-earth.jpg", func(t *testing.T) {
		m, err := NewMediaFile("testdata/animated-earth.jpg")

		if err != nil {
			t.Fatal(err)
		}

		err = m.CreateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.CreateThumbnails(thumbsPath, false))
	})

	t.Run("photoprism.png", func(t *testing.T) {
		m, err := NewMediaFile("testdata/photoprism.png")

		if err != nil {
			t.Fatal(err)
		}

		err = m.CreateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.CreateThumbnails(thumbsPath, false))
	})

	t.Run("broken/animated-earth.jpg", func(t *testing.T) {
		m, err := NewMediaFile("testdata/broken/animated-earth.jpg")

		if err != nil {
			t.Fatal(err)
		}

		err = m.CreateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.CreateThumbnails(thumbsPath, false))
	})
}

func TestMediaFile_ChangeOrientation(t *testing.T) {
	t.Run("JPEG", func(t *testing.T) {
		m, err := NewMediaFile("testdata/orientation.jpg")

		if err != nil {
			t.Fatal(err)
		}

		orig := m.Orientation()

		if err = m.ChangeOrientation(8); err != nil {
			t.Fatal(err)
		}

		if err = m.ChangeOrientation(orig); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("PNG", func(t *testing.T) {
		m, err := NewMediaFile("testdata/orientation.png")

		if err != nil {
			t.Fatal(err)
		}

		orig := m.Orientation()

		if err = m.ChangeOrientation(8); err != nil {
			t.Fatal(err)
		}

		if err = m.ChangeOrientation(orig); err != nil {
			t.Fatal(err)
		}
	})
}
