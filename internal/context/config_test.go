package context

import (
	"testing"

	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/stretchr/testify/assert"
)

func TestConfig_SetValuesFromFile(t *testing.T) {
	c := NewConfig(CliTestContext())

	err := c.SetValuesFromFile(fsutil.ExpandedFilename("../../configs/photoprism.yml"))

	assert.Nil(t, err)

	assert.Equal(t, "/srv/photoprism", c.AssetsPath)
	assert.Equal(t, "/srv/photoprism/cache", c.CachePath)
	assert.Equal(t, "/srv/photoprism/photos/originals", c.OriginalsPath)
	assert.Equal(t, "/srv/photoprism/photos/import", c.ImportPath)
	assert.Equal(t, "/srv/photoprism/photos/export", c.ExportPath)
	assert.Equal(t, "internal", c.DatabaseDriver)
	assert.Equal(t, "root:photoprism@tcp(localhost:4000)/photoprism?parseTime=true", c.DatabaseDsn)
}
