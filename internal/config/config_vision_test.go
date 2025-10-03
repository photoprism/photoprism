package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/vision"
)

func TestConfig_VisionYaml(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/config/vision.yml", c.VisionYaml())
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
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/models/nasnet", path)
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
