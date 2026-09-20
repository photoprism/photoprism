package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/onnx"
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

func TestConfig_TensorFlowDisabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.DisableTensorFlow()
	assert.Equal(t, false, version)
}

func TestConfig_NSFWModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.NsfwModelPath(), "/assets/models/nsfw")
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
		// An unusable value must not stop inference, so it resolves to the default.
		c := NewConfig(CliTestContext())
		c.options.OnnxProvider = "rocm"
		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())
	})
	t.Run("Nil", func(t *testing.T) {
		var c *Config
		assert.Equal(t, onnx.DefaultProvider, c.OnnxProvider())
	})
}

func TestConfig_WarnVisionConfig(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		// The getters run per loaded model and from the config report, so a repeated call
		// must not repeat the warning.
		c := NewConfig(CliTestContext())

		c.warnVisionConfig("test-vision-warning", "config: %s", "first")
		_, warned := c.warnedOnce.Load("test-vision-warning")
		assert.True(t, warned)

		c.warnVisionConfig("test-vision-warning", "config: %s", "second")
		assert.True(t, warned)
	})
	t.Run("DistinctKeys", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		c.warnVisionConfig("test-vision-a", "config: a")
		c.warnVisionConfig("test-vision-b", "config: b")

		_, warnedA := c.warnedOnce.Load("test-vision-a")
		_, warnedB := c.warnedOnce.Load("test-vision-b")
		assert.True(t, warnedA)
		assert.True(t, warnedB)
	})
}
