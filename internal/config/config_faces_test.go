package config

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
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

func TestConfig_FaceAngles(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.DefaultAngles, c.FaceAngles())

	c.options.FaceAngles = []float64{-0.5, 0, 0.5}
	assert.Equal(t, []float64{-0.5, 0, 0.5}, c.FaceAngles())

	c.options.FaceAngles = []float64{math.Pi + 0.1, math.NaN(), 4}
	assert.Equal(t, face.DefaultAngles, c.FaceAngles())
}

func TestConfig_FaceEngineShouldRun(t *testing.T) {
	t.Run("AutoHighThreads", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 4

		assert.True(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.False(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))
		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
	})
	t.Run("AutoLowThreads", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 2

		assert.False(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.True(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))
	})
	t.Run("ExplicitRunModes", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		c.options.FaceEngineRun = vision.RunOnIndex
		assert.True(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.False(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))

		c.options.FaceEngineRun = vision.RunNever
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.False(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))

		c.options.FaceEngineRun = vision.RunManual
		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnDemand))

		c.options.DisableFaces = true
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnIndex))
	})
}

func TestConfig_FaceEngine(t *testing.T) {
	c := NewConfig(CliTestContext())
	tempModels := t.TempDir()
	c.options.ModelsPath = tempModels
	c.options.FaceEngine = face.EnginePigo

	assert.Equal(t, face.EnginePigo, c.FaceEngine())

	modelDir := filepath.Join(tempModels, "scrfd")
	require.NoError(t, os.MkdirAll(modelDir, 0o755))
	modelFile := filepath.Join(modelDir, face.DefaultONNXModelFilename)
	require.NoError(t, os.WriteFile(modelFile, []byte("onnx"), 0o644))

	c.options.FaceEngine = face.EngineAuto
	assert.Equal(t, face.EngineONNX, c.FaceEngine())

	c.options.FaceEngine = face.EnginePigo
	assert.Equal(t, face.EnginePigo, c.FaceEngine())

	c.options.FaceEngine = face.EngineONNX
	assert.Equal(t, face.EngineONNX, c.FaceEngine())
}

func TestConfig_FaceEngineRunType(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "auto", c.FaceEngineRunType())
	assert.Equal(t, "", c.options.FaceEngineRun)

	c.options.FaceEngineRun = vision.RunOnDemand
	assert.Equal(t, vision.RunOnDemand, c.FaceEngineRunType())

	c.options.FaceEngineRun = vision.RunAuto
	c.options.FaceEngine = face.EngineONNX
	c.options.FaceEngineThreads = 1
	assert.Equal(t, "on-demand", c.FaceEngineRunType())
	assert.Equal(t, "on-demand", vision.ParseRunType(c.options.FaceEngineRun))

	c.options.FaceEngineThreads = 4
	c.options.FaceEngineRun = vision.RunAuto
	assert.Equal(t, "auto", c.FaceEngineRunType())
	assert.Equal(t, "", c.options.FaceEngineRun)
}

func TestConfig_FaceEngineThreads(t *testing.T) {
	c := NewConfig(CliTestContext())
	expected := runtime.NumCPU() / 2
	if expected < 1 {
		expected = 1
	}
	assert.Equal(t, expected, c.FaceEngineThreads())

	c.options.FaceEngineThreads = 8
	assert.Equal(t, 8, c.FaceEngineThreads())
}

func TestConfig_FaceEngineModelPath(t *testing.T) {
	t.Run("DefaultPath", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempModels := t.TempDir()
		c.options.ModelsPath = tempModels

		path := c.FaceEngineModelPath()
		assert.Contains(t, path, "scrfd")
		expected := filepath.Join(tempModels, "scrfd", face.DefaultONNXModelFilename)
		assert.Equal(t, expected, path)
	})
}
