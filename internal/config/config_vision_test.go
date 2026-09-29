package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig_VisionYaml(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, ProjectRoot+"/storage/testdata/config/vision.yml", c.VisionYaml())
	})
	t.Run("PreferYamlExtension", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempDir := t.TempDir()
		c.options.ConfigPath = tempDir
		c.options.VisionYaml = ""

		yamlPath := filepath.Join(tempDir, "vision"+fs.ExtYaml)
		if err := os.WriteFile(yamlPath, []byte("models: []\n"), fs.ModeFile); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		assert.Equal(t, yamlPath, c.VisionYaml())
	})
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

func TestConfig_ModelsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.NasnetModelPath()
	assert.True(t, strings.HasPrefix(path, c.ModelsPath()))
	assert.Equal(t, ProjectRoot+"/assets/models/nasnet", path)
}

func TestConfig_TensorFlowDisabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.DisableTensorFlow()
	assert.Equal(t, false, version)
}

func TestConfig_NSFWModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.NsfwModelPath(), "/assets/models/nsfw")
}

func TestConfig_FaceNetModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.FacenetModelPath(), "/assets/models/facenet")
}

func TestConfig_DetectNSFW(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.DetectNSFW()
	assert.Equal(t, true, result)
}

func TestConfig_VisionModelShouldRun(t *testing.T) {
	t.Run("ClassificationDisabledLabels", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DisableClassification = true
		withVisionConfig(t, vision.NewConfig())
		if c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunManual) {
			t.Fatalf("expected false when classification disabled")
		}
	})
	t.Run("DetectNSFWDisabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DetectNSFW = false
		withVisionConfig(t, vision.NewConfig())
		if c.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunManual) {
			t.Fatalf("expected false when detect nsfw disabled")
		}
	})
	t.Run("NilVisionConfig", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		withVisionConfig(t, nil)
		if c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunManual) {
			t.Fatalf("expected false when no vision config is loaded")
		}
	})
	t.Run("DelegatesToVisionConfig", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		withVisionConfig(t, vision.NewConfig())
		if !c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunManual) {
			t.Fatalf("expected labels model to run manually with defaults")
		}
		if !c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex) {
			t.Fatalf("expected labels model to run on index with defaults")
		}
	})
	t.Run("CustomLabelsRunAfterIndex", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		defaultModel := vision.NasnetModel.Clone()
		custom := &vision.Model{Type: vision.ModelTypeLabels, Name: "custom"}
		withVisionConfig(t, &vision.ConfigValues{Models: vision.Models{defaultModel, custom}})
		if !c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunNewlyIndexed) {
			t.Fatalf("expected custom labels model to run after indexing")
		}
		if c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex) {
			t.Fatalf("expected custom labels model to skip on-index runs")
		}
	})
}

func TestConfig_VisionSchedule(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.VisionSchedule = ""
	assert.Equal(t, "", c.VisionSchedule())

	c.options.VisionSchedule = "0 6 * * *"
	assert.Equal(t, "0 6 * * *", c.VisionSchedule())

	c.options.VisionSchedule = "invalid"
	assert.Equal(t, "", c.VisionSchedule())
}

func TestConfig_VisionFilter(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.VisionFilter = "  private:false  "
	assert.Equal(t, "private:false", c.VisionFilter())

	c.options.VisionFilter = ""
	assert.Equal(t, "", c.VisionFilter())
}

func TestConfig_OnnxProvider(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())
	})
	t.Run("CUDA", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.OnnxProvider = "cuda"
		assert.Equal(t, onnx.ProviderCUDA, c.OnnxProvider())
	})
	t.Run("Empty", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.OnnxProvider = ""
		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())
	})
	t.Run("Unsupported", func(t *testing.T) {
		// An unusable value must not stop inference, so it resolves to the default - and must
		// say so, or an operator reads the default as their setting having been applied.
		c := NewConfig(CliTestContext())
		hook := captureConfigLog(t)
		c.options.OnnxProvider = "rocm"

		assert.Equal(t, onnx.ProviderCPU, c.OnnxProvider())

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "rocm")

		// Reported once, not once per loaded model.
		c.OnnxProvider()
		assert.Len(t, hook.AllEntries(), 1)
	})
	t.Run("Nil", func(t *testing.T) {
		var c *Config
		assert.Equal(t, onnx.DefaultProvider, c.OnnxProvider())
	})
}

// captureConfigLog redirects the package logger for the duration of the test and returns its
// entries, so a "report it once" contract can be asserted on what was actually logged.
func captureConfigLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

func TestConfig_WarnVisionConfig(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		// The getters run per loaded model and from the config report, so a repeated call
		// must not repeat the warning. Asserted on the log, not on the map: storing the key
		// and still logging every time would satisfy the map.
		c := NewConfig(CliTestContext())
		hook := captureConfigLog(t)

		c.warnVisionConfig("test-vision-warning", "config: %s", "first")
		c.warnVisionConfig("test-vision-warning", "config: %s", "second")

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "config: first", hook.LastEntry().Message)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	})
	t.Run("DistinctKeys", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		hook := captureConfigLog(t)

		c.warnVisionConfig("test-vision-a", "config: a")
		c.warnVisionConfig("test-vision-b", "config: b")

		assert.Len(t, hook.AllEntries(), 2)
	})
}

// TestVisionKeyWarnings checks which Vision API keys are reported as unable to authenticate requests.
func TestVisionKeyWarnings(t *testing.T) {
	const stripped = "vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API"
	const dollar = "vision key contains $ and is compared as written, without expanding environment variables"

	t.Setenv("VISION_TEST_SECRET", "vision-api-shared-token")
	t.Setenv("VISION_TEST_INVALID", "vision!secret")

	for _, tc := range []struct {
		name     string
		key      string
		incoming bool
		outgoing bool
		want     []string
	}{
		{name: "Empty", key: "", incoming: true, outgoing: true},
		{name: "Valid", key: `Ab3"-+/=#@:;_. ok`, incoming: true, outgoing: true},
		{name: "Generated", key: "vision-api-shared-token", incoming: true},
		{name: "Exclamation", key: "SecretAccessToken!", incoming: true, want: []string{stripped}},
		{name: "Braces", key: "{secret}", incoming: true, want: []string{stripped}},
		{name: "DoubleSpace", key: "vision  key", incoming: true, want: []string{stripped}},
		{name: "NonAscii", key: "schlüssel", incoming: true, want: []string{stripped}},
		{name: "TooLong", key: strings.Repeat("a", 4097), incoming: true, want: []string{stripped}},
		{name: "ExclamationOutgoing", key: "SecretAccessToken!", outgoing: true, want: []string{stripped}},
		{name: "Unused", key: "SecretAccessToken!"},
		{name: "DollarIncoming", key: "$VISION_TEST_SECRET", incoming: true, want: []string{dollar}},
		{name: "DollarOutgoing", key: "$VISION_TEST_SECRET", outgoing: true},
		{name: "BracedIncoming", key: "${VISION_TEST_SECRET}", incoming: true, outgoing: true, want: []string{stripped, dollar}},
		{name: "BracedOutgoing", key: "${VISION_TEST_SECRET}", outgoing: true},
		{name: "DollarBothInvalid", key: "$VISION_TEST_INVALID", incoming: true, outgoing: true, want: []string{stripped, dollar}},
		{name: "BracedOutgoingInvalid", key: "${VISION_TEST_INVALID}", outgoing: true, want: []string{stripped}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, visionKeyWarnings(tc.key, tc.incoming, tc.outgoing))
		})
	}
}

// captureVisionKeyLog redirects the system log and the package logger for the duration of the test, and
// returns the system log entries after checking that the package logger received none.
func captureVisionKeyLog(t *testing.T) *test.Hook {
	t.Helper()

	appHook := captureConfigLog(t)
	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger

	t.Cleanup(func() {
		event.SystemLog = orig
		assert.Empty(t, appHook.AllEntries(), "vision key warnings must not reach the package logger")
	})

	return hook
}

// TestConfig_WarnVisionKey checks that problems with the configured Vision API key are logged once each.
func TestConfig_WarnVisionKey(t *testing.T) {
	t.Run("Warnings", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.VisionKey = "${VISION_SECRET}"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		c.warnVisionKey()

		require.Len(t, hook.AllEntries(), 2)
		for _, entry := range hook.AllEntries() {
			assert.Equal(t, logrus.WarnLevel, entry.Level)
			assert.True(t, strings.HasPrefix(entry.Message, "config: vision key contains "), entry.Message)
			assert.NotContains(t, entry.Message, "VISION_SECRET")
		}
	})
	t.Run("None", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.VisionKey = " vision-api-shared-token "
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Unused", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = false
		c.options.VisionUri = ""
		c.options.VisionKey = "SecretAccessToken!"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Outgoing", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = false
		c.options.VisionUri = "https://vision.example.com/api/v1/vision"
		c.options.VisionKey = "SecretAccessToken!"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		require.Len(t, hook.AllEntries(), 1)
	})
	t.Run("OutgoingDollar", func(t *testing.T) {
		t.Setenv("VISION_SECRET", "")
		c := NewConfig(CliTestContext())
		c.options.VisionApi = false
		c.options.VisionUri = "https://vision.example.com/api/v1/vision"
		c.options.VisionKey = "$VISION_SECRET"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Demo", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.Demo = true
		c.options.VisionKey = "$VISION_SECRET"
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("KeyFile", func(t *testing.T) {
		keyFile := filepath.Join(t.TempDir(), "vision_key")
		require.NoError(t, os.WriteFile(keyFile, []byte("SecretAccessToken!\n"), fs.ModeSecretFile))
		t.Setenv(FlagFileVar("VISION_KEY"), keyFile)

		c := NewConfig(CliTestContext())
		c.options.VisionApi = true
		c.options.VisionKey = ""
		hook := captureVisionKeyLog(t)

		c.warnVisionKey()
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, "config: vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API", hook.LastEntry().Message)
	})
}

// TestConfig_InitWarnVisionKey checks that Init logs the Vision API key warnings to the system log.
func TestConfig_InitWarnVisionKey(t *testing.T) {
	c := NewIsolatedTestConfig("visionkeyinit", t.TempDir(), true)
	c.options.VisionApi = true
	c.options.VisionKey = "Secret Access Token!"

	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger

	// Init registers its database and propagates its settings, so the package's test config is restored.
	t.Cleanup(func() {
		event.SystemLog = orig
		_ = c.CloseDb()
		entity.SetDbProvider(TestConfig())
		TestConfig().Propagate()

		if c.DatabaseDriver() == dsn.DriverSQLite3 {
			for _, suffix := range []string{"", "-journal", "-wal", "-shm"} {
				_ = os.Remove(c.DatabaseDSN() + suffix)
			}
		}
	})

	require.NoError(t, c.Init())

	var warnings []string

	for _, entry := range hook.AllEntries() {
		if strings.HasPrefix(entry.Message, "config: vision key ") {
			warnings = append(warnings, entry.Message)
		}
	}

	assert.Equal(t, []string{"config: vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API"}, warnings)
}
