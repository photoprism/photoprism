package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/urfave/cli/v2"

	_ "github.com/jinzhu/gorm/dialects/mysql" // register mysql dialect
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt/report"
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

// testDataPath resolves the QA fixture directory that ships with the assets
// bundle. Helpers fall back to this location when the caller does not provide
// an explicit storage path.
func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

// PkgNameRegexp normalizes database file names by stripping unsupported
// characters from the Go package identifier supplied by tests.
var PkgNameRegexp = regexp.MustCompile(`[^a-zA-Z\-_]+`)

// NewTestOptions builds fully-populated Options suited for backend tests. It
// creates an isolated storage directory under storage/testdata (or the
// PHOTOPRISM_STORAGE_PATH override) and enables all test-friendly defaults.
func NewTestOptions(dbName string) *Options {
	// Find storage path.
	storagePath := os.Getenv("PHOTOPRISM_STORAGE_PATH")
	if storagePath == "" {
		storagePath = fs.Abs("../../storage")
	}

	dataPath := filepath.Join(storagePath, fs.TestdataDir)

	return NewTestOptionsForPath(dbName, dataPath)
}

// NewTestOptionsForPath returns test Options using the provided storage path.
// When the caller omits the path, it falls back to storage/testdata, discovers
// the repo-level assets directory, and ensures Hub traffic is disabled.
func NewTestOptionsForPath(dbName, dataPath string) *Options {
	// Default to storage/testdata is no path was specified.
	if dataPath == "" {
		storagePath := os.Getenv("PHOTOPRISM_STORAGE_PATH")

		if storagePath == "" {
			storagePath = fs.Abs("../../storage")
		}

		dataPath = filepath.Join(storagePath, fs.TestdataDir)
	}

	// Enable test mode in dependencies.
	hub.ApplyTestConfig()

	// Create specified data path as storage.
	dataPath = fs.Abs(dataPath)
	if err := fs.MkdirAll(dataPath); err != nil {
		log.Errorf("config: %s (create test data path)", err)
		return &Options{}
	}

	// Create a config directory within the data path.
	configPath := filepath.Join(dataPath, "config")
	if err := fs.MkdirAll(configPath); err != nil {
		log.Errorf("config: %s (create test config path)", err)
		return &Options{}
	}

	// Find the assets paths containing models and frontend assets.
	assetsPath := os.Getenv("PHOTOPRISM_ASSETS_PATH")
	if assetsPath == "" {
		if wd, err := os.Getwd(); err == nil {
			for dir := wd; dir != "" && dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
				candidate := filepath.Join(dir, "assets")
				if fs.PathExists(candidate) {
					assetsPath = candidate
					break
				}
			}
		}

		if assetsPath == "" {
			assetsPath = fs.Abs("../../assets")
		}
	}

	// Obtain test database credentials.
	//
	// Example PHOTOPRISM_TEST_DSN for MariaDB / MySQL:
	// - "photoprism:photoprism@tcp(mariadb:4001)/photoprism?parseTime=true"
	dbName = PkgNameRegexp.ReplaceAllString(dbName, "")
	testDriver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	testDsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	// Set default test database driver.
	if testDriver == "test" || testDriver == "sqlite" || testDriver == "" || testDsn == "" {
		testDriver = dsn.DriverSQLite3
	}

	// Set default database DSN.
	if testDriver == dsn.DriverSQLite3 {
		if testDsn == "" && dbName != "" {
			if testDsn = fmt.Sprintf(".%s.db", clean.TypeLower(dbName)); !fs.FileExists(testDsn) {
				log.Tracef("sqlite: test database %s does not already exist", clean.Log(testDsn))
			} else if err := os.Remove(testDsn); err != nil {
				log.Errorf("sqlite: failed to remove existing test database %s (%s)", clean.Log(testDsn), err)
			}
		} else if testDsn == "" || testDsn == dsn.SQLiteTestDB {
			testDsn = dsn.SQLiteTestDB
			if !fs.FileExists(testDsn) {
				log.Tracef("sqlite: test database %s does not already exist", clean.Log(testDsn))
			} else if err := os.Remove(testDsn); err != nil {
				log.Errorf("sqlite: failed to remove existing test database %s (%s)", clean.Log(testDsn), err)
			}
		}
	}

	// Test config options.
	opts := &Options{
		Name:            "PhotoPrism",
		Version:         "0.0.0",
		Copyright:       "(c) 2018-2025 PhotoPrism UG. All rights reserved.",
		Public:          true,
		Sponsor:         true,
		AuthMode:        "",
		Test:            true,
		Debug:           true,
		Trace:           false,
		Experimental:    true,
		ReadOnly:        false,
		UploadNSFW:      false,
		ExifBruteForce:  false,
		AssetsPath:      assetsPath,
		AutoIndex:       -1,
		IndexSchedule:   DefaultIndexSchedule,
		AutoImport:      7200,
		StoragePath:     dataPath,
		CachePath:       filepath.Join(dataPath, "cache"),
		OriginalsPath:   filepath.Join(dataPath, "originals"),
		ImportPath:      filepath.Join(dataPath, "import"),
		ConfigPath:      configPath,
		DefaultsYaml:    filepath.Join(configPath, "defaults.yml"),
		OptionsYaml:     filepath.Join(configPath, "options.yml"),
		SidecarPath:     filepath.Join(dataPath, "sidecar"),
		TempPath:        filepath.Join(dataPath, "temp"),
		BackupRetain:    DefaultBackupRetain,
		BackupSchedule:  DefaultBackupSchedule,
		DatabaseDriver:  testDriver,
		DatabaseDSN:     testDsn,
		AdminPassword:   "photoprism",
		ClusterCIDR:     "",
		JWTScope:        DefaultJWTAllowedScopes,
		OriginalsLimit:  66,
		ResolutionLimit: 33,
		VisionApi:       true,
		DetectNSFW:      true,
	}

	return opts
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
		DatabaseDSN:    ".test-error.db",
	}

	return c
}

// SetNewTestConfig resets the singleton returned by TestConfig() so follow-up
// calls build a fresh fixture-backed config instance.
func SetNewTestConfig() {
	testConfig = NewTestConfig("test")
}

// TestConfig returns the existing test config instance or creates a new instance and returns it.
func TestConfig() *Config {
	testConfigOnce.Do(SetNewTestConfig)

	return testConfig
}

// NewMinimalTestConfig creates a lightweight test Config (no DB, minimal filesystem).
//
// Not suitable for tests requiring a database or pre-created storage directories.
func NewMinimalTestConfig(dataPath string) *Config {
	return NewIsolatedTestConfig("", dataPath, false)
}

var testDbCache []byte
var testDbMutex sync.Mutex

// NewMinimalTestConfigWithDb creates a lightweight test Config (minimal filesystem).
//
// Creates an isolated SQLite DB (cached after first run) without seeding media fixtures.
func NewMinimalTestConfigWithDb(dbName, dataPath string) *Config {
	c := NewIsolatedTestConfig(dbName, dataPath, true)

	cachedDb := false

	// Try to restore test db from cache.
	if len(testDbCache) > 0 && c.DatabaseDriver() == SQLite3 && !fs.FileExists(c.DatabaseDSN()) {
		if err := os.WriteFile(c.DatabaseDSN(), testDbCache, fs.ModeFile); err != nil {
			log.Warnf("config: %s (restore test database)", err)
		} else {
			cachedDb = true
		}
	}

	if err := c.Init(); err != nil {
		log.Fatalf("config: %s (init)", err.Error())
	}

	c.RegisterDb()

	if cachedDb {
		return c
	}

	c.InitTestDb()

	if testDbCache == nil && c.DatabaseDriver() == SQLite3 && fs.FileExistsNotEmpty(c.DatabaseDSN()) {
		testDbMutex.Lock()
		defer testDbMutex.Unlock()

		if testDbCache != nil {
			return c
		}

		if testDb, readErr := os.ReadFile(c.DatabaseDSN()); readErr != nil {
			log.Warnf("config: could not cache test database (%s)", readErr)
		} else {
			testDbCache = testDb
		}
	}

	return c
}

// NewIsolatedTestConfig constructs a lightweight Config backed by the provided config path.
//
// It avoids running migrations or loading test fixtures, making it useful for unit tests that
// only need basic access to config options (for example, JWT helpers). The caller should provide
// an isolated directory (e.g. via testing.T.TempDir) so temporary files are cleaned up automatically.
func NewIsolatedTestConfig(dbName, dataPath string, createDirs bool) *Config {
	if dataPath == "" {
		dataPath = filepath.Join(os.TempDir(), "photoprism-test-"+rnd.Base36(6))
	}

	opts := NewTestOptionsForPath(dbName, dataPath)

	c := &Config{
		options: opts,
		token:   rnd.Base36(8),
		cache:   gc.New(time.Second, time.Minute),
	}

	if !createDirs {
		return c
	}

	if err := c.CreateDirectories(); err != nil {
		log.Errorf("config: %s (create test directories)", err)
	}

	return c
}

// NewTestConfig initializes test data so required directories exist before tests run.
// See AGENTS.md (Test Data & Fixtures) and specs/dev/backend-testing.md for guidance.
func NewTestConfig(dbName string) *Config {
	defer log.Debug(capture.Time(time.Now(), "config: new test config created"))

	testConfigMutex.Lock()
	defer testConfigMutex.Unlock()

	c := &Config{
		cliCtx:  CliTestContext(),
		options: NewTestOptions(dbName),
		token:   rnd.Base36(8),
		cache:   gc.New(time.Second, time.Minute),
	}

	s := customize.NewSettings(c.DefaultTheme(), c.DefaultLocale(), c.DefaultTimezone().String())

	if err := fs.MkdirAll(c.ConfigPath()); err != nil {
		log.Panicf("config: %s", err.Error())
	}

	// Save settings next to the test config path, reusing any existing
	// `.yaml`/`.yml` variant so the tests mirror production behavior.
	if err := s.Save(fs.ConfigFilePath(c.ConfigPath(), "settings", fs.ExtYml)); err != nil {
		log.Panicf("config: %s", err.Error())
	}

	if err := c.Init(); err != nil {
		log.Panicf("config: %s", err.Error())
	}

	if err := c.InitializeTestData(); err != nil {
		log.Errorf("config: %s", err.Error())
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
	c := &Config{
		options: NewTestOptionsError(),
		cache:   gc.New(time.Second, time.Minute),
	}

	return c
}

// NewTestContext creates a new CLI test context with the flags and arguments provided.
func NewTestContext(args []string) *cli.Context {
	// Create new command-line app.
	app := cli.NewApp()
	app.Usage = "PhotoPrism®"
	app.Version = "test"
	app.Copyright = "(c) 2018-2025 PhotoPrism UG. All rights reserved."
	app.EnableBashCompletion = true
	app.Flags = Flags.Cli()
	app.Metadata = Values{
		"Name":    "PhotoPrism",
		"About":   "PhotoPrism®",
		"Edition": "ce",
		"Version": "test",
	}

	// Parse command arguments.
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	LogErr(flags.Parse(args))

	// Create and return new context.
	return cli.NewContext(app, flags, nil)
}

// CliTestContext returns a CLI context for testing.
func CliTestContext() *cli.Context {
	config := NewTestOptions("config-cli")

	globalSet := flag.NewFlagSet("test", flag.ContinueOnError)
	globalSet.String("config-path", config.ConfigPath, "doc")
	globalSet.String("admin-password", config.DarktableBin, "doc")
	globalSet.String("oidc-uri", config.OIDCUri, "doc")
	globalSet.String("oidc-client", config.OIDCClient, "doc")
	globalSet.String("oidc-secret", config.OIDCSecret, "doc")
	globalSet.String("oidc-scopes", config.OIDCScopes, "doc")
	globalSet.String("storage-path", config.StoragePath, "doc")
	globalSet.String("sidecar-path", config.SidecarPath, "doc")
	globalSet.Bool("sidecar-yaml", config.SidecarYaml, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("import-path", config.OriginalsPath, "doc")
	globalSet.String("cache-path", config.OriginalsPath, "doc")
	globalSet.String("temp-path", config.OriginalsPath, "doc")
	globalSet.String("defaults-yaml", config.DefaultsYaml, "doc")
	globalSet.String("cluster-uuid", config.ClusterUUID, "doc")
	globalSet.String("backup-path", config.StoragePath, "doc")
	globalSet.Int("backup-retain", config.BackupRetain, "doc")
	globalSet.String("backup-schedule", config.BackupSchedule, "doc")
	globalSet.String("darktable-cli", config.DarktableBin, "doc")
	globalSet.String("darktable-exclude", config.DarktableExclude, "doc")
	globalSet.String("sips-exclude", config.SipsExclude, "doc")
	globalSet.String("wakeup-interval", "1h34m9s", "doc")
	globalSet.Bool("vision-api", config.VisionApi, "doc")
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
	LogErr(c.Set("oidc-uri", config.OIDCUri))
	LogErr(c.Set("oidc-client", config.OIDCClient))
	LogErr(c.Set("oidc-secret", config.OIDCSecret))
	LogErr(c.Set("oidc-scopes", authn.OidcDefaultScopes))
	LogErr(c.Set("storage-path", config.StoragePath))
	LogErr(c.Set("sidecar-path", config.SidecarPath))
	LogErr(c.Set("sidecar-yaml", fmt.Sprintf("%t", config.SidecarYaml)))
	LogErr(c.Set("assets-path", config.AssetsPath))
	LogErr(c.Set("originals-path", config.OriginalsPath))
	LogErr(c.Set("import-path", config.ImportPath))
	LogErr(c.Set("cache-path", config.CachePath))
	LogErr(c.Set("temp-path", config.TempPath))
	LogErr(c.Set("defaults-yaml", config.DefaultsYaml))
	LogErr(c.Set("backup-path", config.BackupPath))
	LogErr(c.Set("backup-retain", strconv.Itoa(config.BackupRetain)))
	LogErr(c.Set("backup-schedule", config.BackupSchedule))
	LogErr(c.Set("darktable-cli", config.DarktableBin))
	LogErr(c.Set("darktable-exclude", "raf, cr3"))
	LogErr(c.Set("sips-exclude", "avif, avifs, thm"))
	LogErr(c.Set("wakeup-interval", "1h34m9s"))
	LogErr(c.Set("vision-api", "true"))
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
	if _, _, err := fs.Unzip(TestDataZip, c.StoragePath(), 2*fs.GB, -1); err != nil {
		return fmt.Errorf("config: could not unzip test data: %s", err.Error())
	}

	return nil
}

// InitializeTestData resets "storage/testdata" to a clean state.
//
// The function removes prior artifacts, downloads fixtures when missing,
// unzips them, and then calls CreateDirectories so required directories exist.
// See AGENTS.md (Test Data & Fixtures) for details.
func (c *Config) InitializeTestData() (err error) {
	testDataMutex.Lock()
	defer testDataMutex.Unlock()

	start := time.Now()

	// Delete existing test files and directories in "storage/testdata".
	if err = c.RemoveTestData(); err != nil {
		return fmt.Errorf("%s (remove testdata)", err)
	}

	// If the test file archive "/tmp/photoprism/testdata.zip" is missing,
	// download it from https://dl.photoprism.app/qa/testdata.zip.
	if err = c.DownloadTestData(); err != nil {
		return fmt.Errorf("%s (download testdata)", err)
	}

	// Extract "/tmp/photoprism/testdata.zip" in "storage/testdata".
	if err = c.UnzipTestData(); err != nil {
		return fmt.Errorf("%s (unzip testdata)", err)
	}

	// Make sure all the required directories exist in "storage/testdata.
	if err = c.CreateDirectories(); err != nil {
		return fmt.Errorf("%s (create directories)", err)
	}

	log.Infof("config: initialized test data [%s]", time.Since(start))

	return nil
}

// AssertTestData verifies the existence of the required test directories in "storage/testdata".
//
// Use this helper early in tests when diagnosing fixture setup issues. It logs
// presence/emptiness of required directories to testing.T. See the backend testing
// guide for additional patterns.
func (c *Config) AssertTestData(t *testing.T) {
	reportDir := func(dir string) {
		if fs.PathExists(dir) {
			t.Logf("testdata: dir %s exists (%s)", clean.Log(dir),
				report.Bool(fs.DirIsEmpty(dir), "empty", "not empty"))
		} else {
			t.Logf("testdata: dir %s is missing %s, but required", clean.Log(dir), report.Cross)
		}
	}

	reportErr := func(funcName string) {
		t.Errorf("testdata: *Config.%s() must not return an empty string %s", funcName, report.Cross)
	}

	if dir := c.AssetsPath(); dir != "" {
		reportDir(dir)
	} else {
		reportErr("AssetsPath")
	}

	if dir := c.ConfigPath(); dir != "" {
		reportDir(dir)
	} else {
		reportErr("ConfigPath")
	}

	if dir := c.ImportPath(); dir != "" {
		reportDir(dir)
	} else {
		reportErr("ImportPath")
	}

	if dir := c.OriginalsPath(); dir != "" {
		reportDir(dir)
	} else {
		reportErr("OriginalsPath")
	}

	if dir := c.SidecarPath(); dir != "" {
		reportDir(dir)
	} else {
		reportErr("SidecarPath")
	}
}
