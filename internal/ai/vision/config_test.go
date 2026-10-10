package vision

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestOptions(t *testing.T) {
	var configPath = fs.Abs("testdata")
	var configFile = filepath.Join(configPath, "vision.yml")

	t.Run("Save", func(t *testing.T) {
		// Regenerate the committed fixture deterministically, independent of any ambient
		// OLLAMA_* env, so it always reflects the self-hosted defaults (gemma4 + Think: "false").
		origModel := CaptionModel.Model
		origService := CaptionModel.Service
		t.Cleanup(func() {
			CaptionModel.Model = origModel
			CaptionModel.Service = origService
			ensureEnvOnce = sync.Once{}
			registerOllamaEngineDefaults()
		})
		t.Setenv(ollama.APIKeyEnv, "")
		t.Setenv(ollama.BaseUrlEnv, ollama.DefaultBaseUrl)
		ensureEnvOnce = sync.Once{}
		CaptionModel.Model = ""
		CaptionModel.Service = Service{}
		registerOllamaEngineDefaults()

		_ = os.Remove(configFile)
		options := NewConfig()
		err := options.Save(configFile)
		assert.NoError(t, err)
		err = options.Load(configFile)
		assert.NoError(t, err)
	})
	t.Run("LoadMissingFile", func(t *testing.T) {
		options := NewConfig()
		err := options.Load(filepath.Join(configPath, "invalid.yml"))
		assert.Error(t, err)
	})
}

// TestSetOnnxProvider verifies empty values reset the shared inference provider.
func TestSetOnnxProvider(t *testing.T) {
	previous := OnnxProvider
	t.Cleanup(func() { OnnxProvider = previous })

	SetOnnxProvider(onnx.ProviderCUDA)
	assert.Equal(t, onnx.ProviderCUDA, OnnxProvider)
	SetOnnxProvider("")
	assert.Equal(t, onnx.DefaultProvider, OnnxProvider)
}

// TestNewConfigClonesModels verifies per-config runtime state is independent.
func TestNewConfigClonesModels(t *testing.T) {
	first := NewConfig()
	second := NewConfig()

	firstLabel := first.Model(ModelTypeLabels)
	secondLabel := second.Model(ModelTypeLabels)
	require.NotNil(t, firstLabel)
	require.NotNil(t, secondLabel)
	require.NotSame(t, firstLabel, secondLabel)

	firstLabel.classifyModel = &classify.Model{}
	firstLabel.Options = &ModelOptions{Stop: []string{"first"}}
	firstLabel.TensorFlow = &tensorflow.ModelInfo{
		Tags:  []string{"first"},
		Input: &tensorflow.PhotoInput{Intervals: []tensorflow.Interval{{Start: -1, End: 1}}},
	}
	clone := firstLabel.Clone()
	require.NotNil(t, clone)
	assert.Nil(t, clone.classifyModel)
	require.NotNil(t, clone.Options)
	clone.Options.Stop[0] = "second"
	assert.Equal(t, "first", firstLabel.Options.Stop[0])
	require.NotNil(t, clone.TensorFlow)
	clone.TensorFlow.Tags[0] = "second"
	clone.TensorFlow.Input.Intervals[0].Start = 0
	assert.Equal(t, "first", firstLabel.TensorFlow.Tags[0])
	assert.Equal(t, float32(-1), firstLabel.TensorFlow.Input.Intervals[0].Start)

	unsafeClassIndex := 0
	nsfwModel := &Model{UnsafeClassIndex: &unsafeClassIndex}
	nsfwClone := nsfwModel.Clone()
	require.NotNil(t, nsfwClone.UnsafeClassIndex)
	require.NotSame(t, nsfwModel.UnsafeClassIndex, nsfwClone.UnsafeClassIndex)
	assert.Equal(t, 0, *nsfwClone.UnsafeClassIndex)

	firstLabel.Disabled = true
	assert.NotNil(t, second.Model(ModelTypeLabels))
	assert.False(t, DefaultLabelModel.Disabled)
}

func TestConfigValues_Load(t *testing.T) {
	t.Run("RejectsMultipleLocalRuntimes", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")
		err := os.WriteFile(configFile, []byte("Models:\n- Type: labels\n  Name: invalid\n  TensorFlow: {}\n  ONNX: {}\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both TensorFlow and ONNX")
	})
	t.Run("ReplacesOllamaNsfwModel", func(t *testing.T) {
		system := captureSystemLog(t)
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(configFile, []byte("Models:\n- Type: nsfw\n  Name: gemma4\n  Engine: ollama\n  Run: manual\n"), fs.ModeConfigFile))

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		configured := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, configured)
		assert.Equal(t, NsfwModel.Name, configured.Name)
		assert.NotNil(t, configured.ONNX)
		assert.Empty(t, configured.Engine)
		assert.Equal(t, RunManual, configured.Run)

		warnings := warnMessages(system.AllEntries())
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "nsfw model gemma4:latest cannot use the ollama request format, so the built-in detector is used")
	})
	t.Run("ReplacesOpenAINsfwModel", func(t *testing.T) {
		captureSystemLog(t)
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(configFile, []byte("Models:\n- Type: nsfw\n  Name: gpt-5-mini\n  Engine: openai\n"), fs.ModeConfigFile))

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		configured := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, configured)
		assert.Equal(t, NsfwModel.Name, configured.Name)
		assert.Equal(t, NsfwModel.Run, configured.Run)
	})
	t.Run("KeepsVisionApiNsfwModel", func(t *testing.T) {
		system := captureSystemLog(t)
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(configFile, []byte("Models:\n- Type: nsfw\n  Name: remote-nsfw\n  Service:\n    Uri: https://vision.example.com/api/v1/vision/nsfw\n"), fs.ModeConfigFile))

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		configured := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, configured)
		assert.Equal(t, "remote-nsfw", configured.Name)
		assert.Empty(t, warnMessages(system.AllEntries()))
	})
	t.Run("MapsTensorFlowLabelModel", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Models:\n- Type: labels\n  Name: custom\n  TensorFlow: {}\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		require.Len(t, cfg.Models, len(DefaultModels))
		require.NotNil(t, cfg.Model(ModelTypeLabels))
		assert.Equal(t, DefaultLabelModel.Name, cfg.Models[0].Name)
		assert.True(t, cfg.Models[0].Default)
		assert.False(t, cfg.Models[0].Disabled)
		assert.Nil(t, cfg.Models[0].TensorFlow)
	})
	t.Run("MapsTensorFlowNSFWModel", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Models:\n- Type: nsfw\n  Name: nsfw\n  TensorFlow: {}\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		configured := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, configured)
		assert.Equal(t, NsfwModel.Name, configured.Name)
		assert.True(t, configured.Default)
		assert.Nil(t, configured.TensorFlow)
	})
	t.Run("NSFWThresholdForLabelsOnly", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Thresholds:\n  NSFW: 60\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		assert.Equal(t, 60, cfg.Thresholds.NSFW)
		assert.Equal(t, 60, cfg.Thresholds.GetNSFW())
		assert.Nil(t, cfg.Thresholds.NSFWUpload)
		assert.Nil(t, cfg.Thresholds.NSFWIndex)
		assert.False(t, cfg.Thresholds.NSFWUploadIsSet())
		assert.False(t, cfg.Thresholds.NSFWIndexIsSet())
	})
	t.Run("NormalizesNSFWThreshold", func(t *testing.T) {
		cases := map[string]int{
			"":              DefaultNSFWThreshold,
			"  NSFW: 0\n":   DefaultNSFWThreshold,
			"  NSFW: -1\n":  DefaultNSFWThreshold,
			"  NSFW: 1\n":   1,
			"  NSFW: 100\n": 100,
			"  NSFW: 150\n": 100,
		}

		for value, expected := range cases {
			configFile := filepath.Join(t.TempDir(), "vision.yml")
			err := os.WriteFile(configFile, []byte("Thresholds:\n  Confidence: 10\n"+value), fs.ModeConfigFile)
			require.NoError(t, err)

			cfg := NewConfig()
			require.NoError(t, cfg.Load(configFile))
			assert.Equal(t, expected, cfg.Thresholds.NSFW, "%q", value)
			assert.Equal(t, expected, cfg.Thresholds.GetNSFW(), "%q", value)
		}
	})
	t.Run("SavesClampedNSFWThreshold", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Thresholds:\n  NSFW: 150\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		require.NoError(t, cfg.Save(configFile))
		data, err := os.ReadFile(configFile) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(data), "NSFW: 100\n")
	})
	t.Run("LoadsContextNSFWThresholds", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Thresholds:\n  NSFW: 75\n  NSFWUpload: 50\n  NSFWIndex: 110\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		require.NotNil(t, cfg.Thresholds.NSFWUpload)
		require.NotNil(t, cfg.Thresholds.NSFWIndex)
		assert.Equal(t, 50, *cfg.Thresholds.NSFWUpload)
		assert.Equal(t, 100, *cfg.Thresholds.NSFWIndex)
		assert.Equal(t, 75, cfg.Thresholds.NSFW)
		assert.Equal(t, 50, cfg.Thresholds.GetNSFWUpload())
		assert.Equal(t, 100, cfg.Thresholds.GetNSFWIndex())
		assert.Equal(t, 75, cfg.Thresholds.GetNSFW())
	})
	t.Run("NormalizesZeroContextNSFWThresholds", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Thresholds:\n  NSFW: 80\n  NSFWUpload: 0\n  NSFWIndex: 0\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		require.NotNil(t, cfg.Thresholds.NSFWUpload)
		require.NotNil(t, cfg.Thresholds.NSFWIndex)
		assert.Equal(t, NSFWThresholdAuto, *cfg.Thresholds.NSFWUpload)
		assert.Equal(t, NSFWThresholdAuto, *cfg.Thresholds.NSFWIndex)
		assert.False(t, cfg.Thresholds.NSFWUploadIsSet())
		assert.False(t, cfg.Thresholds.NSFWIndexIsSet())
		assert.Equal(t, 80, cfg.Thresholds.GetNSFW())
	})
	t.Run("KeepsSavedNSFWThreshold", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		data, err := os.ReadFile(filepath.Join("testdata", "vision-251130.yml"))
		require.NoError(t, err)
		require.Contains(t, string(data), "  NSFW: 75\n")
		require.NoError(t, os.WriteFile(configFile, data, fs.ModeConfigFile)) //nolint:gosec // test file

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		assert.Equal(t, 75, cfg.Thresholds.NSFW)
		require.NoError(t, cfg.Save(configFile))
		saved, err := os.ReadFile(configFile) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Contains(t, string(saved), "  NSFW: 75\n")
		assert.NotContains(t, string(saved), "NSFWLabels")

		loaded := NewConfig()
		require.NoError(t, loaded.Load(configFile))
		assert.Equal(t, cfg.Thresholds, loaded.Thresholds)
		assert.Equal(t, 75, loaded.Thresholds.GetNSFW())
	})
	t.Run("IgnoresUnknownNSFWLabelsKey", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Thresholds:\n  NSFW: 75\n  NSFWLabels: 60\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		assert.Equal(t, 75, cfg.Thresholds.GetNSFW())
		assert.False(t, cfg.Thresholds.NSFWUploadIsSet())
		assert.False(t, cfg.Thresholds.NSFWIndexIsSet())
		require.NoError(t, cfg.Save(configFile))
		data, err := os.ReadFile(configFile) //nolint:gosec // test file
		require.NoError(t, err)
		assert.NotContains(t, string(data), "NSFWLabels")
	})
	t.Run("PreservesExplicitClassIndexZero", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Models:\n- Type: nsfw\n  Name: custom\n  Reduction: softmax-unsafe\n  UnsafeClassIndex: 0\n  ONNX: {}\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		model := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, model)
		require.NotNil(t, model.UnsafeClassIndex)
		assert.Equal(t, 0, *model.UnsafeClassIndex)
	})
	t.Run("PreservesProbabilityOutput", func(t *testing.T) {
		configFile := filepath.Join(t.TempDir(), "vision.yml")
		err := os.WriteFile(configFile, []byte("Models:\n- Type: labels\n  Name: custom\n  ONNX:\n    Output:\n      Logits: false\n"), fs.ModeConfigFile)
		require.NoError(t, err)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(configFile))
		model := cfg.Model(ModelTypeLabels)
		require.NotNil(t, model)
		require.NotNil(t, model.ONNX)
		require.NotNil(t, model.ONNX.Output)
		require.NotNil(t, model.ONNX.Output.Logits)
		assert.False(t, model.ONNX.Output.OutputsLogits())
	})
	t.Run("DefaultModelWithCustomRun", func(t *testing.T) {
		originalRun := DefaultLabelModel.Run
		t.Cleanup(func() {
			DefaultLabelModel.Run = originalRun
		})

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		err := os.WriteFile(configFile, []byte("Models:\n- Type: labels\n  Default: true\n  Run: on-demand\n"), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		assert.Equal(t, RunOnDemand, cfg.RunType(ModelTypeLabels))
		assert.True(t, cfg.ShouldRun(ModelTypeLabels, RunOnSchedule))
		assert.False(t, cfg.ShouldRun(ModelTypeLabels, RunOnIndex))
	})
	t.Run("AddsMissingDefaults", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		configYml := "Models:\n- Type: caption\n  Name: custom-caption\n"

		err := os.WriteFile(configFile, []byte(configYml), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		assert.Len(t, cfg.Models, len(DefaultModels))

		if labels := cfg.Model(ModelTypeLabels); assert.NotNil(t, labels) {
			assert.Equal(t, DefaultLabelModel.Name, labels.Name)
		}

		if caption := cfg.Model(ModelTypeCaption); assert.NotNil(t, caption) {
			assert.Equal(t, "custom-caption", caption.Name)
		}
	})
	t.Run("AddsDefaultsWhenModelsMissing", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		// Empty config should be populated with all default models.
		err := os.WriteFile(configFile, []byte(""), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		assert.Len(t, cfg.Models, len(DefaultModels))
		assert.True(t, cfg.IsDefault(ModelTypeLabels))
		assert.True(t, cfg.IsDefault(ModelTypeNsfw))
		assert.True(t, cfg.IsDefault(ModelTypeFace))
	})
	t.Run("DefaultModelDisabled", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		err := os.WriteFile(configFile, []byte("Models:\n- Type: labels\n  Default: true\n  Disabled: true\n"), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		if m := cfg.Model(ModelTypeLabels); m != nil {
			t.Fatalf("expected disabled default model to be ignored, got %v", m)
		}

		assert.Equal(t, RunNever, cfg.RunType(ModelTypeLabels))
		assert.False(t, cfg.ShouldRun(ModelTypeLabels, RunManual))
	})
	t.Run("MissingThresholdsUsesDefaults", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		err := os.WriteFile(configFile, []byte("Models:\n- Type: labels\n"), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		assert.Equal(t, DefaultThresholds, cfg.Thresholds)
	})
	t.Run("NormalizePreserved", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		configYml := "Models:\n- Type: labels\n  Name: custom\n  Engine: ollama\n  Normalize: phrase\n"

		err := os.WriteFile(configFile, []byte(configYml), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		m := cfg.Model(ModelTypeLabels)
		require.NotNil(t, m)
		assert.Equal(t, NormalizePhrase, m.Normalize)
		assert.Equal(t, NormalizePhrase, m.GetNormalize())
	})
	t.Run("NormalizeUnquotedFalse", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		// YAML parses a bare "false" as a boolean, which must still reach the string field.
		configYml := "Models:\n- Type: labels\n  Name: custom\n  Engine: ollama\n  Normalize: false\n"

		err := os.WriteFile(configFile, []byte(configYml), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		m := cfg.Model(ModelTypeLabels)
		require.NotNil(t, m)
		assert.Equal(t, NormalizeFalse, m.GetNormalize())
	})
	t.Run("NormalizeSurvivesSave", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		cfg := NewConfig()
		cfg.Models = Models{{Type: ModelTypeLabels, Name: "custom", Engine: "ollama", Normalize: NormalizePhrase}}
		assert.NoError(t, cfg.Save(configFile))

		reloaded := NewConfig()
		assert.NoError(t, reloaded.Load(configFile))

		m := reloaded.Model(ModelTypeLabels)
		require.NotNil(t, m)
		assert.Equal(t, NormalizePhrase, m.Normalize)
	})
	t.Run("NormalizeInvalidCleared", func(t *testing.T) {
		useSelfHostedOllamaDefaults(t)

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "vision.yml")

		configYml := "Models:\n- Type: labels\n  Name: custom\n  Engine: ollama\n  Normalize: bogus\n"

		err := os.WriteFile(configFile, []byte(configYml), fs.ModeConfigFile)
		assert.NoError(t, err)

		cfg := NewConfig()
		err = cfg.Load(configFile)
		assert.NoError(t, err)

		m := cfg.Model(ModelTypeLabels)
		require.NotNil(t, m)
		assert.Equal(t, NormalizeAuto, m.Normalize)
		assert.Equal(t, NormalizeWord, m.GetNormalize())
	})
}

// TestConfigValues_SetModel verifies replacement and append behavior by model type.
func TestConfigValues_SetModel(t *testing.T) {
	cfg := &ConfigValues{Models: Models{{Type: ModelTypeCaption, Name: "caption"}}}
	cfg.SetModel(&Model{Type: ModelTypeLabels, Name: "first", ONNX: &onnx.ModelInfo{}})
	assert.Len(t, cfg.Models, 2)
	assert.Equal(t, "first", cfg.Model(ModelTypeLabels).Name)

	cfg.SetModel(&Model{Type: ModelTypeLabels, Name: "second", ONNX: &onnx.ModelInfo{}})
	assert.Len(t, cfg.Models, 2)
	assert.Equal(t, "second", cfg.Model(ModelTypeLabels).Name)
}

func TestConfigValues_applyDefaultModels(t *testing.T) {
	t.Run("ReplacesPlaceholderAndKeepsOverrides", func(t *testing.T) {
		cfg := &ConfigValues{
			Models: Models{
				{
					Type:     ModelTypeLabels,
					Default:  true,
					Run:      RunOnDemand,
					Disabled: true,
				},
			},
		}

		cfg.applyDefaultModels()

		if got := cfg.Models[0]; got.Name != DefaultLabelModel.Name {
			t.Fatalf("expected placeholder to become the default model, got %s", got.Name)
		} else if got.Run != RunOnDemand {
			t.Fatalf("expected Run to be preserved, got %s", got.Run)
		} else if !got.Disabled {
			t.Fatalf("expected Disabled to be preserved")
		}
	})
	t.Run("IgnoresNonDefaultEntries", func(t *testing.T) {
		original := &Model{Type: ModelTypeLabels, Name: "custom", Default: false}
		cfg := &ConfigValues{Models: Models{original}}

		cfg.applyDefaultModels()

		if cfg.Models[0] != original {
			t.Fatalf("expected non-default model to remain unchanged")
		}
	})
}

func TestConfigValues_ensureDefaultModels(t *testing.T) {
	t.Run("AppendsMissingDefaults", func(t *testing.T) {
		cfg := &ConfigValues{Models: Models{}}

		cfg.ensureDefaultModels()

		if len(cfg.Models) != len(DefaultModels) {
			t.Fatalf("expected %d models, got %d", len(DefaultModels), len(cfg.Models))
		}
	})
	t.Run("SkipsTypesAlreadyPresent", func(t *testing.T) {
		custom := &Model{Type: ModelTypeLabels, Name: "custom"}
		cfg := &ConfigValues{Models: Models{custom}}

		cfg.ensureDefaultModels()

		if len(cfg.Models) != len(DefaultModels) {
			t.Fatalf("expected defaults minus duplicate type, got %d", len(cfg.Models))
		}

		if cfg.Models[0] != custom && cfg.Models[len(cfg.Models)-1] != custom {
			t.Fatalf("expected existing custom model to remain")
		}
	})
	t.Run("TreatsDisabledCustomAsPresent", func(t *testing.T) {
		custom := &Model{Type: ModelTypeNsfw, Name: "custom", Disabled: true}
		cfg := &ConfigValues{Models: Models{custom}}

		cfg.ensureDefaultModels()

		countType := 0
		for _, m := range cfg.Models {
			if m.Type == ModelTypeNsfw {
				countType++
			}
		}
		if countType != 1 {
			t.Fatalf("expected no additional nsfw default when custom exists, got %d entries", countType)
		}
	})
}

func TestConfigModelPrefersLastEnabled(t *testing.T) {
	defaultModel := DefaultLabelModel.Clone()
	defaultModel.Disabled = false
	defaultModel.Name = "nasnet-default"

	customModel := &Model{
		Type:     ModelTypeLabels,
		Name:     "ollama-labels",
		Engine:   "ollama",
		Disabled: false,
	}

	cfg := &ConfigValues{
		Models: Models{
			defaultModel,
			customModel,
		},
	}

	got := cfg.Model(ModelTypeLabels)
	if got != customModel {
		t.Fatalf("expected last enabled model, got %v", got)
	}

	customModel.Disabled = true
	got = cfg.Model(ModelTypeLabels)
	if got == nil || got.Name != defaultModel.Name {
		t.Fatalf("expected fallback to default model, got %v", got)
	}
}

func TestConfigValues_IsDefaultAndIsCustom(t *testing.T) {
	defaultModel := DefaultLabelModel.Clone()
	defaultModel.Default = false

	t.Run("DefaultModel", func(t *testing.T) {
		cfg := &ConfigValues{Models: Models{defaultModel}}
		if !cfg.IsDefault(ModelTypeLabels) {
			t.Fatalf("expected default model to be reported as default")
		}
		if cfg.IsCustom(ModelTypeLabels) {
			t.Fatalf("expected default model not to be reported as custom")
		}
	})
	t.Run("CustomOverridesDefault", func(t *testing.T) {
		custom := &Model{Type: ModelTypeLabels, Name: "custom", Engine: "ollama"}
		cfg := &ConfigValues{Models: Models{defaultModel, custom}}
		if cfg.IsDefault(ModelTypeLabels) {
			t.Fatalf("expected custom model to disable default detection")
		}
		if !cfg.IsCustom(ModelTypeLabels) {
			t.Fatalf("expected custom model to be detected")
		}
	})
	t.Run("DisabledCustomFallsBackToDefault", func(t *testing.T) {
		custom := &Model{Type: ModelTypeLabels, Name: "custom", Engine: "ollama", Disabled: true}
		cfg := &ConfigValues{Models: Models{defaultModel, custom}}
		if !cfg.IsDefault(ModelTypeLabels) {
			t.Fatalf("expected disabled custom model to fall back to default")
		}
		if cfg.IsCustom(ModelTypeLabels) {
			t.Fatalf("expected disabled custom model not to force custom mode")
		}
	})
	t.Run("MissingModel", func(t *testing.T) {
		cfg := &ConfigValues{}
		if cfg.IsDefault(ModelTypeLabels) {
			t.Fatalf("expected missing model to return false for default detection")
		}
		if cfg.IsCustom(ModelTypeLabels) {
			t.Fatalf("expected missing model to return false for custom detection")
		}
	})
}

func TestConfigValues_ShouldRun(t *testing.T) {
	t.Run("MissingModel", func(t *testing.T) {
		cfg := &ConfigValues{}
		if cfg.ShouldRun(ModelTypeLabels, RunManual) {
			t.Fatalf("expected false when no model configured")
		}
	})
	t.Run("DefaultAutoModel", func(t *testing.T) {
		cfg := &ConfigValues{Models: Models{DefaultLabelModel.Clone()}}
		assertConfigShouldRun(t, cfg, RunManual, true)
		assertConfigShouldRun(t, cfg, RunOnSchedule, true)
		assertConfigShouldRun(t, cfg, RunAlways, true)
		assertConfigShouldRun(t, cfg, RunOnIndex, true)
		assertConfigShouldRun(t, cfg, RunNewlyIndexed, false)
		assertConfigShouldRun(t, cfg, RunNever, false)
	})
	t.Run("CustomOverridesDefault", func(t *testing.T) {
		defaultModel := DefaultLabelModel.Clone()
		custom := &Model{Type: ModelTypeLabels, Name: "custom"}
		cfg := &ConfigValues{Models: Models{defaultModel, custom}}
		assertConfigShouldRun(t, cfg, RunManual, true)
		assertConfigShouldRun(t, cfg, RunAlways, false)
		assertConfigShouldRun(t, cfg, RunOnIndex, false)
		assertConfigShouldRun(t, cfg, RunNewlyIndexed, true)
	})
	t.Run("DisabledCustomFallsBack", func(t *testing.T) {
		defaultModel := DefaultLabelModel.Clone()
		custom := &Model{Type: ModelTypeLabels, Name: "custom", Disabled: true}
		cfg := &ConfigValues{Models: Models{defaultModel, custom}}
		assertConfigShouldRun(t, cfg, RunManual, true)
		assertConfigShouldRun(t, cfg, RunAlways, true)
		assertConfigShouldRun(t, cfg, RunOnIndex, true)
		assertConfigShouldRun(t, cfg, RunNewlyIndexed, false)
	})
	t.Run("ManualOnly", func(t *testing.T) {
		model := &Model{Type: ModelTypeLabels, Run: RunManual}
		cfg := &ConfigValues{Models: Models{model}}
		assertConfigShouldRun(t, cfg, RunManual, true)
		assertConfigShouldRun(t, cfg, RunOnDemand, false)
		assertConfigShouldRun(t, cfg, RunOnIndex, false)
	})
}

func assertConfigShouldRun(t *testing.T, cfg *ConfigValues, when RunType, want bool) {
	t.Helper()
	if got := cfg.ShouldRun(ModelTypeLabels, when); got != want {
		t.Fatalf("ConfigValues.ShouldRun(%q) = %v, want %v", when, got, want)
	}
}

// TestNSFWThresholdContexts verifies that each detector path uses its own threshold.
func TestNSFWThresholdContexts(t *testing.T) {
	for _, context := range []string{"NSFWUpload", "NSFWIndex"} {
		t.Run(context, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "vision.yml")
			require.NoError(t, os.WriteFile(filename, []byte("Thresholds:\n  NSFW: 75\n  "+context+": 50\n"), fs.ModeConfigFile))
			cfg := NewConfig()
			require.NoError(t, cfg.Load(filename))
			assert.Equal(t, 75, cfg.Thresholds.GetNSFW())
			if context == "NSFWUpload" {
				assert.Equal(t, 50, cfg.Thresholds.GetNSFWUpload())
				assert.False(t, cfg.Thresholds.NSFWIndexIsSet())
			} else {
				assert.Equal(t, 50, cfg.Thresholds.GetNSFWIndex())
				assert.False(t, cfg.Thresholds.NSFWUploadIsSet())
			}
			require.NoError(t, cfg.Save(filename))
			loaded := NewConfig()
			require.NoError(t, loaded.Load(filename))
			assert.Equal(t, cfg.Thresholds, loaded.Thresholds)
		})
	}
}

// TestModeDisablementPersistence verifies option overrides are not saved as user disablement.
func TestModeDisablementPersistence(t *testing.T) {
	cfg := NewConfig()
	cfg.Models[0].DisabledByMode = true
	assert.Nil(t, cfg.Model(ModelTypeLabels))
	filename := filepath.Join(t.TempDir(), "vision.yml")
	require.NoError(t, cfg.Save(filename))
	loaded := NewConfig()
	require.NoError(t, loaded.Load(filename))
	require.NotNil(t, loaded.Model(ModelTypeLabels))
	assert.False(t, loaded.Models[0].Disabled)
	assert.False(t, loaded.Models[0].DisabledByMode)
}
