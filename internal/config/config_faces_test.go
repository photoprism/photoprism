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

func TestConfig_FaceEngine(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		engine := c.FaceEngine()
		assert.Contains(t, []string{face.EnginePigo, face.EngineONNX}, engine)
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.EngineNone, (*Config)(nil).FaceEngine())
	})
	t.Run("MissingVisionConfig", func(t *testing.T) {
		origVision := vision.Config
		vision.Config = nil
		defer func() { vision.Config = origVision }()

		c := NewConfig(CliTestContext())
		assert.Equal(t, face.EngineNone, c.FaceEngine())
	})
	t.Run("AutoResolvesToONNX", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempModels := t.TempDir()
		c.options.ModelsPath = tempModels

		modelDir := filepath.Join(tempModels, "scrfd")
		require.NoError(t, os.MkdirAll(modelDir, 0o755))
		modelFile := filepath.Join(modelDir, face.DefaultONNXModelFilename)
		require.NoError(t, os.WriteFile(modelFile, []byte("onnx"), 0o644))

		c.options.FaceEngine = face.EngineAuto
		assert.Equal(t, face.EngineONNX, c.FaceEngine())
	})
	t.Run("ExplicitEngine", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngine = face.EnginePigo
		assert.Equal(t, face.EnginePigo, c.FaceEngine())
		c.options.FaceEngine = face.EngineONNX
		assert.Equal(t, face.EngineONNX, c.FaceEngine())
	})
}

func TestConfig_FaceEngineShouldRun(t *testing.T) {
	t.Run("AutoHighThreads", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 4

		assert.True(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.False(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))
		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
		assert.True(t, c.FaceEngineShouldRun(vision.RunAuto))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule))
	})
	t.Run("AutoLowThreads", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 2

		assert.False(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.True(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))
		assert.True(t, c.FaceEngineShouldRun(vision.RunAuto))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule))
	})
	t.Run("ExplicitRunModes", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DisableFaces = true
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		c.options.DisableFaces = false
	})
	t.Run("RunOnDemandSkipsSchedule", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace}}}

		c := NewConfig(CliTestContext())
		m := vision.Config.Model(vision.ModelTypeFace)
		require.NotNil(t, m)
		m.Run = vision.RunOnDemand

		assert.True(t, c.FaceEngineShouldRun(vision.RunOnDemand))
		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
		assert.True(t, c.FaceEngineShouldRun(vision.RunAuto))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule))
	})
}

func TestConfig_FaceEngineRunType(t *testing.T) {
	t.Run("AutoDefaults", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 1
		assert.Equal(t, "auto", vision.ReportRunType(c.FaceEngineRunType()))

		c.options.DisableFaces = true
		assert.Equal(t, "never", vision.ReportRunType(c.FaceEngineRunType()))
		c.options.DisableFaces = false

		c.options.FaceEngineThreads = 4
		assert.Equal(t, "auto", vision.ReportRunType(c.FaceEngineRunType()))
	})
	t.Run("DisabledFaceModel", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace, Disabled: true}}}
		c := NewConfig(CliTestContext())
		assert.Equal(t, vision.RunNever, c.FaceEngineRunType())
	})
	t.Run("NoFaceModel", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{}}
		c := NewConfig(CliTestContext())
		assert.Equal(t, vision.RunNever, c.FaceEngineRunType())
	})
	t.Run("DelegatesToVisionModel", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace}}}
		c := NewConfig(CliTestContext())
		m := vision.Config.Model(vision.ModelTypeFace)
		require.NotNil(t, m)
		m.Run = vision.RunOnSchedule
		require.Equal(t, vision.RunOnSchedule, vision.Config.RunType(vision.ModelTypeFace))
		assert.Equal(t, vision.RunOnSchedule, c.FaceEngineRunType())
	})
	t.Run("VisionModelShouldRunFace", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace}}}
		c := NewConfig(CliTestContext())

		m := vision.Config.Model(vision.ModelTypeFace)
		require.NotNil(t, m)
		m.Run = vision.RunOnSchedule

		assert.True(t, c.VisionModelShouldRun(vision.ModelTypeFace, vision.RunOnSchedule))

		c.options.DisableFaces = true
		assert.False(t, c.VisionModelShouldRun(vision.ModelTypeFace, vision.RunOnSchedule))
		c.options.DisableFaces = false

		m.Disabled = true
		assert.False(t, c.VisionModelShouldRun(vision.ModelTypeFace, vision.RunOnSchedule))
	})
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

func TestConfig_FaceSize(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.SizeThreshold, c.FaceSize())
	c.options.FaceSize = 30
	assert.Equal(t, 30, c.FaceSize())
	c.options.FaceSize = 1
	assert.Equal(t, face.SizeThreshold, c.FaceSize())
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
	assert.Equal(t, face.OverlapThreshold, c.FaceOverlap())
	c.options.FaceOverlap = 300
	assert.Equal(t, face.OverlapThreshold, c.FaceOverlap())
	c.options.FaceOverlap = 1
	assert.Equal(t, 1, c.FaceOverlap())
}

func TestConfig_FaceClusterSize(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.ClusterSizeThreshold, c.FaceClusterSize())
	c.options.FaceClusterSize = 10
	assert.Equal(t, face.ClusterSizeThreshold, c.FaceClusterSize())
	c.options.FaceClusterSize = 66
	assert.Equal(t, 66, c.FaceClusterSize())
}

func TestConfig_FaceClusterScore(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.ClusterScoreThreshold, c.FaceClusterScore())
	c.options.FaceClusterScore = 0
	assert.Equal(t, face.ClusterScoreThreshold, c.FaceClusterScore())
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
	c.options.FaceCollisionDist = 0.05
	c.options.FaceClusterDist = 0.06
	assert.Equal(t, 0.06, c.FaceClusterDist())
	c.options.FaceClusterDist = 0.34
	assert.Equal(t, 0.34, c.FaceClusterDist())
}

func TestConfig_FaceClusterRadius(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.ClusterRadius, c.FaceClusterRadius())
	c.options.FaceClusterRadius = 0.01
	assert.Equal(t, face.ClusterRadius, c.FaceClusterRadius())
	c.options.FaceCollisionDist = 0.05
	c.options.FaceClusterRadius = 0.5
	assert.Equal(t, 0.5, c.FaceClusterRadius())
}

func TestConfig_FaceCollisionDist(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.CollisionDist, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 0.05
	assert.Equal(t, 0.05, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 0
	assert.Equal(t, face.CollisionDist, c.FaceCollisionDist())
}

func TestConfig_FaceEpsilonDist(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.Epsilon, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.02
	assert.Equal(t, 0.02, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.2
	assert.Equal(t, face.Epsilon, c.FaceEpsilonDist())
}

func TestConfig_FaceMatchDist(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.MatchDist, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.1
	assert.Equal(t, 0.1, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.01
	assert.Equal(t, face.MatchDist, c.FaceMatchDist())
}

func TestConfig_FaceMatchChildren(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.FaceMatchChildren())
	c.options.FaceMatchChildren = true
	assert.True(t, c.FaceMatchChildren())
}

func TestConfig_FaceMatchBackground(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.FaceMatchBackground())
	c.options.FaceMatchBackground = true
	assert.True(t, c.FaceMatchBackground())
}

func TestConfig_FaceAngles(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.DefaultAngles, c.FaceAngles())

	c.options.FaceAngles = []float64{-0.5, 0, 0.5}
	assert.Equal(t, []float64{-0.5, 0, 0.5}, c.FaceAngles())

	c.options.FaceAngles = []float64{math.Pi + 0.1, math.NaN(), 4}
	assert.Equal(t, face.DefaultAngles, c.FaceAngles())
}
