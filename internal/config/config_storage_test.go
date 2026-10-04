package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestConfig_FindBin(t *testing.T) {
	assert.Equal(t, "", FindBin("yyy123", "xxx123"))
	assert.Equal(t, "", FindBin("yyy123", "sh"))
	assert.Equal(t, "/usr/bin/sh", FindBin("sh", "yyy123"))
	assert.Equal(t, "/usr/bin/sh", FindBin("", "sh"))
	assert.Equal(t, "/usr/bin/sh", FindBin("", "", "sh"))
	assert.Equal(t, "/usr/bin/sh", FindBin("", "yyy123", "sh"))
	assert.Equal(t, "/usr/bin/sh", FindBin("sh", "bash"))
	assert.Equal(t, "/usr/bin/bash", FindBin("bash", "sh"))
}

func TestConfig_SidecarPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.SidecarPath(), "testdata/sidecar")
	c.options.SidecarPath = ".photoprism"
	assert.Equal(t, ".photoprism", c.SidecarPath())
	c.options.SidecarPath = ""
	assert.Equal(t, ProjectRoot+"/storage/testdata/sidecar", c.SidecarPath())
}

func TestConfig_SidecarYaml(t *testing.T) {
	c := NewConfig(NewTestContext(nil))

	// t.Logf("c.options.DisableBackups = %t", c.options.DisableBackups)
	// t.Logf("c.options.SidecarYaml = %t", c.options.SidecarYaml)

	assert.Equal(t, true, c.SidecarYaml())
	assert.Equal(t, c.DisableBackups(), !c.SidecarYaml())

	c.options.DisableBackups = true

	assert.Equal(t, false, c.SidecarYaml())
	assert.Equal(t, c.DisableBackups(), !c.SidecarYaml())

	c.options.DisableBackups = false
	c.options.SidecarYaml = true

	assert.Equal(t, true, c.SidecarYaml())
	assert.Equal(t, c.DisableBackups(), !c.SidecarYaml())
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
	assert.Contains(t, c.UsersStoragePath(), fs.UsersDir)
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

func TestConfig_UserUploadBatchDir(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.Options().StoragePath = t.TempDir()

	t.Run("Success", func(t *testing.T) {
		dir, err := c.UserUploadBatchDir("urjult03ceelhw6k", "sess/abc")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(c.UsersStoragePath(), "urjult03ceelhw6k", "upload", "sessabc"), dir)
		assert.NoDirExists(t, dir)
		assert.NoDirExists(t, filepath.Join(c.UsersStoragePath(), "urjult03ceelhw6k"))
		created, err := c.UserUploadBatchPath("urjult03ceelhw6k", "sess/abc")
		require.NoError(t, err)
		assert.Equal(t, created, dir)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		for _, uid := range []string{"", "etaetyget", "../urjult03ceelhw6k"} {
			dir, err := c.UserUploadBatchDir(uid, "sessabcdefgh1234567")
			assert.Error(t, err, uid)
			assert.Equal(t, "", dir)
		}
		for _, batch := range []string{"", "../", "/.", strings.Repeat("a", clean.LengthLimit+1)} {
			dir, err := c.UserUploadBatchDir("urjult03ceelhw6k", batch)
			assert.Error(t, err, batch)
			assert.Equal(t, "", dir)
		}
	})
}

func TestConfig_UserUploadBatchPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.Options().StoragePath = t.TempDir()

	t.Run("Success", func(t *testing.T) {
		dir, err := c.UserUploadBatchPath("urjult03ceelhw6k", "sessabcdefgh1234567")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(c.UserStoragePath("urjult03ceelhw6k"), "upload", "sessabcdefgh1234567"), dir)
		assert.DirExists(t, dir)
	})
	t.Run("CleanedName", func(t *testing.T) {
		dir, err := c.UserUploadBatchPath("urjult03ceelhw6k", "sess/abc")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(c.UserStoragePath("urjult03ceelhw6k"), "upload", "sessabc"), dir)
	})
	t.Run("StorageUnavailable", func(t *testing.T) {
		// A file in place of the users folder makes the user's storage folder unavailable.
		s := NewConfig(CliTestContext())
		s.Options().StoragePath = t.TempDir()
		require.NoError(t, os.WriteFile(s.UsersStoragePath(), []byte("file"), fs.ModeFile))
		dir, err := s.UserUploadBatchPath("urjult03ceelhw6k", "sessabcdefgh1234567")
		assert.Error(t, err)
		assert.Equal(t, "", dir)
		dir, err = s.UserUploadPath("urjult03ceelhw6k", "")
		assert.Error(t, err)
		assert.Equal(t, "", dir)
	})
	t.Run("EmptyName", func(t *testing.T) {
		for _, batch := range []string{"", "../", "/.", strings.Repeat("a", clean.LengthLimit+1)} {
			dir, err := c.UserUploadBatchPath("urjult03ceelhw6k", batch)
			assert.Error(t, err, batch)
			assert.Equal(t, "", dir)
		}
	})
	t.Run("InvalidUser", func(t *testing.T) {
		dir, err := c.UserUploadBatchPath("etaetyget", "sessabcdefgh1234567")
		assert.Error(t, err)
		assert.Equal(t, "", dir)
	})
}

func TestConfig_WebStoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.WebStoragePath(), fs.WebDir)
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

func TestConfig_FFprobeBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, strings.Contains(c.FFprobeBin(), "/bin/ffprobe"))
}

func TestConfig_YtDlpBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, strings.Contains(c.YtDlpBin(), "/bin/yt-dlp"))
}

func TestConfig_TempPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	// Reset the cached TempPath() so the result does not depend on execution order.
	tempPath = ""
	t.Cleanup(func() { tempPath = "" })

	// With TempPath configured (default test config), it is returned unchanged.
	want := ProjectRoot + "/storage/testdata/temp"
	assert.Equal(t, want, c.tempPath())

	// TempPath() caches the first resolved value and keeps returning it even after
	// the configured option is cleared.
	assert.Equal(t, want, c.TempPath())
	c.options.TempPath = ""
	assert.Equal(t, want, c.TempPath(), "cached TempPath() should not change")

	// Without a configured option, tempPath() falls back to a "photoprism_" subdirectory
	// of the OS temp dir, which is not necessarily "/tmp" (e.g. when TMPDIR is set).
	wantPrefix := filepath.Join(os.TempDir(), "photoprism_")

	d1 := c.tempPath()
	if d1 == "" {
		t.Fatal("temp path is empty")
	}
	assert.True(t, strings.HasPrefix(d1, wantPrefix), "want prefix %q, got %q", wantPrefix, d1)

	// The fallback must be stable across repeated calls.
	assert.Equal(t, d1, c.tempPath(), "fallback temp paths should match")
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
	assert.Equal(t, ProjectRoot+"/storage/testdata/cache", c.CachePath())
	c.options.CachePath = ""
	assert.Equal(t, ProjectRoot+"/storage/testdata/cache", c.CachePath())
}

func TestConfig_SettingsYaml(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Contains(t, c.SettingsYaml(), "settings.yml")
	})
	t.Run("PreferYamlExtension", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempDir := t.TempDir()
		c.options.ConfigPath = tempDir

		yamlPath := filepath.Join(tempDir, "settings"+fs.ExtYaml)
		if err := os.WriteFile(yamlPath, []byte("ui:\n"), fs.ModeFile); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		assert.Equal(t, yamlPath, c.SettingsYaml())
	})
}

func TestConfig_HubConfigFile(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Contains(t, c.HubConfigFile(), "hub.yml")
	})
	t.Run("PreferYamlExtension", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempDir := t.TempDir()
		c.options.ConfigPath = tempDir

		yamlPath := filepath.Join(tempDir, "hub"+fs.ExtYaml)
		if err := os.WriteFile(yamlPath, []byte("host: example\n"), fs.ModeFile); err != nil {
			t.Fatalf("write %s: %v", yamlPath, err)
		}

		assert.Equal(t, yamlPath, c.HubConfigFile())
	})
}

func TestConfig_StoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ProjectRoot+"/storage/testdata", c.StoragePath())
	c.options.StoragePath = ""
	assert.Equal(t, ProjectRoot+"/storage/testdata/originals/.photoprism/storage", c.StoragePath())
}

func TestConfig_TestdataPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ProjectRoot+"/storage/testdata/testdata", c.TestdataPath())
}

func TestConfig_AlbumsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	// The default albums path has changed from “albums/” to “backup/albums/”.
	//
	// If this test fails, please manually move “albums” to the “backup” folder
	// in the “storage/testdata” directory within your development environment:
	// https://github.com/photoprism/photoprism/discussions/4520
	assert.Equal(t, ProjectRoot+"/storage/testdata/backup/albums", c.BackupAlbumsPath())
}

func TestConfig_OriginalsAlbumsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ProjectRoot+"/storage/testdata/originals/albums", c.OriginalsAlbumsPath())
}

// TestConfig_CreateDirectories checks storage setup and marker placement.
func TestConfig_CreateDirectories(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()

		c := &Config{
			options: NewTestOptions("config"),
			token:   rnd.Base36(8),
			cache:   gc.New(time.Second, time.Minute),
		}

		assert.NoError(t, c.CreateDirectories())
	})
	t.Run("IdenticalPaths", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()

		c := &Config{
			options: NewTestOptions("config"),
			token:   rnd.Base36(8),
			cache:   gc.New(time.Second, time.Minute),
		}

		c.options.StoragePath = "./testdata"
		c.options.OriginalsPath = "./testdata"

		assert.Error(t, c.CreateDirectories())
	})
	t.Run("StorageMarker", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())

		require.NoError(t, c.CreateDirectories())
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(c.StoragePath(), ".ppstorage")))
		assert.NoFileExists(t, filepath.Join(c.OriginalsPath(), ".ppstorage"))
		assert.NoFileExists(t, filepath.Join(c.ImportPath(), ".ppstorage"))
	})
}

/* TODO
	--- FAIL: TestConfig_CreateDirectories2 (0.00s)
    --- FAIL: TestConfig_CreateDirectories2/asset_path_not_found (0.00s)
        fs_test.go:142: error expected

func TestConfig_CreateDirectories2(t *testing.T) {
	t.Run("AssetPathNotFound", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
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
	t.Run("StoragePathError", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
		}

		c.options.StoragePath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})
	t.Run("OriginalsPathNotFound", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
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
	t.Run("ImportPathNotFound", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
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
	t.Run("SidecarPathError", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
		}

		c.options.SidecarPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})
	t.Run("CachePathError", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
		}

		c.options.CachePath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})
	t.Run("ConfigPathError", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
		}

		c.options.ConfigPath = "/-*&^%$#@!`~"
		err2 := c.CreateDirectories()

		if err2 == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err2.Error(), "check config and permissions")
	})
	t.Run("TempPathError", func(t *testing.T) {
		testConfigMutex.Lock()
		defer testConfigMutex.Unlock()
		c := &Config{
			options: NewTestOptions(),
			token:   rnd.Base36(8),
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
	assert.Equal(t, ProjectRoot+"/storage/testdata/photoprism.pid", c.PIDFilename())
	c.options.PIDFilename = ProjectRoot + "/internal/config/testdata/test.pid"
	assert.Equal(t, ProjectRoot+"/internal/config/testdata/test.pid", c.PIDFilename())
}

func TestConfig_LogFilename2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ProjectRoot+"/storage/testdata/photoprism.log", c.LogFilename())
	c.options.LogFilename = ProjectRoot + "/internal/config/testdata/test.log"
	assert.Equal(t, ProjectRoot+"/internal/config/testdata/test.log", c.LogFilename())
}

func TestConfig_OriginalsPath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ProjectRoot+"/storage/testdata/originals", c.OriginalsPath())
	c.options.OriginalsPath = ""
	if s := c.OriginalsPath(); s != "" && s != "/photoprism/originals" {
		t.Errorf("unexpected originals path: %s", s)
	}
}

func TestConfig_OriginalsDeletable(t *testing.T) {
	c := TestConfig()

	c.Settings().Features.Delete = true
	c.Options().ReadOnly = false
	c.AssertTestData(t)

	assert.True(t, c.OriginalsDeletable())
}

func TestConfig_ImportAllow(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.ImportAllow = "jpg, PNG,pdf"

	assert.Equal(t, "jpg, pdf, png", c.ImportAllow().String())

	c.options.ImportAllow = ""

	assert.Len(t, c.ImportAllow(), 0)
	assert.Equal(t, "", c.ImportAllow().String())
}

func TestConfig_AssetsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasSuffix(c.AssetsPath(), "/assets"))
	assert.Equal(t, ProjectRoot+"/assets", c.AssetsPath())
	c.options.AssetsPath = ""
	if s := c.AssetsPath(); s != "" && s != "/opt/photoprism/assets" {
		t.Errorf("unexpected assets path: %s", s)
	}
}

func TestConfig_ProfilesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.ProfilesPath()
	assert.True(t, strings.HasSuffix(result, "/assets/profiles"))
	assert.True(t, fs.PathExists(result))
}

func TestConfig_IccProfilesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.IccProfilesPath()
	assert.True(t, strings.HasSuffix(result, "/assets/profiles/icc"))
}

func TestConfig_CustomAssetsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.CustomAssetsPath())
}

func TestConfig_MariadbBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.MariadbBin(), "mariadb")
}

func TestConfig_MariadbDumpBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.MariadbDumpBin(), "mariadb-dump")
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

func TestDefaultsYamlResolution(t *testing.T) {
	t.Run("ExplicitFlag", func(t *testing.T) {
		ctx := CliTestContext()
		file := filepath.Join(t.TempDir(), "explicit-defaults.yml")
		require.NoError(t, os.WriteFile(file, []byte("Test: true"), fs.ModeFile))
		require.NoError(t, ctx.Set("defaults-yaml", file))
		got := defaultsYaml(ctx)
		require.Equal(t, fs.Abs(file), got)
	})
	t.Run("ConfigFallback", func(t *testing.T) {
		ctx := CliTestContext()
		configDir := filepath.Join(t.TempDir(), "cfg")
		require.NoError(t, os.MkdirAll(configDir, fs.ModeDir))
		file := filepath.Join(configDir, "defaults.yml")
		require.NoError(t, os.WriteFile(file, []byte("SiteUrl: https://example.com"), fs.ModeFile))
		require.NoError(t, ctx.Set("defaults-yaml", ""))
		require.NoError(t, ctx.Set("config-path", configDir))
		got := defaultsYaml(ctx)
		require.Equal(t, fs.Abs(file), got)
	})
	t.Run("MissingReturnsEmpty", func(t *testing.T) {
		ctx := CliTestContext()
		require.NoError(t, ctx.Set("defaults-yaml", filepath.Join(t.TempDir(), "missing.yml")))
		require.NoError(t, ctx.Set("config-path", t.TempDir()))
		require.Equal(t, "", defaultsYaml(ctx))
	})
}

func TestConfig_StorageFree(t *testing.T) {
	c := NewConfig(CliTestContext())

	// Pin the shared default so the "use default" cases are deterministic, and
	// restore both the option and the package-level threshold afterwards.
	origOpt := c.options.StorageFree
	origPct := disk.StorageLowPct
	disk.StorageLowPct = 1.0
	t.Cleanup(func() {
		c.options.StorageFree = origOpt
		disk.StorageLowPct = origPct
	})

	t.Run("Disabled", func(t *testing.T) {
		c.options.StorageFree = -1
		assert.Equal(t, float64(-1), c.StorageFree())
	})
	t.Run("NegativeDisables", func(t *testing.T) {
		c.options.StorageFree = -12.5
		assert.Equal(t, float64(-1), c.StorageFree())
	})
	t.Run("ZeroUsesDefault", func(t *testing.T) {
		c.options.StorageFree = 0
		assert.Equal(t, 1.0, c.StorageFree())
	})
	t.Run("HundredOrMoreUsesDefault", func(t *testing.T) {
		c.options.StorageFree = 100
		assert.Equal(t, 1.0, c.StorageFree())
		c.options.StorageFree = 250
		assert.Equal(t, 1.0, c.StorageFree())
	})
	t.Run("CustomThreshold", func(t *testing.T) {
		c.options.StorageFree = 5
		assert.Equal(t, 5.0, c.StorageFree())
		c.options.StorageFree = 99
		assert.Equal(t, 99.0, c.StorageFree())
	})
}

// stubCaseDetection replaces the case detection of the storage and originals paths until the test ends.
func stubCaseDetection(t *testing.T, storage, originals bool, originalsErr, storageErr error) {
	t.Helper()

	mode, prevStorage, prevOriginals := fs.GetCaseMode(), storageCaseInsensitive, originalsCaseInsensitive

	t.Cleanup(func() {
		fs.RestoreCaseMode(mode)
		storageCaseInsensitive, originalsCaseInsensitive = prevStorage, prevOriginals
	})

	storageCaseInsensitive = func(string) (bool, error) { return storage, storageErr }
	originalsCaseInsensitive = func(string) (bool, error) { return originals, originalsErr }
}

func TestConfig_CaseInsensitive(t *testing.T) {
	c := NewMinimalTestConfig(t.TempDir())
	require.NoError(t, os.MkdirAll(c.StoragePath(), fs.ModeDir))

	insensitive, err := c.CaseInsensitive()
	require.NoError(t, err)
	assert.Equal(t, insensitive, fileExistsAfterCaseSwap(t, c.StoragePath()))
}

func TestConfig_OriginalsCaseInsensitive(t *testing.T) {
	c := NewMinimalTestConfig(t.TempDir())
	require.NoError(t, os.MkdirAll(c.OriginalsPath(), fs.ModeDir))

	_, err := c.OriginalsCaseInsensitive()
	assert.Error(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(c.OriginalsPath(), "IMG_1.jpg"), []byte("x"), fs.ModeFile))

	insensitive, err := c.OriginalsCaseInsensitive()
	assert.NoError(t, err)
	assert.Equal(t, fileExistsAfterCaseSwap(t, c.OriginalsPath()), insensitive)
}

// fileExistsAfterCaseSwap creates a file in dir and reports whether it is found with its name in another case.
func fileExistsAfterCaseSwap(t *testing.T, dir string) bool {
	t.Helper()
	fileName := filepath.Join(dir, "Case_Test.tmp")
	require.NoError(t, os.WriteFile(fileName, []byte("x"), fs.ModeFile))
	t.Cleanup(func() { _ = os.Remove(fileName) })
	return fs.FileExists(filepath.Join(dir, "cASE_tEST.TMP"))
}

func TestConfig_InitCaseModeCalled(t *testing.T) {
	// Init and InitCore return the error of the storage test before they connect to a database.
	for name, initFn := range map[string]func(*Config) error{"Init": (*Config).Init, "InitCore": (*Config).InitCore} {
		t.Run(name, func(t *testing.T) {
			stubCaseDetection(t, false, false, nil, errors.New("storage not writable"))
			c := NewMinimalTestConfig(t.TempDir())
			assert.ErrorContains(t, initFn(c), "storage not writable")
		})
	}
}

func TestConfig_InitCaseMode(t *testing.T) {
	// The host may have no case-insensitive file system, so the detection is stubbed and a lookup in
	// case-insensitive mode misses the lowercase names of the case-sensitive temp folders.
	c := NewMinimalTestConfig(t.TempDir())
	originals, storage := c.OriginalsPath(), c.StoragePath()

	if insensitive, err := fs.CaseInsensitive(t.TempDir()); err != nil {
		t.Fatal(err)
	} else if insensitive {
		t.Skip("requires a case-sensitive file system")
	}

	for name, dir := range map[string]string{"IMG_1.raw": originals, "img_1.jpg": originals, "IMG_2.raw": storage, "img_2.jpg": storage} {
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(name), fs.ModeFile))
	}

	findOriginals := func() string { return fs.ImageJpeg.Find(filepath.Join(originals, "IMG_1.raw"), false) }
	findStorage := func() string { return fs.ImageJpeg.Find(filepath.Join(storage, "IMG_2.raw"), false) }

	t.Run("OriginalsInsensitive", func(t *testing.T) {
		stubCaseDetection(t, false, true, nil, nil)
		require.NoError(t, c.initCaseMode())
		assert.Equal(t, "", findOriginals())
		assert.Equal(t, filepath.Join(storage, "img_2.jpg"), findStorage())
	})
	t.Run("StorageInsensitive", func(t *testing.T) {
		stubCaseDetection(t, true, false, nil, nil)
		require.NoError(t, c.initCaseMode())
		assert.Equal(t, filepath.Join(originals, "img_1.jpg"), findOriginals())
		assert.Equal(t, "", findStorage())
	})
	t.Run("CaseSensitive", func(t *testing.T) {
		stubCaseDetection(t, false, false, nil, nil)
		require.NoError(t, c.initCaseMode())
		assert.Equal(t, filepath.Join(originals, "img_1.jpg"), findOriginals())
		assert.Equal(t, filepath.Join(storage, "img_2.jpg"), findStorage())
	})
	t.Run("OriginalsUnknown", func(t *testing.T) {
		stubCaseDetection(t, true, false, errors.New("no file name with ASCII letters within reach"), nil)
		fs.SetCaseDir(originals, false)
		require.NoError(t, c.initCaseMode())
		assert.Equal(t, "", findOriginals())
		assert.Equal(t, "", findStorage())
	})
	t.Run("OriginalsUnknownStorageSensitive", func(t *testing.T) {
		stubCaseDetection(t, false, true, errors.New("no file name with ASCII letters within reach"), nil)
		require.NoError(t, c.initCaseMode())
		assert.Equal(t, filepath.Join(originals, "img_1.jpg"), findOriginals())
		assert.Equal(t, filepath.Join(storage, "img_2.jpg"), findStorage())
	})
	t.Run("StorageError", func(t *testing.T) {
		stubCaseDetection(t, false, true, nil, errors.New("storage not writable"))
		assert.Error(t, c.initCaseMode())
		assert.Equal(t, filepath.Join(originals, "img_1.jpg"), findOriginals())
	})
}
