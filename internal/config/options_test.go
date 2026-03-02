package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

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
	assert.NotEmpty(t, c.DatabaseDSN)
	assert.Equal(t, 81, c.HttpPort)
}

func TestOptions_LoadDoesNotOverrideEdition(t *testing.T) {
	c := NewOptions(NewTestContext([]string{}))
	assert.Equal(t, "ce", c.Edition)
	assert.Equal(t, "PhotoPrism", c.Name)
	assert.Equal(t, "PhotoPrism®", c.About)
	assert.Equal(t, "test", c.Version)
	assert.Equal(t, "(c) 2018-2025 PhotoPrism UG. All rights reserved.", c.Copyright)

	dir := t.TempDir()
	fileName := filepath.Join(dir, "options.yml")
	content := strings.Join([]string{
		"Edition: portal",
		"Name: Evil Name",
		"About: Evil About",
		"Version: 9.9.9",
		"Copyright: Evil Copyright",
		"HttpPort: 4242",
		"",
	}, "\n")
	assert.NoError(t, os.WriteFile(fileName, []byte(content), fs.ModeFile))

	assert.NoError(t, c.Load(fileName))
	assert.Equal(t, "ce", c.Edition)
	assert.Equal(t, "PhotoPrism", c.Name)
	assert.Equal(t, "PhotoPrism®", c.About)
	assert.Equal(t, "test", c.Version)
	assert.Equal(t, "(c) 2018-2025 PhotoPrism UG. All rights reserved.", c.Copyright)
	assert.Equal(t, 4242, c.HttpPort)
}

func TestOptions_MarshalDoesNotIncludeBuildMetadata(t *testing.T) {
	c := NewOptions(NewTestContext([]string{}))
	c.HttpPort = 4242

	data, err := yaml.Marshal(c)
	assert.NoError(t, err)

	var values map[string]any
	assert.NoError(t, yaml.Unmarshal(data, &values))

	_, hasName := values["Name"]
	_, hasAbout := values["About"]
	_, hasEdition := values["Edition"]
	_, hasVersion := values["Version"]
	_, hasCopyright := values["Copyright"]

	assert.False(t, hasName)
	assert.False(t, hasAbout)
	assert.False(t, hasEdition)
	assert.False(t, hasVersion)
	assert.False(t, hasCopyright)
	assert.Equal(t, 4242, values["HttpPort"])
}

func TestOptions_ExpandFilenames(t *testing.T) {
	p := Options{TempPath: "tmp", ImportPath: "import"}
	assert.Equal(t, "tmp", p.TempPath)
	assert.Equal(t, "import", p.ImportPath)
	p.expandFilenames()
	assert.Equal(t, ProjectRoot+"/internal/config/tmp", p.TempPath)
	assert.Equal(t, ProjectRoot+"/internal/config/import", p.ImportPath)
}
