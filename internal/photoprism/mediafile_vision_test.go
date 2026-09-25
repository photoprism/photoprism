package photoprism

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

func setupVisionMediaFile(t *testing.T) *MediaFile {
	t.Helper()

	cfg := config.TestConfig()
	require.NoError(t, cfg.InitializeTestData())

	mediaFile, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)

	return mediaFile
}

func TestMediaFile_GenerateCaption(t *testing.T) {
	mediaFile := setupVisionMediaFile(t)

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

		caption, err := mediaFile.GenerateCaption(entity.SrcAuto)
		require.NoError(t, err)
		require.NotNil(t, caption)
		assert.Equal(t, captionModel.GetSource(), caption.Source)
	})
	t.Run("CustomSourceOverrides", func(t *testing.T) {
		vision.SetCaptionFunc(func(files vision.Files, mediaSrc media.Src) (*vision.CaptionResult, *vision.Model, error) {
			return &vision.CaptionResult{Text: "stub", Source: captionModel.GetSource()}, captionModel, nil
		})

		caption, err := mediaFile.GenerateCaption(entity.SrcManual)
		require.NoError(t, err)
		require.NotNil(t, caption)
		assert.Equal(t, entity.SrcManual, caption.Source)
	})
	t.Run("MissingModelReturnsError", func(t *testing.T) {
		vision.Config = &vision.ConfigValues{}
		vision.SetCaptionFunc(nil)

		caption, err := mediaFile.GenerateCaption(entity.SrcAuto)
		assert.Error(t, err)
		assert.Nil(t, caption)
	})
}

func TestMediaFile_GenerateLabels(t *testing.T) {
	mediaFile := setupVisionMediaFile(t)

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
			return classify.Labels{{Name: "stub", Source: src}}, nil
		})

		labels := mediaFile.GenerateLabels(entity.SrcAuto)
		assert.NotEmpty(t, labels)
		assert.Equal(t, labelModel.GetSource(), captured)
	})
	t.Run("CustomSourceOverrides", func(t *testing.T) {
		var captured string
		vision.SetLabelsFunc(func(files vision.Files, mediaSrc media.Src, src entity.Src) (classify.Labels, error) {
			captured = src
			return classify.Labels{{Name: "stub", Source: src}}, nil
		})

		labels := mediaFile.GenerateLabels(entity.SrcManual)
		assert.NotEmpty(t, labels)
		assert.Equal(t, entity.SrcManual, captured)
	})
	t.Run("MissingModel", func(t *testing.T) {
		vision.Config = &vision.ConfigValues{}
		vision.SetLabelsFunc(nil)

		labels := mediaFile.GenerateLabels(entity.SrcAuto)
		assert.Empty(t, labels)
	})
}

func TestMediaFile_DetectNSFW(t *testing.T) {
	mediaFile := setupVisionMediaFile(t)

	t.Run("FlagsHighConfidence", func(t *testing.T) {
		vision.SetNSFWFunc(func(files vision.Files, mediaSrc media.Src) ([]nsfw.Result, error) {
			return []nsfw.Result{nsfw.NewResult(0.99, nsfw.DefaultThreshold)}, nil
		})
		t.Cleanup(func() { vision.SetNSFWFunc(nil) })

		result := mediaFile.DetectNSFW()
		assert.True(t, result.IsUnsafe())
		assert.False(t, result.IsSafe())
	})
	t.Run("SafeContent", func(t *testing.T) {
		vision.SetNSFWFunc(func(files vision.Files, mediaSrc media.Src) ([]nsfw.Result, error) {
			return []nsfw.Result{nsfw.NewResult(0.02, nsfw.DefaultThreshold)}, nil
		})
		t.Cleanup(func() { vision.SetNSFWFunc(nil) })

		result := mediaFile.DetectNSFW()
		assert.True(t, result.IsSafe())
		assert.False(t, result.IsUnsafe())
	})
	// A detector that could not answer must not report a clearance, because every caller of
	// this method treats "not unsafe" as permission to leave a photo public.
	t.Run("DetectorError", func(t *testing.T) {
		vision.SetNSFWFunc(func(files vision.Files, mediaSrc media.Src) ([]nsfw.Result, error) {
			return nil, errors.New("model is unavailable")
		})
		t.Cleanup(func() { vision.SetNSFWFunc(nil) })

		result := mediaFile.DetectNSFW()
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
		assert.False(t, result.IsUnsafe())
	})
	t.Run("NoResult", func(t *testing.T) {
		vision.SetNSFWFunc(func(files vision.Files, mediaSrc media.Src) ([]nsfw.Result, error) {
			return []nsfw.Result{}, nil
		})
		t.Cleanup(func() { vision.SetNSFWFunc(nil) })

		result := mediaFile.DetectNSFW()
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
	t.Run("UndecidedResult", func(t *testing.T) {
		vision.SetNSFWFunc(func(files vision.Files, mediaSrc media.Src) ([]nsfw.Result, error) {
			return []nsfw.Result{nsfw.Unavailable("thumbnail is missing")}, nil
		})
		t.Cleanup(func() { vision.SetNSFWFunc(nil) })

		result := mediaFile.DetectNSFW()
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
}
