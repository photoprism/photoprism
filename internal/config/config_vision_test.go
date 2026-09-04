package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig_VisionYaml(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, ProjectRoot+"/storage/testdata/config/vision.yml", c.VisionYaml())
	})
	t.Run("PreferYamlExtension", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempDir := t.TempDir()
		c.options.ConfigPath = tempDir
		c.options.VisionYaml = ""

		yamlPath := filepath.Join(tempDir, "vision"+fs.ExtYaml)
		if err := os.WriteFile(yamlPath, []byte("models: []\n"), fs.ModeFile); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		assert.Equal(t, yamlPath, c.VisionYaml())
	})
}

func TestConfig_VisionApi(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.VisionApi())
}

func TestConfig_VisionUri(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.VisionUri())
	c.options.VisionUri = "https://www.example.com/api/v1/vision"
	assert.Equal(t, "https://www.example.com/api/v1/vision", c.VisionUri())
	c.options.VisionUri = ""
	assert.Equal(t, "", c.VisionUri())
}

func TestConfig_VisionKey(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.VisionKey())
	c.options.VisionKey = "SecretAccessToken!"
	assert.Equal(t, "SecretAccessToken!", c.VisionKey())
	c.options.VisionKey = ""
	assert.Equal(t, "", c.VisionKey())
}

func TestConfig_ModelsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.NasnetModelPath()
	assert.True(t, strings.HasPrefix(path, c.ModelsPath()))
	assert.Equal(t, ProjectRoot+"/assets/models/nasnet", path)
}

// TestConfig_LabelModel verifies automatic, named, disabled, and custom model selection.
func TestConfig_LabelModel(t *testing.T) {
	t.Run("Auto", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.LabelModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, classify.DefaultModelName(), c.EffectiveLabelModel())
		assert.Contains(t, c.LabelModelPath(), string(classify.DefaultModelName()))
		assert.Equal(t, vision.EngineONNX, c.LabelModelRuntime())
	})
	t.Run("Named", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.LabelModel = string(classify.ModelRepViTM10)
		c.applyLabelModel()
		assert.Equal(t, classify.ModelRepViTM10, c.EffectiveLabelModel())
		assert.Equal(t, string(classify.ModelRepViTM10), vision.Config.Model(vision.ModelTypeLabels).Name)
	})
	t.Run("Cli", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("label-model", string(classify.ModelEfficientNetB0)))
		c := NewConfig(ctx)
		c.applyLabelModel()
		assert.Equal(t, classify.ModelEfficientNetB0, c.EffectiveLabelModel())
	})
	t.Run("None", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.LabelModel = "none"
		c.applyLabelModel()
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
		assert.Nil(t, vision.Config.Model(vision.ModelTypeLabels))
		assert.Empty(t, c.LabelModelPath())
		assert.Equal(t, "none", c.LabelModelRuntime())
	})
	t.Run("CustomVisionModel", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom_21k", Path: "custom_21k"}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.LabelModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, classify.ModelName("custom_21k"), c.EffectiveLabelModel())
		assert.Contains(t, c.LabelModelPath(), filepath.Join("custom_21k", "custom_21k.onnx"))
		assert.Equal(t, vision.EngineONNX, c.LabelModelRuntime())
		assert.Zero(t, vision.Config.Model(vision.ModelTypeLabels).Resolution)
	})
	t.Run("CustomDirectoryWithExtension", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom_v2", Path: "custom.v2"}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.LabelModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, filepath.Join(c.ModelsPath(), "custom.v2", "custom.v2.onnx"), c.LabelModelPath())
	})
	t.Run("CustomONNXFile", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom_file", Path: "custom/model.ONNX"}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.LabelModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, filepath.Join(c.ModelsPath(), "custom", "model.ONNX"), c.LabelModelPath())
	})
}

func TestConfig_TensorFlowDisabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.DisableTensorFlow()
	assert.Equal(t, false, version)
}

func TestConfig_NSFWModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.NsfwModelPath(), "/assets/models/nsfw")
}

// TestConfig_reportUnscreenedUploads verifies the missing-detector warning conditions.
func TestConfig_reportUnscreenedUploads(t *testing.T) {
	t.Run("MissingDetector", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{})
		c := NewConfig(CliTestContext())
		c.options.UploadNSFW = false
		hook := captureLog(t)

		c.reportUnscreenedUploads()

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Contains(t, entry.Message, "no nsfw model is configured")
	})
	t.Run("UploadsAllowed", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{})
		c := NewConfig(CliTestContext())
		c.options.UploadNSFW = true
		hook := captureLog(t)

		c.reportUnscreenedUploads()

		assert.Empty(t, hook.AllEntries())
	})
	t.Run("DetectorConfigured", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.UploadNSFW = false
		hook := captureLog(t)

		c.reportUnscreenedUploads()

		assert.Empty(t, hook.AllEntries())
	})
}

func TestConfig_FaceNetModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.FacenetModelPath(), "/assets/models/facenet")
}

func TestConfig_DetectNSFW(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.DetectNSFW()
	assert.Equal(t, true, result)
}

func TestConfig_VisionModelShouldRun(t *testing.T) {
	t.Run("ClassificationDisabledLabels", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DisableClassification = true
		withVisionConfig(t, vision.NewConfig())
		if c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunManual) {
			t.Fatalf("expected false when classification disabled")
		}
	})
	t.Run("DetectNSFWDisabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DetectNSFW = false
		withVisionConfig(t, vision.NewConfig())
		if c.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunManual) {
			t.Fatalf("expected false when detect nsfw disabled")
		}
	})
	t.Run("NilVisionConfig", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		withVisionConfig(t, nil)
		if c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunManual) {
			t.Fatalf("expected false when no vision config is loaded")
		}
	})
	t.Run("DelegatesToVisionConfig", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		withVisionConfig(t, vision.NewConfig())
		if !c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunManual) {
			t.Fatalf("expected labels model to run manually with defaults")
		}
		if !c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex) {
			t.Fatalf("expected labels model to run on index with defaults")
		}
	})
	t.Run("CustomLabelsRunAfterIndex", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		defaultModel := vision.NasnetModel.Clone()
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom"}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{defaultModel, custom}})
		if !c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunNewlyIndexed) {
			t.Fatalf("expected custom labels model to run after indexing")
		}
		if c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex) {
			t.Fatalf("expected custom labels model to skip on-index runs")
		}
	})
}

func TestConfig_VisionSchedule(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.VisionSchedule = ""
	assert.Equal(t, "", c.VisionSchedule())

	c.options.VisionSchedule = "0 6 * * *"
	assert.Equal(t, "0 6 * * *", c.VisionSchedule())

	c.options.VisionSchedule = "invalid"
	assert.Equal(t, "", c.VisionSchedule())
}

func TestConfig_VisionFilter(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.VisionFilter = "  private:false  "
	assert.Equal(t, "private:false", c.VisionFilter())

	c.options.VisionFilter = ""
	assert.Equal(t, "", c.VisionFilter())
}
