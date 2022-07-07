package config

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	ctx := CliTestContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewOptions(ctx)

	assert.IsType(t, new(Options), c)

	assert.Equal(t, fs.Abs("../../assets"), c.AssetsPath)
	assert.Equal(t, "1h34m9s", c.WakeupInterval.String())
	assert.False(t, c.Debug)
	assert.False(t, c.ReadOnly)
}

func TestOptions_SetOptionsFromFile(t *testing.T) {
	c := NewOptions(CliTestContext())

	err := c.Load("testdata/config.yml")

	assert.Nil(t, err)

	assert.False(t, c.Debug)
	assert.False(t, c.ReadOnly)
	assert.Equal(t, "/srv/photoprism", c.AssetsPath)
	assert.Equal(t, "/srv/photoprism/cache", c.CachePath)
	assert.Equal(t, "/srv/photoprism/photos/originals", c.OriginalsPath)
	assert.Equal(t, "/srv/photoprism/photos/import", c.ImportPath)
	assert.Equal(t, "/srv/photoprism/temp", c.TempPath)
	assert.Equal(t, "1h34m9s", c.WakeupInterval.String())
	assert.NotEmpty(t, c.DatabaseDriver)
	assert.NotEmpty(t, c.DatabaseDsn)
	assert.Equal(t, 81, c.HttpPort)
}

func TestOptions_ExpandFilenames(t *testing.T) {
	p := Options{TempPath: "tmp", ImportPath: "import"}
	assert.Equal(t, "tmp", p.TempPath)
	assert.Equal(t, "import", p.ImportPath)
	p.expandFilenames()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/tmp", p.TempPath)
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/import", p.ImportPath)
}
