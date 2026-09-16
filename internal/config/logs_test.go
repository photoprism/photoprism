package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// captureLogChannels runs fn and returns the messages it published on the operator-only system
// channel and on the browser-facing log channel, so a test can assert which of the two a call
// site selected. It proves selection, not delivery: that no subscriber receives the system
// channel is asserted by TestWebsocketTopicsExcludeSystem in internal/api.
func captureLogChannels(t *testing.T, fn func()) (system, browser []string) {
	t.Helper()

	// TestMain points the package logger at the standard logger, which carries no event hook.
	// Without this swap a log.* entry reaches no channel at all and every browser assertion
	// below would hold whatever the code under test did.
	restore := log
	log = event.Log

	t.Cleanup(func() { log = restore })

	sub := event.Subscribe(string(acl.ChannelSystem)+".log.*", "log.*")
	defer event.Unsubscribe(sub)

	fn()

	// Publishing is a synchronous send into a buffered channel, so everything fn emitted is
	// already queued. The deadline covers only the first read, in case that ever changes.
	deadline := time.After(time.Second)

	for {
		var msg event.Message

		if len(system)+len(browser) == 0 {
			select {
			case msg = <-sub.Receiver:
			case <-deadline:
				return system, browser
			}
		} else {
			select {
			case msg = <-sub.Receiver:
			default:
				return system, browser
			}
		}

		text, _ := msg.Fields["message"].(string)

		// An event segment list is a format string assembled from several parts, which no
		// printf vet check can reach, so an arity mistake surfaces only in the rendering.
		assert.NotContains(t, text, "%!", "unrendered format verb in %q", text)

		if ch, _ := event.Topic(msg.Name); ch == string(acl.ChannelSystem) {
			system = append(system, text)
		} else {
			browser = append(browser, text)
		}
	}
}

// namesPath reports whether any of the messages contains the given path.
func namesPath(messages []string, path string) bool {
	for _, s := range messages {
		if strings.Contains(s, path) {
			return true
		}
	}

	return false
}

// blockRegularFile puts a regular file where a directory is expected, so creating anything
// beneath it fails with ENOTDIR for every account. A mode-based denial would not hold for root,
// and this package is run as root often enough that such a case would skip exactly when needed.
func blockRegularFile(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte("not a directory"), fs.ModeFile))
}

func TestNamesPath(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		assert.True(t, namesPath([]string{"a", "read /etc/x (denied)"}, "/etc/x"))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.False(t, namesPath([]string{"a", "read /etc/y (denied)"}, "/etc/x"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, namesPath(nil, "/etc/x"))
	})
}

func TestCaptureLogChannels(t *testing.T) {
	t.Run("SystemChannel", func(t *testing.T) {
		system, browser := captureLogChannels(t, func() {
			event.SystemWarn([]string{"config", "test", "capture system %s"}, "one")
		})
		assert.True(t, namesPath(system, "capture system one"), "expected a system entry, got %v", system)
		assert.False(t, namesPath(browser, "capture system one"))
	})
	t.Run("BrowserChannel", func(t *testing.T) {
		system, browser := captureLogChannels(t, func() {
			log.Warnf("config: capture browser %s", "two")
		})
		assert.True(t, namesPath(browser, "capture browser two"), "expected a browser entry, got %v", browser)
		assert.False(t, namesPath(system, "capture browser two"))
	})
}

func TestConfig_VisionKeyFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		keyFile := filepath.Join(t.TempDir(), "vision_key")
		require.NoError(t, os.WriteFile(keyFile, []byte("SecretAccessToken!"), fs.ModeSecretFile))
		t.Setenv(FlagFileVar("VISION_KEY"), keyFile)

		c := NewConfig(CliTestContext())
		c.options.VisionKey = ""

		assert.Equal(t, "SecretAccessToken!", c.VisionKey())
	})
	t.Run("Unreadable", func(t *testing.T) {
		// A read that fails while the file still stats as non-empty is a permission property,
		// so this one case cannot be built structurally the way the others below are.
		if os.Geteuid() == 0 {
			t.Skip("root reads a file that has no permission bits set")
		}

		keyFile := filepath.Join(t.TempDir(), "vision_key_unreadable")
		require.NoError(t, os.WriteFile(keyFile, []byte("SecretAccessToken!"), fs.ModeSecretFile))
		require.NoError(t, os.Chmod(keyFile, 0o000))
		t.Setenv(FlagFileVar("VISION_KEY"), keyFile)

		c := NewConfig(CliTestContext())
		c.options.VisionKey = ""

		var key string

		system, browser := captureLogChannels(t, func() { key = c.VisionKey() })

		assert.Empty(t, key)
		assert.True(t, namesPath(system, keyFile), "credential file path belongs on the system channel, got %v", system)
		assert.False(t, namesPath(browser, keyFile), "credential file path reached the browser channel: %v", browser)
		assert.False(t, namesPath(system, "***"), "an operator reading the console needs the path, got %v", system)
	})
}

func TestConfig_LoadVisionConfigUnreadable(t *testing.T) {
	// The read failure this covers is a permission property, like VisionKey/Unreadable above.
	if os.Geteuid() == 0 {
		t.Skip("root reads a file that has no permission bits set")
	}

	tempCfg := t.TempDir()
	visionYaml := filepath.Join(tempCfg, "vision.yml")
	require.NoError(t, os.WriteFile(visionYaml, []byte("models:\n"), fs.ModeConfigFile))
	require.NoError(t, os.Chmod(visionYaml, 0o000))

	ctx := CliTestContext()
	require.NoError(t, ctx.Set("config-path", tempCfg))

	c := NewConfig(ctx)
	require.Equal(t, visionYaml, c.VisionYaml())

	restore := vision.Config
	vision.Config = &vision.ConfigValues{}

	t.Cleanup(func() { vision.Config = restore })

	_, browser := captureLogChannels(t, func() { c.LoadVisionConfig() })

	assert.False(t, namesPath(browser, tempCfg), "config path reached the browser channel: %v", browser)
	assert.True(t, namesPath(browser, "***"), "expected the path to be replaced, got %v", browser)
}

func TestNewConfig_OptionsYamlInvalid(t *testing.T) {
	tempCfg := filepath.Join(t.TempDir(), "cfg dir")
	require.NoError(t, os.MkdirAll(tempCfg, 0o700))

	optionsYaml := filepath.Join(tempCfg, "options.yml")
	require.NoError(t, os.WriteFile(optionsYaml, []byte("Debug: not-a-bool\n"), fs.ModeConfigFile))

	ctx := CliTestContext()
	require.NoError(t, ctx.Set("config-path", tempCfg))

	var c *Config

	system, browser := captureLogChannels(t, func() { c = NewConfig(ctx) })

	require.Equal(t, optionsYaml, c.OptionsYaml())
	assert.True(t, namesPath(system, optionsYaml), "config file path belongs on the system channel, got %v", system)
	assert.False(t, namesPath(browser, optionsYaml), "config file path reached the browser channel: %v", browser)
	// The directory name carries a space, which only clean.Log quotes, so the name argument is
	// asserted to be sanitized rather than passed through.
	assert.True(t, namesPath(system, "'"+optionsYaml+"'"), "the file name is not sanitized, got %v", system)
}

func TestNewOptions_DefaultsYamlInvalid(t *testing.T) {
	tempCfg := t.TempDir()
	defaultsYaml := filepath.Join(tempCfg, "defaults.yml")
	require.NoError(t, os.WriteFile(defaultsYaml, []byte("Debug: not-a-bool\n"), fs.ModeConfigFile))

	ctx := CliTestContext()
	require.NoError(t, ctx.Set("config-path", tempCfg))
	require.NoError(t, ctx.Set("defaults-yaml", defaultsYaml))

	var o *Options

	system, browser := captureLogChannels(t, func() { o = NewOptions(ctx) })

	require.Equal(t, defaultsYaml, o.DefaultsYaml)
	assert.True(t, namesPath(system, defaultsYaml), "config file path belongs on the system channel, got %v", system)
	assert.False(t, namesPath(browser, defaultsYaml), "config file path reached the browser channel: %v", browser)
}

func TestConfig_InitSettings(t *testing.T) {
	t.Run("ConfigPathNotCreated", func(t *testing.T) {
		blocker := filepath.Join(t.TempDir(), "blocker")
		blockRegularFile(t, blocker)

		ctx := CliTestContext()
		require.NoError(t, ctx.Set("config-path", filepath.Join(blocker, "config")))
		require.NoError(t, ctx.Set("defaults-yaml", ""))

		c := NewConfig(ctx)
		configPath := c.ConfigPath()

		system, browser := captureLogChannels(t, func() { c.initSettings() })

		assert.True(t, namesPath(system, configPath), "config path belongs on the system channel, got %v", system)
		assert.False(t, namesPath(browser, configPath), "config path reached the browser channel: %v", browser)
		assert.True(t, namesPath(system, "failed to create the directory"), "expected the create failure, got %v", system)
	})
	t.Run("SettingsNotWritten", func(t *testing.T) {
		tempCfg := t.TempDir()

		ctx := CliTestContext()
		require.NoError(t, ctx.Set("config-path", tempCfg))
		require.NoError(t, ctx.Set("defaults-yaml", ""))

		c := NewConfig(ctx)
		settingsYaml := c.SettingsYaml()

		// A directory in place of the settings file makes the write fail for every account.
		require.NoError(t, os.MkdirAll(settingsYaml, 0o700))

		system, browser := captureLogChannels(t, func() { c.initSettings() })

		assert.True(t, namesPath(system, settingsYaml), "config file path belongs on the system channel, got %v", system)
		assert.False(t, namesPath(browser, settingsYaml), "config file path reached the browser channel: %v", browser)
		assert.False(t, namesPath(system, "***"), "an operator reading the console needs the path, got %v", system)
	})
}

func TestConfig_SaveNodeClientSecretError(t *testing.T) {
	tempCfg := t.TempDir()
	ctx := CliTestContext()
	require.NoError(t, ctx.Set("config-path", tempCfg))
	c := NewConfig(ctx)

	secretFile := c.NodeClientSecretFile()
	require.NoError(t, os.MkdirAll(secretFile, 0o700))

	fileName, err := c.SaveNodeClientSecret(cluster.ExampleClientSecret)
	require.Error(t, err)

	// The secret file is named by the wrapped error, so the shared renderer replaces it.
	assert.NotContains(t, clean.Error(err), secretFile)
	assert.Contains(t, clean.Error(err), "***")
	assert.Contains(t, clean.ErrorFull(err), secretFile)

	// The secret stays in memory so the running process keeps working, and the write path
	// reports no file name because it wrote none.
	assert.Empty(t, fileName)
	assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
}

func TestConfig_SaveNodeClientSecretMkdirError(t *testing.T) {
	ctx := CliTestContext()
	require.NoError(t, ctx.Set("config-path", t.TempDir()))
	c := NewConfig(ctx)

	nodeDir := c.NodeConfigPath()
	blockRegularFile(t, nodeDir)

	fileName, err := c.SaveNodeClientSecret(cluster.ExampleClientSecret)
	require.Error(t, err)

	assert.NotContains(t, clean.Error(err), nodeDir)
	assert.Contains(t, clean.Error(err), "***")
	assert.Contains(t, clean.ErrorFull(err), nodeDir)

	assert.Equal(t, c.NodeClientSecretFile(), fileName)
	assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
}

func TestConfig_JoinTokenSaveError(t *testing.T) {
	ctx := CliTestContext()
	require.NoError(t, ctx.Set("config-path", t.TempDir()))

	c := NewConfig(ctx)
	c.options.Edition = Portal
	c.options.JoinToken = ""

	portalDir := c.PortalConfigPath()
	blockRegularFile(t, portalDir)

	var token string

	system, browser := captureLogChannels(t, func() { token = c.JoinToken() })

	assert.Empty(t, token)
	assert.True(t, namesPath(system, portalDir), "join token path belongs on the system channel, got %v", system)
	assert.False(t, namesPath(browser, portalDir), "join token path reached the browser channel: %v", browser)
	assert.False(t, namesPath(system, "***"), "an operator reading the console needs the path, got %v", system)
}

func TestConfig_PropagateLogLevel(t *testing.T) {
	// Both loggers follow the configured level, so an operator who asks for warnings does not
	// receive debug output from the console-only channel the boot stage writes to.
	restoreSystem, restoreLog := event.SystemLog.GetLevel(), log.GetLevel()
	restoreTensorFlow := os.Getenv("TF_CPP_MIN_LOG_LEVEL")

	t.Cleanup(func() {
		event.SystemLog.SetLevel(restoreSystem)
		log.SetLevel(restoreLog)
	})

	c := NewConfig(CliTestContext())
	c.options.Debug = false
	c.options.Trace = false
	c.options.LogLevel = logrus.WarnLevel.String()

	event.SystemLog.SetLevel(logrus.TraceLevel)
	c.Propagate()

	assert.Equal(t, logrus.WarnLevel, event.SystemLog.GetLevel())
	assert.Equal(t, logrus.WarnLevel, log.GetLevel())

	// The TensorFlow variable is process environment rather than a package value, and Propagate
	// runs from a request handler, so only the startup paths may write it.
	assert.Equal(t, restoreTensorFlow, os.Getenv("TF_CPP_MIN_LOG_LEVEL"))
}

func TestSetAppLogLevel(t *testing.T) {
	restoreSystem, restoreLog := event.SystemLog.GetLevel(), log.GetLevel()
	restoreTensorFlow := os.Getenv("TF_CPP_MIN_LOG_LEVEL")

	t.Cleanup(func() {
		event.SystemLog.SetLevel(restoreSystem)
		log.SetLevel(restoreLog)
	})

	t.Run("BothLoggers", func(t *testing.T) {
		SetAppLogLevel(logrus.ErrorLevel)
		assert.Equal(t, logrus.ErrorLevel, log.GetLevel())
		assert.Equal(t, logrus.ErrorLevel, event.SystemLog.GetLevel())
	})
	t.Run("TensorFlowUntouched", func(t *testing.T) {
		SetAppLogLevel(logrus.TraceLevel)
		assert.Equal(t, restoreTensorFlow, os.Getenv("TF_CPP_MIN_LOG_LEVEL"))
	})
}
