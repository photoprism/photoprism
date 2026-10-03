package vision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/fs"
)

// captureVisionLog replaces the package logger for the duration of a test.
func captureVisionLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	previous := log
	log = logger
	t.Cleanup(func() { log = previous })

	return hook
}

// onnxOnlyModelsPath returns a models path that contains only the default ONNX models, as on
// installations that no longer ship the TensorFlow label and NSFW models.
func onnxOnlyModelsPath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	for _, name := range []string{string(classify.DefaultModelName()), string(nsfw.DefaultModelName())} {
		src := filepath.Join(assetsPath, "models", name)
		if !fs.PathExists(src) {
			t.Skipf("%s model is not installed", name)
		}
		require.NoError(t, os.Symlink(src, filepath.Join(dir, name)))
	}

	previous := ModelsPath
	ModelsPath = dir
	t.Cleanup(func() { ModelsPath = previous })

	return dir
}

// configuredModels returns all configured models of a type, including disabled ones.
func configuredModels(cfg *ConfigValues, t ModelType) (result Models) {
	for _, m := range cfg.Models {
		if m != nil && m.Type == t {
			result = append(result, m)
		}
	}

	return result
}

// loadLegacyConfig loads a vision config file and returns it with the log entries it produced.
func loadLegacyConfig(t *testing.T, fileName string) (*ConfigValues, []*logrus.Entry) {
	t.Helper()

	hook := captureVisionLog(t)
	cfg := NewConfig()
	require.NoError(t, cfg.Load(fileName))

	return cfg, hook.AllEntries()
}

// assertStableRoundTrip checks that saving and loading the config again gives the same result.
func assertStableRoundTrip(t *testing.T, cfg *ConfigValues) {
	t.Helper()

	first := filepath.Join(t.TempDir(), "vision.yml")
	require.NoError(t, cfg.Save(first))

	reloaded := NewConfig()
	require.NoError(t, reloaded.Load(first))

	second := filepath.Join(t.TempDir(), "vision.yml")
	require.NoError(t, reloaded.Save(second))

	want, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	got, err := yaml.Marshal(reloaded)
	require.NoError(t, err)
	assert.Equal(t, string(want), string(got))

	firstData, err := os.ReadFile(first) //nolint:gosec // test file
	require.NoError(t, err)
	secondData, err := os.ReadFile(second) //nolint:gosec // test file
	require.NoError(t, err)
	assert.Equal(t, string(firstData), string(secondData))
}

// logMessages returns the messages of the entries with the given level.
func logMessages(entries []*logrus.Entry, level logrus.Level) (result []string) {
	for _, entry := range entries {
		if entry.Level == level {
			result = append(result, entry.Message)
		}
	}

	return result
}

// mappedLogMessages returns the messages with the given level that report a mapped legacy model.
func mappedLogMessages(entries []*logrus.Entry, level logrus.Level) (result []string) {
	for _, message := range logMessages(entries, level) {
		if strings.Contains(message, "in place of") || strings.Contains(message, "not supported") {
			result = append(result, message)
		}
	}

	return result
}

// TestModel_IsLegacy verifies which entries are treated as legacy TensorFlow labels or NSFW models.
func TestModel_IsLegacy(t *testing.T) {
	t.Run("BuiltInLabelName", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: "NASNet"}).IsLegacy())
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: " nasnet "}).IsLegacy())
	})
	t.Run("BuiltInNsfwName", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeNsfw, Name: "Nsfw"}).IsLegacy())
	})
	t.Run("CustomTensorFlow", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: "inception", TensorFlow: &tensorflow.ModelInfo{}}).IsLegacy())
		assert.True(t, (&Model{Type: ModelTypeNsfw, Name: "custom", Engine: EngineTensorFlow, TensorFlow: &tensorflow.ModelInfo{}}).IsLegacy())
	})
	t.Run("ONNX", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", ONNX: &onnx.ModelInfo{}}).IsLegacy())
	})
	t.Run("RegisteredOnnxName", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: string(classify.ModelRepViTM10)}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeNsfw, Name: string(nsfw.DefaultModelName())}).IsLegacy())
	})
	t.Run("ServiceUri", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Service: Service{Uri: "http://photoprism-vision:5000/api/v1/vision/labels"}}).IsLegacy())
	})
	t.Run("UnresolvedServiceUri", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeNsfw, Name: "nsfw", Service: Service{Uri: "${PHOTOPRISM_TEST_UNSET_VISION_URI}"}}).IsLegacy())
	})
	t.Run("DisabledService", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeNsfw, Name: "nsfw", Service: Service{Uri: "http://photoprism-vision:5000/api/v1/vision/nsfw", Disabled: true}}).IsLegacy())
	})
	t.Run("RemoteEngine", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: "ollama"}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: " Ollama "}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: "openai"}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: "ollama", TensorFlow: &tensorflow.ModelInfo{}}).IsLegacy())
	})
	t.Run("LocalEngine", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: "TensorFlow"}).IsLegacy())
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: "custom", Engine: EngineONNX, TensorFlow: &tensorflow.ModelInfo{}}).IsLegacy())
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: EngineLocal}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Engine: EngineONNX}).IsLegacy())
	})
	t.Run("OnnxSettings", func(t *testing.T) {
		index := 1
		assert.False(t, (&Model{Type: ModelTypeNsfw, Name: "nsfw", Path: "my_detector", Reduction: nsfw.ReductionSoftmaxUnsafe, UnsafeClassIndex: &index}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeNsfw, Name: "nsfw", DefaultThreshold: 0.6}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", LabelFile: "labels.txt"}).IsLegacy())
		assert.True(t, (&Model{Type: ModelTypeLabels, Name: "nasnet", Path: "nasnet"}).IsLegacy())
	})
	t.Run("OtherTypes", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeFace, Name: "facenet", TensorFlow: &tensorflow.ModelInfo{}}).IsLegacy())
		assert.False(t, (&Model{Type: ModelTypeCaption, Name: "nasnet"}).IsLegacy())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Model
		assert.False(t, m.IsLegacy())
	})
}

// TestModel_hasService verifies which models are meant to use a service.
func TestModel_hasService(t *testing.T) {
	t.Run("Endpoint", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeLabels, Service: Service{Uri: "http://photoprism-vision:5000/api/v1/vision/labels"}}).hasService())
	})
	t.Run("UnresolvedUri", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeLabels, Service: Service{Uri: "${PHOTOPRISM_TEST_UNSET_VISION_URI}"}}).hasService())
	})
	t.Run("EngineDefault", func(t *testing.T) {
		assert.True(t, (&Model{Type: ModelTypeLabels, Engine: "Ollama"}).hasService())
		assert.True(t, (&Model{Type: ModelTypeLabels, Engine: "openai"}).hasService())
	})
	t.Run("GlobalServiceUri", func(t *testing.T) {
		previous := ServiceUri
		ServiceUri = "http://photoprism-vision:5000/api/v1/vision"
		t.Cleanup(func() { ServiceUri = previous })
		assert.True(t, (&Model{Type: ModelTypeLabels, Engine: EngineVision}).hasService())
	})
	t.Run("None", func(t *testing.T) {
		previous := ServiceUri
		ServiceUri = ""
		t.Cleanup(func() { ServiceUri = previous })
		assert.False(t, (&Model{Type: ModelTypeLabels, Engine: "tf"}).hasService())
		assert.False(t, (&Model{Type: ModelTypeLabels, Engine: EngineVision}).hasService())
		assert.False(t, (&Model{Type: ModelTypeLabels, Service: Service{Uri: "http://photoprism-vision:5000", Disabled: true}}).hasService())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Model
		assert.False(t, m.hasService())
	})
}

// TestModel_hasOnnxSettings verifies which options mark a model as an ONNX model.
func TestModel_hasOnnxSettings(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		assert.False(t, (&Model{Type: ModelTypeNsfw, Name: "nsfw", Path: "nsfw", Resolution: 224}).hasOnnxSettings())
	})
	t.Run("Set", func(t *testing.T) {
		index := 0
		assert.True(t, (&Model{Reduction: nsfw.ReductionSoftmaxUnsafe}).hasOnnxSettings())
		assert.True(t, (&Model{UnsafeClassIndex: &index}).hasOnnxSettings())
		assert.True(t, (&Model{NeutralClassIndex: &index}).hasOnnxSettings())
		assert.True(t, (&Model{DefaultThreshold: 0.5}).hasOnnxSettings())
		assert.True(t, (&Model{LabelFile: "labels.txt"}).hasOnnxSettings())
		assert.True(t, (&Model{CanonicalOrder: true}).hasOnnxSettings())
	})
}

// TestConfigValues_mapLegacyModels verifies that legacy entries become default placeholders.
func TestConfigValues_mapLegacyModels(t *testing.T) {
	t.Run("KeepsRunAndDisabled", func(t *testing.T) {
		hook := captureVisionLog(t)
		labels := &Model{Type: ModelTypeLabels, Name: "NASNet", Run: RunOnSchedule}
		nsfwModel := &Model{Type: ModelTypeNsfw, Name: "custom", TensorFlow: &tensorflow.ModelInfo{}, Disabled: true}
		remote := &Model{Type: ModelTypeLabels, Name: "nasnet", Engine: "ollama"}
		cfg := &ConfigValues{Models: Models{labels, nsfwModel, remote}}
		cfg.mapLegacyModels()
		assert.True(t, labels.Default)
		assert.Equal(t, RunOnSchedule, labels.Run)
		assert.True(t, nsfwModel.Default)
		assert.True(t, nsfwModel.Disabled)
		assert.False(t, remote.Default)
		assert.Len(t, logMessages(hook.AllEntries(), logrus.InfoLevel), 1)
		assert.Len(t, logMessages(hook.AllEntries(), logrus.WarnLevel), 1)
	})
	t.Run("SanitizesNames", func(t *testing.T) {
		hook := captureVisionLog(t)
		cfg := &ConfigValues{Models: Models{{Type: ModelTypeLabels, Name: "custom\nmodel", TensorFlow: &tensorflow.ModelInfo{}}}}
		cfg.mapLegacyModels()
		warn := logMessages(hook.AllEntries(), logrus.WarnLevel)
		require.Len(t, warn, 1)
		assert.NotContains(t, warn[0], "\n")
	})
	t.Run("Empty", func(t *testing.T) {
		hook := captureVisionLog(t)
		cfg := &ConfigValues{}
		cfg.mapLegacyModels()
		assert.Empty(t, hook.AllEntries())
	})
}

// TestConfigValues_LoadLegacy verifies that config files written for the TensorFlow label and
// NSFW models load the default ONNX models instead, and save back unchanged.
func TestConfigValues_LoadLegacy(t *testing.T) {
	t.Run("Release250707", func(t *testing.T) {
		onnxOnlyModelsPath(t)
		cfg, entries := loadLegacyConfig(t, filepath.Join("testdata", "vision-250707.yml"))

		labels := cfg.Model(ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, DefaultLabelModel.Name, labels.Name)
		assert.True(t, labels.Default)
		assert.Equal(t, RunAuto, labels.RunType())
		assert.NotNil(t, labels.ClassifyModel())

		detector := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, NsfwModel.Name, detector.Name)
		assert.True(t, detector.Default)
		assert.Equal(t, RunAuto, detector.RunType())
		assert.NotNil(t, detector.NsfwModel())

		// Face and caption entries are left as they are.
		face := configuredModels(cfg, ModelTypeFace)
		require.Len(t, face, 1)
		assert.Equal(t, "FaceNet", face[0].Name)
		caption := configuredModels(cfg, ModelTypeCaption)
		require.Len(t, caption, 1)
		assert.Equal(t, "http://photoprism-vision:5000/api/v1/vision/caption", caption[0].Service.Uri)

		info := mappedLogMessages(entries, logrus.InfoLevel)
		require.Len(t, info, 2)
		assert.Contains(t, info[0], "NASNet")
		assert.Contains(t, info[1], "Nsfw")
		assert.Empty(t, mappedLogMessages(entries, logrus.WarnLevel))

		assertStableRoundTrip(t, cfg)
	})
	t.Run("Release251130", func(t *testing.T) {
		onnxOnlyModelsPath(t)
		cfg, entries := loadLegacyConfig(t, filepath.Join("testdata", "vision-251130.yml"))

		labels := cfg.Model(ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, DefaultLabelModel.Name, labels.Name)
		assert.Nil(t, labels.TensorFlow)
		assert.Equal(t, RunOnSchedule, labels.Run)
		assert.NotNil(t, labels.ClassifyModel())

		// The NSFW entry was disabled in the file and stays disabled.
		assert.Nil(t, cfg.Model(ModelTypeNsfw))
		detectors := configuredModels(cfg, ModelTypeNsfw)
		require.Len(t, detectors, 1)
		assert.Equal(t, NsfwModel.Name, detectors[0].Name)
		assert.True(t, detectors[0].Disabled)
		assert.Equal(t, RunAuto, detectors[0].RunType())

		face := configuredModels(cfg, ModelTypeFace)
		require.Len(t, face, 1)
		assert.Equal(t, FacenetModel.Name, face[0].Name)
		caption := configuredModels(cfg, ModelTypeCaption)
		require.Len(t, caption, 1)
		assert.Equal(t, "gemma3", caption[0].Name)

		info := mappedLogMessages(entries, logrus.InfoLevel)
		require.Len(t, info, 2)
		assert.Contains(t, info[0], "in place of nasnet")
		assert.Contains(t, info[1], "in place of nsfw")
		assert.Empty(t, mappedLogMessages(entries, logrus.WarnLevel))

		assertStableRoundTrip(t, cfg)
	})
	t.Run("CustomTensorFlow", func(t *testing.T) {
		onnxOnlyModelsPath(t)
		cfg, entries := loadLegacyConfig(t, filepath.Join("testdata", "vision-custom-tensorflow.yml"))

		labels := cfg.Model(ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, DefaultLabelModel.Name, labels.Name)
		assert.Equal(t, RunAuto, labels.RunType())
		assert.NotNil(t, labels.ClassifyModel())

		detector := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, NsfwModel.Name, detector.Name)
		assert.Equal(t, RunAlways, detector.Run, "a Run value written in the file is kept")
		assert.NotNil(t, detector.NsfwModel())

		warn := mappedLogMessages(entries, logrus.WarnLevel)
		require.Len(t, warn, 2)
		assert.Contains(t, warn[0], "inception")
		assert.Contains(t, warn[1], "custom-nsfw")
		assert.Empty(t, mappedLogMessages(entries, logrus.InfoLevel))

		assertStableRoundTrip(t, cfg)
	})
	t.Run("MixedLegacyAndRemote", func(t *testing.T) {
		const legacy = "- Type: labels\n  Name: NASNet\n"
		const remote = "- Type: labels\n  Name: gemma3:latest\n  Engine: ollama\n  Service:\n    Uri: http://ollama:11434/api/generate\n"

		t.Run("RemoteLast", func(t *testing.T) {
			fileName := filepath.Join(t.TempDir(), "vision.yml")
			require.NoError(t, os.WriteFile(fileName, []byte("Models:\n"+legacy+remote), fs.ModeConfigFile))
			cfg, _ := loadLegacyConfig(t, fileName)
			labels := cfg.Model(ModelTypeLabels)
			require.NotNil(t, labels)
			assert.Equal(t, "gemma3:latest", labels.Name)
			assert.Equal(t, "ollama", labels.Engine)
			models := configuredModels(cfg, ModelTypeLabels)
			require.Len(t, models, 2)
			assert.Equal(t, DefaultLabelModel.Name, models[0].Name)
		})
		t.Run("LegacyLast", func(t *testing.T) {
			fileName := filepath.Join(t.TempDir(), "vision.yml")
			require.NoError(t, os.WriteFile(fileName, []byte("Models:\n"+remote+legacy), fs.ModeConfigFile))
			cfg, _ := loadLegacyConfig(t, fileName)
			labels := cfg.Model(ModelTypeLabels)
			require.NotNil(t, labels)
			assert.Equal(t, DefaultLabelModel.Name, labels.Name)
			models := configuredModels(cfg, ModelTypeLabels)
			require.Len(t, models, 2)
			assert.Equal(t, "gemma3:latest", models[0].Name)
		})
	})
	t.Run("DisabledService", func(t *testing.T) {
		data := "Models:\n- Type: nsfw\n  Name: nsfw\n  Service:\n    Uri: http://photoprism-vision:5000/api/v1/vision/nsfw\n    Disabled: true\n"
		fileName := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeConfigFile))
		cfg, _ := loadLegacyConfig(t, fileName)
		detector := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, NsfwModel.Name, detector.Name)
	})
	t.Run("RemoteWithTensorFlowSettings", func(t *testing.T) {
		data := `Models:
- Type: labels
  Name: gpt-5-mini
  Engine: openai
- Type: labels
  Name: nasnet
  Service:
    Uri: http://photoprism-vision:5000/api/v1/vision/labels
  TensorFlow:
    Tags:
    - photoprism
- Type: nsfw
  Name: nsfw
  Engine: Ollama
  TensorFlow:
    Tags:
    - serve
`
		fileName := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeConfigFile))
		cfg, entries := loadLegacyConfig(t, fileName)

		// The later service entry stays active, so the earlier OpenAI entry is not used instead.
		labels := cfg.Model(ModelTypeLabels)
		require.NotNil(t, labels)
		assert.Equal(t, "nasnet", labels.Name)
		assert.Equal(t, "http://photoprism-vision:5000/api/v1/vision/labels", labels.Service.Uri)
		assert.Nil(t, labels.TensorFlow)

		detector := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, "nsfw", detector.Name)
		assert.Equal(t, "ollama", detector.Engine)
		assert.Nil(t, detector.TensorFlow)

		assert.Empty(t, mappedLogMessages(entries, logrus.InfoLevel))
		assert.Empty(t, mappedLogMessages(entries, logrus.WarnLevel))
		assertStableRoundTrip(t, cfg)
	})
	t.Run("UnresolvedServiceWithTensorFlowSettings", func(t *testing.T) {
		data := "Models:\n- Type: nsfw\n  Name: nsfw\n  Service:\n    Uri: ${PHOTOPRISM_TEST_UNSET_VISION_URI}\n  TensorFlow:\n    Tags:\n    - serve\n"
		fileName := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeConfigFile))
		cfg, _ := loadLegacyConfig(t, fileName)
		detector := cfg.Model(ModelTypeNsfw)
		require.NotNil(t, detector)
		assert.Equal(t, "nsfw", detector.Name)
		assert.Nil(t, detector.TensorFlow)
		assert.True(t, detector.Service.UriUnresolved())
	})
	t.Run("NoServiceWithTensorFlowSettings", func(t *testing.T) {
		data := `Models:
- Type: labels
  Name: inception
  Engine: tf
  TensorFlow:
    Tags:
    - serve
- Type: nsfw
  Name: nsfw
  Engine: vision
  TensorFlow:
    Tags:
    - serve
`
		fileName := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeConfigFile))
		cfg, entries := loadLegacyConfig(t, fileName)

		labels := configuredModels(cfg, ModelTypeLabels)
		require.Len(t, labels, 1)
		assert.True(t, labels[0].Disabled)
		assert.Equal(t, "inception", labels[0].Name)
		detectors := configuredModels(cfg, ModelTypeNsfw)
		require.Len(t, detectors, 1)
		assert.True(t, detectors[0].Disabled)

		warn := logMessages(entries, logrus.WarnLevel)
		require.Len(t, warn, 2)
		assert.Contains(t, warn[0], "inception")
		assert.Contains(t, warn[1], "nsfw")
	})
	t.Run("PositiveControls", func(t *testing.T) {
		data := `Models:
- Type: labels
  Name: gemma3:latest
  Engine: ollama
  Service:
    Uri: http://ollama:11434/api/generate
- Type: labels
  Name: gpt-5-mini
  Engine: openai
- Type: labels
  Name: nasnet
  Engine: OpenAI
- Type: labels
  Name: custom_21k
  Path: custom_21k
  ONNX:
    File: model.onnx
- Type: nsfw
  Name: custom_detector
  ONNX:
    File: detector.onnx
- Type: face
  Name: facenet
  TensorFlow:
    Tags:
    - serve
`
		fileName := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeConfigFile))
		cfg, entries := loadLegacyConfig(t, fileName)

		labels := configuredModels(cfg, ModelTypeLabels)
		require.Len(t, labels, 4)
		assert.Equal(t, "gemma3:latest", labels[0].Name)
		assert.Equal(t, "ollama", labels[0].Engine)
		assert.False(t, labels[0].Default)
		assert.Equal(t, "gpt-5-mini", labels[1].Name)
		assert.Equal(t, "openai", labels[1].Engine)
		assert.False(t, labels[1].Default)
		assert.Equal(t, "nasnet", labels[2].Name)
		assert.Equal(t, "openai", labels[2].Engine)
		assert.False(t, labels[2].Default)
		assert.Equal(t, "custom_21k", labels[3].Name)
		require.NotNil(t, labels[3].ONNX)
		assert.Equal(t, "model.onnx", labels[3].ONNX.File)
		assert.False(t, labels[3].Default)

		detectors := configuredModels(cfg, ModelTypeNsfw)
		require.Len(t, detectors, 1)
		assert.Equal(t, "custom_detector", detectors[0].Name)
		assert.False(t, detectors[0].Default)

		face := configuredModels(cfg, ModelTypeFace)
		require.Len(t, face, 1)
		assert.NotNil(t, face[0].TensorFlow)
		assert.False(t, face[0].Disabled)

		for _, entry := range entries {
			assert.NotContains(t, entry.Message, "in place of")
			assert.False(t, strings.Contains(entry.Message, "not supported"), entry.Message)
		}
	})
	t.Run("DefaultConfigHasNoRun", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "vision.yml")
		require.NoError(t, NewConfig().Save(fileName))
		data, err := os.ReadFile(fileName) //nolint:gosec // test file
		require.NoError(t, err)

		var saved struct {
			Models []struct {
				Type string `yaml:"Type"`
				Run  string `yaml:"Run"`
			} `yaml:"Models"`
		}
		require.NoError(t, yaml.Unmarshal(data, &saved))
		found := 0
		for _, m := range saved.Models {
			if m.Type == ModelTypeLabels || m.Type == ModelTypeNsfw {
				found++
				assert.Empty(t, m.Run, "%s model", m.Type)
			}
		}
		assert.Equal(t, 2, found)

		cfg := NewConfig()
		require.NoError(t, cfg.Load(fileName))
		for _, m := range cfg.Models {
			if m.Type == ModelTypeLabels || m.Type == ModelTypeNsfw {
				assert.Equal(t, RunAuto, m.RunType(), "%s model", m.Type)
			}
		}
	})
}
