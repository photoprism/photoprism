package photoprism

import (
	"flag"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/context"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/test"
	"github.com/urfave/cli"

	"github.com/stretchr/testify/assert"
)

func getTestCliContext() *cli.Context {
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("config-file", test.ConfigFile, "doc")
	globalSet.String("assets-path", test.AssetsPath, "doc")
	globalSet.String("originals-path", test.OriginalsPath, "doc")
	globalSet.String("darktable-cli", test.DarktableCli, "doc")

	app := cli.NewApp()

	c := cli.NewContext(app, globalSet, nil)

	c.Set("config-file", test.ConfigFile)
	c.Set("assets-path", test.AssetsPath)
	c.Set("originals-path", test.OriginalsPath)
	c.Set("darktable-cli", test.DarktableCli)

	return c
}

func TestNewConfig(t *testing.T) {
	ctx := getTestCliContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := context.NewConfig(ctx)

	assert.IsType(t, new(context.Config), c)

	assert.Equal(t, test.AssetsPath, c.GetAssetsPath())
	assert.False(t, c.Debug())
}

func TestContextConfig_SetValuesFromFile(t *testing.T) {
	c := context.NewConfig(getTestCliContext())

	c.SetValuesFromFile(fsutil.ExpandedFilename(test.ConfigFile))

	assert.Equal(t, "/srv/photoprism", c.GetAssetsPath())
	assert.Equal(t, "/srv/photoprism/cache", c.CachePath())
	assert.Equal(t, "/srv/photoprism/cache/thumbnails", c.GetThumbnailsPath())
	assert.Equal(t, "/srv/photoprism/photos/originals", c.OriginalsPath())
	assert.Equal(t, "/srv/photoprism/photos/import", c.ImportPath())
	assert.Equal(t, "/srv/photoprism/photos/export", c.ExportPath())
	assert.Equal(t, "tidb", c.DatabaseDriver())
	assert.Equal(t, "root:@tcp(localhost:4000)/photoprism?parseTime=true", c.DatabaseDsn())
}

func TestTestConfig_ConnectToDatabase(t *testing.T) {
	c := test.NewConfig()

	db := c.Db()

	assert.IsType(t, &gorm.DB{}, db)
}
