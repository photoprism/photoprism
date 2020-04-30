package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/fs"
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

	testDataPath := testDataPath(assetsPath)

	c := &Params{
		Debug:          true,
		Public:         true,
		ReadOnly:       false,
		DetectNSFW:     true,
		UploadNSFW:     false,
		DarktableBin:   "/usr/bin/darktable-cli",
		ExifToolBin:    "/usr/bin/exiftool",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		TempPath:       testDataPath + "/temp",
		DatabaseDriver: "mysql",
		DatabaseDsn:    "photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true",
	}

	return c
}

// NewTestParamsError inits invalid params used for testing
func NewTestParamsError() *Params {
	assetsPath := fs.Abs("../..")

	testDataPath := testDataPath("../../assets")

	c := &Params{
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		TempPath:       testDataPath + "/temp",
		DatabaseDriver: "mysql",
		DatabaseDsn:    "photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true",
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

	c := &Config{params: NewTestParams()}
	c.initSettings()

	if err := c.Init(context.Background()); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	c.DropTables()

	// Make sure changes have been written to disk.
	time.Sleep(250 * time.Millisecond)

	c.MigrateDb()

	// Make sure changes have been written to disk.
	time.Sleep(250 * time.Millisecond)

	entity.CreateTestFixtures(c.Db())

	// TODO: Remove when new test fixtures are ready
	c.ImportSQL(c.ExamplesPath() + "/fixtures.sql")

	thumb.JpegQuality = c.ThumbQuality()
	thumb.PreRenderSize = c.ThumbSize()
	thumb.MaxRenderSize = c.ThumbLimit()
	thumb.Filter = c.ThumbFilter()

	return c
}

// NewTestErrorConfig inits invalid config used for testing
func NewTestErrorConfig() *Config {
	c := &Config{params: NewTestParamsError()}
	err := c.Init(context.Background())
	if err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	c.MigrateDb()
	return c
}

// CliTestContext returns example cli config for testing
func CliTestContext() *cli.Context {
	config := NewTestParams()

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("config-file", config.ConfigFile, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("import-path", config.OriginalsPath, "doc")
	globalSet.String("temp-path", config.OriginalsPath, "doc")
	globalSet.String("cache-path", config.OriginalsPath, "doc")
	globalSet.String("darktable-cli", config.DarktableBin, "doc")
	globalSet.Bool("detect-nsfw", config.DetectNSFW, "doc")

	app := cli.NewApp()
	app.Version = "test"

	c := cli.NewContext(app, globalSet, nil)

	LogError(c.Set("config-file", config.ConfigFile))
	LogError(c.Set("assets-path", config.AssetsPath))
	LogError(c.Set("originals-path", config.OriginalsPath))
	LogError(c.Set("import-path", config.ImportPath))
	LogError(c.Set("temp-path", config.TempPath))
	LogError(c.Set("cache-path", config.CachePath))
	LogError(c.Set("darktable-cli", config.DarktableBin))
	LogError(c.Set("detect-nsfw", "true"))

	return c
}

// RemoveTestData deletes files in import, export, originals and cache folders
func (c *Config) RemoveTestData(t *testing.T) {
	LogError(os.RemoveAll(c.ImportPath()))
	LogError(os.RemoveAll(c.TempPath()))
	LogError(os.RemoveAll(c.OriginalsPath()))
	LogError(os.RemoveAll(c.CachePath()))
}

// DownloadTestData downloads test data from photoprism.org server
func (c *Config) DownloadTestData(t *testing.T) {
	if fs.FileExists(TestDataZip) {
		hash := fs.Hash(TestDataZip)

		if hash != TestDataHash {
			LogError(os.Remove(TestDataZip))
			t.Logf("removed outdated test data zip file (fingerprint %s)\n", hash)
		}
	}

	if !fs.FileExists(TestDataZip) {
		fmt.Printf("downloading latest test data zip file from %s\n", TestDataURL)

		if err := fs.Download(TestDataZip, TestDataURL); err != nil {
			fmt.Printf("Download failed: %s\n", err.Error())
		}
	}
}

// UnzipTestData in default test folder
func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := fs.Unzip(TestDataZip, testDataPath(c.AssetsPath())); err != nil {
		t.Logf("could not unzip test data: %s\n", err.Error())
	}
}

// InitializeTestData using testing constant
func (c *Config) InitializeTestData(t *testing.T) {
	t.Log("initializing test data")

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}
