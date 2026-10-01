package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/dsn"
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
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(classify.ModelEfficientFormerV2S2), classify.FindModel(classify.ModelEfficientFormerV2S2).ONNX.File)
		c.options.LabelsModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, classify.DefaultModelName(), c.EffectiveLabelModel())
		assert.Contains(t, c.LabelModelPath(), string(classify.DefaultModelName()))
		assert.Equal(t, vision.EngineONNX, c.LabelModelRuntime())
	})
	t.Run("AutoPreservesDisabled", func(t *testing.T) {
		config := vision.NewConfig()
		config.Models[0].Disabled = true
		withVisionConfig(t, config)
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(classify.ModelEfficientFormerV2S2), classify.FindModel(classify.ModelEfficientFormerV2S2).ONNX.File)
		c.options.LabelsModel = "auto"
		c.applyLabelModel()
		assert.True(t, vision.Config.Models[0].Disabled)
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
		assert.Empty(t, c.LabelModelPath())
		assert.Equal(t, "none", c.LabelModelRuntime())
	})
	t.Run("AutoMissingRemainsEnabled", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.LabelsModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
		require.False(t, vision.Config.Models[0].Disabled)
		visionFile := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, vision.Config.Save(visionFile))
		reloaded := vision.NewConfig()
		require.NoError(t, reloaded.Load(visionFile))
		require.NotNil(t, configuredVisionModel(reloaded, vision.ModelTypeLabels))
		assert.False(t, configuredVisionModel(reloaded, vision.ModelTypeLabels).Disabled)

		installVisionTestArtifact(t, c.ModelsPath(), string(classify.DefaultModelName()), classify.FindModel(classify.DefaultModelName()).ONNX.File)
		c.applyLabelModel()
		assert.Equal(t, classify.DefaultModelName(), c.EffectiveLabelModel())
		require.False(t, vision.Config.Models[0].Disabled)
	})
	t.Run("AutoDisablesCustomTensorFlow", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom", TensorFlow: &tensorflow.ModelInfo{}, Disabled: true}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.LabelsModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
		assert.Empty(t, c.LabelModelPath())
		assert.Equal(t, "none", c.LabelModelRuntime())
	})
	t.Run("Named", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{vision.NewLabelModel(classify.ModelRepViTM10)}})
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.LabelsModel = "auto"
		hook := captureVisionSystemLog(t)
		c.applyLabelModel()
		c.reportVisionModes()
		assert.Equal(t, classify.ModelRepViTM10, c.EffectiveLabelModel())
		assert.Equal(t, string(classify.ModelRepViTM10), vision.Config.Model(vision.ModelTypeLabels).Name)
		require.NotNil(t, hook.LastEntry())
		assert.Contains(t, hook.LastEntry().Message, "scripts/dist/download-models.sh repvit_m1_0")
	})
	t.Run("NamedDisabledCustom", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom_21k", Path: "custom_21k", Disabled: true}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.LabelsModel = "custom_21k"
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
	})
	t.Run("Cli", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{vision.NewLabelModel(classify.ModelEfficientNetB0)}})
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("labels-model", "auto"))
		c := NewConfig(ctx)
		c.applyLabelModel()
		assert.Equal(t, classify.ModelEfficientNetB0, c.EffectiveLabelModel())
	})
	t.Run("None", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.LabelsModel = "none"
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
		c.options.LabelsModel = "auto"
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
		c.options.LabelsModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, filepath.Join(c.ModelsPath(), "custom.v2", "custom.v2.onnx"), c.LabelModelPath())
	})
	t.Run("CustomONNXFile", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom_file", Path: "custom/model.ONNX"}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.LabelsModel = "auto"
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

	assert.Contains(t, c.NsfwModelPath(), filepath.Join("assets", "models", string(nsfw.DefaultModelName())))
	assert.Equal(t, vision.EngineONNX, c.NsfwModelRuntime())
}

// TestConfig_NSFWModel verifies automatic, named, disabled, and custom model selection.
func TestConfig_NSFWModel(t *testing.T) {
	t.Run("Auto", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.ModelYahoo), nsfw.FindModel(nsfw.ModelYahoo).ONNX.File)
		c.options.NsfwModel = "auto"
		c.applyNSFWModel()
		assert.Equal(t, nsfw.DefaultModelName(), c.EffectiveNSFWModel())
		assert.Equal(t, vision.EngineONNX, c.NsfwModelRuntime())
	})
	t.Run("AutoPreservesDisabled", func(t *testing.T) {
		config := vision.NewConfig()
		config.Models[1].Disabled = true
		withVisionConfig(t, config)
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.ModelYahoo), nsfw.FindModel(nsfw.ModelYahoo).ONNX.File)
		c.options.NsfwModel = "auto"
		c.applyNSFWModel()
		assert.True(t, vision.Config.Models[1].Disabled)
		assert.Equal(t, nsfw.ModelNone, c.EffectiveNSFWModel())
	})
	t.Run("AutoMissingRemainsEnabled", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.NsfwModel = "auto"
		c.applyNSFWModel()
		assert.Equal(t, nsfw.ModelNone, c.EffectiveNSFWModel())
		require.False(t, vision.Config.Models[1].Disabled)

		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.ModelYahoo), nsfw.FindModel(nsfw.ModelYahoo).ONNX.File)
		c.applyNSFWModel()
		c.reportUnscreenedUploads()
		assert.Equal(t, nsfw.ModelYahoo, c.EffectiveNSFWModel())
		require.False(t, vision.Config.Models[1].Disabled)
	})
	t.Run("Named", func(t *testing.T) {
		model := vision.NewNsfwModel(nsfw.ModelYahoo)
		model.Default = false
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{model}})
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.NsfwModel = "auto"
		hook := captureVisionSystemLog(t)
		c.applyNSFWModel()
		c.reportUnscreenedUploads()
		assert.Equal(t, nsfw.ModelYahoo, c.EffectiveNSFWModel())
		assert.Contains(t, c.NsfwModelPath(), string(nsfw.ModelYahoo))
		require.NotNil(t, hook.LastEntry())
		assert.Contains(t, hook.LastEntry().Message, "scripts/dist/download-models.sh yahoo_open_nsfw")
	})
	t.Run("NamedDisabledCustom", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeNsfw, Name: "custom_nsfw", Path: "custom/model.onnx", Disabled: true}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.NsfwModel = "custom_nsfw"
		assert.Equal(t, nsfw.ModelNone, c.EffectiveNSFWModel())
	})
	t.Run("None", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.NsfwModel = "none"
		c.applyNSFWModel()
		assert.Equal(t, nsfw.ModelNone, c.EffectiveNSFWModel())
		assert.Empty(t, c.NsfwModelPath())
		assert.Equal(t, "none", c.NsfwModelRuntime())
		assert.Nil(t, vision.Config.Model(vision.ModelTypeNsfw))
		require.False(t, vision.Config.Models[1].Disabled)
		require.True(t, vision.Config.Models[1].DisabledByMode)
	})
	t.Run("CustomFromVisionYaml", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{&vision.Model{
			Type: vision.ModelTypeNsfw, Name: "custom_nsfw", Path: "custom/model.onnx",
		}}})
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.NsfwModel = "auto"
		c.applyNSFWModel()
		assert.Equal(t, nsfw.ModelName("custom_nsfw"), c.EffectiveNSFWModel())
		assert.Equal(t, filepath.Join(c.ModelsPath(), "custom", "model.onnx"), c.NsfwModelPath())
	})
}

// TestConfig_installedVisionModels verifies automatic selection follows installed artifacts.
func TestConfig_installedVisionModels(t *testing.T) {
	t.Run("Labels", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		assert.Equal(t, classify.ModelNone, c.installedLabelModel())

		installVisionTestArtifact(t, c.ModelsPath(), string(classify.ModelEfficientFormerV2S1), classify.FindModel(classify.ModelEfficientFormerV2S1).ONNX.File)
		assert.Equal(t, classify.ModelEfficientFormerV2S1, c.installedLabelModel())

		installVisionTestArtifact(t, c.ModelsPath(), string(classify.ModelEfficientFormerV2S2), classify.FindModel(classify.ModelEfficientFormerV2S2).ONNX.File)
		assert.Equal(t, classify.ModelEfficientFormerV2S2, c.installedLabelModel())
	})
	t.Run("NSFW", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		assert.Equal(t, nsfw.ModelNone, c.installedNSFWModel())

		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.ModelAdamCoddINT8), nsfw.FindModel(nsfw.ModelAdamCoddINT8).ONNX.File)
		assert.Equal(t, nsfw.ModelAdamCoddINT8, c.installedNSFWModel())

		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.ModelYahoo), nsfw.FindModel(nsfw.ModelYahoo).ONNX.File)
		assert.Equal(t, nsfw.ModelYahoo, c.installedNSFWModel())
	})
}

// installVisionTestArtifact creates a non-empty file for installed-model selection tests.
func installVisionTestArtifact(t *testing.T, modelsPath, name, fileName string) {
	t.Helper()
	dir := filepath.Join(modelsPath, name)
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(dir, fileName), []byte("model"), fs.ModeFile))
}

// TestConfig_reportUnscreenedUploads verifies the missing-detector warning conditions.
func TestConfig_reportUnscreenedUploads(t *testing.T) {
	t.Run("MissingDetector", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{})
		c := NewConfig(CliTestContext())
		c.options.NsfwModel = "none"
		c.options.UploadNSFW = false
		hook := captureVisionSystemLog(t)

		c.reportUnscreenedUploads()

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Contains(t, entry.Message, "no nsfw model is configured")
	})
	t.Run("AutoMissingArtifact", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.NsfwModel = "auto"
		c.options.UploadNSFW = false
		hook := captureVisionSystemLog(t)

		c.reportUnscreenedUploads()

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Contains(t, entry.Message, "scripts/dist/download-models.sh yahoo_open_nsfw")
		assert.Contains(t, entry.Message, "restart PhotoPrism")
	})
	t.Run("UploadsAllowed", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{})
		c := NewConfig(CliTestContext())
		c.options.UploadNSFW = true
		hook := captureVisionSystemLog(t)

		c.reportUnscreenedUploads()

		assert.Empty(t, hook.AllEntries())
	})
	t.Run("DetectorConfigured", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.DefaultModelName()), nsfw.FindModel(nsfw.DefaultModelName()).ONNX.File)
		c.options.UploadNSFW = false
		hook := captureVisionSystemLog(t)

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
	t.Run("NamedDefaultLabelsRunOnIndex", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		named := &vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.DefaultModelName())}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{named}})
		assert.True(t, c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex))
		assert.False(t, c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunNewlyIndexed))
	})
	t.Run("DefaultLabelsRunAlways", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		model := vision.DefaultLabelModel.Clone()
		model.Run = vision.RunAlways
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{model}})
		assert.True(t, c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex))
		assert.True(t, c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunNewlyIndexed))
	})
	t.Run("CustomLabelsRunAfterIndex", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		defaultModel := vision.DefaultLabelModel.Clone()
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

func TestConfig_OnnxProvider(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())
	})
	t.Run("CUDA", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.OnnxProvider = "cuda"
		assert.Equal(t, onnx.ProviderCUDA, c.OnnxProvider())
	})
	t.Run("Empty", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.OnnxProvider = ""
		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())
	})
	t.Run("Unsupported", func(t *testing.T) {
		// An unusable value must not stop inference, so it resolves to the default - and must
		// say so, or an operator reads the default as their setting having been applied.
		c := NewConfig(CliTestContext())
		hook := captureConfigLog(t)
		c.options.OnnxProvider = "rocm"

		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "rocm")

		// Reported once, not once per loaded model.
		c.OnnxProvider()
		assert.Len(t, hook.AllEntries(), 1)
	})
	t.Run("Nil", func(t *testing.T) {
		var c *Config
		assert.Equal(t, onnx.DefaultProvider, c.OnnxProvider())
	})
}

// TestConfig_PropagateOnnxProvider verifies label and NSFW factories receive the configured provider.
func TestConfig_PropagateOnnxProvider(t *testing.T) {
	previous := vision.OnnxProvider
	t.Cleanup(func() { vision.OnnxProvider = previous })

	c := NewConfig(CliTestContext())
	c.options.OnnxProvider = string(onnx.ProviderCUDA)
	c.Propagate()
	assert.Equal(t, onnx.ProviderCUDA, vision.OnnxProvider)
}

// captureConfigLog redirects the package logger for the duration of the test and returns its
// entries, so a "report it once" contract can be asserted on what was actually logged.
func captureConfigLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

func TestConfig_WarnVisionConfig(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		// The getters run per loaded model and from the config report, so a repeated call
		// must not repeat the warning. Asserted on the log, not on the map: storing the key
		// and still logging every time would satisfy the map.
		c := NewConfig(CliTestContext())
		hook := captureConfigLog(t)

		c.warnVisionConfig("test-vision-warning", "config: %s", "first")
		c.warnVisionConfig("test-vision-warning", "config: %s", "second")

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "config: first", hook.LastEntry().Message)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	})
	t.Run("DistinctKeys", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		hook := captureConfigLog(t)

		c.warnVisionConfig("test-vision-a", "config: a")
		c.warnVisionConfig("test-vision-b", "config: b")

		assert.Len(t, hook.AllEntries(), 2)
	})
}

// TestVisionKeyWarnings checks which Vision API keys are reported as unable to authenticate requests.
func TestVisionKeyWarnings(t *testing.T) {
	const stripped = "vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API"
	const dollar = "vision key contains $ and is compared as written, without expanding environment variables"

	t.Setenv("VISION_TEST_SECRET", "vision-api-shared-token")
	t.Setenv("VISION_TEST_INVALID", "vision!secret")

	for _, tc := range []struct {
		name     string
		key      string
		incoming bool
		outgoing bool
		want     []string
	}{
		{name: "Empty", key: "", incoming: true, outgoing: true},
		{name: "Valid", key: `Ab3"-+/=#@:;_. ok`, incoming: true, outgoing: true},
		{name: "Generated", key: "vision-api-shared-token", incoming: true},
		{name: "Exclamation", key: "SecretAccessToken!", incoming: true, want: []string{stripped}},
		{name: "Braces", key: "{secret}", incoming: true, want: []string{stripped}},
		{name: "DoubleSpace", key: "vision  key", incoming: true, want: []string{stripped}},
		{name: "NonAscii", key: "schlüssel", incoming: true, want: []string{stripped}},
		{name: "TooLong", key: strings.Repeat("a", 4097), incoming: true, want: []string{stripped}},
		{name: "ExclamationOutgoing", key: "SecretAccessToken!", outgoing: true, want: []string{stripped}},
		{name: "Unused", key: "SecretAccessToken!"},
		{name: "DollarIncoming", key: "$VISION_TEST_SECRET", incoming: true, want: []string{dollar}},
		{name: "DollarOutgoing", key: "$VISION_TEST_SECRET", outgoing: true},
		{name: "BracedIncoming", key: "${VISION_TEST_SECRET}", incoming: true, outgoing: true, want: []string{stripped, dollar}},
		{name: "BracedOutgoing", key: "${VISION_TEST_SECRET}", outgoing: true},
		{name: "DollarBothInvalid", key: "$VISION_TEST_INVALID", incoming: true, outgoing: true, want: []string{stripped, dollar}},
		{name: "BracedOutgoingInvalid", key: "${VISION_TEST_INVALID}", outgoing: true, want: []string{stripped}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, visionKeyWarnings(tc.key, tc.incoming, tc.outgoing))
		})
	}
}

// captureVisionKeyLog redirects the system log and the package logger for the duration of the test, and
// returns the system log entries after checking that the package logger received none.
func captureVisionKeyLog(t *testing.T) *test.Hook {
	t.Helper()

	appHook := captureConfigLog(t)
	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger

	t.Cleanup(func() {
		event.SystemLog = orig
		assert.Empty(t, appHook.AllEntries(), "vision key warnings must not reach the package logger")
	})

	return hook
}

// TestConfig_WarnVisionKey checks that problems with the configured Vision API key are logged once each.
func TestConfig_WarnVisionKey(t *testing.T) {
	t.Run("Warnings", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.VisionKey = "${VISION_SECRET}"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		c.warnVisionKey()

		require.Len(t, hook.AllEntries(), 2)
		for _, entry := range hook.AllEntries() {
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.True(t, strings.HasPrefix(entry.Message, "config: vision key contains "), entry.Message)
			assert.NotContains(t, entry.Message, "VISION_SECRET")
		}
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.VisionKey = " vision-api-shared-token "
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Unused", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = false
		c.options.VisionUri = ""
		c.options.VisionKey = "SecretAccessToken!"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Outgoing", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = false
		c.options.VisionUri = "https://vision.example.com/api/v1/vision"
		c.options.VisionKey = "SecretAccessToken!"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		require.Len(t, hook.AllEntries(), 1)
	})
	t.Run("OutgoingDollar", func(t *testing.T) {
		t.Setenv("VISION_SECRET", "")
		c := NewConfig(CliTestContext())
		c.options.VisionApi = false
		c.options.VisionUri = "https://vision.example.com/api/v1/vision"
		c.options.VisionKey = "$VISION_SECRET"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Demo", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.Demo = true
		c.options.VisionKey = "$VISION_SECRET"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("KeyFile", func(t *testing.T) {
		keyFile := filepath.Join(t.TempDir(), "vision_key")
		require.NoError(t, os.WriteFile(keyFile, []byte("SecretAccessToken!\n"), fs.ModeSecretFile))
		t.Setenv(FlagFileVar("VISION_KEY"), keyFile)

		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.VisionKey = ""
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "config: vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API", hook.LastEntry().Message)
	})
}

// TestConfig_InitWarnVisionKey checks that Init logs the Vision API key warnings to the system log.
func TestConfig_InitWarnVisionKey(t *testing.T) {
	c := NewIsolatedTestConfig("visionkeyinit", t.TempDir(), true)
	c.options.VisionApi = true
	c.options.VisionKey = "Secret Access Token!"

	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger

	// Init registers its database and propagates its settings, so the package's test config is restored.
	t.Cleanup(func() {
		event.SystemLog = orig
		_ = c.CloseDb()
		entity.SetDbProvider(TestConfig())
		TestConfig().Propagate()

		if c.DatabaseDriver() == dsn.DriverSQLite3 {
			for _, suffix := range []string{"", "-journal", "-wal", "-shm"} {
				_ = os.Remove(c.DatabaseDSN() + suffix)
			}
		}
	})

	require.NoError(t, c.Init())

	var warnings []string

	for _, entry := range hook.AllEntries() {
		if strings.HasPrefix(entry.Message, "config: vision key ") {
			warnings = append(warnings, entry.Message)
		}
	}

	assert.Equal(t, []string{"config: vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API"}, warnings)
}

// TestVisionModes verifies mode precedence and exclusive NSFW sources.
func TestVisionModes(t *testing.T) {
	for _, mode := range []string{"auto", "none", "labels"} {
		t.Run(mode, func(t *testing.T) {
			withVisionConfig(t, vision.NewConfig())
			c := NewConfig(CliTestContext())
			c.options.NsfwModel = mode
			c.options.DetectNSFW = true
			c.options.Experimental = false
			c.applyNSFWModel()
			assert.Equal(t, mode == "labels", c.DetectNSFWLabels())
			assert.Equal(t, mode == "auto", c.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunOnIndex))
			assert.Equal(t, mode == "auto", vision.Config.Model(vision.ModelTypeNsfw) != nil)
			c.options.DetectNSFW = false
			assert.False(t, c.DetectNSFWLabels())
			assert.False(t, c.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunOnIndex))
			c.options.DetectNSFW = true
			c.Propagate()
			assert.Equal(t, mode == "labels", vision.DetectNSFWLabels)
			t.Cleanup(func() { vision.DetectNSFWLabels = false })
		})
	}
	t.Run("DeprecatedPrecedence", func(t *testing.T) {
		c := TestConfig()
		previousMode, previousDisabled := c.options.LabelsModel, c.options.DisableClassification
		t.Cleanup(func() {
			c.options.LabelsModel, c.options.DisableClassification = previousMode, previousDisabled
		})
		c.options.DisableClassification = true
		c.options.LabelsModel = ""
		assert.Equal(t, classify.ModelNone, c.LabelModelSetting())
		assert.True(t, c.DisableClassification())
		assert.True(t, c.ClientUser(false).Disable.Classification)
		c.options.LabelsModel = "auto"
		assert.Equal(t, classify.ModelAuto, c.LabelModelSetting())
		assert.False(t, c.DisableClassification())
		assert.False(t, c.ClientUser(false).Disable.Classification)
		c.options.DisableClassification = false
		c.options.LabelsModel = "none"
		assert.True(t, c.DisableClassification())
	})
	t.Run("DeprecatedIgnoredReport", func(t *testing.T) {
		// ignoredReports returns the reports that the deprecated option is ignored.
		ignoredReports := func(hook *test.Hook) (result []*logrus.Entry) {
			for _, entry := range hook.AllEntries() {
				if strings.Contains(entry.Message, "disable-classification is ignored") {
					result = append(result, entry)
				}
			}
			return result
		}

		cases := []struct {
			name     string
			disabled bool
			mode     string
			expected classify.ModelName
			reports  int
		}{
			{"Ignored", true, "auto", classify.ModelAuto, 1},
			{"Honored", true, "", classify.ModelNone, 0},
			{"Whitespace", true, "  ", classify.ModelNone, 0},
			{"Unsupported", true, "off", classify.ModelNone, 0},
			{"NotSet", false, "auto", classify.ModelAuto, 0},
			{"UnsupportedNotSet", false, "off", classify.ModelAuto, 0},
			{"PaddedNone", false, " None ", classify.ModelNone, 0},
			{"PaddedAuto", true, " AUTO ", classify.ModelAuto, 1},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				c := NewConfig(CliTestContext())
				c.options.DisableClassification = tc.disabled
				c.options.LabelsModel = tc.mode
				hook := captureLog(t)
				assert.Equal(t, tc.expected, c.LabelModelSetting())
				assert.Equal(t, tc.expected, c.LabelModelSetting())
				reports := ignoredReports(hook)
				require.Len(t, reports, tc.reports)
				for _, entry := range reports {
					assert.Equal(t, logrus.InfoLevel, entry.Level)
				}
			})
		}
	})
	t.Run("UnsupportedLabelsModeReport", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{})
		c := NewConfig(CliTestContext())
		c.options.LabelsModel, c.options.DisableClassification = "off", true
		system, _ := captureLogChannels(t, c.reportVisionModes)
		require.NotEmpty(t, system)
		assert.Contains(t, system[0], "using none")
	})
	t.Run("InvalidModes", func(t *testing.T) {
		withVisionConfig(t, &vision.ConfigValues{})
		c := NewConfig(CliTestContext())
		c.options.LabelsModel, c.options.NsfwModel = "repvit_m1_0", "yahoo_open_nsfw"
		assert.Equal(t, classify.ModelAuto, c.LabelModelSetting())
		assert.Equal(t, nsfw.ModelAuto, c.NSFWModelSetting())
		system, _ := captureLogChannels(t, c.reportVisionModes)
		assert.Len(t, system, 2)
	})
	t.Run("ModeSaveRoundTrip", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.LabelsModel, c.options.NsfwModel = "none", "labels"
		c.applyLabelModel()
		c.applyNSFWModel()
		filename := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, vision.Config.Save(filename))
		cfg := vision.NewConfig()
		require.NoError(t, cfg.Load(filename))
		require.NotNil(t, cfg.Model(vision.ModelTypeLabels))
		require.NotNil(t, cfg.Model(vision.ModelTypeNsfw))
	})
}

// TestVisionModeWarnings verifies warnings stay on the system channel and respect remote detectors.
func TestVisionModeWarnings(t *testing.T) {
	t.Run("RemoteDetector", func(t *testing.T) {
		model := &vision.Model{Type: vision.ModelTypeNsfw, Name: "remote", Service: vision.Service{Uri: "https://example.com", Method: "POST"}}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{model}})
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.NsfwModel = "auto"
		c.options.UploadNSFW = false
		system, browser := captureLogChannels(t, c.reportUnscreenedUploads)
		assert.Empty(t, system)
		assert.Empty(t, browser)
	})
	t.Run("LabelsUploads", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.NsfwModel = "labels"
		c.options.UploadNSFW = false
		system, browser := captureLogChannels(t, c.reportUnscreenedUploads)
		require.Len(t, system, 1)
		assert.Contains(t, system[0], "uploads are not screened")
		assert.Empty(t, browser)
		c.options.UploadNSFW = true
		system, _ = captureLogChannels(t, c.reportUnscreenedUploads)
		assert.Empty(t, system)
	})
	t.Run("LocalLabelsCannotDetect", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.NsfwModel = "labels"
		c.options.DetectNSFW = true
		system, browser := captureLogChannels(t, c.reportVisionModes)
		assert.True(t, namesPath(system, "no nsfw detection takes place"))
		assert.Empty(t, browser)
	})
	t.Run("RemoteLabelsCanDetect", func(t *testing.T) {
		model := &vision.Model{Type: vision.ModelTypeLabels, Engine: vision.ApiFormatOllama}
		model.ApplyEngineDefaults()
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{model}})
		c := NewConfig(CliTestContext())
		c.options.NsfwModel = "labels"
		c.options.DetectNSFW = true
		system, _ := captureLogChannels(t, c.reportVisionModes)
		assert.Empty(t, system)
	})
}

// captureVisionSystemLog captures operator-only configuration messages.
func captureVisionSystemLog(t *testing.T) *test.Hook {
	t.Helper()
	previous := event.SystemLog
	logger, hook := test.NewNullLogger()
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = previous })
	return hook
}

// TestModeDisablesAllEntries verifies earlier entries cannot bypass an option override.
func TestModeDisablesAllEntries(t *testing.T) {
	cfg := vision.NewConfig()
	cfg.Models = append(cfg.Models, &vision.Model{Type: vision.ModelTypeLabels, Name: "custom"}, &vision.Model{Type: vision.ModelTypeNsfw, Name: "custom"})
	withVisionConfig(t, cfg)
	c := NewConfig(CliTestContext())
	c.options.LabelsModel, c.options.NsfwModel = "none", "labels"
	c.applyLabelModel()
	c.applyNSFWModel()
	assert.Nil(t, vision.Config.Model(vision.ModelTypeLabels))
	assert.Nil(t, vision.Config.Model(vision.ModelTypeNsfw))
	c.options.LabelsModel, c.options.NsfwModel = "auto", "auto"
	c.applyLabelModel()
	c.applyNSFWModel()
	require.NotNil(t, vision.Config.Model(vision.ModelTypeLabels))
	require.NotNil(t, vision.Config.Model(vision.ModelTypeNsfw))
	assert.Equal(t, "custom", vision.Config.Model(vision.ModelTypeLabels).Name)
}
