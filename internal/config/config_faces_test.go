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

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
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

func TestConfig_FaceModelReport(t *testing.T) {
	t.Run("ProvisionalWithoutDatabase", func(t *testing.T) {
		// A configuration report has to stay usable when the database is unreachable, so an
		// unresolved "auto" is named as provisional rather than printed as if it were in force.
		c := NewConfig(CliTestContext())
		assert.Equal(t, face.ModelSFace+" (auto, unresolved)", c.FaceModelReport())
	})
	t.Run("ResolvedAgainstLibrary", func(t *testing.T) {
		c := TestConfig()
		c.options.FaceModel = face.ModelAuto

		assert.Equal(t, face.ModelFaceNet, c.FaceModelReport())
	})
	t.Run("Explicit", func(t *testing.T) {
		// An explicitly named model needs no library lookup, so it carries no caveat.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelArcFaceR50)
		c.options.FaceModel = "ArcFace-R50"

		assert.Equal(t, face.ModelArcFaceR50, c.FaceModelReport())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Equal(t, face.ModelNone, c.FaceModelReport())
	})
	t.Run("NilConfig", func(t *testing.T) {
		var c *Config
		assert.Equal(t, face.ModelNone, c.FaceModelReport())
	})
}

func TestConfig_FaceModel(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		// Without a database there is no library to ask, so the preference list decides.
		c := NewConfig(CliTestContext())
		assert.Equal(t, face.ModelSFace, c.FaceModel())
	})
	t.Run("LibraryKeepsItsModel", func(t *testing.T) {
		// The fixture markers hold vectors without a recorded model, which is what a
		// library indexed before the provenance column looks like. Resolving those to
		// SFace on upgrade would leave every stored cluster incomparable.
		c := TestConfig()
		c.options.FaceModel = face.ModelAuto

		assert.Equal(t, face.ModelFaceNet, c.FaceModel())
	})
	t.Run("Explicit", func(t *testing.T) {
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
		c.options.FaceModel = face.ModelArcFaceR50

		assert.Equal(t, face.ModelNone, c.FaceModel())
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
		assert.Equal(t, face.ModelSFace, c.FaceModel())
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

func TestInstalledFaceModel(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, face.ModelNone, installedFaceModel(t.TempDir()))
	})
	t.Run("FollowsPreferenceOrder", func(t *testing.T) {
		// FaceNet is installed too, but SFace comes first for a library with no vectors.
		modelsPath := installTestModels(t, face.ModelFaceNet, face.ModelSFace)
		assert.Equal(t, face.ModelSFace, installedFaceModel(modelsPath))
	})
	t.Run("SkipsMissing", func(t *testing.T) {
		modelsPath := installTestModels(t, face.ModelArcFaceR50)
		assert.Equal(t, face.ModelArcFaceR50, installedFaceModel(modelsPath))
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
		assert.Equal(t, face.ModelFaceNet, TestConfig().libraryFaceModel())
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
		// schema that predates the column had neither. RemoveIndex keeps this portable
		// across the drivers the suite runs on.
		require.NoError(t, db.Model(&entity.Marker{}).RemoveIndex("idx_markers_embed_model").Error)
		require.NoError(t, db.Exec("ALTER TABLE "+table+" DROP COLUMN embed_model").Error)

		t.Cleanup(func() {
			require.NoError(t, db.Exec("ALTER TABLE "+table+" ADD COLUMN embed_model VARBINARY(32) DEFAULT ''").Error)
			require.NoError(t, db.Model(&entity.Marker{}).AddIndex("idx_markers_embed_model", "embed_model").Error)
		})

		assert.Equal(t, face.ModelFaceNet, c.libraryFaceModel())
	})
}

func TestConfig_FaceEmbeddingModel(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		m := c.FaceEmbeddingModel()
		require.NotNil(t, m)
		assert.Equal(t, face.ModelSFace, m.Name)
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
	// Thresholds follow the resolved model, which is SFace without a library to ask.
	sface := face.FindEmbeddingModel(face.ModelSFace).ClusterDist

	c := NewConfig(CliTestContext())
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
	// Thresholds follow the resolved model, which is SFace without a library to ask.
	sface := face.FindEmbeddingModel(face.ModelSFace).ClusterRadius

	c := NewConfig(CliTestContext())
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

		assert.Equal(t, 0.91, c.FaceClusterDist())
		assert.Equal(t, 0.67, c.FaceClusterRadius())
		assert.Equal(t, 0.39, c.FaceMatchDist())
		assert.Equal(t, 0.071, c.FaceCollisionDist())
		assert.Equal(t, 0.014, c.FaceEpsilonDist())
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
		assert.Equal(t, 0.91, c.FaceClusterDist())
		assert.Equal(t, 0.67, c.FaceClusterRadius())
		assert.Equal(t, 0.39, c.FaceMatchDist())
		assert.Equal(t, 0.071, c.FaceCollisionDist())
		assert.Equal(t, 0.014, c.FaceEpsilonDist())
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
	t.Run("OutOfRangeWarnsOnce", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		// Propagate and the config report both resolve every threshold, so an
		// unguarded warning would repeat for the lifetime of the process.
		hook := captureLog(t)

		assert.Equal(t, 0.39, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))
		assert.Equal(t, 0.39, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))

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
		// read, leaving the config report echoing a value that never applies. One constant
		// bounds both, so the value is refused here and the calibrated one is used.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		hook := captureLog(t)
		above := float64(face.AcceptDistMax) + 0.05

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
	// Like the three calibrated thresholds, this follows the resolved model, which is
	// SFace without a library to ask.
	sface := face.FindEmbeddingModel(face.ModelSFace).CollisionDist

	c := NewConfig(CliTestContext())
	assert.Equal(t, sface, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 0.04
	assert.Equal(t, 0.04, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 0
	assert.Equal(t, sface, c.FaceCollisionDist())
	c.options.FaceCollisionDist = 1.5
	assert.Equal(t, sface, c.FaceCollisionDist())

	t.Run("OutOfRangeWarns", func(t *testing.T) {
		c := NewConfig(CliTestContext())
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
		c := NewConfig(CliTestContext())
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

	c := NewConfig(CliTestContext())
	assert.Equal(t, sface, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.02
	assert.Equal(t, 0.02, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.2
	assert.Equal(t, sface, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0
	assert.Equal(t, sface, c.FaceEpsilonDist())

	t.Run("OutOfRangeWarns", func(t *testing.T) {
		c := NewConfig(CliTestContext())
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
	// Thresholds follow the resolved model, which is SFace without a library to ask.
	sface := face.FindEmbeddingModel(face.ModelSFace).MatchDist

	c := NewConfig(CliTestContext())
	assert.Equal(t, sface, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.1
	assert.Equal(t, 0.1, c.FaceMatchDist())
	c.options.FaceMatchDist = 0.01
	assert.Equal(t, sface, c.FaceMatchDist())
}
