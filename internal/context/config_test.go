package context

import (
	"flag"
	"testing"

	"github.com/photoprism/photoprism/internal/fsutil"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testDataPath = "testdata"
const testConfigFile = "../../configs/photoprism.yml"

var darktableCli = "/usr/bin/darktable-cli"
var assetsPath = fsutil.ExpandedFilename("../../assets")
var originalsPath = fsutil.ExpandedFilename(testDataPath + "/originals")
var databaseDriver = "mysql"
var databaseDsn = "photoprism:photoprism@tcp(database:3306)/photoprism?parseTime=true"

func getTestCliContext() *cli.Context {
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("debug", false, "doc")
	globalSet.String("config-file", testConfigFile, "doc")
	globalSet.String("assets-path", assetsPath, "doc")
	globalSet.String("originals-path", originalsPath, "doc")
	globalSet.String("darktable-cli", darktableCli, "doc")

	app := cli.NewApp()

	c := cli.NewContext(app, globalSet, nil)

	c.Set("config-file", testConfigFile)
	c.Set("assets-path", assetsPath)
	c.Set("originals-path", originalsPath)
	c.Set("darktable-cli", darktableCli)

	return c
}

func TestNewConfig(t *testing.T) {
	ctx := getTestCliContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewConfig(ctx)

	assert.IsType(t, new(Config), c)

	assert.Equal(t, assetsPath, c.GetAssetsPath())
	assert.False(t, c.IsDebug())
}

func TestConfig_SetValuesFromFile(t *testing.T) {
	c := NewConfig(getTestCliContext())

	c.SetValuesFromFile(fsutil.ExpandedFilename(testConfigFile))

	assert.Equal(t, "/srv/photoprism", c.GetAssetsPath())
	assert.Equal(t, "/srv/photoprism/cache", c.GetCachePath())
	assert.Equal(t, "/srv/photoprism/cache/thumbnails", c.GetThumbnailsPath())
	assert.Equal(t, "/srv/photoprism/photos/originals", c.GetOriginalsPath())
	assert.Equal(t, "/srv/photoprism/photos/import", c.GetImportPath())
	assert.Equal(t, "/srv/photoprism/photos/export", c.GetExportPath())
	assert.Equal(t, databaseDriver, c.GetDatabaseDriver())
	assert.Equal(t, databaseDsn, c.GetDatabaseDsn())
}
