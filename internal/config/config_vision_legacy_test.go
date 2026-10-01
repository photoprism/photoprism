package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
)

// TestConfig_applyVisionModelsLegacy verifies that config files written for the TensorFlow label
// and NSFW models select the installed default ONNX models in automatic mode.
func TestConfig_applyVisionModelsLegacy(t *testing.T) {
	// newLegacyConfig loads a legacy vision config file and applies the automatic model modes.
	newLegacyConfig := func(t *testing.T, fileName string) *Config {
		t.Helper()

		cfg := vision.NewConfig()
		require.NoError(t, cfg.Load(filepath.Join("..", "ai", "vision", "testdata", fileName)))
		withVisionConfig(t, cfg)

		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(classify.DefaultModelName()), classify.FindModel(classify.DefaultModelName()).ONNX.File)
		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.DefaultModelName()), nsfw.FindModel(nsfw.DefaultModelName()).ONNX.File)
		c.options.LabelsModel = "auto"
		c.options.NsfwModel = "auto"
		c.applyLabelModel()
		c.applyNSFWModel()

		return c
	}

	t.Run("Release250707", func(t *testing.T) {
		c := newLegacyConfig(t, "vision-250707.yml")

		labels := vision.Config.Model(vision.ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, string(classify.ModelEfficientFormerV2S2), labels.Name)
		assert.False(t, labels.Disabled || labels.DisabledByMode)
		assert.True(t, classify.FindModel(classify.ModelName(labels.Name)).Installed(c.ModelsPath()))
		assert.Equal(t, vision.RunAuto, labels.RunType())
		assert.Equal(t, classify.ModelEfficientFormerV2S2, c.EffectiveLabelModel())

		detector := vision.Config.Model(vision.ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, string(nsfw.ModelYahoo), detector.Name)
		assert.False(t, detector.Disabled || detector.DisabledByMode)
		assert.True(t, nsfw.FindModel(nsfw.ModelName(detector.Name)).Installed(c.ModelsPath()))
		assert.Equal(t, vision.RunAuto, detector.RunType())
	})
	t.Run("Release251130", func(t *testing.T) {
		newLegacyConfig(t, "vision-251130.yml")

		labels := vision.Config.Model(vision.ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, string(classify.ModelEfficientFormerV2S2), labels.Name)
		assert.Equal(t, vision.RunOnSchedule, labels.Run)

		// The NSFW entry stays disabled, as written in the file.
		assert.Nil(t, vision.Config.Model(vision.ModelTypeNsfw))
		detector := configuredVisionModel(vision.Config, vision.ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, string(nsfw.ModelYahoo), detector.Name)
		assert.True(t, detector.Disabled)
		assert.Equal(t, vision.RunAuto, detector.RunType())
	})
	t.Run("CustomTensorFlow", func(t *testing.T) {
		newLegacyConfig(t, "vision-custom-tensorflow.yml")

		labels := vision.Config.Model(vision.ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, string(classify.ModelEfficientFormerV2S2), labels.Name)
		assert.Equal(t, vision.RunAuto, labels.RunType())

		detector := vision.Config.Model(vision.ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, string(nsfw.ModelYahoo), detector.Name)
		assert.Equal(t, vision.RunAlways, detector.Run)
	})
}
