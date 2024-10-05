package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestIndex_MediaFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("flash.jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		cfg.InitializeTestData()

		tf := classify.New(cfg.AssetsPath(), cfg.DisableTensorFlow())
		nd := nsfw.New(cfg.NSFWModelPath())
		fn := face.NewNet(cfg.FaceNetModelPath(), "", cfg.DisableTensorFlow())
		convert := NewConvert(cfg)

		ind := NewIndex(cfg, tf, nd, fn, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll()
		mediaFile, err := NewMediaFile("testdata/flash.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.Keywords.String())

		result := ind.MediaFile(mediaFile, indexOpt, "flash.jpg", "")

		words := mediaFile.metaData.Keywords.String()

		t.Logf("size in megapixel: %d", mediaFile.Megapixels())

		limitErr, _ := mediaFile.ExceedsResolution(cfg.ResolutionLimit())
		t.Logf("index: %s", limitErr)

		assert.Contains(t, words, "marienkäfer")
		assert.Contains(t, words, "burst")
		assert.Contains(t, words, "flash")
		assert.Contains(t, words, "panorama")
		assert.Equal(t, "Animal with green eyes on table burst", mediaFile.metaData.Description)
		assert.Equal(t, IndexStatus("added"), result.Status)
	})

	t.Run("blue-go-video.mp4", func(t *testing.T) {
		cfg := config.TestConfig()

		cfg.InitializeTestData()

		tf := classify.New(cfg.AssetsPath(), cfg.DisableTensorFlow())
		nd := nsfw.New(cfg.NSFWModelPath())
		fn := face.NewNet(cfg.FaceNetModelPath(), "", cfg.DisableTensorFlow())
		convert := NewConvert(cfg)

		ind := NewIndex(cfg, tf, nd, fn, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll()
		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/blue-go-video.mp4")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.metaData.Title)

		result := ind.UserMediaFile(mediaFile, indexOpt, "blue-go-video.mp4", "", entity.Admin.GetUID())

		assert.Equal(t, "Blue Gopher", mediaFile.metaData.Title)
		assert.Equal(t, IndexStatus("added"), result.Status)
	})

	t.Run("twoFiles", func(t *testing.T) {
		cfg := config.TestConfig()

		cfg.InitializeTestData()

		tf := classify.New(cfg.AssetsPath(), cfg.DisableTensorFlow())
		nd := nsfw.New(cfg.NSFWModelPath())
		fn := face.NewNet(cfg.FaceNetModelPath(), "", cfg.DisableTensorFlow())
		convert := NewConvert(cfg)

		ind := NewIndex(cfg, tf, nd, fn, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll()
		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/beach_sand.jpg")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.Title)
		assert.Equal(t, "", mediaFile.metaData.CameraMake)

		result := ind.UserMediaFile(mediaFile, indexOpt, "beach_sand.jpg", "", entity.Admin.GetUID())

		assert.Equal(t, "", mediaFile.metaData.Title)
		assert.Equal(t, "Apple", mediaFile.metaData.CameraMake)
		log.Debugf("mediaFile.metaData = %v", mediaFile.metaData)

		photo := entity.Photo{}
		entity.Db().Debug().Model(entity.Photo{}).Preload("Details").Where("original_name = 'beach_sand'").First(&photo)
		assert.Equal(t, "beach_sand", photo.OriginalName)
		quality := photo.PhotoQuality
		cameraid := photo.CameraID
		placeid := photo.PlaceID
		assert.Contains(t, photo.Details.Keywords, "beach")
		assert.Contains(t, photo.Details.Keywords, "sand")
		assert.Contains(t, photo.Details.Keywords, "blue")
		assert.Equal(t, IndexStatus("added"), result.Status)

		mediaFile, err = NewMediaFile(cfg.ExamplesPath() + "/beach_sand.json")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.CameraMake)
		assert.Equal(t, "", mediaFile.metaData.Title)

		result = ind.UserMediaFile(mediaFile, indexOpt, "beach_sand.jpg", "", entity.Admin.GetUID())

		// This isn't a Primary file, so these should NOT be updated.
		assert.Equal(t, "", mediaFile.metaData.CameraMake)
		assert.Equal(t, "", mediaFile.metaData.Title)

		photo = entity.Photo{}
		entity.Db().Model(entity.Photo{}).Preload("Details").Where("original_name = 'beach_sand'").First(&photo)
		assert.Equal(t, "beach_sand", photo.OriginalName)
		assert.Contains(t, photo.Details.Keywords, "beach")
		assert.Contains(t, photo.Details.Keywords, "sand")
		assert.Contains(t, photo.Details.Keywords, "blue")
		// Make sure that reading in a json file with the same details as the photo hasn't changed the data.
		assert.Equal(t, quality, photo.PhotoQuality)
		assert.Equal(t, cameraid, photo.CameraID)
		assert.Equal(t, placeid, photo.PlaceID)

		assert.Equal(t, IndexStatus("added"), result.Status)
	})

	t.Run("error", func(t *testing.T) {
		cfg := config.TestConfig()

		cfg.InitializeTestData()

		tf := classify.New(cfg.AssetsPath(), cfg.DisableTensorFlow())
		nd := nsfw.New(cfg.NSFWModelPath())
		fn := face.NewNet(cfg.FaceNetModelPath(), "", cfg.DisableTensorFlow())
		convert := NewConvert(cfg)

		ind := NewIndex(cfg, tf, nd, fn, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll()

		result := ind.MediaFile(nil, indexOpt, "blue-go-video.mp4", "")
		assert.Equal(t, IndexStatus("failed"), result.Status)
	})
}

func TestIndexResult_Archived(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		r := &IndexResult{IndexArchived, nil, 5, "", 5, ""}
		assert.True(t, r.Archived())
	})

	t.Run("false", func(t *testing.T) {
		r := &IndexResult{IndexAdded, nil, 5, "", 5, ""}
		assert.False(t, r.Archived())
	})
}

func TestIndexResult_Skipped(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		r := &IndexResult{IndexSkipped, nil, 5, "", 5, ""}
		assert.True(t, r.Skipped())
	})

	t.Run("false", func(t *testing.T) {
		r := &IndexResult{IndexAdded, nil, 5, "", 5, ""}
		assert.False(t, r.Skipped())
	})
}
