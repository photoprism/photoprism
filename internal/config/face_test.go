package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_FaceSize(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 50, c.FaceSize())
	c.options.FaceSize = 30
	assert.Equal(t, 30, c.FaceSize())
	c.options.FaceSize = 1
	assert.Equal(t, 50, c.FaceSize())
}

func TestConfig_FaceScore(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 9.0, c.FaceScore())
	c.options.FaceScore = 8.5
	assert.Equal(t, 8.5, c.FaceScore())
	c.options.FaceScore = 0.1
	assert.Equal(t, 9.0, c.FaceScore())
}

func TestConfig_FaceOverlap(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 42, c.FaceOverlap())
	c.options.FaceOverlap = 300
	assert.Equal(t, 42, c.FaceOverlap())
	c.options.FaceOverlap = 1
	assert.Equal(t, 1, c.FaceOverlap())
}

func TestConfig_FaceClusterSize(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 80, c.FaceClusterSize())
	c.options.FaceClusterSize = 10
	assert.Equal(t, 80, c.FaceClusterSize())
	c.options.FaceClusterSize = 66
	assert.Equal(t, 66, c.FaceClusterSize())
}

func TestConfig_FaceClusterScore(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 15, c.FaceClusterScore())
	c.options.FaceClusterScore = 0
	assert.Equal(t, 15, c.FaceClusterScore())
	c.options.FaceClusterScore = 55
	assert.Equal(t, 55, c.FaceClusterScore())
}

func TestConfig_FaceClusterCore(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 4, c.FaceClusterCore())
	c.options.FaceClusterCore = 1000
	assert.Equal(t, 4, c.FaceClusterCore())
	c.options.FaceClusterCore = 1
	assert.Equal(t, 1, c.FaceClusterCore())
}

func TestConfig_FaceClusterDist(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 0.64, c.FaceClusterDist())
	c.options.FaceClusterDist = 0.01
	assert.Equal(t, 0.64, c.FaceClusterDist())
	c.options.FaceClusterDist = 0.34
	assert.Equal(t, 0.34, c.FaceClusterDist())
}

func TestConfig_FaceMatchDist(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 0.46, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.1
	assert.Equal(t, 0.1, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.01
	assert.Equal(t, 0.46, c.FaceMatchDist())
}
