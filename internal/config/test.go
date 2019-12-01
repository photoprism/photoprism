package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	TestDataZip  = "/tmp/photoprism/testdata.zip"
	TestDataURL  = "https://dl.photoprism.org/fixtures/testdata.zip"
	TestDataHash = "a217ac5242de2189ffb414d819b628c7957c67d7"
)

var testConfig *Config

func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

func NewTestParams() *Params {
	assetsPath := util.ExpandedFilename("../../assets")

	testDataPath := testDataPath(assetsPath)

	c := &Params{
		Public:         true,
		ReadOnly:       false,
		DarktableBin:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		ExportPath:     testDataPath + "/export",
		DatabaseDriver: "mysql",
		DatabaseDsn:    "photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true",
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
		DatabaseDsn:    "photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true",
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
	log.SetLevel(logrus.DebugLevel)

	c := &Config{config: NewTestParams()}
	err := c.Init(context.Background())
	if err != nil {
		log.Fatalf("failed init config: %v", err)
	}

	c.MigrateDb()
	return c
}

func NewTestErrorConfig() *Config {
	log.SetLevel(logrus.DebugLevel)

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
