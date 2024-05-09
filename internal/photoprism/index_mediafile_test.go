package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/nsfw"
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
