package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestRelatedFiles_HasPreview(t *testing.T) {
	cfg := config.TestConfig()

	t.Run("JPEG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/gopher-video.mp4")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.True(t, relatedFiles.HasPreview())
	})
	t.Run("PNG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/gopher-video.mp4")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.True(t, relatedFiles.HasPreview())
	})
	t.Run("False", func(t *testing.T) {
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/gopher-video.mp4")
		if err2 != nil {
			t.Fatal(err2)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile3, mediaFile2},
			Main:  nil,
		}
		assert.False(t, relatedFiles.HasPreview())
	})
}

func TestRelatedFiles_ContainsPreview(t *testing.T) {
	cfg := config.TestConfig()

	jpeg, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
	if err != nil {
		t.Fatal(err)
	}
	heic, err := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
	if err != nil {
		t.Fatal(err)
	}
	video, err := NewMediaFile(cfg.SamplesPath() + "/gopher-video.mp4")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("PreviewPresent", func(t *testing.T) {
		r := RelatedFiles{Files: MediaFiles{jpeg, video}, Main: heic}
		assert.True(t, r.ContainsPreview())
	})
	t.Run("PreviewFilteredOut", func(t *testing.T) {
		// Main is a preview JPEG but it was filtered out of the list (incremental
		// sidecar-only update): ContainsPreview is false where HasPreview would be
		// true via the Main fallback.
		r := RelatedFiles{Files: MediaFiles{}, Main: jpeg}
		assert.False(t, r.ContainsPreview())
		assert.True(t, r.HasPreview())
	})
	t.Run("NoPreview", func(t *testing.T) {
		r := RelatedFiles{Files: MediaFiles{heic, video}, Main: heic}
		assert.False(t, r.ContainsPreview())
	})
}

func TestRelatedFiles_String(t *testing.T) {
	cfg := config.TestConfig()

	t.Run("True", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
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
	cfg := config.TestConfig()

	t.Run("True", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
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

func TestRelatedFiles_Count(t *testing.T) {
	cfg := config.TestConfig()
	t.Run("NoMainFile", func(t *testing.T) {
		relatedFiles := RelatedFiles{
			Files: MediaFiles{},
			Main:  nil,
		}
		assert.Equal(t, 0, relatedFiles.Count())
	})
	t.Run("None", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{},
			Main:  mediaFile,
		}
		assert.Equal(t, 0, relatedFiles.Count())
	})
	t.Run("One", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.Equal(t, 1, relatedFiles.Count())
	})
}

func TestRelatedFiles_MainFileType(t *testing.T) {
	cfg := config.TestConfig()
	t.Run("None", func(t *testing.T) {
		relatedFiles := RelatedFiles{
			Files: MediaFiles{},
			Main:  nil,
		}
		assert.Equal(t, "", relatedFiles.MainFileType())
	})
	t.Run("Primary", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{},
			Main:  mediaFile,
		}
		assert.Equal(t, string(fs.ImageJpeg), relatedFiles.MainFileType())
	})
	t.Run("Heif", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2},
			Main:  mediaFile3,
		}
		assert.Equal(t, string(fs.ImageHeic), relatedFiles.MainFileType())
	})
}

func TestRelatedFiles_MainLogName(t *testing.T) {
	cfg := config.TestConfig()
	t.Run("None", func(t *testing.T) {
		relatedFiles := RelatedFiles{
			Files: MediaFiles{},
			Main:  nil,
		}
		assert.Equal(t, "", relatedFiles.MainFileType())
	})
	t.Run("Telegram", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{},
			Main:  mediaFile,
		}
		assert.Equal(t, "telegram_2020-01-30_09-57-18.jpg", relatedFiles.MainLogName())
	})
	t.Run("IPhoneSeven", func(t *testing.T) {
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		mediaFile2, err2 := NewMediaFile(cfg.SamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err2 != nil {
			t.Fatal(err2)
		}
		mediaFile3, err3 := NewMediaFile(cfg.SamplesPath() + "/iphone_7.heic")
		if err3 != nil {
			t.Fatal(err3)
		}
		relatedFiles := RelatedFiles{
			Files: MediaFiles{mediaFile, mediaFile2, mediaFile3},
			Main:  mediaFile3,
		}
		assert.Equal(t, "iphone_7.heic", relatedFiles.MainLogName())
	})
}
