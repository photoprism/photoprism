package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestConfig_FindExecutable(t *testing.T) {
	assert.Equal(t, "", findBin("yyy", "xxx"))
}

func TestConfig_SidecarPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.SidecarPath(), "testdata/sidecar")
	c.options.SidecarPath = ".photoprism"
	assert.Equal(t, ".photoprism", c.SidecarPath())
	c.options.SidecarPath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/sidecar", c.SidecarPath())
}

func TestConfig_UsersPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.UsersPath(), "users")
}

func TestConfig_UsersOriginalsPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.UsersOriginalsPath(), "users")
}

func TestConfig_UsersStoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.UsersStoragePath(), "users")
}

func TestConfig_UserStoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.UserStoragePath(""))
	assert.Equal(t, "", c.UserStoragePath("etaetyget"))
	assert.Contains(t, c.UserStoragePath("urjult03ceelhw6k"), "users/urjult03ceelhw6k")
}

func TestConfig_UserUploadPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	if dir, err := c.UserUploadPath("", ""); err == nil {
		t.Error("error expected")
	} else {
		assert.Equal(t, "", dir)
	}
	if dir, err := c.UserUploadPath("etaetyget", ""); err == nil {
		t.Error("error expected")
	} else {
		assert.Equal(t, "", dir)
	}
	if dir, err := c.UserUploadPath("urjult03ceelhw6k", ""); err != nil {
		t.Fatal(err)
	} else {
		assert.Contains(t, dir, "users/urjult03ceelhw6k/upload")
	}
	if dir, err := c.UserUploadPath("urjult03ceelhw6k", "foo"); err != nil {
		t.Fatal(err)
	} else {
		assert.Contains(t, dir, "users/urjult03ceelhw6k/upload/foo")
	}
}

func TestConfig_SidecarPathIsAbs(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.SidecarPathIsAbs())
	c.options.SidecarPath = ".photoprism"
	assert.Equal(t, false, c.SidecarPathIsAbs())
}

func TestConfig_SidecarWritable(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.SidecarWritable())
}

func TestConfig_FFmpegBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.Contains(c.FFmpegBin(), "/bin/ffmpeg"))
}

func TestConfig_TempPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	d0 := c.tempPath()

	t.Logf("c.options.TempPath: '%s'", c.options.TempPath)
	t.Logf("c.tempPath(): '%s'", d0)

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/temp", c.tempPath())

	c.options.TempPath = ""

	d1 := c.tempPath()

	if d1 == "" {
		t.Fatal("temp path is empty")
	}

	if !strings.HasPrefix(d1, "/tmp/photoprism_") {
		t.Fatalf("unexpected temp path: %s", d1)
	}

	d2 := c.tempPath()

	if d2 == "" {
		t.Fatal("temp path is empty")
	}

	if !strings.HasPrefix(d2, "/tmp/photoprism_") {
		t.Fatalf("unexpected temp path: %s", d2)
	}

	if d1 != d2 {
		t.Fatalf("temp paths should match: '%s' <=> '%s'", d1, d2)
	} else {
		t.Logf("temp paths match: '%s' == '%s'", d1, d2)
	}

	if d4 := c.TempPath(); d4 != d0 {
		t.Fatalf("temp paths should match: '%s' <=> '%s'", d4, d0)
	} else {
		t.Logf("temp paths match: '%s' == '%s'", d4, d0)
	}
}

func TestConfig_CmdCachePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	if dir := c.CmdCachePath(); dir == "" {
		t.Fatal("cmd cache path is empty")
	} else if !strings.HasPrefix(dir, c.CachePath()) {
		t.Fatalf("unexpected cmd cache path: %s", dir)
	}
}

func TestConfig_CmdLibPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	if dir := c.CmdLibPath(); dir == "" {
		t.Fatal("cmd lib path is empty")
	} else if !strings.HasPrefix(dir, "/usr") {
		t.Fatalf("unexpected cmd lib path: %s", dir)
	}
}

func TestConfig_CachePath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/cache", c.CachePath())
	c.options.CachePath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/cache", c.CachePath())
}

func TestConfig_StoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata", c.StoragePath())
	c.options.StoragePath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/originals/.photoprism/storage", c.StoragePath())
}

func TestConfig_TestdataPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/testdata", c.TestdataPath())
}

func TestConfig_AlbumsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/albums", c.AlbumsPath())
}

func TestConfig_OriginalsAlbumsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/originals/albums", c.OriginalsAlbumsPath())
}

func TestConfig_CreateDirectories(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()

		c := &Config{
			options: NewTestOptions("config"),
			token:   rnd.GenerateToken(8),
		}

		if err := c.CreateDirectories(); err != nil {
			t.Fatal(err)
		}
	})
}

/* TODO Doesn't fail on https://drone.photoprism.app/!
	--- FAIL: TestConfig_CreateDirectories2 (0.00s)
    --- FAIL: TestConfig_CreateDirectories2/asset_path_not_found (0.00s)
        fs_test.go:142: error expected

func TestConfig_CreateDirectories2(t *testing.T) {
	t.Run("asset path not found", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}
		c.options.AssetsPath = ""

		err := c.CreateDirectories()
		if err == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err.Error(), "assets path not found")

		c.options.AssetsPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("storage path error", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}

		c.options.StoragePath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("originals path not found", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}
		c.options.OriginalsPath = ""

		err := c.CreateDirectories()
		if err == nil {
			t.Fatal("error expected")
		}

		assert.Contains(t, err.Error(), "originals path not found")

		c.options.OriginalsPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("import path not found", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}
		c.options.ImportPath = ""

		err := c.CreateDirectories()
		if err == nil {
			t.Fatal("error expected")
		}

		assert.Contains(t, err.Error(), "import path not found")

		c.options.ImportPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("sidecar path error", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}

		c.options.SidecarPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("cache path error", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}

		c.options.CachePath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("config path error", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}

		c.options.ConfigPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})

	t.Run("temp path error", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.GenerateToken(8),
		}

		c.options.TempPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})
}
*/

func TestConfig_PIDFilename2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/photoprism.pid", c.PIDFilename())
	c.options.PIDFilename = "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.pid"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.pid", c.PIDFilename())
}

func TestConfig_LogFilename2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/photoprism.log", c.LogFilename())
	c.options.LogFilename = "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.log"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.log", c.LogFilename())
}

func TestConfig_OriginalsPath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/originals", c.OriginalsPath())
	c.options.OriginalsPath = ""
	if s := c.OriginalsPath(); s != "" && s != "/photoprism/originals" {
		t.Errorf("unexpected originals path: %s", s)
	}
}

func TestConfig_OriginalsDeletable(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.Settings().Features.Delete = true
	c.options.ReadOnly = false

	assert.True(t, c.OriginalsDeletable())
}

func TestConfig_ImportPath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/import", c.ImportPath())
	c.options.ImportPath = ""
	if s := c.ImportPath(); s != "" && s != "/photoprism/import" {
		t.Errorf("unexpected import path: %s", s)
	}
}

func TestConfig_AssetsPath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets", c.AssetsPath())
	c.options.AssetsPath = ""
	if s := c.AssetsPath(); s != "" && s != "/opt/photoprism/assets" {
		t.Errorf("unexpected assets path: %s", s)
	}
}

func TestConfig_MysqlBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.MysqlBin(), "mysql")
}

func TestConfig_MysqldumpBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.MysqldumpBin(), "mysqldump")
}

func TestConfig_SqliteBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.SqliteBin(), "sqlite")
}

func TestConfig_SettingsYamlDefaults(t *testing.T) {
	c := NewConfig(CliTestContext())
	name1 := c.SettingsYamlDefaults(c.SettingsYaml())
	t.Logf("(1) DefaultsYaml: %s", c.DefaultsYaml())
	t.Logf("(1) SettingsYaml: %s", c.SettingsYaml())
	t.Logf("(1) SettingsYamlDefaults: %s", name1)
	assert.Equal(t, c.SettingsYaml(), name1)
	c.options.ConfigPath = "/tmp/012345678ABC"
	c.options.DefaultsYaml = "testdata/etc/defaults.yml"
	name2 := c.SettingsYamlDefaults("")
	t.Logf("(2) DefaultsYaml: %s", c.DefaultsYaml())
	t.Logf("(2) SettingsYaml: %s", c.SettingsYaml())
	t.Logf("(2) SettingsYamlDefaults: %s", name2)
	assert.True(t, strings.HasSuffix(name2, "testdata/etc/settings.yml"))
	name3 := c.SettingsYamlDefaults(c.SettingsYaml())
	t.Logf("(3) DefaultsYaml: %s", c.DefaultsYaml())
	t.Logf("(3) SettingsYaml: %s", c.SettingsYaml())
	t.Logf("(3) SettingsYamlDefaults: %s", name3)
	assert.True(t, strings.HasSuffix(name3, "testdata/etc/settings.yml"))
	assert.NotEqual(t, c.SettingsYaml(), name1)
	assert.NotEqual(t, c.SettingsYaml(), name3)
}
