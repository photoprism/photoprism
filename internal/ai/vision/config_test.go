package vision

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestOptions(t *testing.T) {
	var configPath = fs.Abs("testdata")
	var configFile = filepath.Join(configPath, "vision.yml")

	t.Run("Save", func(t *testing.T) {
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

func TestConfigValues_LoadDefaultModelWithCustomRun(t *testing.T) {
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
}

func TestConfigValues_LoadDefaultModelDisabled(t *testing.T) {
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
}

func TestConfigModelPrefersLastEnabled(t *testing.T) {
	defaultModel := *NasnetModel
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
