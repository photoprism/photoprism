package config

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
)

// installTestModels creates the files that make the named face embedding models appear
// installed and returns the temporary models path holding them.
func installTestModels(t *testing.T, names ...face.ModelName) string {
	t.Helper()

	modelsPath := t.TempDir()

	for _, name := range names {
		m := face.FindEmbeddingModel(name)
		require.NotNil(t, m)
		require.NoError(t, os.MkdirAll(filepath.Join(modelsPath, m.Dir), fs.ModeDir))

		if m.FileName != "" {
			require.NoError(t, os.WriteFile(m.FilePath(modelsPath), []byte("onnx"), fs.ModeFile))
		}
	}

	return modelsPath
}

func TestConfig_FaceEngine(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		engine := c.FaceEngine()
		assert.Contains(t, []string{face.EngineNone, face.EngineONNX}, engine)
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
		require.NoError(t, os.MkdirAll(modelDir, 0o750))
		modelFile := filepath.Join(modelDir, face.DefaultONNXModelFilename)
		require.NoError(t, os.WriteFile(modelFile, []byte("onnx"), 0o600))

		c.options.FaceEngine = face.EngineAuto
		assert.Equal(t, face.EngineONNX, c.FaceEngine())
	})
	t.Run("ExplicitEngine", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngine = face.EngineONNX
		assert.Equal(t, face.EngineONNX, c.FaceEngine())
		c.options.FaceEngine = face.EngineNone
		assert.Equal(t, face.EngineNone, c.FaceEngine())
	})
	t.Run("LegacyPigoAlias", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngine = "pigo"
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
	expected := max(runtime.NumCPU()/2, 1)
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

func TestConfig_FaceModel(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, face.ModelFaceNet, c.FaceModel())
	})
	t.Run("Explicit", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)
		c.options.FaceModel = "ArcFace-R50"

		assert.Equal(t, face.ModelArcFaceR50, c.FaceModel())
	})
	t.Run("ExplicitNotInstalled", func(t *testing.T) {
		// Falling back keeps provenance honest: the weights are missing, so embeddings
		// would be generated by the model that is actually installed.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelArcFaceR50

		assert.Equal(t, face.ModelFaceNet, c.FaceModel())
	})
	t.Run("ExplicitNotInstalledWithoutFallback", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, face.ModelNone, c.FaceModel())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone
		assert.Equal(t, face.ModelNone, c.FaceModel())
	})
	t.Run("Unsupported", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = "dlib"
		assert.Equal(t, face.ModelFaceNet, c.FaceModel())
	})
	t.Run("AutoWithoutModels", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = face.ModelAuto
		assert.Equal(t, face.ModelNone, c.FaceModel())
	})
	t.Run("AutoPrefersSFaceWhenFaceNetMissing", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelAuto

		assert.Equal(t, face.ModelSFace, c.FaceModel())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.ModelNone, (*Config)(nil).FaceModel())
	})
}

func TestConfig_FaceEmbeddingModel(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		m := c.FaceEmbeddingModel()
		require.NotNil(t, m)
		assert.Equal(t, face.ModelFaceNet, m.Name)
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone
		assert.Nil(t, c.FaceEmbeddingModel())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Nil(t, (*Config)(nil).FaceEmbeddingModel())
	})
}

func TestConfig_FaceModelPath(t *testing.T) {
	t.Run("SavedModelDir", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempModels := installTestModels(t, face.ModelFaceNet)
		c.options.ModelsPath = tempModels
		c.options.FaceModel = face.ModelFaceNet
		assert.Equal(t, filepath.Join(tempModels, "facenet"), c.FaceModelPath())
	})
	t.Run("ONNXFile", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempModels := installTestModels(t, face.ModelSFace)
		c.options.ModelsPath = tempModels
		c.options.FaceModel = face.ModelSFace
		assert.Equal(t, filepath.Join(tempModels, "sface", "face_recognition_sface_2021dec.onnx"), c.FaceModelPath())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone
		assert.Equal(t, "", c.FaceModelPath())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, "", (*Config)(nil).FaceModelPath())
	})
}

func TestConfig_FaceModelLicense(t *testing.T) {
	t.Run("SFace", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace
		assert.Equal(t, face.LicenseApache2, c.FaceModelLicense())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone
		assert.Equal(t, "", c.FaceModelLicense())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, "", (*Config)(nil).FaceModelLicense())
	})
}

func TestConfig_FaceModelDims(t *testing.T) {
	t.Run("FaceNet", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelFaceNet
		assert.Equal(t, 512, c.FaceModelDims())
	})
	t.Run("SFace", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace
		assert.Equal(t, 128, c.FaceModelDims())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone
		assert.Equal(t, 0, c.FaceModelDims())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, 0, (*Config)(nil).FaceModelDims())
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

func TestConfig_FaceThresholdsPerModel(t *testing.T) {
	t.Run("SFace", func(t *testing.T) {
		// Distances are not comparable across models, so the thresholds must follow the
		// configured model rather than the values FaceNet was tuned with.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, 0.91, c.FaceClusterDist())
		assert.Equal(t, 0.67, c.FaceClusterRadius())
		assert.Equal(t, 0.39, c.FaceMatchDist())
	})
	t.Run("FaceNetKeepsShippedValues", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelFaceNet

		assert.Equal(t, face.ClusterDistDefault, c.FaceClusterDist())
		assert.Equal(t, face.ClusterRadiusDefault, c.FaceClusterRadius())
		assert.Equal(t, face.MatchDistDefault, c.FaceMatchDist())
	})
	t.Run("NoModel", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Equal(t, face.ClusterDistDefault, c.FaceClusterDist())
		assert.Equal(t, face.ClusterRadiusDefault, c.FaceClusterRadius())
		assert.Equal(t, face.MatchDistDefault, c.FaceMatchDist())
	})
	t.Run("ExplicitOptionWins", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace
		c.options.FaceClusterDist = 0.7
		c.options.FaceClusterRadius = 0.5
		c.options.FaceMatchDist = 0.3

		assert.Equal(t, 0.7, c.FaceClusterDist())
		assert.Equal(t, 0.5, c.FaceClusterRadius())
		assert.Equal(t, 0.3, c.FaceMatchDist())
	})
	t.Run("CliDefaultsDoNotWin", func(t *testing.T) {
		// The CLI flags carry the FaceNet defaults so "--help" documents them, and those
		// defaults reach the options even when the operator sets nothing.
		ctx := cliContextWithFlagDefaults(t)
		c := &Config{cliCtx: ctx, options: NewOptions(ctx)}
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		require.Equal(t, face.ClusterDistDefault, c.options.FaceClusterDist)
		assert.Equal(t, 0.91, c.FaceClusterDist())
		assert.Equal(t, 0.67, c.FaceClusterRadius())
		assert.Equal(t, 0.39, c.FaceMatchDist())
	})
}

// cliContextWithFlagDefaults returns a CLI context that carries the default values of the
// registered flags, which is what the app does for every option the operator leaves unset.
func cliContextWithFlagDefaults(t *testing.T) *cli.Context {
	t.Helper()

	set := flag.NewFlagSet("test", flag.ContinueOnError)

	for _, f := range Flags.Cli() {
		require.NoError(t, f.Apply(set))
	}

	return cli.NewContext(cli.NewApp(), set, nil)
}

func TestConfig_FaceThresholdIsSet(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Default", func(t *testing.T) {
		assert.False(t, c.faceThresholdIsSet("face-cluster-dist", face.ClusterDistDefault, face.ClusterDistDefault))
	})
	t.Run("CustomValue", func(t *testing.T) {
		assert.True(t, c.faceThresholdIsSet("face-cluster-dist", 0.5, face.ClusterDistDefault))
	})
	t.Run("FlagSetToDefault", func(t *testing.T) {
		set := flag.NewFlagSet("test", flag.ContinueOnError)
		set.Float64("face-cluster-dist", face.ClusterDistDefault, "doc")
		assert.NoError(t, set.Parse([]string{"--face-cluster-dist", "0.64"}))

		explicit := &Config{cliCtx: cli.NewContext(cli.NewApp(), set, nil), options: NewOptions(nil)}

		assert.True(t, explicit.faceThresholdIsSet("face-cluster-dist", face.ClusterDistDefault, face.ClusterDistDefault))
	})
	t.Run("NoContext", func(t *testing.T) {
		none := &Config{options: NewOptions(nil)}

		assert.False(t, none.faceThresholdIsSet("face-cluster-dist", face.ClusterDistDefault, face.ClusterDistDefault))
	})
}

func TestConfig_FaceThreshold(t *testing.T) {
	pick := func(m *face.EmbeddingModel) float64 { return m.MatchDist }

	t.Run("OutOfRange", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, 0.39, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))
		assert.Equal(t, 0.39, c.faceThreshold("face-match-dist", 0.001, face.MatchDistDefault, pick))
	})
	t.Run("InRange", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, 0.5, c.faceThreshold("face-match-dist", 0.5, face.MatchDistDefault, pick))
	})
}

func TestFaceModelThreshold(t *testing.T) {
	pick := func(m *face.EmbeddingModel) float64 { return m.MatchDist }

	t.Run("Model", func(t *testing.T) {
		assert.Equal(t, 0.39, faceModelThreshold(face.FindEmbeddingModel(face.ModelSFace), pick, 0.4))
	})
	t.Run("NilModel", func(t *testing.T) {
		assert.Equal(t, 0.4, faceModelThreshold(nil, pick, 0.4))
	})
	t.Run("Uncalibrated", func(t *testing.T) {
		assert.Equal(t, 0.4, faceModelThreshold(&face.EmbeddingModel{Name: "test"}, pick, 0.4))
	})
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

func TestConfig_FaceSkipChildren(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.FaceSkipChildren())
	c.options.FaceSkipChildren = true
	assert.True(t, c.FaceSkipChildren())
}

func TestConfig_FaceAllowBackground(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.FaceAllowBackground())
	c.options.FaceAllowBackground = true
	assert.True(t, c.FaceAllowBackground())
}
