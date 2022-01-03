package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/urfave/cli"
)

// Download URL and ZIP hash for test files.
const (
	TestDataZip  = "/tmp/photoprism/testdata.zip"
	TestDataURL  = "https://dl.photoprism.app/qa/testdata.zip"
	TestDataHash = "be394d5bee8a5634d415e9e0663eef20b5604510" // sha1sum
)

var testConfig *Config
var testConfigOnce sync.Once
var testConfigMutex sync.Mutex
var testDataMutex sync.Mutex

func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

// NewTestOptions returns valid config options for tests.
func NewTestOptions() *Options {
	assetsPath := fs.Abs("../../assets")
	storagePath := fs.Abs("../../storage")
	testDataPath := filepath.Join(storagePath, "testdata")

	dbDriver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	dbDsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	// Config example for MySQL / MariaDB:
	//   dbDriver = MySQL,
	//   dbDsn = "photoprism:photoprism@tcp(mariadb:4001)/photoprism?parseTime=true",

	if dbDriver == "test" || dbDriver == "sqlite" || dbDriver == "" || dbDsn == "" {
		dbDriver = SQLite3
		dbDsn = ".test.db"
	}

	c := &Options{
		Name:           "PhotoPrism",
		Version:        "0.0.0",
		Copyright:      "(c) 2018-2022 Michael Mayer",
		Test:           true,
		Debug:          true,
		Public:         true,
		Experimental:   true,
		ReadOnly:       false,
		DetectNSFW:     true,
		UploadNSFW:     false,
		AssetsPath:     assetsPath,
		AutoIndex:      -1,
		AutoImport:     7200,
		StoragePath:    testDataPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		TempPath:       testDataPath + "/temp",
		ConfigPath:     testDataPath + "/config",
		SidecarPath:    testDataPath + "/sidecar",
		DatabaseDriver: dbDriver,
		DatabaseDsn:    dbDsn,
		AdminPassword:  "photoprism",
	}

	return c
}

// NewTestOptionsError returns invalid config options for tests.
func NewTestOptionsError() *Options {
	assetsPath := fs.Abs("../..")
	testDataPath := fs.Abs("../../storage/testdata")

	c := &Options{
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		StoragePath:    testDataPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		TempPath:       testDataPath + "/temp",
		DatabaseDriver: SQLite3,
		DatabaseDsn:    ".test-error.db",
	}

	return c
}

func SetNewTestConfig() {
	testConfig = NewTestConfig()
}

// TestConfig returns the existing test config instance or creates a new instance and returns it.
func TestConfig() *Config {
	testConfigOnce.Do(SetNewTestConfig)

	return testConfig
}

// NewTestConfig returns a valid test config.
func NewTestConfig() *Config {
	defer log.Debug(capture.Time(time.Now(), "config: new test config created"))

	testConfigMutex.Lock()
	defer testConfigMutex.Unlock()

	c := &Config{
		options: NewTestOptions(),
		token:   rnd.Token(8),
	}

	s := NewSettings(c)

	if err := os.MkdirAll(c.ConfigPath(), os.ModePerm); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	if err := s.Save(filepath.Join(c.ConfigPath(), "settings.yml")); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	if err := c.Init(); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	c.InitTestDb()

	thumb.SizePrecached = c.ThumbSizePrecached()
	thumb.SizeUncached = c.ThumbSizeUncached()
	thumb.Filter = c.ThumbFilter()
	thumb.JpegQuality = c.JpegQuality()

	return c
}

// NewTestErrorConfig returns an invalid test config.
func NewTestErrorConfig() *Config {
	c := &Config{options: NewTestOptionsError()}

	return c
}

// CliTestContext returns a CLI context for testing.
func CliTestContext() *cli.Context {
	config := NewTestOptions()

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("storage-path", config.StoragePath, "doc")
	globalSet.String("backup-path", config.StoragePath, "doc")
	globalSet.String("sidecar-path", config.SidecarPath, "doc")
	globalSet.String("config-file", config.ConfigFile, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("import-path", config.OriginalsPath, "doc")
	globalSet.String("temp-path", config.OriginalsPath, "doc")
	globalSet.String("cache-path", config.OriginalsPath, "doc")
	globalSet.String("darktable-cli", config.DarktableBin, "doc")
	globalSet.String("darktable-blacklist", config.DarktableBlacklist, "doc")
	globalSet.String("admin-password", config.DarktableBin, "doc")
	globalSet.Bool("detect-nsfw", config.DetectNSFW, "doc")
	globalSet.Int("auto-index", config.AutoIndex, "doc")
	globalSet.Int("auto-import", config.AutoImport, "doc")

	app := cli.NewApp()
	app.Version = "0.0.0"

	c := cli.NewContext(app, globalSet, nil)

	LogError(c.Set("storage-path", config.StoragePath))
	LogError(c.Set("backup-path", config.BackupPath))
	LogError(c.Set("sidecar-path", config.SidecarPath))
	LogError(c.Set("config-file", config.ConfigFile))
	LogError(c.Set("assets-path", config.AssetsPath))
	LogError(c.Set("originals-path", config.OriginalsPath))
	LogError(c.Set("import-path", config.ImportPath))
	LogError(c.Set("temp-path", config.TempPath))
	LogError(c.Set("cache-path", config.CachePath))
	LogError(c.Set("darktable-cli", config.DarktableBin))
	LogError(c.Set("darktable-blacklist", "raf,cr3"))
	LogError(c.Set("admin-password", config.AdminPassword))
	LogError(c.Set("detect-nsfw", "true"))
	LogError(c.Set("auto-index", strconv.Itoa(config.AutoIndex)))
	LogError(c.Set("auto-import", strconv.Itoa(config.AutoImport)))

	return c
}

// RemoveTestData deletes files in import, export, originals, and cache folders.
func (c *Config) RemoveTestData(t *testing.T) {
	if err := os.RemoveAll(c.ImportPath()); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(c.TempPath()); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(c.OriginalsPath()); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(c.CachePath()); err != nil {
		t.Fatal(err)
	}
}

// DownloadTestData downloads the test files from the file server.
func (c *Config) DownloadTestData(t *testing.T) {
	if fs.FileExists(TestDataZip) {
		hash := fs.Hash(TestDataZip)

		if hash != TestDataHash {
			if err := os.Remove(TestDataZip); err != nil {
				t.Fatalf("config: %s", err.Error())
			}

			t.Logf("config: removed outdated test data zip file (fingerprint %s)", hash)
		}
	}

	if !fs.FileExists(TestDataZip) {
		t.Logf("config: downloading latest test data zip file from %s", TestDataURL)

		if err := fs.Download(TestDataZip, TestDataURL); err != nil {
			t.Fatalf("config: test data download failed: %s", err.Error())
		}
	}
}

// UnzipTestData extracts tests files from the zip archive.
func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := fs.Unzip(TestDataZip, c.StoragePath()); err != nil {
		t.Fatalf("config: could not unzip test data: %s", err.Error())
	}
}

// InitializeTestData resets the test file directory.
func (c *Config) InitializeTestData(t *testing.T) {
	testDataMutex.Lock()
	defer testDataMutex.Unlock()

	start := time.Now()

	c.RemoveTestData(t)
	c.DownloadTestData(t)
	c.UnzipTestData(t)

	t.Logf("config: initialized test data [%s]", time.Since(start))
}
