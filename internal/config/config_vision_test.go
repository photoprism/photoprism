package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
