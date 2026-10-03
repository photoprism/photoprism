package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestIndexCaptionSource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping vision-dependent test in short mode")
	}

	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)

	originalConfig := vision.Config
	t.Cleanup(func() {
		vision.Config = originalConfig
		vision.SetCaptionFunc(nil)
	})

	captionModel := &vision.Model{Type: vision.ModelTypeCaption, Engine: vision.ApiFormatOpenAI}
	captionModel.ApplyEngineDefaults()
	vision.Config = &vision.ConfigValues{Models: vision.Models{captionModel}}

	t.Run("AutoUsesModelSource", func(t *testing.T) {
		vision.SetCaptionFunc(func(files vision.Files, mediaSrc media.Src) (*vision.CaptionResult, *vision.Model, error) {
			return &vision.CaptionResult{Text: "stub", Source: captionModel.GetSource()}, captionModel, nil
		})
		t.Cleanup(func() { vision.SetCaptionFunc(nil) })

		caption, captionErr := mediaFile.GenerateCaption(entity.SrcAuto)
		require.NoError(t, captionErr)
		require.NotNil(t, caption)
		assert.Equal(t, captionModel.GetSource(), caption.Source)
	})
	t.Run("CustomSource", func(t *testing.T) {
		originalSource := captionModel.GetSource()
		vision.SetCaptionFunc(func(files vision.Files, mediaSrc media.Src) (*vision.CaptionResult, *vision.Model, error) {
			return &vision.CaptionResult{Text: "stub", Source: originalSource}, captionModel, nil
		})
		t.Cleanup(func() { vision.SetCaptionFunc(nil) })

		caption, captionErr := mediaFile.GenerateCaption(entity.SrcManual)
		require.NoError(t, captionErr)
		require.NotNil(t, caption)
		assert.Equal(t, entity.SrcManual, caption.Source)
	})
}

// TestLabelsMarkNSFW verifies label-derived NSFW flags apply only when NSFW detection is enabled.
func TestLabelsMarkNSFW(t *testing.T) {
	labels := classify.Labels{{Name: "test", NSFW: true}}
	assert.False(t, labelsMarkNSFW(labels, false))
	assert.True(t, labelsMarkNSFW(labels, true))
}

// TestLabelsNSFWThreshold verifies that labels models use the labels threshold, not the detector's.
func TestLabelsNSFWThreshold(t *testing.T) {
	previous := vision.Config
	t.Cleanup(func() { vision.Config = previous })

	t.Run("Configured", func(t *testing.T) {
		index := 90
		vision.Config = vision.NewConfig()
		vision.Config.Thresholds.NSFW = 40
		vision.Config.Thresholds.NSFWIndex = &index
		assert.Equal(t, 40, labelsNSFWThreshold())
		assert.True(t, labelsMarkNSFW(classify.Labels{{Name: "beach", NSFWConfidence: 50}}, true))
	})
	t.Run("NoConfig", func(t *testing.T) {
		vision.Config = nil
		assert.Equal(t, vision.DefaultNSFWThreshold, labelsNSFWThreshold())
	})
}

// TestIndex_UserMediaFile_LabelsNSFW verifies that NSFW flags from a labels model mark a new
// photo private only in labels mode, while automatic mode runs the dedicated detector instead.
func TestIndex_UserMediaFile_LabelsNSFW(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	previousConfig := vision.Config
	previousMode, previousDetect := cfg.Options().NsfwModel, cfg.Options().DetectNSFW
	t.Cleanup(func() {
		vision.Config = previousConfig
		vision.SetLabelsFunc(nil)
		vision.SetNSFWFunc(nil)
		cfg.Options().NsfwModel, cfg.Options().DetectNSFW = previousMode, previousDetect
	})

	labelName := "labels-nsfw-" + rnd.Base36(8)
	labelModel := &vision.Model{Type: vision.ModelTypeLabels, Engine: vision.ApiFormatOllama, Run: vision.RunOnIndex}
	labelModel.ApplyEngineDefaults()
	vision.Config = &vision.ConfigValues{Models: vision.Models{labelModel, vision.NsfwModel.Clone()}}
	vision.SetLabelsFunc(func(vision.Files, media.Src, entity.Src) (classify.Labels, error) {
		return classify.Labels{{Name: labelName, NSFW: true, NSFWConfidence: 90}}, nil
	})

	detected := 0
	vision.SetNSFWFunc(func(files vision.Files, _ media.Src) ([]nsfw.Result, error) {
		detected++
		results := make([]nsfw.Result, len(files))
		for i := range results {
			results[i] = nsfw.NewResult(0.01, nsfw.DefaultThreshold)
		}
		return results, nil
	})
	t.Cleanup(func() {
		if label, findErr := entity.FindLabel(labelName, false); findErr == nil && label != nil {
			_ = entity.UnscopedDb().Delete(label).Error
		}
	})

	data, err := os.ReadFile(filepath.Join("testdata", "flash.jpg"))
	require.NoError(t, err)

	for _, mode := range []struct {
		name     string
		labels   bool
		detector int
	}{
		{"Auto", false, 1},
		{"Labels", true, 0},
	} {
		t.Run(mode.name, func(t *testing.T) {
			setting := strings.ToLower(mode.name)
			cfg.Options().NsfwModel, cfg.Options().DetectNSFW = setting, true
			detected = 0
			opt := NewIndexOptions("/", false, false, false, false, true, cfg)
			require.True(t, opt.GenerateLabels)
			require.Equal(t, mode.labels, opt.DetectNSFWLabels)

			// A random name and trailing bytes make each file a new photo, also in repeated runs.
			relName := filepath.Join("labels-nsfw", setting+"-"+rnd.Base36(8)+".jpg")
			fileName := filepath.Join(cfg.OriginalsPath(), relName)
			require.NoError(t, os.MkdirAll(filepath.Dir(fileName), fs.ModeDir))
			require.NoError(t, os.WriteFile(fileName, append(append([]byte{}, data...), []byte(relName)...), fs.ModeFile)) //nolint:gosec // isolated test path
			t.Cleanup(func() {
				_ = os.Remove(fileName)
				_ = os.Remove(filepath.Dir(fileName))
			})

			mediaFile, err := NewMediaFile(fileName)
			require.NoError(t, err)
			ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
			result := ind.UserMediaFile(mediaFile, opt, relName, "", entity.OwnerUnknown)
			require.Equal(t, IndexAdded, result.Status, "%s", result.Err)

			photo := entity.FindPhoto(entity.Photo{PhotoUID: result.PhotoUID})
			require.NotNil(t, photo)
			t.Cleanup(func() { _, _ = photo.DeletePermanently() })
			assert.Equal(t, mode.labels, photo.PhotoPrivate)
			assert.Equal(t, mode.detector, detected)
		})
	}
}

func TestIndexLabelsSource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping vision-dependent test in short mode")
	}

	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)

	originalConfig := vision.Config
	t.Cleanup(func() {
		vision.Config = originalConfig
		vision.SetLabelsFunc(nil)
	})

	labelModel := &vision.Model{Type: vision.ModelTypeLabels, Engine: vision.ApiFormatOllama}
	labelModel.ApplyEngineDefaults()
	vision.Config = &vision.ConfigValues{Models: vision.Models{labelModel}}

	t.Run("AutoUsesModelSource", func(t *testing.T) {
		var captured string
		vision.SetLabelsFunc(func(files vision.Files, mediaSrc media.Src, src entity.Src) (classify.Labels, error) {
			captured = src
			return classify.Labels{{Name: "stub", Source: src, Uncertainty: 0}}, nil
		})
		t.Cleanup(func() { vision.SetLabelsFunc(nil) })

		labels := mediaFile.GenerateLabels(entity.SrcAuto)
		assert.NotEmpty(t, labels)
		assert.Equal(t, labelModel.GetSource(), captured)
	})
	t.Run("CustomSource", func(t *testing.T) {
		var captured string
		vision.SetLabelsFunc(func(files vision.Files, mediaSrc media.Src, src entity.Src) (classify.Labels, error) {
			captured = src
			return classify.Labels{{Name: "stub", Source: src, Uncertainty: 0}}, nil
		})
		t.Cleanup(func() { vision.SetLabelsFunc(nil) })

		labels := mediaFile.GenerateLabels(entity.SrcManual)
		assert.NotEmpty(t, labels)
		assert.Equal(t, entity.SrcManual, captured)
	})
}
