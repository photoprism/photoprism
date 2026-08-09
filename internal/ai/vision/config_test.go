package vision

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestConfigValues_Load(t *testing.T) {
	t.Run("DefaultModelWithCustomRun", func(t *testing.T) {
		originalRun := NasnetModel.Run
		t.Cleanup(func() {
			NasnetModel.Run = originalRun
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
			assert.Equal(t, NasnetModel.Name, labels.Name)
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

		if got := cfg.Models[0]; got.Name != NasnetModel.Name {
			t.Fatalf("expected placeholder to become nasnet, got %s", got.Name)
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
	defaultModel := *NasnetModel //nolint:govet // copy for test to avoid mutating shared model
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
			&defaultModel,
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
	defaultModel := NasnetModel.Clone()
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
		cfg := &ConfigValues{Models: Models{NasnetModel.Clone()}}
		assertConfigShouldRun(t, cfg, RunManual, true)
		assertConfigShouldRun(t, cfg, RunOnSchedule, true)
		assertConfigShouldRun(t, cfg, RunAlways, true)
		assertConfigShouldRun(t, cfg, RunOnIndex, true)
		assertConfigShouldRun(t, cfg, RunNewlyIndexed, false)
		assertConfigShouldRun(t, cfg, RunNever, false)
	})
	t.Run("CustomOverridesDefault", func(t *testing.T) {
		defaultModel := NasnetModel.Clone()
		custom := &Model{Type: ModelTypeLabels, Name: "custom"}
		cfg := &ConfigValues{Models: Models{defaultModel, custom}}
		assertConfigShouldRun(t, cfg, RunManual, true)
		assertConfigShouldRun(t, cfg, RunAlways, false)
		assertConfigShouldRun(t, cfg, RunOnIndex, false)
		assertConfigShouldRun(t, cfg, RunNewlyIndexed, true)
	})
	t.Run("DisabledCustomFallsBack", func(t *testing.T) {
		defaultModel := NasnetModel.Clone()
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
