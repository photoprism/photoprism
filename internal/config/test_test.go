package config

import (
	"bytes"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/event"
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

// captureSystemLog swaps event.SystemLog for a buffer-backed logger, runs fn, and
// returns what fn wrote to the system log. CleanupTestFolder logs via event.System*
// (not the DB-hooked default logger), so tests must read the system log to see it.
func captureSystemLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)
	orig := event.SystemLog
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = orig })
	fn()
	return buf.String()
}

func TestCleanupTestFolder(t *testing.T) {
	t.Run("OptionsNil", func(t *testing.T) {
		var c Config
		out := captureSystemLog(t, c.CleanupTestFolder)
		assert.Contains(t, out, "c.options is nil in CleanupTestFolder")
	})
	t.Run("NotExpectedPath", func(t *testing.T) {
		c := Config{options: &Options{StoragePath: "/tmp/photoprism/testdata"}}
		out := captureSystemLog(t, c.CleanupTestFolder)
		assert.Contains(t, out, "cleanup /tmp/photoprism/testdata")
		assert.Contains(t, out, "failed")
	})
	t.Run("Success", func(t *testing.T) {
		c := Config{options: &Options{StoragePath: "/tmp/photoprism/test-photoprism-1394931550/testdata"}}
		out := captureSystemLog(t, c.CleanupTestFolder)
		assert.Contains(t, out, "cleanup /tmp/photoprism/test-photoprism-1394931550")
		assert.Contains(t, out, "succeeded")
	})
	t.Run("MarkerNotPrefix", func(t *testing.T) {
		// The isolated-folder marker must be the prefix of the parent directory, not merely
		// a substring, so a path like ".../my-test-photoprism-x/testdata" is refused.
		c := Config{options: &Options{StoragePath: "/tmp/photoprism/my-test-photoprism-x/testdata"}}
		out := captureSystemLog(t, c.CleanupTestFolder)
		assert.Contains(t, out, "cleanup /tmp/photoprism/my-test-photoprism-x/testdata")
		assert.Contains(t, out, "failed")
	})
}
