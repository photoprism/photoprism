package context

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestTestCliContext(t *testing.T) {
	result := CliTestContext()

	assert.IsType(t, new(cli.Context), result)
}

func TestTestContext(t *testing.T) {
	result := TestContext()

	assert.IsType(t, new(Context), result)
}

func TestNewTestConfig(t *testing.T) {
	c := NewTestConfig()

	assert.IsType(t, new(Config), c)

	assert.Equal(t, fsutil.ExpandedFilename("../../assets"), c.AssetsPath)
	assert.False(t, c.Debug)
}

func TestNewTestContext_Db(t *testing.T) {
	c := NewTestContext()

	db := c.Db()

	assert.IsType(t, &gorm.DB{}, db)
}
