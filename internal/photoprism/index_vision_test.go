package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
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
