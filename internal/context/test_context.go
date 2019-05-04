package context

import (
	"flag"
	"fmt"
	"os"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/urfave/cli"

	log "github.com/sirupsen/logrus"
)

const (
	TestDataZip  = "/tmp/photoprism/testdata.zip"
	TestDataURL  = "https://dl.photoprism.org/fixtures/testdata.zip"
	TestDataHash = "1a59b358b80221ab3e76efb683ad72402f0b0844"
)

var testContext *Context

func testDataPath(assetsPath string) string {
	return assetsPath + "/testdata"
}

func NewTestConfig() *Config {
	assetsPath := fsutil.ExpandedFilename("../../assets")

	testDataPath := testDataPath(assetsPath)

	c := &Config{
		DarktableCli:   "/usr/bin/darktable-cli",
		AssetsPath:     assetsPath,
		CachePath:      testDataPath + "/cache",
		OriginalsPath:  testDataPath + "/originals",
		ImportPath:     testDataPath + "/import",
		ExportPath:     testDataPath + "/export",
		DatabaseDriver: "mysql",
		DatabaseDsn:    "photoprism:photoprism@tcp(database:3306)/photoprism?parseTime=true",
	}

	return c
}

func TestContext() *Context {
	if testContext == nil {
		testContext = NewTestContext()
	}

	return testContext
}

func NewTestContext() *Context {
	log.SetLevel(log.DebugLevel)

	c := &Context{config: NewTestConfig()}

	c.MigrateDb()

	return c
}

// Returns example cli context for testing
func CliTestContext() *cli.Context {
	config := NewTestConfig()

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("config-file", config.ConfigFile, "doc")
	globalSet.String("assets-path", config.AssetsPath, "doc")
	globalSet.String("originals-path", config.OriginalsPath, "doc")
	globalSet.String("darktable-cli", config.DarktableCli, "doc")

	app := cli.NewApp()

	c := cli.NewContext(app, globalSet, nil)

	c.Set("config-file", config.ConfigFile)
	c.Set("assets-path", config.AssetsPath)
	c.Set("originals-path", config.OriginalsPath)
	c.Set("darktable-cli", config.DarktableCli)

	return c
}

func (c *Context) RemoveTestData(t *testing.T) {
	os.RemoveAll(c.ImportPath())
	os.RemoveAll(c.ExportPath())
	os.RemoveAll(c.OriginalsPath())
	os.RemoveAll(c.CachePath())
}

func (c *Context) DownloadTestData(t *testing.T) {
	if fsutil.Exists(TestDataZip) {
		hash := fsutil.Hash(TestDataZip)

		if hash != TestDataHash {
			os.Remove(TestDataZip)
			t.Logf("removed outdated test data zip file (fingerprint %s)\n", hash)
		}
	}

	if !fsutil.Exists(TestDataZip) {
		fmt.Printf("downloading latest test data zip file from %s\n", TestDataURL)

		if err := fsutil.Download(TestDataZip, TestDataURL); err != nil {
			fmt.Printf("Download failed: %s\n", err.Error())
		}
	}
}

func (c *Context) UnzipTestData(t *testing.T) {
	if _, err := fsutil.Unzip(TestDataZip, testDataPath(c.AssetsPath())); err != nil {
		t.Logf("could not unzip test data: %s\n", err.Error())
	}
}

func (c *Context) InitializeTestData(t *testing.T) {
	t.Log("initializing test data")

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}
