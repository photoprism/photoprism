package config

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
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
	c := NewTestOptions()

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
