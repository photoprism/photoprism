package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
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

func TestOptions_Load(t *testing.T) {
	t.Run("SetOptionsFromFile", func(t *testing.T) {
		o := NewOptions(CliTestContext())

		err := o.Load("testdata/config.yml")

		assert.Nil(t, err)

		assert.False(t, o.Debug)
		assert.False(t, o.ReadOnly)
		assert.Equal(t, "/srv/photoprism", o.AssetsPath)
		assert.Equal(t, "/srv/photoprism/cache", o.CachePath)
		assert.Equal(t, "/srv/photoprism/photos/originals", o.OriginalsPath)
		assert.Equal(t, "/srv/photoprism/photos/import", o.ImportPath)
		assert.Equal(t, "/srv/photoprism/temp", o.TempPath)
		assert.Equal(t, "1h34m9s", o.WakeupInterval.String())
		assert.NotEmpty(t, o.DatabaseDriver)
		assert.NotEmpty(t, o.DatabaseDSN)
		assert.Equal(t, 81, o.HttpPort)
	})
	t.Run("ChangeDSN", func(t *testing.T) {
		o := NewOptions(CliTestContext())
		o.DatabaseDSN = ""
		o.DatabaseDriver = "sqlite"
		o.DatabaseName = "photoprism"
		o.DatabasePassword = "photoprism"
		o.DatabaseServer = "mariadb:4001"
		o.DatabaseUser = "root"

		err := o.Load("testdata/dsnoptions.yml")

		assert.Nil(t, err)

		assert.True(t, o.DisableBackups)
		assert.Equal(t, uint64(2), o.FilesQuota)
		assert.Equal(t, "019a0ec6-e01e-7170-b22f-c962af0c93a5", o.NodeUUID)
		assert.Equal(t, "./storage/acceptance/originals", o.OriginalsPath)
		assert.Equal(t, "./storage/acceptance/import", o.ImportPath)
		assert.Equal(t, "http://localhost:2343/", o.SiteUrl)
		assert.Equal(t, "mysql", o.DatabaseDriver)
		assert.Equal(t, "acceptance:accpass@tcp(mariadb:12345)/accdb?charset=utf8mb4,utf8&parseTime=true", o.Deprecated.DatabaseDsn)
		assert.Empty(t, o.DatabaseName)
		assert.Empty(t, o.DatabaseServer)
		assert.Empty(t, o.DatabaseUser)
		assert.Empty(t, o.DatabasePassword)

		c := &Config{
			options: o,
		}

		assert.Equal(t, "accdb", c.DatabaseName())
		assert.Equal(t, "mariadb:12345", c.DatabaseServer())
		assert.Equal(t, "acceptance", c.DatabaseUser())
		assert.Equal(t, "accpass", c.DatabasePassword())
	})
}

func TestOptions_ExpandFilenames(t *testing.T) {
	p := Options{TempPath: "tmp", ImportPath: "import"}
	assert.Equal(t, "tmp", p.TempPath)
	assert.Equal(t, "import", p.ImportPath)
	p.expandFilenames()
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/tmp", p.TempPath)
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/import", p.ImportPath)
}
