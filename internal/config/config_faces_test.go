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
	"github.com/photoprism/photoprism/internal/mutex"
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

		modelFile := face.DefaultDetector().Path(tempModels)
		require.NoError(t, os.MkdirAll(filepath.Dir(modelFile), fs.ModeDir))
		require.NoError(t, os.WriteFile(modelFile, []byte("onnx"), fs.ModeFile))

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
		c.options.FaceModelThreads = 4

		assert.True(t, c.FaceEngineShouldRun(vision.RunOnIndex))
		assert.False(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))
		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
		assert.True(t, c.FaceEngineShouldRun(vision.RunAuto))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule))
	})
	t.Run("AutoLowThreads", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModelThreads = 2

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
	t.Run("AutoSkipsTheScheduledPass", func(t *testing.T) {
		// Left on auto, faces stay out of scheduled cron runs so a background job does not start
		// detecting unannounced. FACE_RUN is what changes that, not a Run value in "vision.yml".
		c := NewConfig(CliTestContext())
		c.options.FaceRun = ""

		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule))
	})
	t.Run("OnDemandDoesNotSweepOnSchedule", func(t *testing.T) {
		// Re-detecting pictures an earlier pass examined finds nothing while the detector is
		// unchanged, and the sweep costs a full decode per file. "On demand" means a person or an
		// import asked; changing detector is a migration or an explicit run, not a cron tick.
		c := NewConfig(CliTestContext())
		c.options.FaceRun = vision.RunOnDemand

		assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule))
		assert.True(t, c.FaceEngineShouldRun(vision.RunNewlyIndexed))
		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
		assert.False(t, c.FaceEngineShouldRun(vision.RunOnIndex), "on-demand does not detect inline")
	})
	t.Run("TheScheduledSweepHasToBeAskedForByName", func(t *testing.T) {
		// Only these two name it, so nothing else can start a library-wide pass by accident.
		c := NewConfig(CliTestContext())

		for _, run := range []vision.RunType{vision.RunOnSchedule, vision.RunAlways} {
			c.options.FaceRun = run
			assert.True(t, c.FaceEngineShouldRun(vision.RunOnSchedule), run)
		}

		for _, run := range []vision.RunType{"", vision.RunOnDemand, vision.RunNewlyIndexed, vision.RunOnIndex, vision.RunManual, vision.RunNever} {
			c.options.FaceRun = run
			assert.False(t, c.FaceEngineShouldRun(vision.RunOnSchedule), run)
		}
	})
	t.Run("VisionYamlDoesNotSchedule", func(t *testing.T) {
		// FaceEngineRunType never consults it, so a test that sets one is testing "auto".
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace}}}

		c := NewConfig(CliTestContext())
		m := vision.Config.Model(vision.ModelTypeFace)
		require.NotNil(t, m)
		m.Run = vision.RunNever

		assert.True(t, c.FaceEngineShouldRun(vision.RunManual))
	})
}

func TestConfig_LoadVisionConfig(t *testing.T) {
	newVisionYaml := func(t *testing.T, body string) *Config {
		t.Helper()

		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		c.options.VisionYaml = filepath.Join(c.options.ConfigPath, "vision.yml")
		require.NoError(t, os.WriteFile(c.options.VisionYaml, []byte(body), fs.ModeConfigFile))

		return c
	}

	t.Run("FaceRunIsIgnored", func(t *testing.T) {
		// Warned rather than noted: this used to be the documented way to turn face detection
		// off, so an operator who set "never" has it running again after an upgrade.
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })
		vision.Config = vision.NewConfig()

		c := newVisionYaml(t, "Models:\n  - Type: face\n    Default: true\n    Run: never\n")

		hook := test.NewGlobal()
		t.Cleanup(hook.Reset)

		c.LoadVisionConfig()

		assert.Equal(t, vision.RunAuto, c.FaceEngineRunType(), "the file must not decide the schedule")

		var reported bool

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "FACE_RUN") {
				reported = true
			}
		}

		assert.True(t, reported, "an ignored setting must not be silent")
	})
	t.Run("NoFaceRun", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })
		vision.Config = vision.NewConfig()

		c := newVisionYaml(t, "Models:\n  - Type: labels\n    Default: true\n    Run: manual\n")

		assert.NotPanics(t, c.LoadVisionConfig)
	})
	t.Run("MissingFile", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionYaml = filepath.Join(t.TempDir(), "absent.yml")

		assert.NotPanics(t, c.LoadVisionConfig)
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.NotPanics(t, (*Config)(nil).LoadVisionConfig)
	})
}

func TestConfig_FaceEngineRunType(t *testing.T) {
	t.Run("AutoDefaults", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModelThreads = 1
		assert.Equal(t, "auto", vision.ReportRunType(c.FaceEngineRunType()))

		c.options.DisableFaces = true
		assert.Equal(t, "never", vision.ReportRunType(c.FaceEngineRunType()))
		c.options.DisableFaces = false

		c.options.FaceModelThreads = 4
		assert.Equal(t, "auto", vision.ReportRunType(c.FaceEngineRunType()))
	})
	t.Run("FollowsFaceRun", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceRun = vision.RunOnSchedule
		t.Cleanup(func() { c.options.FaceRun = "" })

		assert.Equal(t, vision.RunOnSchedule, c.FaceEngineRunType())
	})
	t.Run("IgnoresTheVisionModel", func(t *testing.T) {
		// "vision.yml" no longer configures faces. Two ways to set one schedule raise a
		// precedence question nobody can answer from the outside, so the file is read and
		// ignored rather than obeyed; FACE_RUN and DISABLE_FACES are the way.
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace, Run: vision.RunNever}}}
		c := NewConfig(CliTestContext())

		assert.Equal(t, vision.RunAuto, c.FaceEngineRunType())
	})
	t.Run("VisionSubsystemUnavailable", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = nil
		c := NewConfig(CliTestContext())

		assert.Equal(t, vision.RunNever, c.FaceEngineRunType())
	})
	t.Run("VisionModelShouldRunFace", func(t *testing.T) {
		origVision := vision.Config
		t.Cleanup(func() { vision.Config = origVision })

		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeFace}}}
		c := NewConfig(CliTestContext())
		c.options.FaceRun = vision.RunOnSchedule
		t.Cleanup(func() { c.options.FaceRun = "" })

		assert.True(t, c.VisionModelShouldRun(vision.ModelTypeFace, vision.RunOnSchedule))

		c.options.DisableFaces = true
		assert.False(t, c.VisionModelShouldRun(vision.ModelTypeFace, vision.RunOnSchedule))
		c.options.DisableFaces = false
	})
}

func TestConfig_FaceDetectorThreads(t *testing.T) {
	t.Run("SharedWithIndexWorkers", func(t *testing.T) {
		// Detection takes no lock, so one pool of this size runs per indexing worker.
		c := NewConfig(CliTestContext())
		expected := max(runtime.NumCPU()/max(c.IndexWorkers(), 1), 1)
		assert.Equal(t, expected, c.FaceDetectorThreads())
		assert.LessOrEqual(t, c.FaceDetectorThreads()*c.IndexWorkers(), max(runtime.NumCPU(), c.IndexWorkers()))
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetectorThreads = 8
		assert.Equal(t, 8, c.FaceDetectorThreads())
	})
	t.Run("KeepsOptionUnset", func(t *testing.T) {
		// Writing the derived value back would freeze it for FaceModelThreads as well.
		c := NewConfig(CliTestContext())
		c.FaceDetectorThreads()
		assert.LessOrEqual(t, c.options.FaceDetectorThreads, 0)
	})
	t.Run("DeprecatedOption", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 6
		assert.Equal(t, 6, c.FaceDetectorThreads())
		assert.Equal(t, 6, c.FaceModelThreads())
	})
	t.Run("SpecificOptionWins", func(t *testing.T) {
		// The two derive different defaults, so a value that set both must not override the
		// one an operator configured for detection alone.
		c := NewConfig(CliTestContext())
		c.options.FaceEngineThreads = 6
		c.options.FaceDetectorThreads = 3
		assert.Equal(t, 3, c.FaceDetectorThreads())
		assert.Equal(t, 6, c.FaceModelThreads())
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
		c.options.FaceModelThreads = 8
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
		c.options.FaceModelThreads = 4
		assert.True(t, c.faceEngineRunsOnIndex())
	})
	t.Run("FewCores", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModelThreads = 2
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
	t.Run("DetectionDisabled", func(t *testing.T) {
		// A path beside a detector of "none" reads as though weights were loaded, and that row is
		// what an operator checks to find out whether any are.
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone

		assert.Empty(t, c.FaceEngineModelPath())
	})
	t.Run("RefusedDetectorStillNamesItsOwn", func(t *testing.T) {
		// A detector that was asked for and could not be loaded resolves to none, so this row is
		// the only place the artifact that was looked for is reported.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceDetector = face.DetectorYuNet

		assert.Equal(t, face.DetectorNone, c.FaceDetector())
		assert.Equal(t, face.FindDetector(face.DetectorYuNet).Path(c.options.ModelsPath), c.FaceEngineModelPath())
	})
	t.Run("NothingInstalled", func(t *testing.T) {
		// The path names what would have been loaded, so the caller reports a missing
		// detector rather than an empty string.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()

		assert.Equal(t, face.DefaultDetector().Path(c.options.ModelsPath), c.FaceEngineModelPath())
	})
	t.Run("LegacyArtifactName", func(t *testing.T) {
		// An operator who installed the detector under its earlier name keeps it, rather than
		// having it treated as absent because the registry now names the publisher's artifact.
		c := NewConfig(CliTestContext())
		models := t.TempDir()
		c.options.ModelsPath = models
		c.options.FaceDetector = face.DetectorSCRFD

		t.Setenv(face.LicenseAcceptanceVar, "1")

		scrfd := face.FindDetector(face.DetectorSCRFD)
		require.NotEmpty(t, scrfd.Legacy)

		legacy := installTestDetectorFile(t, filepath.Join(models, scrfd.Dir, scrfd.Legacy[0]))

		assert.Equal(t, legacy, c.FaceEngineModelPath())
	})
	t.Run("SelectedDetector", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		models := t.TempDir()
		c.options.ModelsPath = models

		t.Setenv(face.LicenseAcceptanceVar, "1")

		yunet := face.FindDetector(face.DetectorYuNet)
		scrfd := face.FindDetector(face.DetectorSCRFD)
		installTestDetectorFile(t, yunet.Path(models))
		installTestDetectorFile(t, scrfd.Path(models))

		assert.Equal(t, yunet.Path(models), c.FaceEngineModelPath(), "derivation must not select gated weights")

		c.options.FaceDetector = face.DetectorSCRFD
		assert.Equal(t, scrfd.Path(models), c.FaceEngineModelPath(), "an explicit detector is loaded")
	})
}

// installTestDetectorFile writes a placeholder detector artifact and returns its path. It is
// never loaded, so its contents only have to exist.
func installTestDetectorFile(t *testing.T, path string) string {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), fs.ModeDir))
	require.NoError(t, os.WriteFile(path, []byte("onnx"), fs.ModeFile))

	return path
}

func TestConfig_FaceDetectorSetting(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, face.DetectorAuto, c.FaceDetectorSetting())
	})
	t.Run("DetectIsNoLongerAccepted", func(t *testing.T) {
		// It was an accepted spelling of "auto" during development and never shipped, so it now
		// reads as a typo: reported once, and applied as a request to derive a detector.
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = "detect"
		assert.Equal(t, face.DetectorAuto, c.FaceDetectorSetting())
		assert.False(t, face.KnownDetectorName("detect"))
	})
	t.Run("Named", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = "YuNet"
		assert.Equal(t, face.DetectorYuNet, c.FaceDetectorSetting())
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone
		assert.Equal(t, face.DetectorNone, c.FaceDetectorSetting())
	})
	t.Run("Unsupported", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = "nonexistent"
		assert.Equal(t, face.DetectorAuto, c.FaceDetectorSetting())
	})
	t.Run("DeprecatedEngineDisables", func(t *testing.T) {
		// "none" is the one FACE_ENGINE value that means the same in both options.
		c := NewConfig(CliTestContext())
		c.options.FaceEngine = face.EngineNone
		assert.Equal(t, face.DetectorNone, c.FaceDetectorSetting())
	})
	t.Run("DeprecatedEngineHasNoOtherOpinion", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngine = face.EngineONNX
		assert.Equal(t, face.DetectorAuto, c.FaceDetectorSetting())
		c.options.FaceEngine = "pigo"
		assert.Equal(t, face.DetectorAuto, c.FaceDetectorSetting())
	})
	t.Run("DetectorWinsOverDeprecatedEngine", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceEngine = face.EngineNone
		c.options.FaceDetector = face.DetectorYuNet
		assert.Equal(t, face.DetectorYuNet, c.FaceDetectorSetting())
	})
}

// TestConfig_FaceDetectorLeavesOptionsYaml pins that a persisted deprecated value is read rather
// than rewritten. Writing the option that replaces it would be one line, and removing it from every
// operator's file afterwards would not be, so the deprecated key is simply consulted where it sits.
func TestConfig_FaceDetectorLeavesOptionsYaml(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.ConfigPath = t.TempDir()

	body := "# operator notes\nFaceEngine: none\nFaceModel: sface\n"
	require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte(body), fs.ModeConfigFile))
	require.NoError(t, c.options.Load(c.OptionsYaml()))

	assert.Equal(t, face.DetectorNone, c.FaceDetectorSetting(), "the deprecated value still disables detection")
	assert.Equal(t, face.EngineNone, c.FaceEngine())

	after, err := os.ReadFile(c.OptionsYaml())
	require.NoError(t, err)
	assert.Equal(t, body, string(after), "options.yml must be left exactly as the operator wrote it")
}

func TestConfig_FaceDetector(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.DetectorNone, (*Config)(nil).FaceDetector())
	})
	t.Run("NothingInstalled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		assert.Equal(t, face.DetectorNone, c.FaceDetector())
	})
	t.Run("Derived", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		models := t.TempDir()
		c.options.ModelsPath = models
		installTestDetectorFile(t, face.FindDetector(face.DetectorYuNet).Path(models))

		assert.Equal(t, face.DetectorYuNet, c.FaceDetector())
	})
	t.Run("DerivationSkipsGatedWeights", func(t *testing.T) {
		// Accepting the license is not selecting the detector, so a build that holds only
		// gated weights detects nothing rather than reaching for them on its own.
		c := NewConfig(CliTestContext())
		models := t.TempDir()
		c.options.ModelsPath = models
		installTestDetectorFile(t, face.FindDetector(face.DetectorSCRFD).Path(models))

		t.Setenv(face.LicenseAcceptanceVar, "1")

		assert.Equal(t, face.DetectorNone, c.FaceDetector())
	})
	t.Run("GatedWeightsRefusedWithoutAcceptance", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		models := t.TempDir()
		c.options.ModelsPath = models
		c.options.FaceDetector = face.DetectorSCRFD
		installTestDetectorFile(t, face.FindDetector(face.DetectorSCRFD).Path(models))

		t.Setenv(face.LicenseAcceptanceVar, "")

		assert.Equal(t, face.DetectorNone, c.FaceDetector())
	})
	t.Run("SelectedButNotInstalled", func(t *testing.T) {
		// Falling forward to another detector would move the landmarks, so a detector that
		// was asked for and cannot run disables detection.
		c := NewConfig(CliTestContext())
		models := t.TempDir()
		c.options.ModelsPath = models
		c.options.FaceDetector = face.DetectorSCRFD
		installTestDetectorFile(t, face.FindDetector(face.DetectorYuNet).Path(models))

		t.Setenv(face.LicenseAcceptanceVar, "1")

		assert.Equal(t, face.DetectorNone, c.FaceDetector())
	})
	t.Run("Disabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone
		assert.Equal(t, face.DetectorNone, c.FaceDetector())
	})
}

func TestConfig_FaceModelSetting(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""

		assert.Equal(t, face.ModelAuto, c.FaceModelSetting())
	})
	t.Run("Detect", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = "Detect"

		assert.Equal(t, face.ModelAuto, c.FaceModelSetting())
	})
	t.Run("AutoIsDetect", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelAuto

		assert.Equal(t, face.ModelAuto, c.FaceModelSetting())
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

		assert.Equal(t, face.ModelAuto, c.FaceModelSetting())
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
		c.faceModelFlag = face.ModelAuto
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
		c.options.FaceModel = face.ModelAuto
		c.faceModel = ""

		assert.Equal(t, entity.MarkerFixtures.Get("1000003-4").EmbedModel, c.ResolveFaceModel())
		assert.Equal(t, face.ModelAuto, c.FaceModelSetting())
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
		c.options.FaceModel = face.ModelAuto
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
		// schema that predates the column had neither. RemoveIndex keeps this portable
		// across the drivers the suite runs on.
		require.NoError(t, db.Model(&entity.Marker{}).RemoveIndex("idx_markers_embed_model").Error)
		require.NoError(t, db.Exec("ALTER TABLE "+table+" DROP COLUMN embed_model").Error)

		t.Cleanup(func() {
			require.NoError(t, db.Exec("ALTER TABLE "+table+" ADD COLUMN embed_model VARBINARY(32) DEFAULT ''").Error)

			// The column comes back empty, so what the fixtures recorded has to be written
			// again: a later test that asks the library which model it holds would otherwise
			// be answered by this one.
			require.NoError(t, db.Model(&entity.Marker{}).
				Where("marker_type = ? AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
				UpdateColumn("embed_model", entity.MarkerFixtures.Get("1000003-4").EmbedModel).Error)

			require.NoError(t, db.Model(&entity.Marker{}).AddIndex("idx_markers_embed_model", "embed_model").Error)
		})

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
		// A report runs before the library can be asked, so the calibration that applies is the
		// default model's rather than none: leaving it nil blanked every distance in the report.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""
		m := c.FaceEmbeddingModel()
		require.NotNil(t, m)
		assert.Equal(t, c.installedFaceModel(), m.Name)
	})
	t.Run("NotInstalled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
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
	t.Run("SmallestSupported", func(t *testing.T) {
		// The bound is where the detectors stop being trained, so a deliberate operator can
		// ask for the small faces a crowd photograph is made of.
		c.options.FaceSize = face.MinSizeThreshold
		assert.Equal(t, face.MinSizeThreshold, c.FaceSize())
		c.options.FaceSize = face.MinSizeThreshold - 1
		assert.Equal(t, face.SizeThreshold, c.FaceSize())
	})
	c.options.FaceSize = 0
}

func TestConfig_FaceSizeRetry(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.RetrySizeThreshold, c.FaceSizeRetry())
	t.Run("Disabled", func(t *testing.T) {
		c.options.FaceSizeRetry = -1
		assert.Zero(t, c.FaceSizeRetry())
	})
	t.Run("UnsetSelectsTheDefault", func(t *testing.T) {
		// Zero is what a configuration that never named the option holds, and it must not
		// read as a request to turn the fallback off.
		c.options.FaceSizeRetry = 0
		assert.Equal(t, face.RetrySizeThreshold, c.FaceSizeRetry())
	})
	t.Run("OutOfRange", func(t *testing.T) {
		c.options.FaceSizeRetry = 100000
		assert.Equal(t, face.RetrySizeThreshold, c.FaceSizeRetry())
	})
	t.Run("NeverAboveTheOrdinaryThreshold", func(t *testing.T) {
		// A retry asking for larger faces than the first pass could only find fewer, so it
		// is clamped rather than allowed to make the fallback pointless.
		c.options.FaceSize = 25
		c.options.FaceSizeRetry = 40
		assert.Equal(t, 25, c.FaceSizeRetry())
	})
	c.options.FaceSize = 0
	c.options.FaceSizeRetry = 0
}

func TestConfig_FaceScore(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.ScoreThresholdDefault, c.FaceScore())
	c.options.FaceScore = 80
	assert.Equal(t, 80.0, c.FaceScore())
	c.options.FaceScore = 0.1
	assert.Equal(t, face.ScoreThresholdDefault, c.FaceScore(), "an out-of-range value falls back to the default")
}

func TestConfig_FaceRun(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceRun = ""
		assert.Equal(t, vision.RunAuto, c.FaceRun())
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceRun = "on-schedule"
		assert.Equal(t, vision.RunOnSchedule, c.FaceRun())
	})
	t.Run("Alias", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceRun = "Schedule"
		assert.Equal(t, vision.RunOnSchedule, c.FaceRun())
	})
	t.Run("Unsupported", func(t *testing.T) {
		// Reported and applied as if nothing were set, rather than turning face work off over
		// a typo. The warning fires once, so the getter may be called from Propagate.
		c := NewConfig(CliTestContext())
		c.options.FaceRun = "whenever"
		assert.Equal(t, vision.RunAuto, c.FaceRun())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, vision.RunNever, (*Config)(nil).FaceRun())
	})
}

func TestConfig_FaceClusterScoreEffective(t *testing.T) {
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = 55
		assert.Equal(t, 55, c.FaceClusterScore())
		assert.Equal(t, 55, c.FaceClusterScoreEffective())
	})
	t.Run("Detector", func(t *testing.T) {
		// Unset must stay distinguishable from a value: the bar is per marker, and collapsing it
		// to the detector in force would apply that calibration to markers another one scored.
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = 0
		assert.Zero(t, c.FaceClusterScore())
		assert.Equal(t, c.detectorClusterScore(), c.FaceClusterScoreEffective())
		assert.Positive(t, c.FaceClusterScoreEffective())
	})
	t.Run("Disabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = -1
		assert.Equal(t, -1, c.FaceClusterScore())
		assert.Equal(t, -1, c.FaceClusterScoreEffective())
	})
	t.Run("OutOfRange", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = 300
		assert.Zero(t, c.FaceClusterScore())
	})
}

func TestConfig_FaceScoreEffective(t *testing.T) {
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 80
		assert.Equal(t, 80.0, c.FaceScoreEffective())
	})
	t.Run("Detector", func(t *testing.T) {
		// The raw option is zero in the ordinary case, which reads as "nothing is filtered", so
		// a report has to resolve the cutoff the detector actually enforces.
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 0
		assert.Equal(t, face.DetectorScore(c.FaceDetector()), c.FaceScoreEffective())
		assert.Positive(t, c.FaceScoreEffective(), "a report must never state a cutoff of zero")
	})
	t.Run("DetectionDisabled", func(t *testing.T) {
		// Detection is off, so no detector states a cutoff. Falling back to the default one
		// still beats reporting zero, which would read as a setting rather than as an absence.
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone
		assert.Positive(t, c.FaceScoreEffective())
	})
	t.Run("Disabled", func(t *testing.T) {
		// Negative follows the convention the other numeric options use, and is the only way to
		// ask for no cutoff at all: zero is taken by "let the detector decide".
		c := NewConfig(CliTestContext())
		c.options.FaceScore = -1
		assert.Equal(t, face.NoScoreThreshold, c.FaceScore())
		assert.Equal(t, face.NoScoreThreshold, c.FaceScoreEffective())
	})
}

func TestDetectorScoreThreshold(t *testing.T) {
	t.Run("Rescaled", func(t *testing.T) {
		// The option is on the 0-100 scale operators read scores in, the detector on 0-1.
		assert.InDelta(t, 0.55, detectorScoreThreshold(55), 0.0001)
	})
	t.Run("SentinelsCarryThrough", func(t *testing.T) {
		// Neither is a score, so neither may be divided into a value the engine reads as one.
		assert.Zero(t, detectorScoreThreshold(0))
		assert.EqualValues(t, -1, detectorScoreThreshold(-1))
	})
}

func TestConfig_ConfigureFaceDetector(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.NoError(t, c.ConfigureFaceDetector(0))
	})
	t.Run("LowerScore", func(t *testing.T) {
		// A migration lowers the cutoff below the detector's own, which is the case that had no
		// reachable value before: the option is on the 0-100 scale and the detector on 0-1.
		c := NewConfig(CliTestContext())
		assert.NoError(t, c.ConfigureFaceDetector(face.DetectorMigrateScore(c.FaceDetector())))
	})
	t.Run("DetectionDisabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone
		assert.NoError(t, c.ConfigureFaceDetector(0))
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.NoError(t, (*Config)(nil).ConfigureFaceDetector(0))
	})
}

func TestConfig_EffectiveFaceModel(t *testing.T) {
	t.Run("InForce", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		assert.Equal(t, face.ModelSFace, c.EffectiveFaceModel())
	})
	t.Run("Undetected", func(t *testing.T) {
		// FaceModel reports none until the library has been asked, and a report runs before
		// that, so the model whose calibration applies is the one detection would settle on.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""
		assert.Equal(t, c.installedFaceModel(), c.EffectiveFaceModel())
	})
	t.Run("Disabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone
		assert.Equal(t, face.ModelNone, c.EffectiveFaceModel())
	})
	t.Run("NoneInstalled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = ""
		assert.Equal(t, face.ModelNone, c.EffectiveFaceModel())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.ModelNone, (*Config)(nil).EffectiveFaceModel())
	})
}

func TestConfig_FacesLockFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, filepath.Join(c.StoragePath(), "faces.lock"), c.FacesLockFile())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Empty(t, (*Config)(nil).FacesLockFile())
	})
}

func TestConfig_FacesLocked(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Unlocked", func(t *testing.T) {
		assert.Empty(t, c.FacesLocked())
	})
	t.Run("Locked", func(t *testing.T) {
		lock, err := mutex.AcquireFileLock(c.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		assert.Contains(t, c.FacesLocked(), "faces migration")
	})
	t.Run("Released", func(t *testing.T) {
		assert.Empty(t, c.FacesLocked())
	})
}

func TestConfig_ClearFaceModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		require.NoError(t, c.SetFaceModel(face.ModelSFace))
		require.Equal(t, face.ModelSFace, c.options.FaceModel)

		require.NoError(t, c.ClearFaceModel())

		assert.Empty(t, c.options.FaceModel)
		assert.Equal(t, face.ModelAuto, c.FaceModelSetting())

		b, err := os.ReadFile(c.OptionsYaml())
		require.NoError(t, err)
		assert.NotContains(t, string(b), "FaceModel")
	})
	t.Run("UnwritableFile", func(t *testing.T) {
		// The pin has to stay in memory when it could not be removed from the file, or the run
		// resolves a detected model while a restart pins the old one again.
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		require.NoError(t, c.SetFaceModel(face.ModelFaceNet))
		require.NoError(t, os.Chmod(c.OptionsYaml(), 0o400))
		t.Cleanup(func() { _ = os.Chmod(c.OptionsYaml(), fs.ModeConfigFile) })
		require.NoError(t, os.Chmod(c.options.ConfigPath, 0o500)) //nolint:gosec // read-only by design
		t.Cleanup(func() { _ = os.Chmod(c.options.ConfigPath, fs.ModeDir) })

		require.Error(t, c.ClearFaceModel())

		assert.Equal(t, face.ModelFaceNet, c.options.FaceModel, "memory must agree with the file")
	})
	t.Run("NothingPinned", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		assert.NoError(t, c.ClearFaceModel())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.NoError(t, (*Config)(nil).ClearFaceModel())
	})
}

// TestConfig_FaceMigrateScore pins that the migration's floor is tunable without a rebuild, and
// that it is a floor of its own rather than the detector's calibrated one.
func TestConfig_FaceMigrateScore(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, face.DetectorMigrateScore(c.FaceDetector()), c.FaceMigrateScore())
		assert.Less(t, c.FaceMigrateScore(), face.DetectorScore(c.FaceDetector()),
			"a migration floor at the calibrated cutoff recovers nothing the index would not have found")
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceMigrateScore = 25
		assert.Equal(t, 25.0, c.FaceMigrateScore())
	})
	t.Run("Disabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceMigrateScore = -1
		assert.Equal(t, face.NoScoreThreshold, c.FaceMigrateScore())
	})
	t.Run("FallsBackToFaceScore", func(t *testing.T) {
		// A cutoff an operator set for detection as a whole still stands here.
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 40
		assert.Equal(t, 40.0, c.FaceMigrateScore())
	})
	t.Run("OutranksFaceScore", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 40
		c.options.FaceMigrateScore = 12
		assert.Equal(t, 12.0, c.FaceMigrateScore())
	})
	t.Run("OutOfRange", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceMigrateScore = 300
		assert.Equal(t, face.DetectorMigrateScore(c.FaceDetector()), c.FaceMigrateScore())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.DefaultDetectorMigrateScore(), (*Config)(nil).FaceMigrateScore())
	})
}

func TestConfig_FaceMigrateSize(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, face.MinSizeThreshold, c.FaceMigrateSize())
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceMigrateSize = 18
		assert.Equal(t, 18, c.FaceMigrateSize())
	})
	t.Run("DoesNotInheritFaceSize", func(t *testing.T) {
		// A legacy marker can describe a face well under FACE_SIZE, which is the case this
		// floor exists for.
		c := NewConfig(CliTestContext())
		c.options.FaceSize = 40
		assert.Equal(t, face.MinSizeThreshold, c.FaceMigrateSize())
	})
	t.Run("BelowWhatTheDetectorsAreTrainedFor", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceMigrateSize = 2
		assert.Equal(t, face.MinSizeThreshold, c.FaceMigrateSize())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, face.MinSizeThreshold, (*Config)(nil).FaceMigrateSize())
	})
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

	t.Run("DerivedFromTheModelThroughTheFlagLayer", func(t *testing.T) {
		// Built from the registered flags rather than by assigning the option, because that is
		// where the derivation was defeated: a flag Value made the option non-zero on every
		// start, so the getter never reached the per-model branch. Literals, not the propagated
		// global, which Propagate assigns from this getter and would compare against itself.
		ctx := cliContextWithFlagDefaults(t)

		for model, want := range map[string]int{face.ModelFaceNet: 160, face.ModelSFace: 112} {
			c := &Config{cliCtx: ctx, options: NewOptions(ctx)}
			c.options.FaceModel = model

			require.Zero(t, c.options.FaceClusterSize, "an unset option must stay zero to mean derive it")
			assert.Equal(t, want, c.FaceClusterSize(), model)
		}
	})
}

func TestConfig_FaceClusterScore(t *testing.T) {
	c := NewConfig(CliTestContext())
	// Unset reports zero rather than a bar, because the bar that applies is looked up per marker
	// from the detector that scored it.
	assert.Zero(t, c.FaceClusterScore())
	c.options.FaceClusterScore = 0
	assert.Zero(t, c.FaceClusterScore())
	c.options.FaceClusterScore = 55
	assert.Equal(t, 55, c.FaceClusterScore())
}

func TestConfig_FaceClusterCore(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, face.ClusterCoreDefault, c.FaceClusterCore())
	c.options.FaceClusterCore = 1000
	assert.Equal(t, face.ClusterCoreDefault, c.FaceClusterCore())
	c.options.FaceClusterCore = 1
	assert.Equal(t, 1, c.FaceClusterCore())
}

// TestConfig_FaceClusterPercentile covers the calibration knob for how wide a cluster's stored
// radius is, whose two ends are the shipped percentile and the maximum it replaced.
func TestConfig_FaceRecomputeStats(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("DefaultsOff", func(t *testing.T) {
		// Off while the width guard is calibrated, so a percentile sweep moves one variable.
		assert.False(t, c.FaceRecomputeStats())
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.FaceRecomputeStats = true
		defer func() { c.options.FaceRecomputeStats = false }()

		assert.True(t, c.FaceRecomputeStats())
	})
	t.Run("Nil", func(t *testing.T) {
		var nilConf *Config
		assert.False(t, nilConf.FaceRecomputeStats())
	})
}

func TestConfig_FaceClusterPercentile(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, face.ClusterPercentileDefault, c.FaceClusterPercentile())
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.FaceClusterPercentile = 90
		assert.Equal(t, 90, c.FaceClusterPercentile())
	})
	t.Run("TheMaximum", func(t *testing.T) {
		// 100 is what the radius meant before the percentile, so a run can be compared against it.
		c.options.FaceClusterPercentile = 100
		assert.Equal(t, 100, c.FaceClusterPercentile())
	})
	t.Run("OutOfRange", func(t *testing.T) {
		// Below 1 selects no member at all, so it reads as unset rather than as an empty cluster.
		for _, value := range []int{0, -1, 101} {
			c.options.FaceClusterPercentile = value
			assert.Equal(t, face.ClusterPercentileDefault, c.FaceClusterPercentile(), "%d", value)
		}
	})
	t.Run("Propagated", func(t *testing.T) {
		restore := face.ClusterPercentile
		t.Cleanup(func() { face.ClusterPercentile = restore })

		c.options.FaceClusterPercentile = 80
		c.Propagate()

		assert.Equal(t, 80, face.ClusterPercentile)
	})
}

// setFlagContext returns a context in which the named integer flag was given on the command line,
// which is what tells an explicit value from the zero value of an Options built in code.
func setFlagContext(t *testing.T, name, value string) *cli.Context {
	t.Helper()

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Int(name, -99, "doc")

	require.NoError(t, flags.Parse([]string{"--" + name, value}))

	return cli.NewContext(nil, flags, nil)
}

// TestConfig_FlagIsSet covers the predicate that tells a flag an operator gave from one nobody did,
// which is the only way a zero with a meaning can be distinguished from an unset struct field.
func TestConfig_FlagIsSet(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.cliCtx = setFlagContext(t, "face-cluster-split-rounds", "0")

		assert.True(t, c.flagIsSet("face-cluster-split-rounds"))
	})
	t.Run("Unset", func(t *testing.T) {
		assert.False(t, NewConfig(CliTestContext()).flagIsSet("face-cluster-split-rounds"))
	})
	t.Run("NoContext", func(t *testing.T) {
		assert.False(t, (&Config{}).flagIsSet("face-cluster-split-rounds"))
		assert.False(t, (*Config)(nil).flagIsSet("face-cluster-split-rounds"))
	})
}

// TestConfig_FaceClusterSplitRounds covers the width guard's round budget, whose three settings do
// not order the way their numbers suggest: zero is the strictest and the sentinel is the loosest.
func TestConfig_FaceClusterSplitRounds(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, face.ClusterSplitRoundsDefault, c.FaceClusterSplitRounds())
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.FaceClusterSplitRounds = 8
		assert.Equal(t, 8, c.FaceClusterSplitRounds())
	})
	t.Run("Off", func(t *testing.T) {
		c.options.FaceClusterSplitRounds = face.ClusterSplitOff
		assert.Equal(t, face.ClusterSplitOff, c.FaceClusterSplitRounds())
	})
	t.Run("ZeroIsNotUnset", func(t *testing.T) {
		// The trap this getter exists for. Zero is a value - it discards a wide group - and it is
		// also what an Options built without the flag defaults holds, so it counts only when the
		// flag or its environment variable actually carried it.
		c.options.FaceClusterSplitRounds = 0
		assert.Equal(t, face.ClusterSplitRoundsDefault, c.FaceClusterSplitRounds())

		explicit := NewConfig(CliTestContext())
		explicit.cliCtx = setFlagContext(t, "face-cluster-split-rounds", "0")
		explicit.options.FaceClusterSplitRounds = 0

		assert.Equal(t, 0, explicit.FaceClusterSplitRounds(), "an explicit zero has to survive")
	})
	t.Run("BelowTheSentinel", func(t *testing.T) {
		// Only one negative value carries a meaning, so a typo cannot remove the guard silently.
		c.options.FaceClusterSplitRounds = -2
		assert.Equal(t, face.ClusterSplitRoundsDefault, c.FaceClusterSplitRounds())
	})
}

// TestConfig_FaceClusterSplitShrink covers the factor a split round shortens the link distance by.
func TestConfig_FaceClusterSplitShrink(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Default", func(t *testing.T) {
		assert.InDelta(t, face.ClusterSplitShrinkDefault, c.FaceClusterSplitShrink(), 1e-9)
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.FaceClusterSplitShrink = 0.8
		assert.InDelta(t, 0.8, c.FaceClusterSplitShrink(), 1e-9)
	})
	t.Run("OutOfRange", func(t *testing.T) {
		// One or more never shortens the distance, so every round would repeat the previous pass.
		for _, value := range []float64{0, 1, 1.5, -0.5} {
			c.options.FaceClusterSplitShrink = value
			assert.InDelta(t, face.ClusterSplitShrinkDefault, c.FaceClusterSplitShrink(), 1e-9, "%g", value)
		}
	})
	t.Run("Propagated", func(t *testing.T) {
		rounds, shrink := face.ClusterSplitRounds, face.ClusterSplitShrink
		t.Cleanup(func() { face.ClusterSplitRounds, face.ClusterSplitShrink = rounds, shrink })

		c.options.FaceClusterSplitShrink = 0.8
		c.options.FaceClusterSplitRounds = face.ClusterSplitOff
		c.Propagate()

		assert.InDelta(t, 0.8, face.ClusterSplitShrink, 1e-9)
		assert.True(t, face.ClusterSplitDisabled())
	})
}

// TestConfig_PropagateSampleThreshold pins the clustering trigger to FACE_CLUSTER_CORE. It is
// derived rather than configured, and leaving it at the package initializer froze it at the
// shipped default, so raising the core size moved what a cluster is but not what starts a pass.
func TestConfig_PropagateSampleThreshold(t *testing.T) {
	core, samples := face.ClusterCore, face.SampleThreshold
	t.Cleanup(func() { face.ClusterCore, face.SampleThreshold = core, samples })

	c := NewConfig(CliTestContext())
	c.options.FaceClusterCore = 7
	c.Propagate()

	assert.Equal(t, 7, face.ClusterCore)
	assert.Equal(t, 14, face.SampleThreshold)
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

		assert.Equal(t, 0.72, c.FaceClusterDist())
		assert.Equal(t, 0.70, c.FaceClusterRadius())
		assert.Equal(t, 0.25, c.FaceMatchDist())
		// Neither gap scales with the model: a collision floor that did would exclude the members
		// of both clusters, and a wider Epsilon strands embeddings rather than telling anyone apart.
		assert.Equal(t, face.CollisionDistDefault, c.FaceCollisionDist())
		assert.Equal(t, face.EpsilonDefault, c.FaceEpsilonDist())
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
		// The distance flags carry no default, because the value that applies is calibrated per
		// embedding model and "--help" cannot name one. An unset option therefore reads as zero
		// and must resolve to the model's own value rather than being treated as configured.
		ctx := cliContextWithFlagDefaults(t)
		c := &Config{cliCtx: ctx, options: NewOptions(ctx)}
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		require.Zero(t, c.options.FaceClusterDist)
		require.Zero(t, c.options.FaceCollisionDist)
		assert.Equal(t, 0.72, c.FaceClusterDist())
		assert.Equal(t, 0.70, c.FaceClusterRadius())
		assert.Equal(t, 0.25, c.FaceMatchDist())
		assert.Equal(t, face.CollisionDistDefault, c.FaceCollisionDist())
		assert.Equal(t, face.EpsilonDefault, c.FaceEpsilonDist())
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

	t.Run("Unset", func(t *testing.T) {
		assert.False(t, c.faceThresholdIsSet("face-cluster-dist", 0))
	})
	t.Run("CustomValue", func(t *testing.T) {
		assert.True(t, c.faceThresholdIsSet("face-cluster-dist", 0.5))
	})
	t.Run("FlagSetToZero", func(t *testing.T) {
		// An operator who passes the flag configured it, whatever the value.
		set := flag.NewFlagSet("test", flag.ContinueOnError)
		set.Float64("face-cluster-dist", 0, "doc")
		assert.NoError(t, set.Parse([]string{"--face-cluster-dist", "0"}))

		explicit := &Config{cliCtx: cli.NewContext(cli.NewApp(), set, nil), options: NewOptions(nil)}

		assert.True(t, explicit.faceThresholdIsSet("face-cluster-dist", 0))
	})
	t.Run("NoContext", func(t *testing.T) {
		none := &Config{options: NewOptions(nil)}

		assert.False(t, none.faceThresholdIsSet("face-cluster-dist", 0))
	})
}

func TestConfig_FaceThreshold(t *testing.T) {
	pick := func(m *face.EmbeddingModel) float64 { return m.MatchDist }

	t.Run("OutOfRange", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelSFace)
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, 0.25, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))
		assert.Equal(t, 0.25, c.faceThreshold("face-match-dist", 0.001, face.MatchDistDefault, pick))
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

		assert.Equal(t, 0.25, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))
		assert.Equal(t, 0.25, c.faceThreshold("face-match-dist", 1.6, face.MatchDistDefault, pick))

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
		assert.Equal(t, 0.25, faceModelThreshold(face.FindEmbeddingModel(face.ModelSFace), pick, 0.4))
	})
	t.Run("NilModel", func(t *testing.T) {
		assert.Equal(t, 0.4, faceModelThreshold(nil, pick, 0.4))
	})
	t.Run("Uncalibrated", func(t *testing.T) {
		assert.Equal(t, 0.4, faceModelThreshold(&face.EmbeddingModel{Name: "test"}, pick, 0.4))
	})
}

// TestConfig_FaceMatchMargin covers the gap the nearest cluster has to beat the runner-up by. It
// is the difference between two distances rather than one, so it is not floored at the collision
// distance the calibrated thresholds are floored at.
func TestConfig_FaceMatchMargin(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		assert.Equal(t, float64(face.MatchMarginDefault), c.FaceMatchMargin())
	})
	t.Run("Configured", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		c.options.FaceMatchMargin = 0.1
		assert.Equal(t, 0.1, c.FaceMatchMargin())
	})
	t.Run("BelowTheCollisionDistance", func(t *testing.T) {
		// A margin of a hundredth is a legitimate setting, where a cluster radius that small
		// would not be, so this must not inherit the floor faceThreshold applies.
		c := newSFaceTestConfig(t)
		c.options.FaceMatchMargin = 0.01
		assert.Equal(t, 0.01, c.FaceMatchMargin())
	})
	t.Run("Disabled", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		c.options.FaceMatchMargin = face.NoMatchMargin
		assert.Zero(t, c.FaceMatchMargin())
	})
	t.Run("OutOfRangeWarns", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		hook := captureLog(t)

		c.options.FaceMatchMargin = 1.5
		assert.Equal(t, float64(face.MatchMarginDefault), c.FaceMatchMargin())

		var warnings []string
		for _, e := range hook.AllEntries() {
			if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "face-match-margin") {
				warnings = append(warnings, e.Message)
			}
		}

		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "out of range")
	})
	t.Run("UnsetStaysSilent", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		hook := captureLog(t)

		assert.Equal(t, float64(face.MatchMarginDefault), c.FaceMatchMargin())

		for _, e := range hook.AllEntries() {
			assert.NotContains(t, e.Message, "face-match-margin")
		}
	})
}

func TestConfig_FaceCollisionDist(t *testing.T) {
	// A gap rather than a distance in the model's scale, so it is the same for every model.
	sface := face.FindEmbeddingModel(face.ModelSFace).CollisionDist

	require.Equal(t, float64(face.CollisionDistDefault), sface)

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
	c.options.FaceEpsilonDist = 0.004
	assert.Equal(t, 0.004, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0.2
	assert.Equal(t, sface, c.FaceEpsilonDist())
	c.options.FaceEpsilonDist = 0
	assert.Equal(t, sface, c.FaceEpsilonDist())

	t.Run("CappedAtTheConfigurableCeiling", func(t *testing.T) {
		// The gap is a void where nothing matches, and twice it retires a colliding cluster for
		// good. EpsilonDistMax holds the ambiguity cutoff at or below 0.02, so the option can
		// narrow that door but never widen it.
		c := newSFaceTestConfig(t)

		c.options.FaceEpsilonDist = face.EpsilonDistMax
		assert.Equal(t, face.EpsilonDistMax, c.FaceEpsilonDist())

		c.options.FaceEpsilonDist = face.EpsilonDistMax * 2
		assert.Equal(t, sface, c.FaceEpsilonDist(), "above the cap resolves to the model value")
	})
	t.Run("AboveTheDefaultButInRange", func(t *testing.T) {
		// The ceiling does not follow the default, so narrowing the default leaves a setting that
		// was in range still valid.
		c := newSFaceTestConfig(t)

		c.options.FaceEpsilonDist = face.EpsilonDefault * 5
		assert.Equal(t, face.EpsilonDefault*5, c.FaceEpsilonDist())
	})

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
