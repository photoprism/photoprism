package photoprism

import (
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndex_MediaFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("/blue-go-video.mp4", func(t *testing.T) {
		conf := config.TestConfig()

		conf.InitializeTestData(t)

		tf := classify.New(conf.AssetsPath(), conf.TensorFlowOff())
		nd := nsfw.New(conf.NSFWModelPath())
		convert := NewConvert(conf)

		ind := NewIndex(conf, tf, nd, convert)
		indexOpt := IndexOptionsAll()
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.metaData.Title)

		result := ind.MediaFile(mediaFile, indexOpt, "blue-go-video.mp4")
		assert.Equal(t, "Blue Gopher", mediaFile.metaData.Title)
		assert.Equal(t, IndexStatus("added"), result.Status)
	})
	t.Run("error", func(t *testing.T) {
		conf := config.TestConfig()

		conf.InitializeTestData(t)

		tf := classify.New(conf.AssetsPath(), conf.TensorFlowOff())
		nd := nsfw.New(conf.NSFWModelPath())
		convert := NewConvert(conf)

		ind := NewIndex(conf, tf, nd, convert)
		indexOpt := IndexOptionsAll()

		result := ind.MediaFile(nil, indexOpt, "blue-go-video.mp4")
		assert.Equal(t, IndexStatus("failed"), result.Status)
	})

}
