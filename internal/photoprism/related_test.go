package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRelatedFiles_ContainsJpeg(t *testing.T) {
	conf := config.TestConfig()

	t.Run("true", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(conf.ExamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.True(t, relatedFiles.ContainsJpeg())
	})
	t.Run("false", func(t *testing.T) {
		mediaFile3, err3 := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		mediaFile2, err2 := NewMediaFile(conf.ExamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile3, mediaFile2},
			Main:  nil,
		}
		assert.False(t, relatedFiles.ContainsJpeg())
	})
}

func TestRelatedFiles_String(t *testing.T) {
	conf := config.TestConfig()

	t.Run("true", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(conf.ExamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.Equal(t, "telegram_2020-01-30_09-57-18.jpg, Screenshot 2019-05-21 at 10.45.52.png", relatedFiles.String())
	})
}

func TestRelatedFiles_Len(t *testing.T) {
	conf := config.TestConfig()

	t.Run("true", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(conf.ExamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.Equal(t, 2, relatedFiles.Len())
	})
}
