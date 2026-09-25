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
		c.options.LabelModel = "auto"
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
		c.options.LabelModel = "auto"
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
		c.options.LabelModel = "auto"
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
		c.options.LabelModel = "auto"
		c.applyLabelModel()
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
		assert.Empty(t, c.LabelModelPath())
		assert.Equal(t, "none", c.LabelModelRuntime())
	})
	t.Run("Named", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.LabelModel = string(classify.ModelRepViTM10)
		hook := captureLog(t)
		c.applyLabelModel()
		assert.Equal(t, classify.ModelRepViTM10, c.EffectiveLabelModel())
		assert.Equal(t, string(classify.ModelRepViTM10), vision.Config.Model(vision.ModelTypeLabels).Name)
		require.NotNil(t, hook.LastEntry())
		assert.Contains(t, hook.LastEntry().Message, "scripts/dist/download-models.sh repvit_m1_0")
	})
	t.Run("NamedDisabledCustom", func(t *testing.T) {
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom_21k", Path: "custom_21k", Disabled: true}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{custom}})
		c := NewConfig(CliTestContext())
		c.options.LabelModel = "custom_21k"
		assert.Equal(t, classify.ModelNone, c.EffectiveLabelModel())
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
		assert.Equal(t, nsfw.ModelYahoo, c.EffectiveNSFWModel())
		require.False(t, vision.Config.Models[1].Disabled)
	})
	t.Run("Named", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.NsfwModel = string(nsfw.ModelYahoo)
		hook := captureLog(t)
		c.applyNSFWModel()
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
		require.True(t, vision.Config.Models[1].Disabled)
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
		hook := captureLog(t)

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
		hook := captureLog(t)

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
		hook := captureLog(t)

		c.reportUnscreenedUploads()

		assert.Empty(t, hook.AllEntries())
	})
	t.Run("DetectorConfigured", func(t *testing.T) {
		withVisionConfig(t, vision.NewConfig())
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		installVisionTestArtifact(t, c.ModelsPath(), string(nsfw.DefaultModelName()), nsfw.FindModel(nsfw.DefaultModelName()).ONNX.File)
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
