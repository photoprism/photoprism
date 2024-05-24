package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/urfave/cli"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/photoprism/photoprism/internal/customize"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
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

var PkgNameRegexp = regexp.MustCompile("[^a-zA-Z\\-_]+")

// NewTestOptions returns valid config options for tests.
func NewTestOptions(pkg string) *Options {
	// Find assets path.
	assetsPath := os.Getenv("PHOTOPRISM_ASSETS_PATH")
	if assetsPath == "" {
		fs.Abs("../../assets")
	}

	// Find storage path.
	storagePath := os.Getenv("PHOTOPRISM_STORAGE_PATH")
	if storagePath == "" {
		storagePath = fs.Abs("../../storage")
	}

	dataPath := filepath.Join(storagePath, "testdata")

	pkg = PkgNameRegexp.ReplaceAllString(pkg, "")
	driver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	// Config example for MySQL / MariaDB:
	//   driver = MySQL,
	//   dsn = "photoprism:photoprism@tcp(mariadb:4001)/photoprism?parseTime=true",

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dsn == "" {
		driver = SQLite3
	}

	// Set default database DSN.
	if driver == SQLite3 {
		if dsn == "" && pkg != "" {
			dsn = fmt.Sprintf("file:%s?mode=memory&cache=shared", pkg)
		} else if dsn == "" {
			dsn = SQLiteMemoryDSN
		} else if dsn != SQLiteTestDB {
			// Continue.
		} else if err := os.Remove(dsn); err == nil {
			log.Debugf("sqlite: test file %s removed", clean.Log(dsn))
		}
	}

	// Test config options.
	c := &Options{
		Name:            "PhotoPrism",
		Version:         "0.0.0",
		Copyright:       "(c) 2018-2024 PhotoPrism UG. All rights reserved.",
		Public:          true,
		Sponsor:         true,
		AuthMode:        "",
		Test:            true,
		Debug:           true,
		Trace:           false,
		Experimental:    true,
		ReadOnly:        false,
		DetectNSFW:      true,
		UploadNSFW:      false,
		ExifBruteForce:  false,
		AssetsPath:      assetsPath,
		AutoIndex:       -1,
		IndexSchedule:   DefaultIndexSchedule,
		AutoImport:      7200,
		StoragePath:     dataPath,
		CachePath:       dataPath + "/cache",
		OriginalsPath:   dataPath + "/originals",
		ImportPath:      dataPath + "/import",
		ConfigPath:      dataPath + "/config",
		SidecarPath:     dataPath + "/sidecar",
		TempPath:        dataPath + "/temp",
		BackupRetain:    DefaultBackupRetain,
		BackupSchedule:  DefaultBackupSchedule,
		DatabaseDriver:  driver,
		DatabaseDsn:     dsn,
		AdminPassword:   "photoprism",
		OriginalsLimit:  66,
		ResolutionLimit: 33,
	}

	return c
}

// NewTestOptionsError returns invalid config options for tests.
func NewTestOptionsError() *Options {
	assetsPath := fs.Abs("../..")
	dataPath := fs.Abs("../../storage/testdata")

	c := &Options{
		DarktableBin:   "/bin/darktable-cli",
		AssetsPath:     assetsPath,
		StoragePath:    dataPath,
		CachePath:      dataPath + "/cache",
		OriginalsPath:  dataPath + "/originals",
		ImportPath:     dataPath + "/import",
		TempPath:       dataPath + "/temp",
		DatabaseDriver: SQLite3,
		DatabaseDsn:    ".test-error.db",
	}

	return c
}

func SetNewTestConfig() {
	testConfig = NewTestConfig("test")
}

// TestConfig returns the existing test config instance or creates a new instance and returns it.
func TestConfig() *Config {
	testConfigOnce.Do(SetNewTestConfig)

	return testConfig
}

// NewTestConfig returns a valid test config.
func NewTestConfig(pkg string) *Config {
	defer log.Debug(capture.Time(time.Now(), "config: new test config created"))

	testConfigMutex.Lock()
	defer testConfigMutex.Unlock()

	c := &Config{
		cliCtx:  CliTestContext(),
		options: NewTestOptions(pkg),
		token:   rnd.Base36(8),
	}

	s := customize.NewSettings(c.DefaultTheme(), c.DefaultLocale())

	if err := fs.MkdirAll(c.ConfigPath()); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	if err := s.Save(filepath.Join(c.ConfigPath(), "settings.yml")); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	if err := c.Init(); err != nil {
		log.Fatalf("config: %s", err.Error())
	}

	c.RegisterDb()
	c.InitTestDb()

	thumb.SizeCached = c.ThumbSizePrecached()
	thumb.SizeOnDemand = c.ThumbSizeUncached()
	thumb.Filter = c.ThumbFilter()
	thumb.JpegQualityDefault = c.JpegQuality()

	return c
}

// NewTestErrorConfig returns an invalid test config.
func NewTestErrorConfig() *Config {
	c := &Config{options: NewTestOptionsError()}

	return c
}

// NewTestContext creates a new CLI test context with the flags and arguments provided.
func NewTestContext(args []string) *cli.Context {
	// Create new command-line app.
	app := cli.NewApp()
	app.Usage = "PhotoPrism®"
	app.Version = "test"
	app.Copyright = "(c) 2018-2024 PhotoPrism UG. All rights reserved."
	app.EnableBashCompletion = true
	app.Flags = Flags.Cli()
	app.Metadata = map[string]interface{}{
		"Name":    "PhotoPrism",
		"About":   "PhotoPrism®",
		"Edition": "ce",
		"Version": "test",
	}

	// Parse command arguments.
	flags := flag.NewFlagSet("test", 0)
	LogErr(flags.Parse(args))

	// Create and return new context.
	return cli.NewContext(app, flags, nil)
}

// CliTestContext returns a CLI context for testing.
func CliTestContext() *cli.Context {
	config := NewTestOptions("config-cli")

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.String("config-path", config.ConfigPath, "doc")
	globalSet.String("admin-password", config.DarktableBin, "doc")
	globalSet.String("storage-path", config.StoragePath, "doc")
	globalSet.String("sidecar-path", config.SidecarPath, "doc")
	globalSet.Bool("sidecar-yaml", config.SidecarYaml, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("import-path", config.OriginalsPath, "doc")
	globalSet.String("cache-path", config.OriginalsPath, "doc")
	globalSet.String("temp-path", config.OriginalsPath, "doc")
	globalSet.String("backup-path", config.StoragePath, "doc")
	globalSet.Int("backup-retain", config.BackupRetain, "doc")
	globalSet.String("backup-schedule", config.BackupSchedule, "doc")
	globalSet.String("darktable-cli", config.DarktableBin, "doc")
	globalSet.String("darktable-skip", config.DarktableSkip, "doc")
	globalSet.String("wakeup-interval", "1h34m9s", "doc")
	globalSet.Bool("detect-nsfw", config.DetectNSFW, "doc")
	globalSet.Bool("debug", false, "doc")
	globalSet.Bool("sponsor", true, "doc")
	globalSet.Bool("test", true, "doc")
	globalSet.Int("auto-index", config.AutoIndex, "doc")
	globalSet.String("auto-index-schedule", config.IndexSchedule, "doc")
	globalSet.Int("auto-import", config.AutoImport, "doc")

	app := cli.NewApp()
	app.Version = "0.0.0"

	c := cli.NewContext(app, globalSet, nil)

	LogErr(c.Set("config-path", config.ConfigPath))
	LogErr(c.Set("admin-password", config.AdminPassword))
	LogErr(c.Set("storage-path", config.StoragePath))
	LogErr(c.Set("sidecar-path", config.SidecarPath))
	LogErr(c.Set("sidecar-yaml", fmt.Sprintf("%t", config.SidecarYaml)))
	LogErr(c.Set("assets-path", config.AssetsPath))
	LogErr(c.Set("originals-path", config.OriginalsPath))
	LogErr(c.Set("import-path", config.ImportPath))
	LogErr(c.Set("cache-path", config.CachePath))
	LogErr(c.Set("temp-path", config.TempPath))
	LogErr(c.Set("backup-path", config.BackupPath))
	LogErr(c.Set("backup-retain", strconv.Itoa(config.BackupRetain)))
	LogErr(c.Set("backup-schedule", config.BackupSchedule))
	LogErr(c.Set("darktable-cli", config.DarktableBin))
	LogErr(c.Set("darktable-skip", "raf,cr3"))
	LogErr(c.Set("wakeup-interval", "1h34m9s"))
	LogErr(c.Set("detect-nsfw", "true"))
	LogErr(c.Set("debug", "false"))
	LogErr(c.Set("sponsor", "true"))
	LogErr(c.Set("test", "true"))
	LogErr(c.Set("auto-index", strconv.Itoa(config.AutoIndex)))
	LogErr(c.Set("auto-index-schedule", config.IndexSchedule))
	LogErr(c.Set("auto-import", strconv.Itoa(config.AutoImport)))

	return c
}

// RemoveTestData deletes files in import, export, originals, and cache folders.
func (c *Config) RemoveTestData() error {
	if err := os.RemoveAll(c.ImportPath()); err != nil {
		return err
	}

	if err := os.RemoveAll(c.TempPath()); err != nil {
		return err
	}

	if err := os.RemoveAll(c.OriginalsPath()); err != nil {
		return err
	}

	if err := os.RemoveAll(c.CachePath()); err != nil {
		log.Warnf("test: %s (remove cache)", err)
	}

	return nil
}

// DownloadTestData downloads the test files from the file server.
func (c *Config) DownloadTestData() error {
	if fs.FileExists(TestDataZip) {
		hash := fs.Hash(TestDataZip)

		if hash != TestDataHash {
			if err := os.Remove(TestDataZip); err != nil {
				return fmt.Errorf("config: %s", err.Error())
			}

			log.Debugf("config: removed outdated test data zip file (fingerprint %s)", hash)
		}
	}

	if !fs.FileExists(TestDataZip) {
		log.Debugf("config: downloading latest test data zip file from %s", TestDataURL)

		if err := fs.Download(TestDataZip, TestDataURL); err != nil {
			return fmt.Errorf("config: test data download failed: %s", err.Error())
		}
	}

	return nil
}

// UnzipTestData extracts tests files from the zip archive.
func (c *Config) UnzipTestData() error {
	if _, err := fs.Unzip(TestDataZip, c.StoragePath()); err != nil {
		return fmt.Errorf("config: could not unzip test data: %s", err.Error())
	}

	return nil
}

// InitializeTestData resets the test file directory.
func (c *Config) InitializeTestData() (err error) {
	testDataMutex.Lock()
	defer testDataMutex.Unlock()

	start := time.Now()

	if err = c.RemoveTestData(); err != nil {
		return err
	}

	if err = c.DownloadTestData(); err != nil {
		return err
	}

	if err = c.UnzipTestData(); err != nil {
		return err
	}

	log.Infof("config: initialized test data [%s]", time.Since(start))

	return nil
}
