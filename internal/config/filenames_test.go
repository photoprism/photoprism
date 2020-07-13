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
}

func TestConfig_SidecarYaml(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, false, c.SidecarYaml())
}

func TestConfig_SidecarPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ".photoprism", c.SidecarPath())
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
}

func TestConfig_StoragePath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata", c.StoragePath())
}

func TestConfig_TestdataPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/go/src/github.com/photoprism/photoprism/storage/testdata/testdata", c.TestdataPath())
}
