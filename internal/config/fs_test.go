package config

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/assert"
)

func TestConfig_FindExecutable(t *testing.T) {
	assert.Equal(t, "", findExecutable("yyy", "xxx"))
}

func TestConfig_SidecarJson(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.ExifToolJson())
	assert.Equal(t, c.DisableExifTool(), !c.ExifToolJson())

	c.options.DisableExifTool = true

	assert.Equal(t, false, c.ExifToolJson())
	assert.Equal(t, c.DisableExifTool(), !c.ExifToolJson())
}

func TestConfig_SidecarYaml(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.BackupYaml())
	assert.Equal(t, c.DisableBackups(), !c.BackupYaml())

	c.options.DisableBackups = true

	assert.Equal(t, false, c.BackupYaml())
	assert.Equal(t, c.DisableBackups(), !c.BackupYaml())
}

func TestConfig_SidecarPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.SidecarPath(), "testdata/sidecar")
	c.options.SidecarPath = ".photoprism"
	assert.Equal(t, ".photoprism", c.SidecarPath())
	c.options.SidecarPath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/sidecar", c.SidecarPath())
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
	assert.Equal(t, "/usr/bin/ffmpeg", c.FFmpegBin())
}

func TestConfig_TempPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/temp", c.TempPath())
	c.options.TempPath = ""
	assert.Equal(t, "/tmp/photoprism", c.TempPath())
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
			options: NewTestOptions(),
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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
			token:   rnd.Token(8),
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

func TestConfig_ConfigFile2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.ConfigFile(), "options.yml")
	c.options.ConfigFile = "/go/src/github.com/photoprism/photoprism/internal/config/testdata/config.yml"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/testdata/config.yml", c.ConfigFile())
}

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
