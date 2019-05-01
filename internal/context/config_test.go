package context

import (
	"testing"

	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	ctx := test.CliContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewConfig(ctx)

	assert.IsType(t, new(Config), c)

	assert.Equal(t, fsutil.ExpandedFilename("../../assets"), c.AssetsPath())
	assert.False(t, c.Debug())
}

func TestConfig_SetValuesFromFile(t *testing.T) {
	c := NewConfig(test.CliContext())

	err := c.SetValuesFromFile(fsutil.ExpandedFilename("../../configs/photoprism.yml"))

	assert.Nil(t, err)

	assert.Equal(t, "/srv/photoprism", c.AssetsPath())
	assert.Equal(t, "/srv/photoprism/cache", c.CachePath())
	assert.Equal(t, "/srv/photoprism/cache/thumbnails", c.ThumbnailsPath())
	assert.Equal(t, "/srv/photoprism/photos/originals", c.OriginalsPath())
	assert.Equal(t, "/srv/photoprism/photos/import", c.ImportPath())
	assert.Equal(t, "/srv/photoprism/photos/export", c.ExportPath())
	assert.Equal(t, "internal", c.DatabaseDriver())
	assert.Equal(t, "root:photoprism@tcp(localhost:4000)/photoprism?parseTime=true", c.DatabaseDsn())
}
