package config

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestTestCliContext(t *testing.T) {
	result := CliTestContext()

	assert.IsType(t, new(cli.Context), result)
}

func TestTestConfig(t *testing.T) {
	result := TestConfig()

	assert.IsType(t, new(Config), result)
}

func TestNewTestParams(t *testing.T) {
	c := NewTestParams()

	assert.IsType(t, new(Params), c)

	assert.Equal(t, util.ExpandedFilename("../../assets"), c.AssetsPath)
	assert.False(t, c.Debug)
}

func TestNewTestConfig(t *testing.T) {
	c := NewTestConfig()

	db := c.Db()

	assert.IsType(t, &gorm.DB{}, db)
}
