package config

import (
	"context"
	"flag"
	"os"
	"path/filepath"
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

// define constants used for testing the config package
const (
	TestDataZip  = "/tmp/photoprism/testdata.zip"
	TestDataURL  = "https://dl.photoprism.org/fixtures/testdata.zip"
	TestDataHash = "be394d5bee8a5634d415e9e0663eef20b5604510" // sha1sum
)

var testConfig *Config
var testConfigOnce sync.Once
var testConfigMutex sync.Mutex

func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

// NewTestParams inits valid params used for testing
func NewTestParams() *Params {
	assetsPath := fs.Abs("../../assets")
	storagePath := fs.Abs("../../storage")
	testDataPath := filepath.Join(storagePath, "testdata")

	dbDriver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	dbDsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	// Config example for MySQL / MariaDB:
	//   dbDriver = MySQL,
	//   dbDsn = "photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true",

	if dbDriver == "test" || dbDriver == "sqlite" || dbDriver == "" || dbDsn == "" {
		dbDriver = SQLite
		dbDsn = ".test.db"
	}

	c := &Params{
		Name:           "PhotoPrism",
		Version:        "0.0.0",
		Copyright:      "(c) 2018-2020 Michael Mayer",
		Debug:          true,
		Public:         true,
		ReadOnly:       false,
		DetectNSFW:     true,
		UploadNSFW:     false,
		SidecarPath:    fs.HiddenPath,
		DarktableBin:   "/usr/bin/darktable-cli",
		ExifToolBin:    "/usr/bin/exiftool",
		AssetsPath:     assetsPath,
		StoragePath:    testDataPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		TempPath:       testDataPath + "/temp",
		SettingsPath:   testDataPath + "/settings",
		DatabaseDriver: dbDriver,
		DatabaseDsn:    dbDsn,
		AdminPassword:  "photoprism",
	}

	return c
}

// NewTestParamsError inits invalid params used for testing
func NewTestParamsError() *Params {
	assetsPath := fs.Abs("../..")
	testDataPath := fs.Abs("../../storage/testdata")

	c := &Params{
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		StoragePath:    testDataPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		TempPath:       testDataPath + "/temp",
		DatabaseDriver: SQLite,
		DatabaseDsn:    ".test-error.db",
	}

	return c
}

func SetNewTestConfig() {
	testConfig = NewTestConfig()
}

// TestConfig inits the global testConfig if it was not already initialised
func TestConfig() *Config {
	testConfigOnce.Do(SetNewTestConfig)

	return testConfig
}

// NewTestConfig inits valid config used for testing
func NewTestConfig() *Config {
	defer log.Debug(capture.Time(time.Now(), "config: new test config created"))

	testConfigMutex.Lock()
	defer testConfigMutex.Unlock()

	c := &Config{
		params: NewTestParams(),
		token:  rnd.Token(8),
	}

	s := NewSettings()

	if err := os.MkdirAll(c.SettingsPath(), os.ModePerm); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	if err := s.Save(filepath.Join(c.SettingsPath(), "settings.yml")); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	c.initSettings()

	if err := c.Init(context.Background()); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	c.InitTestDb()

	thumb.Size = c.ThumbSize()
	thumb.Limit = c.ThumbSizeUncached()
	thumb.Filter = c.ThumbFilter()
	thumb.JpegQuality = c.JpegQuality()

	return c
}

// NewTestErrorConfig inits invalid config used for testing
func NewTestErrorConfig() *Config {
	c := &Config{params: NewTestParamsError()}

	c.initSettings()

	if err := c.Init(context.Background()); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	return c
}

// CliTestContext returns example cli config for testing
func CliTestContext() *cli.Context {
	config := NewTestParams()

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("storage-path", config.StoragePath, "doc")
	globalSet.String("sidecar-path", config.SidecarPath, "doc")
	globalSet.String("config-file", config.ConfigFile, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("import-path", config.OriginalsPath, "doc")
	globalSet.String("temp-path", config.OriginalsPath, "doc")
	globalSet.String("cache-path", config.OriginalsPath, "doc")
	globalSet.String("darktable-cli", config.DarktableBin, "doc")
	globalSet.String("admin-password", config.DarktableBin, "doc")
	globalSet.Bool("detect-nsfw", config.DetectNSFW, "doc")

	app := cli.NewApp()
	app.Version = "test"

	c := cli.NewContext(app, globalSet, nil)

	LogError(c.Set("storage-path", config.StoragePath))
	LogError(c.Set("sidecar-path", config.SidecarPath))
	LogError(c.Set("config-file", config.ConfigFile))
	LogError(c.Set("assets-path", config.AssetsPath))
	LogError(c.Set("originals-path", config.OriginalsPath))
	LogError(c.Set("import-path", config.ImportPath))
	LogError(c.Set("temp-path", config.TempPath))
	LogError(c.Set("cache-path", config.CachePath))
	LogError(c.Set("darktable-cli", config.DarktableBin))
	LogError(c.Set("admin-password", config.AdminPassword))
	LogError(c.Set("detect-nsfw", "true"))

	return c
}

// RemoveTestData deletes files in import, export, originals and cache folders
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

// DownloadTestData downloads test data from photoprism.org server
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

// UnzipTestData in default test folder
func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := fs.Unzip(TestDataZip, c.StoragePath()); err != nil {
		t.Fatalf("config: could not unzip test data: %s", err.Error())
	}
}

// InitializeTestData using testing constant
func (c *Config) InitializeTestData(t *testing.T) {
	defer t.Logf(capture.Time(time.Now(), "config: initialized test data"))

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}
