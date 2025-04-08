package config

import (
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

func TestConfig_TensorFlowVersion(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.TensorFlowVersion()
	assert.IsType(t, "2.18.0", version)
}

func TestConfig_TensorFlowModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.NasnetModelPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/nasnet", path)
}

func TestConfig_TensorFlowDisabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.DisableTensorFlow()
	assert.Equal(t, false, version)
}

func TestConfig_NSFWModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.NSFWModelPath(), "/assets/nsfw")
}

func TestConfig_FaceNetModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.FaceNetModelPath(), "/assets/facenet")
}

func TestConfig_DetectNSFW(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.DetectNSFW()
	assert.Equal(t, true, result)
}
