package photoprism

import (
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

const testDataPath = "testdata"
const testDataUrl = "https://www.dropbox.com/s/na9p9wwt98l7m5b/import.zip?dl=1"
const testDataHash = "ed3bdb2fe86ea662bc863b63e219b47b8d9a74024757007f7979887d"

var darktableCli = "/Applications/darktable.app/Contents/MacOS/darktable-cli"
var testDataZip = getExpandedFilename(testDataPath + "/import.zip")
var originalsPath = getExpandedFilename(testDataPath + "/originals")
var thumbnailsPath = getExpandedFilename(testDataPath + "/thumbnails")
var importPath = getExpandedFilename(testDataPath + "/import")
var exportPath = getExpandedFilename(testDataPath + "/export")

func (c *Config) RemoveTestData(t *testing.T) {
	os.RemoveAll(c.ImportPath)
	os.RemoveAll(c.ExportPath)
	os.RemoveAll(c.OriginalsPath)
	os.RemoveAll(c.ThumbnailsPath)
}

func (c *Config) DownloadTestData(t *testing.T) {
	if fileExists(testDataZip) {
		hash := fileHash(testDataZip)

		if hash != testDataHash {
			os.Remove(testDataZip)
			t.Logf("Removed outdated test data zip file (fingerprint %s)\n", hash)
		}
	}

	if !fileExists(testDataZip) {
		fmt.Printf("Downloading latest test data zip file from %s\n", testDataUrl)

		if err := downloadFile(testDataZip, testDataUrl); err != nil {
			fmt.Printf("Download failed: %s\n", err.Error())
		}
	}
}

func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := unzip(testDataZip, testDataPath); err != nil {
		t.Logf("Could not unzip test data: %s\n", err.Error())
	}
}

func (c *Config) InitializeTestData(t *testing.T) {
	t.Log("Initializing test data")

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}

func NewTestConfig() *Config {
	return &Config{
		DarktableCli:   darktableCli,
		OriginalsPath:  originalsPath,
		ThumbnailsPath: thumbnailsPath,
		ImportPath:     importPath,
		ExportPath:     exportPath,
	}
}

func TestNewConfig(t *testing.T) {
	c := NewConfig()

	assert.IsType(t, &Config{}, c)
}

func TestConfig_SetValuesFromFile(t *testing.T) {
	c := NewConfig()

	c.SetValuesFromFile(getExpandedFilename("config.example.yml"))

	assert.Equal(t, getExpandedFilename("photos/originals"), c.OriginalsPath)
	assert.Equal(t, getExpandedFilename("photos/thumbnails"), c.ThumbnailsPath)
	assert.Equal(t, getExpandedFilename("photos/import"), c.ImportPath)
	assert.Equal(t, getExpandedFilename("photos/export"), c.ExportPath)
}
