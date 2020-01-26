package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	TestDataZip  = "/tmp/photoprism/testdata.zip"
	TestDataURL  = "https://dl.photoprism.org/fixtures/testdata.zip"
	TestDataHash = "be394d5bee8a5634d415e9e0663eef20b5604510" // sha1sum
)

var testConfig *Config
var once sync.Once

func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

func NewTestParams() *Params {
	assetsPath := fs.ExpandFilename("../../assets")

	testDataPath := testDataPath(assetsPath)

	c := &Params{
		Public:         true,
		ReadOnly:       false,
		DetectNSFW:     true,
		UploadNSFW:     false,
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
	assetsPath := fs.ExpandFilename("../..")

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
	once.Do(func() {
		testConfig = NewTestConfig()
	})

	return testConfig
}

func NewTestConfig() *Config {
	log.SetLevel(logrus.DebugLevel)

	c := &Config{config: NewTestParams()}
	err := c.Init(context.Background())
	if err != nil {
		log.Fatalf("failed init config: %v", err)
	}

	c.DropTables()

	c.MigrateDb()

	c.ImportSQL(c.ExamplesPath() + "/fixtures.sql")

	thumb.JpegQuality = c.ThumbQuality()
	thumb.PreRenderSize = c.ThumbSize()
	thumb.MaxRenderSize = c.ThumbLimit()
	thumb.Filter = c.ThumbFilter()

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
	globalSet.Bool("detect-nsfw", config.DetectNSFW, "doc")

	app := cli.NewApp()

	c := cli.NewContext(app, globalSet, nil)

	c.Set("config-file", config.ConfigFile)
	c.Set("assets-path", config.AssetsPath)
	c.Set("originals-path", config.OriginalsPath)
	c.Set("darktable-cli", config.DarktableBin)
	c.Set("detect-nsfw", "true")

	return c
}

func (c *Config) RemoveTestData(t *testing.T) {
	os.RemoveAll(c.ImportPath())
	os.RemoveAll(c.ExportPath())
	os.RemoveAll(c.OriginalsPath())
	os.RemoveAll(c.CachePath())
}

func (c *Config) DownloadTestData(t *testing.T) {
	if fs.FileExists(TestDataZip) {
		hash := fs.Hash(TestDataZip)

		if hash != TestDataHash {
			os.Remove(TestDataZip)
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

func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := fs.Unzip(TestDataZip, testDataPath(c.AssetsPath())); err != nil {
		t.Logf("could not unzip test data: %s\n", err.Error())
	}
}

func (c *Config) InitializeTestData(t *testing.T) {
	t.Log("initializing test data")

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}
