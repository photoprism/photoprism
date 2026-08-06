package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ProjectRoot references the project root directory for use in tests.
var ProjectRoot = fs.Abs("../../")

// Runs first when package is tested.
func init() {
	hub.ApplyTestConfig()
}

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	_ = os.Setenv("PHOTOPRISM_TEST", "true")
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	c := TestConfig()
	defer c.CleanupTestFolder()
	defer func() {
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	return testextras.TestDbCleanup(m.Run())
}

func TestNewConfig(t *testing.T) {
	ctx := CliTestContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewConfig(ctx)

	assert.IsType(t, new(Config), c)
	assert.Equal(t, fs.Abs("../../assets"), c.AssetsPath())
	assert.False(t, c.Prod())
	assert.False(t, c.Debug())
	assert.False(t, c.ReadOnly())
	assert.Equal(t, Develop, c.Develop())
	assert.False(t, c.Experimental())
}

func TestConfig_Prod(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.Prod())
	assert.False(t, c.Debug())
	assert.False(t, c.Trace())

	c.options.Prod = true
	c.options.Debug = true

	assert.True(t, c.Prod())
	assert.False(t, c.Debug())
	assert.False(t, c.Trace())

	c.options.Prod = false

	assert.True(t, c.Debug())
	assert.False(t, c.Trace())

	c.options.Debug = false

	assert.False(t, c.Debug())
	assert.False(t, c.Debug())
	assert.False(t, c.Trace())
}

func TestConfig_Name(t *testing.T) {
	c := NewConfig(CliTestContext())

	name := c.Name()
	assert.Equal(t, "PhotoPrism", name)
}

func TestConfig_About(t *testing.T) {
	c := NewConfig(CliTestContext())

	name := c.About()
	assert.Equal(t, "PhotoPrism®", name)
}

func TestConfig_Edition(t *testing.T) {
	c := NewConfig(CliTestContext())

	name := c.Edition()
	assert.NotEmpty(t, name)
}

func TestConfig_Version(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.Version()
	assert.Equal(t, "0.0.0", version)
}

func TestConfig_VersionChecksum(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, uint32(0x2e5b4b86), c.VersionChecksum())
}

func TestConfig_Copyright(t *testing.T) {
	c := NewConfig(CliTestContext())

	copyright := c.Copyright()
	assert.Equal(t, "", copyright)
}

func TestConfig_OptionsYaml(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Contains(t, c.OptionsYaml(), "options.yml")
	})
	t.Run("ChangePath", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Contains(t, c.OptionsYaml(), "options.yml")
		c.options.ConfigPath = ProjectRoot + "/internal/config/testdata/"
		assert.Equal(t, ProjectRoot+"/internal/config/testdata/options.yml", c.OptionsYaml())
	})
	t.Run("PreferYamlExtension", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempDir := t.TempDir()
		c.options.ConfigPath = tempDir
		c.options.OptionsYaml = ""

		yamlPath := filepath.Join(tempDir, "options"+fs.ExtYaml)
		if err := os.WriteFile(yamlPath, []byte("foo: bar\n"), fs.ModeFile); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		assert.Equal(t, yamlPath, c.OptionsYaml())
	})
}

func TestConfig_PIDFilename(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.PIDFilename(), "/storage/testdata/photoprism.pid")
}

func TestConfig_LogFilename(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.LogFilename(), "/storage/testdata/photoprism.log")
}

func TestConfig_DetachServer(t *testing.T) {
	c := NewConfig(CliTestContext())

	detachServer := c.DetachServer()
	assert.Equal(t, false, detachServer)
}

func TestConfig_OriginalsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.OriginalsPath()
	assert.True(t, strings.HasPrefix(result, "/"))
	assert.True(t, strings.HasSuffix(result, "/storage/testdata/originals"))
}

func TestConfig_ImportPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.AssertTestData(t)

	assert.Equal(t, ProjectRoot+"/storage/testdata/import", c.ImportPath())
	result := c.ImportPath()
	assert.True(t, strings.HasPrefix(result, "/"))
	assert.True(t, strings.HasSuffix(result, "/storage/testdata/import"))

	c.options.ImportPath = ""
	if s := c.ImportPath(); s != "" && s != "/photoprism/import" {
		t.Errorf("unexpected import path: %s", s)
	}

	c.options.ImportPath = result
}

func TestConfig_CachePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasSuffix(c.CachePath(), "storage/testdata/cache"))
}

func TestConfig_MediaCachePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasPrefix(c.MediaCachePath(), "/"))
	assert.True(t, strings.HasSuffix(c.MediaCachePath(), "storage/testdata/cache/media"))
}

func TestConfig_MediaFileCachePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, c.MediaCachePath(), c.MediaFileCachePath(""))
	assert.Equal(t, filepath.Join(c.MediaCachePath(), "a"), c.MediaFileCachePath("a"))
	assert.Equal(t, filepath.Join(c.MediaCachePath(), "0", "b", "5"), c.MediaFileCachePath("0b57b50fe3f6d12bbbf5f1abda3ebcc8bb5ebcee"))
}

func TestConfig_ThumbCachePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasPrefix(c.ThumbCachePath(), "/"))
	assert.True(t, strings.HasSuffix(c.ThumbCachePath(), "storage/testdata/cache/thumbnails"))
}

func TestConfig_AdminUser(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.AdminUser = "foo  "
	assert.Equal(t, "foo", c.AdminUser())
	c.options.AdminUser = "  Admin"
	assert.Equal(t, "admin", c.AdminUser())
}

func TestConfig_SamplesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.SamplesPath()
	assert.Equal(t, ProjectRoot+"/assets/samples", path)
}

func TestConfig_TemplatesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.TemplatesPath()
	assert.Equal(t, ProjectRoot+"/assets/templates", path)
}

func TestConfig_CustomTemplatesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.CustomTemplatesPath()
	assert.Equal(t, "", path)
}

func TestConfig_TemplatesFiles(t *testing.T) {
	c := NewConfig(CliTestContext())

	files := c.TemplateFiles()

	t.Logf("TemplateFiles: %#v", files)
}

func TestConfig_StaticPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.StaticPath()
	assert.Equal(t, ProjectRoot+"/assets/static", path)
}

func TestConfig_StaticFile(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.StaticFile("video/404.mp4")
	assert.Equal(t, ProjectRoot+"/assets/static/video/404.mp4", path)

	path = c.StaticFile("/img/logo.png")
	assert.Equal(t, filepath.Join(c.StaticPath(), "img/logo.png"), path)
}

func TestConfig_StaticBuildPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.StaticBuildPath()
	assert.Equal(t, ProjectRoot+"/assets/static/build", path)
}

func TestConfig_StaticBuildFile(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, filepath.Join(c.StaticBuildPath(), fs.SwJsFile), c.StaticBuildFile(fs.SwJsFile))
	assert.Equal(t, filepath.Join(c.StaticBuildPath(), "chunk/app.js"), c.StaticBuildFile("chunk/app.js"))
}

func TestConfig_StaticImgPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.StaticImgPath()
	assert.Equal(t, ProjectRoot+"/assets/static/img", result)
}

func TestConfig_StaticImgFile(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, filepath.Join(c.StaticImgPath(), "favicon.ico"), c.StaticImgFile("favicon.ico"))
	assert.Equal(t, filepath.Join(c.StaticImgPath(), "wallpapers/default.jpg"), c.StaticImgFile("/wallpapers/default.jpg"))
}

func TestConfig_ThemePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ProjectRoot+"/storage/testdata/config/theme", c.ThemePath())
	c.SetThemePath("testdata/static/img/wallpaper")
	assert.Equal(t, ProjectRoot+"/internal/config/testdata/static/img/wallpaper", c.ThemePath())
	c.SetThemePath("")
	assert.Equal(t, ProjectRoot+"/storage/testdata/config/theme", c.ThemePath())
}

func TestConfig_Serial(t *testing.T) {
	c := TestConfig() // Use complete test context, as NewConfig may not have the required file.

	result := c.Serial()

	t.Logf("Serial: %s", result)

	assert.NotEmpty(t, result)
}

func TestReadSerialFile(t *testing.T) {
	valid := rnd.GenerateUID(serialPrefix)
	write := func(t *testing.T, data string) string {
		t.Helper()
		fileName := filepath.Join(t.TempDir(), serialName)
		require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeSecretFile))
		return fileName
	}
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, valid, readSerialFile(write(t, valid)))
	})
	t.Run("TrailingNewline", func(t *testing.T) {
		// A stray newline must not discard the serial, as that would rotate the derived preview token.
		assert.Equal(t, valid, readSerialFile(write(t, valid+"\n")))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Empty(t, readSerialFile(filepath.Join(t.TempDir(), serialName)))
	})
	t.Run("Truncated", func(t *testing.T) {
		assert.Empty(t, readSerialFile(write(t, valid[:8])))
	})
	t.Run("WrongPrefix", func(t *testing.T) {
		assert.Empty(t, readSerialFile(write(t, "a"+valid[1:])))
	})
	t.Run("NotAlnum", func(t *testing.T) {
		assert.Empty(t, readSerialFile(write(t, "z!!!!!!!!!!!!!!!")))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, readSerialFile(write(t, "")))
	})
}

func TestConfig_InitSerial(t *testing.T) {
	t.Run("GeneratesAndPersists", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		require.NoError(t, c.CreateDirectories())
		require.NoError(t, c.InitSerial())
		serial := c.Serial()
		assert.True(t, rnd.IsUID(serial, serialPrefix))
		// Written to the storage path and mirrored to the backup path, both readable back.
		assert.Equal(t, serial, readSerialFile(filepath.Join(c.StoragePath(), serialName)))
		assert.Equal(t, serial, readSerialFile(c.BackupPath(serialName)))
		// Stable across calls: an existing serial is never regenerated.
		require.NoError(t, c.InitSerial())
		assert.Equal(t, serial, c.Serial())
	})
	t.Run("RecoversFromBackup", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		require.NoError(t, c.CreateDirectories())
		require.NoError(t, c.InitSerial())
		serial := c.Serial()
		// Losing the storage copy must not change the serial, or every preview URL would break.
		require.NoError(t, os.Remove(filepath.Join(c.StoragePath(), serialName)))
		c.serial = ""
		assert.Equal(t, serial, c.Serial())
	})
	t.Run("BackupFailureIsNotFatal", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		require.NoError(t, c.CreateDirectories())
		// Block the backup copy by putting a directory where the file belongs; startup must continue,
		// since the backup only adds redundancy.
		require.NoError(t, os.MkdirAll(c.BackupPath(serialName), fs.ModeDir))
		assert.NoError(t, c.InitSerial())
		assert.True(t, rnd.IsUID(c.Serial(), serialPrefix))
	})
}

func TestConfig_reportDownloadTokenOptions(t *testing.T) {
	// capture swaps the console-only system logger for a buffer, so the assertions do not depend on the
	// application log level or on what another test left behind.
	capture := func(t *testing.T) *bytes.Buffer {
		t.Helper()
		var buf bytes.Buffer
		logger := logrus.New()
		logger.Out = &buf
		logger.SetLevel(logrus.InfoLevel)
		origLog := event.SystemLog
		event.SystemLog = logger
		t.Cleanup(func() { event.SystemLog = origLog })
		return &buf
	}
	newConfig := func(downloadToken string, public bool) *Config {
		c := NewMinimalTestConfig(t.TempDir())
		c.options.Public = public
		c.options.Demo = false
		c.options.DownloadToken = downloadToken
		return c
	}
	t.Run("StaticTokenConfigured", func(t *testing.T) {
		buf := capture(t)
		newConfig("static-download-token", false).reportDownloadTokenOptions()
		// Reported at warning level so it stands out in an operator's log, and named after the option so
		// it can be traced back to the setting that caused it.
		assert.Contains(t, buf.String(), "level=warning")
		assert.Contains(t, buf.String(), "config: download-token")
		assert.Contains(t, buf.String(), "without identifying a session")
	})
	t.Run("NotConfigured", func(t *testing.T) {
		buf := capture(t)
		newConfig("", false).reportDownloadTokenOptions()
		assert.Empty(t, buf.String())
	})
	t.Run("PublicMode", func(t *testing.T) {
		// Nothing is scoped in public mode, so the trade-off does not apply.
		buf := capture(t)
		newConfig("static-download-token", true).reportDownloadTokenOptions()
		assert.Empty(t, buf.String())
	})
	t.Run("MaxAgeBelowMinimum", func(t *testing.T) {
		buf := capture(t)
		c := newConfig("", false)
		c.options.DownloadTokenMaxAge = 60
		c.reportDownloadTokenOptions()
		// Names both the configured value and the floor it was raised to, so the operator can see what
		// the instance actually uses; the clamp itself is covered by TestConfig_DownloadTokenMaxAge.
		assert.Contains(t, buf.String(), "config: download-token-maxage")
		assert.Contains(t, buf.String(), "60s is below the 900s minimum")
	})
}

func TestConfig_SerialChecksum(t *testing.T) {
	c := NewConfig(CliTestContext())

	serial := "zr2g80wvjmm1zwzg"
	expected := "c7dcdb1c"

	c.serial = serial

	result := c.SerialChecksum()

	// t.Logf("Serial: %s", c.serial)
	// t.Logf("SerialChecksum: %s", result)

	assert.NotEmpty(t, result)
	assert.Equal(t, expected, result)
}

func TestConfig_Public(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.Demo = false
	c.options.Public = false
	c.options.AuthMode = "public"

	assert.True(t, c.Public())

	c.options.Demo = true
	c.options.Public = false
	c.options.AuthMode = "public"

	assert.True(t, c.Public())

	c.options.Demo = true
	c.options.Public = true
	c.options.AuthMode = "public"

	assert.True(t, c.Public())

	c.options.Demo = false
	c.options.Public = false
	c.options.AuthMode = "other"

	assert.False(t, c.Public())

	c.options.Demo = false
	c.options.Public = false
	c.options.AuthMode = "password"

	assert.False(t, c.Public())

	c.options.Demo = false
	c.options.Public = true
	c.options.AuthMode = "password"

	assert.True(t, c.Public())

	c.options.Demo = true
	c.options.Public = false
	c.options.AuthMode = "password"

	assert.True(t, c.Public())
}

func TestConfig_Auth(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.Demo = false
	c.options.Public = false
	c.options.AuthMode = "public"

	assert.False(t, c.Auth())

	c.options.Demo = true
	c.options.Public = false
	c.options.AuthMode = "public"

	assert.False(t, c.Auth())

	c.options.Demo = true
	c.options.Public = true
	c.options.AuthMode = "public"

	assert.False(t, c.Auth())

	c.options.Demo = false
	c.options.Public = false
	c.options.AuthMode = "other"

	assert.True(t, c.Auth())

	c.options.Demo = false
	c.options.Public = false
	c.options.AuthMode = "password"

	assert.True(t, c.Auth())

	c.options.Demo = false
	c.options.Public = true
	c.options.AuthMode = "password"

	assert.False(t, c.Auth())

	c.options.Demo = true
	c.options.Public = false
	c.options.AuthMode = "password"

	assert.False(t, c.Auth())
}

func TestConfigOptions(t *testing.T) {
	c := NewConfig(CliTestContext())
	r := c.Options()

	assert.False(t, r.DisableExifTool)
	assert.Equal(t, r.AutoImport, 7200)
	assert.Equal(t, r.AutoIndex, -1)

	c.options = nil
	r2 := c.Options()

	assert.Equal(t, r2.AutoImport, 0)
	assert.Equal(t, r2.AutoIndex, 0)
}
