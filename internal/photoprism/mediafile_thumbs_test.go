package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMediaFile_Thumbnail(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	thumbsPath := conf.CachePath() + "/.test_mediafile_thumbnail"

	defer os.RemoveAll(thumbsPath)

	t.Run("ElephantsJpg", func(t *testing.T) {
		image, err := NewMediaFile(conf.SamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, "tile_500")

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbnail)
	})
	t.Run("Layered16BitTiff", func(t *testing.T) {
		image, err := NewMediaFile(testSamplesPath + "/layered-16bit-small.tif")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, thumb.Fit720)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbnail)

		img, _, err := fs.DecodeImageFile(thumbnail)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 236, img.Bounds().Dx())
		assert.Equal(t, 158, img.Bounds().Dy())
	})
	t.Run("RotateSixTiff", func(t *testing.T) {
		image, err := NewMediaFile("testdata/rotate/6.tiff")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, thumb.Fit720)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbnail)

		img, _, err := fs.DecodeImageFile(thumbnail)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 43, img.Bounds().Dx())
		assert.Equal(t, 65, img.Bounds().Dy())
	})
	t.Run("InvalidImageFormat", func(t *testing.T) {
		image, err := NewMediaFile(conf.SamplesPath() + "/canon_eos_6d.xmp")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Thumbnail(thumbsPath, "tile_500")

		assert.EqualError(t, err, "media: failed to create thumbnail for canon_eos_6d.xmp (unsupported image format)")

		t.Log(thumbnail)
	})
	t.Run("InvalidThumbnailType", func(t *testing.T) {
		image, err := NewMediaFile(conf.SamplesPath() + "/elephants.jpg")

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

	t.Run("ElephantsJpg", func(t *testing.T) {
		image, err := NewMediaFile(conf.SamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Resample(thumbsPath, thumb.Tile500)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, thumbnail)

	})
	t.Run("InvalidType", func(t *testing.T) {
		image, err := NewMediaFile(conf.SamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := image.Resample(thumbsPath, "xxx_500")

		if err == nil {
			t.Fatal("err must not be nil")
		}

		assert.Equal(t, "media: invalid type xxx_500", err.Error())
		assert.Empty(t, thumbnail)
	})

}

func TestMediaFile_SkipThumbnailSize(t *testing.T) {
	t.Run("ElephantsJpg", func(t *testing.T) {
		m, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "elephants.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, m.SkipThumbnailSize(thumb.SizeColors))
		assert.False(t, m.SkipThumbnailSize(thumb.SizeTile100))
		assert.False(t, m.SkipThumbnailSize(thumb.SizeTile224))
		assert.False(t, m.SkipThumbnailSize(thumb.SizeTile500))
		assert.False(t, m.SkipThumbnailSize(thumb.SizeFit720))
		assert.True(t, m.SkipThumbnailSize(thumb.SizeFit1280))
		assert.True(t, m.SkipThumbnailSize(thumb.SizeFit1920))
	})
}

func TestMediaFile_GenerateThumbnails(t *testing.T) {
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

	t.Run("ElephantsJpg", func(t *testing.T) {
		m, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "elephants.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		err = m.GenerateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.GenerateThumbnails(thumbsPath, false))
	})
	t.Run("AnimatedEarthJpg", func(t *testing.T) {
		m, err := NewMediaFile("testdata/animated-earth.jpg")

		if err != nil {
			t.Fatal(err)
		}

		err = m.GenerateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.GenerateThumbnails(thumbsPath, false))
	})
	t.Run("PhotoPrismPng", func(t *testing.T) {
		m, err := NewMediaFile("testdata/photoprism.png")

		if err != nil {
			t.Fatal(err)
		}

		err = m.GenerateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.GenerateThumbnails(thumbsPath, false))
	})
	t.Run("BrokenAnimatedEarthJpg", func(t *testing.T) {
		m, err := NewMediaFile("testdata/broken/animated-earth.jpg")

		if err != nil {
			t.Fatal(err)
		}

		err = m.GenerateThumbnails(thumbsPath, true)

		if err != nil {
			t.Fatal(err)
		}

		thumbFilename, err := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, thumbFilename)
		assert.NoError(t, m.GenerateThumbnails(thumbsPath, false))
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

func TestMediaFile_configBounds(t *testing.T) {
	c := config.TestConfig()

	t.Run("Image", func(t *testing.T) {
		m, err := NewMediaFile(filepath.Join(c.SamplesPath(), "elephants.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		b := m.configBounds()
		cfg, cfgErr := m.DecodeConfig()

		if cfgErr != nil {
			t.Fatal(cfgErr)
		}

		assert.Equal(t, cfg.Width, b.Max.X)
		assert.Equal(t, cfg.Height, b.Max.Y)
		assert.False(t, b.Empty())
	})
	t.Run("NotAnImage", func(t *testing.T) {
		m, err := NewMediaFile(filepath.Join(c.SamplesPath(), "blue-go-video.mp4"))

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, m.configBounds().Empty())
	})
}

// TestMediaFile_GenerateAvatarThumbnails covers that avatar thumbnails are written without
// consulting the file metadata, which an avatar does not need.
func TestMediaFile_GenerateAvatarThumbnails(t *testing.T) {
	c := config.TestConfig()

	thumbsPath, err := filepath.Abs("./.test_mediafile_avatarthumbnails")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.RemoveAll(thumbsPath)
	}()

	if err = c.CreateDirectories(); err != nil {
		t.Fatal(err)
	}

	t.Run("Success", func(t *testing.T) {
		m, mErr := NewMediaFile(filepath.Join(c.SamplesPath(), "elephants.jpg"))

		if mErr != nil {
			t.Fatal(mErr)
		}

		if genErr := m.GenerateAvatarThumbnails(thumbsPath, true); genErr != nil {
			t.Fatal(genErr)
		}

		thumbFilename, fErr := thumb.FileName(m.Hash(), thumbsPath, thumb.Sizes[thumb.Tile50].Width, thumb.Sizes[thumb.Tile50].Height, thumb.Sizes[thumb.Tile50].Options...)

		if fErr != nil {
			t.Fatal(fErr)
		}

		assert.FileExists(t, thumbFilename)
		// The point of the avatar path: the file metadata is never read.
		assert.Equal(t, meta.Data{}, m.metaData)
	})
	t.Run("RegularPathReadsMetadata", func(t *testing.T) {
		m, mErr := NewMediaFile(filepath.Join(c.SamplesPath(), "elephants.jpg"))

		if mErr != nil {
			t.Fatal(mErr)
		}

		if genErr := m.GenerateThumbnails(thumbsPath, true); genErr != nil {
			t.Fatal(genErr)
		}

		assert.NotEqual(t, meta.Data{}, m.metaData)
	})
	t.Run("CachedRunReadsNothing", func(t *testing.T) {
		m, mErr := NewMediaFile(filepath.Join(c.SamplesPath(), "elephants.jpg"))

		if mErr != nil {
			t.Fatal(mErr)
		}

		// Every size is already cached by the subtests above, so no size needs generating
		// and the source bounds are never resolved.
		if genErr := m.GenerateThumbnails(thumbsPath, false); genErr != nil {
			t.Fatal(genErr)
		}

		assert.Equal(t, meta.Data{}, m.metaData)
	})
}
