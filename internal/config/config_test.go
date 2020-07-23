package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	c := TestConfig()

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}

func TestNewConfig(t *testing.T) {
	ctx := CliTestContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewConfig(ctx)

	assert.IsType(t, new(Config), c)

	assert.Equal(t, fs.Abs("../../assets"), c.AssetsPath())
	assert.False(t, c.Debug())
	assert.False(t, c.ReadOnly())
}

func TestConfig_Name(t *testing.T) {
	c := NewConfig(CliTestContext())

	name := c.Name()
	assert.Equal(t, "config.test", name)
}

func TestConfig_Version(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.Version()
	assert.Equal(t, "test", version)
}

func TestConfig_TensorFlowVersion(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.TensorFlowVersion()
	assert.IsType(t, "1.15.0", version)
}

func TestConfig_TensorFlowDisabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	version := c.TensorFlowOff()
	assert.Equal(t, false, version)
}

func TestConfig_Copyright(t *testing.T) {
	c := NewConfig(CliTestContext())

	copyright := c.Copyright()
	assert.Equal(t, "", copyright)
}

func TestConfig_ConfigFile(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.ConfigFile(), "/storage/testdata/settings/photoprism.yml")
}

func TestConfig_SettingsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.SettingsPath(), "/storage/testdata/settings")
}

func TestConfig_PIDFilename(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.PIDFilename(), "/storage/testdata/photoprism.pid")
}

func TestConfig_LogFilename(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.LogFilename(), "/storage/testdata/photoprism.log")
}

func TestConfig_DetachServer(t *testing.T) {
	c := NewConfig(CliTestContext())

	detachServer := c.DetachServer()
	assert.Equal(t, false, detachServer)
}

func TestConfig_HttpServerHost(t *testing.T) {
	c := NewConfig(CliTestContext())

	host := c.HttpServerHost()
	assert.Equal(t, "0.0.0.0", host)
}

func TestConfig_HttpServerPort(t *testing.T) {
	c := NewConfig(CliTestContext())

	port := c.HttpServerPort()
	assert.Equal(t, 2342, port)
}

func TestConfig_HttpServerMode(t *testing.T) {
	c := NewConfig(CliTestContext())

	mode := c.HttpServerMode()
	assert.Equal(t, "release", mode)
}

func TestConfig_HttpServerPassword(t *testing.T) {
	c := NewConfig(CliTestContext())

	password := c.HttpServerPassword()
	assert.Equal(t, "", password)
}

func TestConfig_OriginalsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.OriginalsPath()
	assert.True(t, strings.HasPrefix(result, "/"))
	assert.True(t, strings.HasSuffix(result, "/storage/testdata/originals"))
}

func TestConfig_ImportPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.ImportPath()
	assert.True(t, strings.HasPrefix(result, "/"))
	assert.True(t, strings.HasSuffix(result, "/storage/testdata/import"))
}

func TestConfig_SipsBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.SipsBin()
	assert.Equal(t, "", bin)
}

func TestConfig_DarktableBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.DarktableBin()
	assert.Equal(t, "/usr/bin/darktable-cli", bin)
}

func TestConfig_HeifConvertBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.HeifConvertBin()
	assert.Equal(t, "/usr/bin/heif-convert", bin)
}

func TestConfig_ExifToolBin(t *testing.T) {
	c := NewConfig(CliTestContext())

	bin := c.ExifToolBin()
	assert.Equal(t, "/usr/bin/exiftool", bin)
}

func TestConfig_DatabaseDriver(t *testing.T) {
	c := NewConfig(CliTestContext())

	driver := c.DatabaseDriver()
	assert.Equal(t, SQLite, driver)
}

func TestConfig_DatabaseDsn(t *testing.T) {
	c := NewConfig(CliTestContext())

	dsn := c.DatabaseDriver()
	assert.Equal(t, SQLite, dsn)
}

func TestConfig_CachePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasSuffix(c.CachePath(), "storage/testdata/cache"))
}

func TestConfig_ThumbnailsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasPrefix(c.ThumbPath(), "/"))
	assert.True(t, strings.HasSuffix(c.ThumbPath(), "storage/testdata/cache/thumbnails"))
}

func TestConfig_AssetsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, strings.HasSuffix(c.AssetsPath(), "/assets"))
}

func TestConfig_DetectNSFW(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.DetectNSFW()
	assert.Equal(t, true, result)
}

func TestConfig_AdminPassword(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.AdminPassword()
	assert.Equal(t, "photoprism", result)
}

func TestConfig_NSFWModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Contains(t, c.NSFWModelPath(), "/assets/nsfw")
}

func TestConfig_ExamplesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.ExamplesPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/examples", path)
}

func TestConfig_TensorFlowModelPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.TensorFlowModelPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/nasnet", path)
}

func TestConfig_TemplatesPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.TemplatesPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/templates", path)
}

func TestConfig_StaticPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.StaticPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/static", path)
}

func TestConfig_BuildPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.BuildPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/static/build", path)
}

func TestConfig_ImgPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	path := c.ImgPath()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/static/img", path)
}

func TestConfig_ClientConfig(t *testing.T) {
	c := TestConfig()

	cc := c.UserConfig()

	assert.IsType(t, ClientConfig{}, cc)
	assert.NotEmpty(t, cc.Name)
	assert.NotEmpty(t, cc.Version)
	assert.NotEmpty(t, cc.Copyright)
	assert.NotEmpty(t, cc.Thumbs)
	assert.NotEmpty(t, cc.JSHash)
	assert.NotEmpty(t, cc.CSSHash)
	assert.Equal(t, true, cc.Debug)
	assert.Equal(t, false, cc.ReadOnly)
}

func TestConfig_Workers(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.GreaterOrEqual(t, c.Workers(), 1)
}

func TestConfig_WakeupInterval(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, time.Duration(900000000000), c.WakeupInterval())
}

func TestConfig_GeoCodingApi(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.GeoCodingApi())
	c.params.GeoCodingApi = "places"
	assert.Equal(t, "places", c.GeoCodingApi())
	c.params.GeoCodingApi = "osm"
	assert.Equal(t, "osm", c.GeoCodingApi())
}

func TestConfig_OriginalsLimit(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, int64(-1), c.OriginalsLimit())
	c.params.OriginalsLimit = 800
	assert.Equal(t, int64(838860800), c.OriginalsLimit())
}

func TestConfig_SiteUrl(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "http://localhost:2342/", c.SiteUrl())
	c.params.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "http://superhost:2342/", c.SiteUrl())
}

func TestConfig_SitePreview(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "http://localhost:2342/static/img/preview.jpg", c.SitePreview())
	c.params.SitePreview = "http://preview.jpg"
	assert.Equal(t, "http://preview.jpg", c.SitePreview())
	c.params.SitePreview = "preview123.jpg"
	assert.Equal(t, "http://localhost:2342/preview123.jpg", c.SitePreview())
}

func TestConfig_SiteTitle(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "config.test", c.SiteTitle())
	c.params.SiteTitle = "Cats"
	assert.Equal(t, "Cats", c.SiteTitle())
}
