package config

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// captureLog records what the shared logger emits for the duration of a test.
func captureLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)

	hook := test.NewLocal(logger)
	t.Cleanup(hook.Reset)

	return hook
}

// installTestModels creates the files that make the named face embedding models appear
// installed and returns the temporary models path holding them.
func installTestModels(t *testing.T, names ...face.ModelName) string {
	t.Helper()

	modelsPath := t.TempDir()

	for _, name := range names {
		m := face.FindEmbeddingModel(name)
		require.NotNil(t, m)
		require.NoError(t, os.MkdirAll(filepath.Join(modelsPath, m.Dir), fs.ModeDir))

		if m.ONNX != nil {
			require.NoError(t, os.WriteFile(m.FilePath(modelsPath), []byte("onnx"), fs.ModeFile))
		}
	}

	return modelsPath
}

// newSFaceTestConfig returns a test config with SFace installed and in force, which is what
// the model-calibrated thresholds follow.
func newSFaceTestConfig(t *testing.T) *Config {
	t.Helper()

	c := NewConfig(CliTestContext())
	c.options.ModelsPath = installTestModels(t, face.ModelSFace)
	c.options.FaceModel = face.ModelSFace

	return c
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
	t.Run("SharedWithIndexWorkers", func(t *testing.T) {
		// Detection takes no lock, so one pool of this size runs per indexing worker.
		c := NewConfig(CliTestContext())
		expected := max(runtime.NumCPU()/max(c.IndexWorkers(), 1), 1)
		assert.Equal(t, expected, c.FaceEngineThreads())
		assert.LessOrEqual(t, c.FaceEngineThreads()*c.IndexWorkers(), max(runtime.NumCPU(), c.IndexWorkers()))
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 8
		assert.Equal(t, 8, c.FaceEngineThreads())
	})
	t.Run("KeepsOptionUnset", func(t *testing.T) {
		// Writing the derived value back would freeze it for FaceModelThreads as well.
		c := NewConfig(CliTestContext())
		c.FaceEngineThreads()
		assert.LessOrEqual(t, c.options.FaceEngineThreads, 0)
	})
}

func TestConfig_FaceModelThreads(t *testing.T) {
	t.Run("HalfTheCores", func(t *testing.T) {
		// Embeddings are generated behind the model session lock, so this count is not
		// divided among the indexing workers the way detection is.
		c := NewConfig(CliTestContext())
		assert.Equal(t, max(runtime.NumCPU()/2, 1), c.FaceModelThreads())
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 8
		assert.Equal(t, 8, c.FaceModelThreads())
	})
	t.Run("IndependentOfDatabaseDriver", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		originalDriver := c.options.DatabaseDriver
		t.Cleanup(func() { c.options.DatabaseDriver = originalDriver })

		c.options.DatabaseDriver = dsn.DriverMySQL
		mysqlThreads := c.FaceModelThreads()
		c.options.DatabaseDriver = dsn.DriverSQLite3
		assert.Equal(t, mysqlThreads, c.FaceModelThreads())
	})
	t.Run("Nil", func(t *testing.T) {
		var c *Config
		assert.Equal(t, 1, c.FaceModelThreads())
	})
}

func TestConfig_faceEngineRunsOnIndex(t *testing.T) {
	t.Run("ManyCores", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 4
		assert.True(t, c.faceEngineRunsOnIndex())
	})
	t.Run("FewCores", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 2
		assert.False(t, c.faceEngineRunsOnIndex())
	})
	t.Run("SurvivesIndexWorkerCount", func(t *testing.T) {
		// Dividing by the indexing workers yields exactly 2 on the automatic MySQL path,
		// which would move detection off the index for every such host.
		c := NewConfig(CliTestContext())
		originalDriver := c.options.DatabaseDriver
		t.Cleanup(func() { c.options.DatabaseDriver = originalDriver })

		c.options.DatabaseDriver = dsn.DriverMySQL

		if runtime.NumCPU() < 6 {
			t.Skip("needs at least six cores")
		}

		assert.True(t, c.faceEngineRunsOnIndex())
	})
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

func TestConfig_FaceModelSetting(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""

		assert.Equal(t, face.ModelDetect, c.FaceModelSetting())
	})
	t.Run("Detect", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = "Detect"

		assert.Equal(t, face.ModelDetect, c.FaceModelSetting())
	})
	t.Run("AutoIsDetect", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelAuto

		assert.Equal(t, face.ModelDetect, c.FaceModelSetting())
	})
	t.Run("Named", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = "SFace"

		assert.Equal(t, face.ModelSFace, c.FaceModelSetting())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Equal(t, face.ModelNone, c.FaceModelSetting())
	})
	t.Run("Unsupported", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = "sfase"

		assert.Equal(t, face.ModelDetect, c.FaceModelSetting())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.ModelNone, (*Config)(nil).FaceModelSetting())
	})
}

func TestConfig_FaceModel(t *testing.T) {
	t.Run("UnsetIsNotResolvedHere", func(t *testing.T) {
		// Detection runs once in Init, so a configuration that has not reached it names no
		// model rather than deriving one that the instance may never use.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""

		assert.Equal(t, face.ModelNone, c.FaceModel())
	})
	t.Run("Detected", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.faceModel = face.ModelSFace

		assert.Equal(t, face.ModelSFace, c.FaceModel())
	})
	t.Run("Explicit", func(t *testing.T) {
		t.Setenv(face.LicenseAcceptanceVar, "1")
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)
		c.options.FaceModel = "ArcFace-R50"

		assert.Equal(t, face.ModelArcFaceR50, c.FaceModel())
	})
	t.Run("ExplicitNotInstalled", func(t *testing.T) {
		// Resolving to whatever else is installed would start a second vector space the
		// library cannot compare with, so a missing explicit model disables embeddings.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, face.ModelNone, c.FaceModel())
	})
	t.Run("ExplicitWithoutLicenseAcceptance", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)
		c.options.FaceModel = face.ModelArcFaceR50

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

		assert.Equal(t, face.ModelNone, c.FaceModel())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.ModelNone, (*Config)(nil).FaceModel())
	})
}

func TestConfig_usableFaceModel(t *testing.T) {
	t.Run("Installed", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)

		assert.Equal(t, face.ModelSFace, c.usableFaceModel(face.ModelSFace))
	})
	t.Run("NotInstalled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()

		assert.Equal(t, face.ModelNone, c.usableFaceModel(face.ModelSFace))
	})
	t.Run("LicenseNotAccepted", func(t *testing.T) {
		// Weights that are installed are still not usable until their terms are accepted.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)

		assert.Equal(t, face.ModelNone, c.usableFaceModel(face.ModelArcFaceR50))
	})
	t.Run("IneligibleEdition", func(t *testing.T) {
		t.Setenv(face.LicenseAcceptanceVar, "1")
		c := NewConfig(CliTestContext())
		c.options.Edition = "pro"
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)

		assert.Equal(t, face.ModelNone, c.usableFaceModel(face.ModelArcFaceR50))
	})
}

func TestConfig_detectFaceModel(t *testing.T) {
	t.Run("LibraryKeepsItsModel", func(t *testing.T) {
		// Vectors written before the provenance column can only be FaceNet, and resolving
		// them to the first installed model would leave every stored cluster incomparable.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet, face.ModelSFace)

		counts := []query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: 12}}

		assert.Equal(t, face.ModelFaceNet, c.detectFaceModel(counts))
	})
	t.Run("LibraryModelIsNotFiltered", func(t *testing.T) {
		// A library the operator deliberately built on gated weights still resolves to them,
		// because a model that cannot read its vectors is no substitute.
		t.Setenv(face.LicenseAcceptanceVar, "1")
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace, face.ModelArcFaceR50)

		counts := []query.MarkerEmbeddingModelCount{{EmbedModel: face.ModelArcFaceR50, Markers: 7}}

		assert.Equal(t, face.ModelArcFaceR50, c.detectFaceModel(counts))
	})
	t.Run("EmptyLibraryUsesPreference", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet, face.ModelSFace)

		assert.Equal(t, face.ModelSFace, c.detectFaceModel(nil))
	})
}

func TestConfig_installedFaceModel(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()

		assert.Equal(t, face.ModelNone, c.installedFaceModel())
	})
	t.Run("FollowsPreferenceOrder", func(t *testing.T) {
		// FaceNet is installed too, but SFace comes first for a library with no vectors.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet, face.ModelSFace)

		assert.Equal(t, face.ModelSFace, c.installedFaceModel())
	})
	t.Run("SkipsGatedWeights", func(t *testing.T) {
		// Nothing has been chosen yet, so selecting weights whose terms nobody accepted is
		// exactly what the gate exists to prevent - even when they are the only ones there.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)

		assert.Equal(t, face.ModelNone, c.installedFaceModel())
	})
	t.Run("AcceptedGatedWeightsStayLast", func(t *testing.T) {
		t.Setenv(face.LicenseAcceptanceVar, "1")
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)

		assert.Equal(t, face.ModelArcFaceR50, c.installedFaceModel())
	})
}

func TestConfig_checkFaceModelMismatch(t *testing.T) {
	t.Cleanup(face.UnblockEmbeddings)

	t.Run("Comparable", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.faceModel = face.ModelSFace
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)

		c.checkFaceModelMismatch([]query.MarkerEmbeddingModelCount{{EmbedModel: face.ModelSFace, Markers: 9}})

		assert.False(t, face.EmbeddingsBlocked())
	})
	t.Run("LegacyVectorsAreFaceNet", func(t *testing.T) {
		// A blank name is a vector written before the provenance column, which FaceNet can
		// read by definition, so it is not a mismatch against it.
		c := NewConfig(CliTestContext())
		c.faceModel = face.ModelFaceNet
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)

		c.checkFaceModelMismatch([]query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: 9}})

		assert.False(t, face.EmbeddingsBlocked())
	})
	t.Run("Incomparable", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.faceModel = face.ModelSFace
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)

		c.checkFaceModelMismatch([]query.MarkerEmbeddingModelCount{
			{EmbedModel: "", Markers: 1031},
			{EmbedModel: face.ModelSFace, Markers: 4},
		})

		require.True(t, face.EmbeddingsBlocked())
		assert.Contains(t, face.EmbeddingsBlockedReason(), "1031 marker(s) use facenet")
		assert.Contains(t, face.EmbeddingsBlockedReason(), "configured for sface")
	})
	t.Run("EmbeddingsDisabled", func(t *testing.T) {
		// Nothing is generated when embeddings are turned off, and the vectors a library
		// already holds stay comparable with each other, so there is nothing to block.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		c.checkFaceModelMismatch([]query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: 12}})

		assert.False(t, face.EmbeddingsBlocked())
	})
	t.Run("UnusableModelBlocks", func(t *testing.T) {
		// A model that could not be loaded reads none of the stored vectors, so clustering
		// them would rewrite the library at another model's distances.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = face.ModelSFace

		require.Equal(t, face.ModelNone, c.FaceModel())

		c.checkFaceModelMismatch([]query.MarkerEmbeddingModelCount{{EmbedModel: face.ModelSFace, Markers: 12}})

		require.True(t, face.EmbeddingsBlocked())
		assert.Contains(t, face.EmbeddingsBlockedReason(), "cannot load")
	})
	t.Run("UnusableModelWithoutVectors", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = face.ModelSFace

		c.checkFaceModelMismatch(nil)

		assert.False(t, face.EmbeddingsBlocked())
	})
	t.Run("NamesEachModelOnce", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.faceModel = face.ModelSFace
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)

		c.checkFaceModelMismatch([]query.MarkerEmbeddingModelCount{
			{EmbedModel: "", Markers: 30},
			{EmbedModel: face.ModelFaceNet, Markers: 12},
		})

		require.True(t, face.EmbeddingsBlocked())
		assert.Contains(t, face.EmbeddingsBlockedReason(), "42 marker(s) use facenet,")
	})
}

func TestStaleFaceModels(t *testing.T) {
	t.Run("Comparable", func(t *testing.T) {
		stale, models := staleFaceModels([]query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: 9}}, face.ModelFaceNet)

		assert.Equal(t, 0, stale)
		assert.Empty(t, models)
	})
	t.Run("Incomparable", func(t *testing.T) {
		stale, models := staleFaceModels([]query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: 9}}, face.ModelSFace)

		assert.Equal(t, 9, stale)
		assert.Equal(t, []string{face.ModelFaceNet}, models)
	})
	t.Run("NoModelReadsNothing", func(t *testing.T) {
		counts := []query.MarkerEmbeddingModelCount{
			{EmbedModel: face.ModelSFace, Markers: 4},
			{EmbedModel: face.ModelFaceNet, Markers: 2},
		}

		stale, models := staleFaceModels(counts, face.ModelNone)

		assert.Equal(t, 6, stale)
		assert.Equal(t, []string{face.ModelSFace, face.ModelFaceNet}, models)
	})
	t.Run("EmptyCountsAreSkipped", func(t *testing.T) {
		stale, models := staleFaceModels([]query.MarkerEmbeddingModelCount{{EmbedModel: face.ModelSFace, Markers: 0}}, face.ModelFaceNet)

		assert.Equal(t, 0, stale)
		assert.Empty(t, models)
	})
}

func TestConfig_reportIgnoredFaceModel(t *testing.T) {
	// The message is all this does, so each case has to read it.
	ignoredLines := func(t *testing.T, hook *test.Hook) []string {
		t.Helper()

		var lines []string

		for _, e := range hook.AllEntries() {
			if strings.Contains(e.Message, "is configured but ignored") {
				lines = append(lines, e.Message)
			}
		}

		return lines
	}

	t.Run("FileWins", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		hook := captureLog(t)
		c.faceModelFlag = face.ModelSFace
		c.options.FaceModel = face.ModelFaceNet

		c.reportIgnoredFaceModel()

		lines := ignoredLines(t, hook)

		require.Len(t, lines, 1)
		assert.Contains(t, lines[0], "face model sface is configured but ignored")
		assert.Contains(t, lines[0], "this library uses facenet")
		assert.Contains(t, lines[0], "faces migrate --to sface")
	})
	t.Run("DisabledPointsAtTheFile", func(t *testing.T) {
		// A migration cannot turn embeddings off, so naming it as the way to apply this
		// value would send the operator to a command that refuses.
		c := NewConfig(CliTestContext())
		hook := captureLog(t)
		c.faceModelFlag = face.ModelNone
		c.options.FaceModel = face.ModelSFace

		c.reportIgnoredFaceModel()

		lines := ignoredLines(t, hook)

		require.Len(t, lines, 1)
		assert.NotContains(t, lines[0], "faces migrate")
		assert.Contains(t, lines[0], "options.yml")
	})
	t.Run("NothingConfigured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		hook := captureLog(t)
		c.faceModelFlag = ""
		c.options.FaceModel = face.ModelFaceNet

		c.reportIgnoredFaceModel()

		assert.Empty(t, ignoredLines(t, hook))
	})
	t.Run("Detect", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		hook := captureLog(t)
		c.faceModelFlag = face.ModelDetect
		c.options.FaceModel = face.ModelSFace

		c.reportIgnoredFaceModel()

		assert.Empty(t, ignoredLines(t, hook))
	})
	t.Run("Agree", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		hook := captureLog(t)
		c.faceModelFlag = face.ModelSFace
		c.options.FaceModel = face.ModelSFace

		c.reportIgnoredFaceModel()

		assert.Empty(t, ignoredLines(t, hook))
	})
}

func TestConfig_libraryFaceModels(t *testing.T) {
	t.Run("SeededLibrary", func(t *testing.T) {
		counts, ok := TestConfig().libraryFaceModels()

		assert.True(t, ok)
		assert.NotEmpty(t, counts)
	})
	t.Run("CountsLegacyVectors", func(t *testing.T) {
		// Leaving the rows that predate the column out would let a handful of migrated
		// markers outvote a whole legacy library, and hide it from the mismatch check.
		c := TestConfig()
		uids := seedLegacyMarkers(t)

		counts, ok := c.libraryFaceModels()
		require.True(t, ok)

		legacy := 0
		for _, count := range counts {
			if count.EmbedModel == "" {
				legacy = count.Markers
			}
		}

		assert.Equal(t, len(uids), legacy)
		assert.Equal(t, face.ModelFaceNet, dominantFaceModel(counts))
	})
	t.Run("WithoutDatabase", func(t *testing.T) {
		// Not being able to ask is not the same answer as an empty library.
		counts, ok := NewConfig(CliTestContext()).libraryFaceModels()

		assert.Nil(t, counts)
		assert.False(t, ok)
	})
	t.Run("NilConfig", func(t *testing.T) {
		counts, ok := (*Config)(nil).libraryFaceModels()

		assert.Nil(t, counts)
		assert.False(t, ok)
	})
}

// seedEmptyLibrary clears the stored vectors of every face marker for the duration of the test,
// which is what a library that has been indexed but never embedded looks like.
func seedEmptyLibrary(t *testing.T) {
	t.Helper()

	db := entity.Db()

	var uids []string
	require.NoError(t, db.Model(&entity.Marker{}).
		Where("marker_type = ? AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
		Pluck("marker_uid", &uids).Error)
	require.NotEmpty(t, uids, "the fixtures must hold face vectors for this to mean anything")

	values := make(map[string][]byte, len(uids))

	for _, uid := range uids {
		m := entity.Marker{}
		require.NoError(t, db.Where("marker_uid = ?", uid).First(&m).Error)
		values[uid] = m.EmbeddingsJSON
	}

	// Both columns, because a vector and its provenance are only ever written together.
	model := entity.MarkerFixtures.Get("1000003-4").EmbedModel

	require.NoError(t, db.Model(&entity.Marker{}).Where("marker_uid IN (?)", uids).
		UpdateColumns(entity.Values{"embeddings_json": []byte(""), "embed_model": ""}).Error)

	t.Cleanup(func() {
		for uid, embeddings := range values {
			assert.NoError(t, db.Model(&entity.Marker{}).Where("marker_uid = ?", uid).
				UpdateColumns(entity.Values{"embeddings_json": embeddings, "embed_model": model}).Error)
		}
	})
}

// seedLegacyMarkers gives the fixture library a majority of markers that record no model, which
// is what a library indexed before the provenance column looks like after a partial migration.
func seedLegacyMarkers(t *testing.T) []string {
	t.Helper()

	db := entity.Db()
	uids := make([]string, 0, 24)

	for i := 0; i < 24; i++ {
		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			EmbedModel:     "",
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
			W:              0.1,
			H:              0.1,
		}

		require.NoError(t, db.Create(m).Error)
		uids = append(uids, m.MarkerUID)
	}

	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Where("marker_uid IN (?)", uids).Delete(&entity.Marker{}).Error)
	})

	return uids
}

func TestConfig_SetFaceModel(t *testing.T) {
	t.Run("Persisted", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		require.NoError(t, c.SetFaceModel(face.ModelSFace))

		assert.Equal(t, face.ModelSFace, c.faceModel)

		values := Values{}
		b, err := os.ReadFile(c.OptionsYaml())
		require.NoError(t, err)
		require.NoError(t, yaml.Unmarshal(b, &values))
		assert.Equal(t, face.ModelSFace, values["FaceModel"])
	})
	t.Run("ReportsAFailedWrite", func(t *testing.T) {
		// A caller that changed the data has to be able to tell that the setting did not
		// follow it, or it reports a migration as done that the next start refuses.
		c := NewMinimalTestConfig(t.TempDir())
		require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte("FaceModel: [\n"), fs.ModeFile))

		err := c.SetFaceModel(face.ModelSFace)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed saving face model")
	})
	t.Run("Empty", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		require.NoError(t, c.SetFaceModel(""))

		assert.Equal(t, "", c.faceModel)
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.NoError(t, (*Config)(nil).SetFaceModel(face.ModelSFace))
	})
}

func TestConfig_ResolveFaceModel(t *testing.T) {
	t.Run("WithoutDatabase", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""

		assert.Equal(t, face.ModelNone, c.ResolveFaceModel())
	})
	t.Run("DoesNotPersist", func(t *testing.T) {
		// A report states what is configured, so resolving for display must leave the
		// setting untouched.
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		c.options.FaceModel = face.ModelDetect
		c.faceModel = ""

		assert.Equal(t, entity.MarkerFixtures.Get("1000003-4").EmbedModel, c.ResolveFaceModel())
		assert.Equal(t, face.ModelDetect, c.FaceModelSetting())
	})
	t.Run("ReportsAMismatch", func(t *testing.T) {
		// An operator who saw the warning at startup looks here next, so the report has to
		// answer with the same verdict rather than "ok".
		t.Cleanup(face.UnblockEmbeddings)
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		c.options.FaceModel = otherLibraryFaceModel(t, c)
		c.faceModel = ""

		c.ResolveFaceModel()

		assert.True(t, face.EmbeddingsBlocked())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.ModelNone, (*Config)(nil).ResolveFaceModel())
	})
}

// otherLibraryFaceModel returns an installed model whose vectors the library does not hold,
// which is the mismatch an instance pointed at the wrong library sees.
func otherLibraryFaceModel(t *testing.T, c *Config) face.ModelName {
	t.Helper()

	held := c.libraryFaceModel()

	for _, name := range []face.ModelName{face.ModelSFace, face.ModelFaceNet} {
		if name != held && face.FindEmbeddingModel(name).Installed(c.ModelsPath()) {
			return name
		}
	}

	t.Skip("faces: no other embedding model is installed")

	return ""
}

func TestConfig_initFaceModel(t *testing.T) {
	t.Run("DetectsAndPersists", func(t *testing.T) {
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		c.options.FaceModel = face.ModelDetect
		c.faceModel = ""

		c.initFaceModel()

		expected := entity.MarkerFixtures.Get("1000003-4").EmbedModel
		assert.Equal(t, expected, c.faceModel)
		assert.Equal(t, expected, c.options.FaceModel)
	})
	t.Run("NamedModelIsAnAnswer", func(t *testing.T) {
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		c.options.FaceModel = face.ModelNone
		c.faceModel = ""

		c.initFaceModel()

		assert.Equal(t, face.ModelNone, c.options.FaceModel)
		assert.Equal(t, "", c.faceModel)
	})
	t.Run("UnsupportedValueIsNotPinned", func(t *testing.T) {
		// A typo applies to the run as if nothing were set - turning face embeddings off
		// instead would be a silent cost for a value the operator can fix in a second - but
		// writing the detected name down would outlive the typo.
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		c.options.FaceModel = "sfase"
		c.faceModel = ""

		c.initFaceModel()

		assert.Equal(t, "sfase", c.options.FaceModel)
		assert.Equal(t, entity.MarkerFixtures.Get("1000003-4").EmbedModel, c.faceModel)
	})
	t.Run("UnreadableLibraryIsNotPinned", func(t *testing.T) {
		// A database that could not be asked is not an empty library. Pinning the preference
		// list there survives the outage and refuses the library once it comes back.
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		c.options.FaceModel = ""
		c.faceModel = ""

		db := entity.Db()
		table := entity.Marker{}.TableName()
		require.NoError(t, db.Exec("ALTER TABLE "+table+" RENAME TO "+table+"_hidden").Error)

		t.Cleanup(func() {
			require.NoError(t, db.Exec("ALTER TABLE "+table+"_hidden RENAME TO "+table).Error)
		})

		c.initFaceModel()

		assert.Equal(t, "", c.options.FaceModel)
		assert.False(t, face.EmbeddingsBlocked())
	})
	t.Run("EmptyLibraryIsNotPinned", func(t *testing.T) {
		// There is nothing to learn from a library with no vectors, so the preference list
		// answers for the run and the name is recorded once the faces exist. A restore into
		// a fresh instance would otherwise be refused by a model nothing in it produced.
		c := TestConfig()
		defer restoreFaceModel(t, c)()
		seedEmptyLibrary(t)
		c.options.FaceModel = ""
		c.faceModel = ""

		c.initFaceModel()

		assert.Equal(t, "", c.options.FaceModel)
		assert.Equal(t, face.ModelSFace, c.faceModel)
	})
	t.Run("WithoutDatabase", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""

		c.initFaceModel()

		assert.Equal(t, "", c.options.FaceModel)
	})
}

func TestConfig_ConfigureFaceEmbedder(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		require.NoError(t, c.ConfigureFaceEmbedder(face.ModelNone))
		assert.Equal(t, face.ModelNone, face.ConfiguredModel())

		t.Cleanup(func() { _ = c.ConfigureFaceEmbedder(TestConfig().FaceModel()) })
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.NoError(t, (*Config)(nil).ConfigureFaceEmbedder(face.ModelSFace))
	})
}

func TestRecordedFaceModel(t *testing.T) {
	t.Run("Blank", func(t *testing.T) {
		assert.Equal(t, face.ModelFaceNet, recordedFaceModel(""))
	})
	t.Run("Named", func(t *testing.T) {
		assert.Equal(t, face.ModelSFace, recordedFaceModel("SFace"))
	})
}

// restoreFaceModel resets the face model of the shared test config when the test ends, since
// the tests that need a detectable library have to use the fixture-backed singleton.
func restoreFaceModel(t *testing.T, c *Config) func() {
	t.Helper()

	setting, resolved := c.options.FaceModel, c.faceModel

	return func() {
		c.options.FaceModel, c.faceModel = setting, resolved
	}
}

func TestDominantFaceModel(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", dominantFaceModel(nil))
	})
	t.Run("NoVectors", func(t *testing.T) {
		assert.Equal(t, "", dominantFaceModel([]query.MarkerEmbeddingModelCount{{EmbedModel: face.ModelSFace, Markers: 0}}))
	})
	t.Run("LegacyIsFaceNet", func(t *testing.T) {
		// The query only counts markers that hold a vector, so a blank name is a vector
		// written before the provenance column existed rather than a missing embedding.
		assert.Equal(t, face.ModelFaceNet, dominantFaceModel([]query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: 12}}))
	})
	t.Run("SingleModel", func(t *testing.T) {
		assert.Equal(t, face.ModelSFace, dominantFaceModel([]query.MarkerEmbeddingModelCount{{EmbedModel: "SFace", Markers: 3}}))
	})
	t.Run("MixedPrefersMajority", func(t *testing.T) {
		counts := []query.MarkerEmbeddingModelCount{
			{EmbedModel: face.ModelSFace, Markers: 4},
			{EmbedModel: "", Markers: 91},
		}
		assert.Equal(t, face.ModelFaceNet, dominantFaceModel(counts))
	})
	t.Run("MixedPrefersMajorityModel", func(t *testing.T) {
		counts := []query.MarkerEmbeddingModelCount{
			{EmbedModel: "", Markers: 4},
			{EmbedModel: "ArcFace-R50", Markers: 91},
		}
		assert.Equal(t, face.ModelArcFaceR50, dominantFaceModel(counts))
	})
}

// seedLegacyLibrary blanks the recorded embedding model of every marker that has one, which
// is what a library indexed before the provenance column looks like, and restores them when
// the test ends. The fixtures record the configured model, so a test that needs the legacy
// case has to create it.
func seedLegacyLibrary(t *testing.T) {
	t.Helper()

	db := entity.Db()

	var uids []string
	require.NoError(t, db.Model(&entity.Marker{}).Where("embed_model <> ''").Pluck("marker_uid", &uids).Error)
	require.NotEmpty(t, uids, "the fixtures must record an embedding model for this to mean anything")

	model := entity.MarkerFixtures.Get("1000003-4").EmbedModel

	require.NoError(t, db.Model(&entity.Marker{}).Where("marker_uid IN (?)", uids).
		UpdateColumn("embed_model", "").Error)

	t.Cleanup(func() {
		assert.NoError(t, db.Model(&entity.Marker{}).Where("marker_uid IN (?)", uids).
			UpdateColumn("embed_model", model).Error)
	})
}

func TestConfig_LibraryFaceModel(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, "", (*Config)(nil).libraryFaceModel())
	})
	t.Run("WithoutDatabase", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, "", c.libraryFaceModel())
	})
	t.Run("SeededLibrary", func(t *testing.T) {
		// The fixtures are generated for the configured model and record it, so the library
		// reports back what it was seeded with.
		assert.Equal(t, entity.MarkerFixtures.Get("1000003-4").EmbedModel, TestConfig().libraryFaceModel())
	})
	t.Run("LegacyLibrary", func(t *testing.T) {
		// A marker that records no model can only hold a FaceNet vector, because FaceNet was
		// the only bundled model before the column existed.
		c := TestConfig()
		seedLegacyLibrary(t)

		assert.Equal(t, face.ModelFaceNet, c.libraryFaceModel())
	})
	t.Run("SchemaWithoutProvenanceColumn", func(t *testing.T) {
		// The schema is migrated after the configuration is propagated, so the first start
		// after an upgrade reads a markers table that still has no embed_model column. If
		// that resolved to the preference list instead of FaceNet, every existing library
		// would quietly start writing vectors into a second, incomparable space.
		c := TestConfig()
		db := entity.Db()
		table := entity.Marker{}.TableName()

		// The index has to go first: a column an index references cannot be dropped, and a
		// schema that predates the column had neither. DropIndex keeps this portable
		// across the drivers the suite runs on.
		require.NoError(t, db.Migrator().DropIndex(&entity.Marker{}, "EmbedModel"))
		require.NoError(t, db.Exec("ALTER TABLE "+table+" DROP COLUMN embed_model").Error)

		t.Cleanup(func() {
			require.NoError(t, db.Migrator().AddColumn(&entity.Marker{}, "EmbedModel"))

			// The column comes back empty, so what the fixtures recorded has to be written
			// again: a later test that asks the library which model it holds would otherwise
			// be answered by this one.
			require.NoError(t, db.Model(&entity.Marker{}).
				Where("marker_type = ? AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
				UpdateColumn("embed_model", entity.MarkerFixtures.Get("1000003-4").EmbedModel).Error)

			require.NoError(t, db.Migrator().CreateIndex(&entity.Marker{}, "EmbedModel"))
		})

		t.Log("Expect column embed_model missing error for markers")
		assert.Equal(t, face.ModelFaceNet, c.libraryFaceModel())
	})
}

func TestConfig_FaceEmbeddingModel(t *testing.T) {
	t.Run("InForce", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		m := c.FaceEmbeddingModel()
		require.NotNil(t, m)
		assert.Equal(t, face.ModelSFace, m.Name)
	})
	t.Run("NotDetectedYet", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""
		assert.Nil(t, c.FaceEmbeddingModel())
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
	// Thresholds follow the model in force, so the test has to put one there.
	sface := face.FindEmbeddingModel(face.ModelSFace).ClusterDist

	c := newSFaceTestConfig(t)
	assert.Equal(t, sface, c.FaceClusterDist())
	c.options.FaceClusterDist = 0.01
	assert.Equal(t, sface, c.FaceClusterDist())
	// The collision distance is the lower bound, and it follows the model too, so it has
	// to be lowered explicitly before a value this small is accepted.
	c.options.FaceCollisionDist = 0.04
	c.options.FaceClusterDist = 0.06
	assert.Equal(t, 0.06, c.FaceClusterDist())
	c.options.FaceClusterDist = 0.34
	assert.Equal(t, 0.34, c.FaceClusterDist())
}

func TestConfig_FaceClusterRadius(t *testing.T) {
	// Thresholds follow the model in force, so the test has to put one there.
	sface := face.FindEmbeddingModel(face.ModelSFace).ClusterRadius

	c := newSFaceTestConfig(t)
	assert.Equal(t, sface, c.FaceClusterRadius())
	c.options.FaceClusterRadius = 0.01
	assert.Equal(t, sface, c.FaceClusterRadius())
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

		assert.Equal(t, 0.78, c.FaceClusterDist())
		assert.Equal(t, 0.60, c.FaceClusterRadius())
		assert.Equal(t, 0.35, c.FaceMatchDist())
		assert.Equal(t, 0.061, c.FaceCollisionDist())
		assert.Equal(t, 0.012, c.FaceEpsilonDist())
	})
	t.Run("FaceNetKeepsShippedValues", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelFaceNet

		assert.Equal(t, face.ClusterDistDefault, c.FaceClusterDist())
		assert.Equal(t, face.ClusterRadiusDefault, c.FaceClusterRadius())
		assert.Equal(t, face.MatchDistDefault, c.FaceMatchDist())
		assert.Equal(t, face.CollisionDistDefault, c.FaceCollisionDist())
		assert.Equal(t, face.EpsilonDefault, c.FaceEpsilonDist())
	})
	t.Run("NoModel", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Equal(t, face.ClusterDistDefault, c.FaceClusterDist())
		assert.Equal(t, face.ClusterRadiusDefault, c.FaceClusterRadius())
		assert.Equal(t, face.MatchDistDefault, c.FaceMatchDist())
		assert.Equal(t, face.CollisionDistDefault, c.FaceCollisionDist())
		assert.Equal(t, face.EpsilonDefault, c.FaceEpsilonDist())
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
		require.Equal(t, face.CollisionDistDefault, c.options.FaceCollisionDist)
		assert.Equal(t, 0.78, c.FaceClusterDist())
		assert.Equal(t, 0.60, c.FaceClusterRadius())
		assert.Equal(t, 0.35, c.FaceMatchDist())
		assert.Equal(t, 0.061, c.FaceCollisionDist())
		assert.Equal(t, 0.012, c.FaceEpsilonDist())
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

		assert.Equal(t, 0.35, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))
		assert.Equal(t, 0.35, c.faceThreshold("face-match-dist", 0.001, face.MatchDistDefault, pick))
	})
	t.Run("InRange", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, 0.5, c.faceThreshold("face-match-dist", 0.5, face.MatchDistDefault, pick))
	})
	t.Run("OutOfRangeWarnsOnce", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		// Propagate and the config report both resolve every threshold, so an
		// unguarded warning would repeat for the lifetime of the process.
		hook := captureLog(t)

		assert.Equal(t, 0.35, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))
		assert.Equal(t, 0.35, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))

		var warnings []string

		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "face-match-dist") {
				warnings = append(warnings, e.Message)
			}
		}

		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "out of range")
	})
	t.Run("InRangeStaysSilent", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		hook := captureLog(t)

		assert.Equal(t, 0.5, c.faceThreshold("face-match-dist", 0.5, face.MatchDistDefault, pick))

		for _, e := range hook.AllEntries() {
			assert.NotContains(t, e.Message, "out of range")
		}
	})
	t.Run("AboveTheCeilingIsRefused", func(t *testing.T) {
		// A threshold accepted above the ceiling would only be clipped again where it is
		// read, leaving the config report echoing a value that never applies, so the value
		// is refused here and the calibrated one is used instead.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		hook := captureLog(t)
		above := float64(face.ConfigDistMax) + 0.05

		assert.NotEqual(t, above, c.faceThreshold("face-match-dist", above, face.MatchDistDefault, pick))

		var warned bool
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "out of range") {
				warned = true
			}
		}
		assert.True(t, warned, "a value above the ceiling warns rather than being silently clipped")
	})
}

func TestFaceModelThreshold(t *testing.T) {
	pick := func(m *face.EmbeddingModel) float64 { return m.MatchDist }

	t.Run("Model", func(t *testing.T) {
		assert.Equal(t, 0.35, faceModelThreshold(face.FindEmbeddingModel(face.ModelSFace), pick, 0.4))
	})
	t.Run("NilModel", func(t *testing.T) {
		assert.Equal(t, 0.4, faceModelThreshold(nil, pick, 0.4))
	})
	t.Run("Uncalibrated", func(t *testing.T) {
		assert.Equal(t, 0.4, faceModelThreshold(&face.EmbeddingModel{Name: "test"}, pick, 0.4))
	})
}

func TestConfig_FaceCollisionDist(t *testing.T) {
	// Like the three calibrated thresholds, this follows the model in force.
	sface := face.FindEmbeddingModel(face.ModelSFace).CollisionDist

	c := newSFaceTestConfig(t)
	assert.Equal(t, sface, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 0.04
	assert.Equal(t, 0.04, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 0
	assert.Equal(t, sface, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 1.5
	assert.Equal(t, sface, c.FaceCollisionDist())

	t.Run("OutOfRangeWarns", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		hook := captureLog(t)

		c.options.FaceCollisionDist = 1.5
		assert.Equal(t, sface, c.FaceCollisionDist())

		var warnings []string
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "face-collision-dist") {
				warnings = append(warnings, e.Message)
			}
		}

		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "out of range")
	})
	t.Run("UnsetStaysSilent", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		hook := captureLog(t)

		c.options.FaceCollisionDist = 0
		assert.Equal(t, sface, c.FaceCollisionDist())

		for _, e := range hook.AllEntries() {
			assert.NotContains(t, e.Message, "face-collision-dist")
		}
	})
}

func TestConfig_FaceEpsilonDist(t *testing.T) {
	sface := face.FindEmbeddingModel(face.ModelSFace).Epsilon

	c := newSFaceTestConfig(t)
	assert.Equal(t, sface, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.02
	assert.Equal(t, 0.02, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.2
	assert.Equal(t, sface, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0
	assert.Equal(t, sface, c.FaceEpsilonDist())

	t.Run("OutOfRangeWarns", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		hook := captureLog(t)

		c.options.FaceEpsilonDist = 0.2
		assert.Equal(t, sface, c.FaceEpsilonDist())

		var warnings []string
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "face-epsilon-dist") {
				warnings = append(warnings, e.Message)
			}
		}

		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "out of range")
	})
}

func TestConfig_FaceMatchDist(t *testing.T) {
	// Thresholds follow the model in force, so the test has to put one there.
	sface := face.FindEmbeddingModel(face.ModelSFace).MatchDist

	c := newSFaceTestConfig(t)
	assert.Equal(t, sface, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.1
	assert.Equal(t, 0.1, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.01
	assert.Equal(t, sface, c.FaceMatchDist())
}

func TestConfig_faceAcceptThresholds(t *testing.T) {
	model := face.FindEmbeddingModel(face.ModelSFace)

	newTestConfig := func(t *testing.T) *Config {
		t.Helper()
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace
		return c
	}

	t.Run("Calibrated", func(t *testing.T) {
		radius, matchDist := newTestConfig(t).faceAcceptThresholds()
		assert.Equal(t, model.ClusterRadius, radius)
		assert.Equal(t, model.MatchDist, matchDist)
	})
	t.Run("ConfiguredPairWithinLimit", func(t *testing.T) {
		c := newTestConfig(t)
		// Not 0.4, which is the flag default a value has to differ from to count as set.
		c.options.FaceClusterRadius = 0.7
		c.options.FaceMatchDist = 0.45

		radius, matchDist := c.faceAcceptThresholds()
		assert.Equal(t, 0.7, radius)
		assert.Equal(t, 0.45, matchDist)
	})
	t.Run("OneWideOptionIsRefused", func(t *testing.T) {
		// A cluster accepts at the sum of the two, so a single wide value reaches the limit
		// on its own. Both fall back, because the pair is what was out of range.
		c := newTestConfig(t)
		hook := captureLog(t)
		c.options.FaceMatchDist = 1.0

		radius, matchDist := c.faceAcceptThresholds()
		assert.Equal(t, model.ClusterRadius, radius)
		assert.Equal(t, model.MatchDist, matchDist)

		var warned bool
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "face-cluster-radius") {
				warned = true
			}
		}
		assert.True(t, warned, "a pair above the limit warns rather than being silently clipped")
	})
	t.Run("WarnsOnce", func(t *testing.T) {
		// Propagate and the config report both resolve the pair, so an unguarded warning
		// would repeat for the lifetime of the process.
		c := newTestConfig(t)
		hook := captureLog(t)
		c.options.FaceClusterRadius = 0.9
		c.options.FaceMatchDist = 0.9

		c.faceAcceptThresholds()
		c.faceAcceptThresholds()

		var warnings int
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "face-cluster-radius") {
				warnings++
			}
		}
		assert.Equal(t, 1, warnings)
	})
}
