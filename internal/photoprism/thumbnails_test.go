package photoprism

import (
	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/models"
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestThumbnails_Thumbnail(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	t.Run("/elephants.jpg", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)

		thumbnail, err := image.Thumbnail(thumbsPath, "tile_500")

		assert.Empty(t, err)

		assert.FileExists(t, thumbnail)
	})
	t.Run("invalid image format", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.xmp")
		assert.Nil(t, err)

		thumbnail, err := image.Thumbnail(thumbsPath, "tile_500")

		assert.Equal(t, "could not create thumbnail: image: unknown format", err.Error())
		t.Log(thumbnail)
	})
	t.Run("invalid thumbnail type", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)

		thumbnail, err := image.Thumbnail(thumbsPath, "invalid_500")

		assert.Equal(t, "invalid type: invalid_500", err.Error())
		t.Log(thumbnail)
	})
}

func TestThumbnails_CreateThumbnailsFromOriginals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	conf.InitializeTestData(t)

	tensorFlow := NewTensorFlow(conf)

	indexer := NewIndexer(conf, tensorFlow)

	converter := NewConverter(conf)

	importer := NewImporter(conf, indexer, converter)

	importer.ImportPhotosFromDirectory(conf.ImportPath())

	err := CreateThumbnailsFromOriginals(conf.OriginalsPath(), conf.ThumbnailsPath(), true)

	if err != nil {
		t.Error(err)
	}
}

func TestThumbnails_Resample(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)
	t.Run("/elephants.jpg", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)

		thumbnail, err := image.Resample(thumbsPath, "tile_500")

		assert.Empty(t, err)
		assert.NotEmpty(t, thumbnail)

	})
	t.Run("invalid type", func(t *testing.T) {
		image, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)

		thumbnail, err := image.Resample(thumbsPath, "xxx_500")

		assert.Equal(t, "invalid type: xxx_500", err.Error())
		assert.Empty(t, thumbnail)

	})

}

func TestThumbnails_ThumbnailFilename(t *testing.T) {
	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	t.Run("", func(t *testing.T) {
		filename, err := ThumbnailFilename("99988", thumbsPath, 150, 150, ResampleFit, ResampleNearestNeighbor)
		assert.Nil(t, err)
		assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/testdata/cache/_tmp/9/9/9/99988_150x150_fit.jpg", filename)
	})
	t.Run("hash too short", func(t *testing.T) {
		_, err := ThumbnailFilename("999", thumbsPath, 150, 150, ResampleFit, ResampleNearestNeighbor)
		assert.Equal(t, "file hash is empty or too short: 999", err.Error())
	})
	t.Run("invalid width", func(t *testing.T) {
		_, err := ThumbnailFilename("99988", thumbsPath, -4, 150, ResampleFit, ResampleNearestNeighbor)
		assert.Equal(t, "width has an invalid value: -4", err.Error())
	})
	t.Run("invalid height", func(t *testing.T) {
		_, err := ThumbnailFilename("99988", thumbsPath, 200, -1, ResampleFit, ResampleNearestNeighbor)
		assert.Equal(t, "height has an invalid value: -1", err.Error())
	})
	t.Run("empty thumbpath", func(t *testing.T) {
		path := ""
		_, err := ThumbnailFilename("99988", path, 200, 150, ResampleFit, ResampleNearestNeighbor)
		assert.Equal(t, "thumbnail path is empty: ", err.Error())
	})
}

func TestThumbnails_ThumbnailFromFile(t *testing.T) {
	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	t.Run("valid parameter", func(t *testing.T) {
		fileModel := &models.File{
			FileName: conf.ExamplesPath() + "/elephants.jpg",
			FileHash: "1234568889",
		}

		thumb, err := ThumbnailFromFile(fileModel.FileName, fileModel.FileHash, thumbsPath, 224, 224)
		assert.Nil(t, err)
		assert.FileExists(t, thumb)
	})

	t.Run("hash too short", func(t *testing.T) {
		fileModel := &models.File{
			FileName: conf.ExamplesPath() + "/elephants.jpg",
			FileHash: "123",
		}

		_, err := ThumbnailFromFile(fileModel.FileName, fileModel.FileHash, thumbsPath, 224, 224)
		assert.Equal(t, "file hash is empty or too short: 123", err.Error())
	})
	t.Run("filename too short", func(t *testing.T) {
		fileModel := &models.File{
			FileName: "xxx",
			FileHash: "12367890",
		}

		_, err := ThumbnailFromFile(fileModel.FileName, fileModel.FileHash, thumbsPath, 224, 224)
		assert.Equal(t, "image filename is empty or too short: xxx", err.Error())
	})
}

func TestThumbnails_CreateThumbnail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	t.Run("valid parameter", func(t *testing.T) {
		expectedFilename, err := ThumbnailFilename("12345", thumbsPath, 150, 150, ResampleFit, ResampleNearestNeighbor)

		if err != nil {
			t.Error(err)
		}

		img, err := imaging.Open(conf.ExamplesPath()+"/elephants.jpg", imaging.AutoOrientation(true))

		if err != nil {
			t.Errorf("can't open original: %s", err)
		}

		thumb, err := CreateThumbnail(img, expectedFilename, 150, 150, ResampleFit, ResampleNearestNeighbor)

		assert.Empty(t, err)

		assert.NotNil(t, thumb)

		bounds := thumb.Bounds()

		assert.Equal(t, 150, bounds.Dx())
		assert.Equal(t, 99, bounds.Dy())

		assert.FileExists(t, expectedFilename)
	})
	t.Run("invalid width", func(t *testing.T) {
		expectedFilename, err := ThumbnailFilename("12345", thumbsPath, 150, 150, ResampleFit, ResampleNearestNeighbor)

		if err != nil {
			t.Error(err)
		}

		img, err := imaging.Open(conf.ExamplesPath()+"/elephants.jpg", imaging.AutoOrientation(true))

		if err != nil {
			t.Errorf("can't open original: %s", err)
		}

		thumbnail, err := CreateThumbnail(img, expectedFilename, -1, 150, ResampleFit, ResampleNearestNeighbor)

		assert.Equal(t, "width has an invalid value: -1", err.Error())
		bounds := thumbnail.Bounds()
		assert.NotEqual(t, 150, bounds.Dx())
	})

	t.Run("invalid height", func(t *testing.T) {
		expectedFilename, err := ThumbnailFilename("12345", thumbsPath, 150, 150, ResampleFit, ResampleNearestNeighbor)

		if err != nil {
			t.Error(err)
		}

		img, err := imaging.Open(conf.ExamplesPath()+"/elephants.jpg", imaging.AutoOrientation(true))

		if err != nil {
			t.Errorf("can't open original: %s", err)
		}

		thumbnail, err := CreateThumbnail(img, expectedFilename, 150, -1, ResampleFit, ResampleNearestNeighbor)

		assert.Equal(t, "height has an invalid value: -1", err.Error())
		bounds := thumbnail.Bounds()
		assert.NotEqual(t, 150, bounds.Dx())
	})
}

func TestThumbnails_CreateDefaultThumbnails(t *testing.T) {
	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	m, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
	assert.Nil(t, err)

	err = m.CreateDefaultThumbnails(thumbsPath, true)

	assert.Empty(t, err)

	thumbFilename, err := ThumbnailFilename(m.Hash(), thumbsPath, ThumbnailTypes["tile_50"].Width, ThumbnailTypes["tile_50"].Height, ThumbnailTypes["tile_50"].Options...)

	assert.Empty(t, err)

	assert.FileExists(t, thumbFilename)

	err = m.CreateDefaultThumbnails(thumbsPath, false)

	assert.Empty(t, err)
}
