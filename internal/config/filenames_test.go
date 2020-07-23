package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_FindExecutable(t *testing.T) {
	assert.Equal(t, "", findExecutable("yyy", "xxx"))
}

func TestConfig_SidecarJson(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, false, c.SidecarJson())
	c.params.ReadOnly = true
	assert.Equal(t, false, c.SidecarJson())
}

func TestConfig_SidecarYaml(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, false, c.SidecarYaml())
	c.params.ReadOnly = true
	assert.Equal(t, false, c.SidecarJson())
}

func TestConfig_SidecarPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, ".photoprism", c.SidecarPath())
	c.params.SidecarPath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/sidecar", c.SidecarPath())
}

func TestConfig_SidecarPathIsAbs(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, false, c.SidecarPathIsAbs())
}

func TestConfig_SidecarWritable(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, true, c.SidecarWritable())
}

func TestConfig_FFmpegBin(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/usr/bin/ffmpeg", c.FFmpegBin())
}

func TestConfig_TempPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/temp", c.TempPath())
	c.params.TempPath = ""
	assert.Equal(t, "/tmp/photoprism", c.TempPath())
}

func TestConfig_CachePath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/cache", c.CachePath())
	c.params.CachePath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/cache", c.CachePath())
}

func TestConfig_StoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata", c.StoragePath())
	c.params.StoragePath = ""
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/originals/.photoprism/storage", c.StoragePath())
}

func TestConfig_TestdataPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/testdata", c.TestdataPath())
}

func TestConfig_CreateDirectories(t *testing.T) {
	c := NewConfig(CliTestContext())
	err := c.CreateDirectories()

	if err != nil {
		t.Fatal(err)
	}
}

func TestConfig_ConfigFile2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/settings/photoprism.yml", c.ConfigFile())
	c.params.ConfigFile = "/go/src/github.com/photoprism/photoprism/internal/config/testdata/config.yml"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/testdata/config.yml", c.ConfigFile())
}

func TestConfig_PIDFilename2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/photoprism.pid", c.PIDFilename())
	c.params.PIDFilename = "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.pid"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.pid", c.PIDFilename())
}

func TestConfig_LogFilename2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/photoprism.log", c.LogFilename())
	c.params.LogFilename = "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.log"
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/internal/config/testdata/test.log", c.LogFilename())
}

func TestConfig_OriginalsPath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/originals", c.OriginalsPath())
	c.params.OriginalsPath = ""
	assert.Equal(t, "", c.OriginalsPath())
}

func TestConfig_ImportPath2(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/import", c.ImportPath())
	c.params.ImportPath = ""
	assert.Equal(t, "", c.ImportPath())
}
