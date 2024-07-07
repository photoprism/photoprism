package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_TensorFlowVersion(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.TensorFlowVersion()
	assert.IsType(t, "1.15.0", version)
}

func TestConfig_TensorFlowModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.TensorFlowModelPath()
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
