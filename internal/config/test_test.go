package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig_TestdataPath2(t *testing.T) {
	assert.Equal(t, "/xxx/testdata", testDataPath("/xxx"))
}

func TestTestCliContext(t *testing.T) {
	result := CliTestContext()

	assert.IsType(t, new(cli.Context), result)
}

func TestTestConfig(t *testing.T) {
	c := TestConfig()

	assert.IsType(t, new(Config), c)
	assert.IsType(t, &gorm.DB{}, c.Db())
}

func TestNewTestOptions(t *testing.T) {
	c := NewTestOptions("config")

	assert.IsType(t, new(Options), c)

	assert.Equal(t, fs.Abs("../../assets"), c.AssetsPath)
	assert.True(t, c.Debug)
}

func TestNewTestOptionsError(t *testing.T) {
	c := NewTestOptionsError()

	assert.IsType(t, new(Options), c)

	assert.Equal(t, fs.Abs("../.."), c.AssetsPath)
	assert.Equal(t, fs.Abs("../../storage/testdata/cache"), c.CachePath)
	assert.False(t, c.Debug)
}

func TestNewTestErrorConfig(t *testing.T) {
	c := NewTestErrorConfig()

	if err := c.connectDb(); err != nil {
		t.Fatal(err)
	}

	db := c.Db()

	assert.IsType(t, &gorm.DB{}, db)
}

func TestCleanupTestFolder(t *testing.T) {
	t.Run("OptionsNil", func(t *testing.T) {
		// Setup and capture log output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		var c Config
		c.CleanupTestFolder()

		// Reset logger
		log.SetOutput(os.Stdout)

		assert.Contains(t, buffer.String(), "config: c.options is nil in CleanupTestFolder")
	})

	t.Run("NotExpectedPath", func(t *testing.T) {
		// Setup and capture log output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		c := Config{options: &Options{StoragePath: "/tmp/photoprism/testdata"}}
		c.CleanupTestFolder()

		// Reset logger
		log.SetOutput(os.Stdout)

		assert.Contains(t, buffer.String(), "config: /tmp/photoprism/testdata not cleaned up")
	})

	t.Run("Success", func(t *testing.T) {
		// Setup and capture log output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		c := Config{options: &Options{StoragePath: "/tmp/photoprism/test-photoprism-1394931550/testdata"}}
		c.CleanupTestFolder()

		// Reset logger
		log.SetOutput(os.Stdout)

		assert.Contains(t, buffer.String(), "config: cleaned up /tmp/photoprism/test-photoprism-1394931550")
	})
}
