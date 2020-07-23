package config

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestNewParams(t *testing.T) {
	ctx := CliTestContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewParams(ctx)

	assert.IsType(t, new(Params), c)

	assert.Equal(t, fs.Abs("../../assets"), c.AssetsPath)
	assert.False(t, c.Debug)
	assert.False(t, c.ReadOnly)
}

func TestParams_SetValuesFromFile(t *testing.T) {
	c := NewParams(CliTestContext())

	err := c.Load("testdata/config.yml")

	assert.Nil(t, err)

	assert.False(t, c.Debug)
	assert.False(t, c.ReadOnly)
	assert.Equal(t, "/srv/photoprism", c.AssetsPath)
	assert.Equal(t, "/srv/photoprism/cache", c.CachePath)
	assert.Equal(t, "/srv/photoprism/photos/originals", c.OriginalsPath)
	assert.Equal(t, "/srv/photoprism/photos/import", c.ImportPath)
	assert.Equal(t, "/srv/photoprism/temp", c.TempPath)
	assert.NotEmpty(t, c.DatabaseDriver)
	assert.NotEmpty(t, c.DatabaseDsn)
	assert.Equal(t, 81, c.HttpServerPort)
}

func TestParams_ExpandFilenames(t *testing.T) {
	p := Params{TempPath: "tmp", ImportPath: "import"}
	assert.Equal(t, "tmp", p.TempPath)
	assert.Equal(t, "import", p.ImportPath)
	p.expandFilenames()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/tmp", p.TempPath)
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/import", p.ImportPath)
}
