package config

import (
	"context"
	"flag"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	TestDataZip  = "/tmp/photoprism/testdata.zip"
	TestDataURL  = "https://dl.photoprism.org/fixtures/testdata.zip"
	TestDataHash = "a217ac5242de2189ffb414d819b628c7957c67d7"
)

var testConfig *Config
var mockDB sqlmock.Sqlmock

func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

func NewTestParams() *Params {
	assetsPath := util.ExpandedFilename("../../assets")

	testDataPath := testDataPath(assetsPath)

	c := &Params{
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		ExportPath:     testDataPath + "/export",
		DatabaseDriver: "mysql",
		DatabaseDsn:    "photoprism:photoprism@tcp(photoprism-mysql:4001)/photoprism?parseTime=true",
	}

	return c
}

func NewTestParamsError() *Params {
	assetsPath := util.ExpandedFilename("../..")

	testDataPath := testDataPath("../../assets")

	c := &Params{
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		ExportPath:     testDataPath + "/export",
		DatabaseDriver: "mysql",
		DatabaseDsn:    "photoprism:photoprism@tcp(photoprism-mysql:4001)/photoprism?parseTime=true",
	}

	return c
}

// NewTestParamsMockDB create instance for *Params for testing with mock database
func NewTestParamsMockDB() *Params {
	assetsPath := util.ExpandedFilename("../../assets")

	testDataPath := testDataPath(assetsPath)

	c := &Params{
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		ExportPath:     testDataPath + "/export",
		DatabaseDriver: "",
		DatabaseDsn:    "",
	}

	return c
}

func TestConfig() *Config {
	if testConfig == nil {
		testConfig = NewTestConfig()
	}

	return testConfig
}

func NewTestConfig() *Config {
	log.SetLevel(log.DebugLevel)

	c := &Config{config: NewTestParams()}
	err := c.Init(context.Background())
	if err != nil {
		log.Fatalf("failed init config: %v", err)
	}

	c.MigrateDb()
	return c
}

func TestConfigMockDB(t *testing.T) (*Config, sqlmock.Sqlmock) {
	if testConfig == nil {
		testConfig, mockDB = NewTestConfigMockDB(t)
	}

	return testConfig, mockDB
}

// NewTestConfigMockDB create instance of *Config for testing with Mock database
func NewTestConfigMockDB(t *testing.T) (*Config, sqlmock.Sqlmock) {
	log.SetLevel(log.DebugLevel)

	c := &Config{config: NewTestParams()}
	mock := c.InitMockDB(t)

	return c, mock
}

func CleanTestConfig() {
	testConfig = nil
}

func NewTestErrorConfig() *Config {
	log.SetLevel(log.DebugLevel)

	c := &Config{config: NewTestParamsError()}
	err := c.Init(context.Background())
	if err != nil {
		log.Fatalf("failed init config: %v", err)
	}

	c.MigrateDb()
	return c
}

// Returns example cli config for testing
func CliTestContext() *cli.Context {
	config := NewTestParams()

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("config-file", config.ConfigFile, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("darktable-cli", config.DarktableBin, "doc")

	app := cli.NewApp()

	c := cli.NewContext(app, globalSet, nil)

	c.Set("config-file", config.ConfigFile)
	c.Set("assets-path", config.AssetsPath)
	c.Set("originals-path", config.OriginalsPath)
	c.Set("darktable-cli", config.DarktableBin)

	return c
}

func (c *Config) RemoveTestData(t *testing.T) {
	os.RemoveAll(c.ImportPath())
	os.RemoveAll(c.ExportPath())
	os.RemoveAll(c.OriginalsPath())
	os.RemoveAll(c.CachePath())
}

func (c *Config) DownloadTestData(t *testing.T) {
	if util.Exists(TestDataZip) {
		hash := util.Hash(TestDataZip)

		if hash != TestDataHash {
			os.Remove(TestDataZip)
			t.Logf("removed outdated test data zip file (fingerprint %s)\n", hash)
		}
	}

	if !util.Exists(TestDataZip) {
		fmt.Printf("downloading latest test data zip file from %s\n", TestDataURL)

		if err := util.Download(TestDataZip, TestDataURL); err != nil {
			fmt.Printf("Download failed: %s\n", err.Error())
		}
	}
}

func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := util.Unzip(TestDataZip, testDataPath(c.AssetsPath())); err != nil {
		t.Logf("could not unzip test data: %s\n", err.Error())
	}
}

func (c *Config) InitializeTestData(t *testing.T) {
	t.Log("initializing test data")

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}

func (c *Config) InitMockDB(t *testing.T) sqlmock.Sqlmock {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	c.db, err = gorm.Open("postgres", db)
	require.NoError(t, err)

	c.db.LogMode(true)
	return mock
}
