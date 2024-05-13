package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_BackupPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupPath(""), "/storage/testdata/backup")
}

func TestConfig_BackupBasePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupBasePath(), "/storage/testdata/backup")
}

func TestConfig_BackupAlbumsPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupAlbumsPath(), "/albums")
}

func TestConfig_BackupIndexPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupIndexPath(), "/storage/testdata/backup/sqlite")
}

func TestConfig_BackupSchedule(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultBackupSchedule, c.BackupSchedule())
}

func TestConfig_BackupRetain(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultBackupRetain, c.BackupRetain())
}

func TestConfig_BackupIndex(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.BackupIndex())
	c.options.BackupIndex = true
	assert.True(t, c.BackupIndex())
	c.options.BackupIndex = false
	assert.False(t, c.BackupIndex())
}

func TestConfig_BackupAlbums(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.BackupAlbums())
	c.options.BackupAlbums = false
	assert.False(t, c.BackupAlbums())
	c.options.BackupAlbums = true
	assert.True(t, c.BackupAlbums())

}

func TestConfig_DisableBackups(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableBackups())
	c.options.DisableBackups = true
	assert.True(t, c.DisableBackups())
	c.options.DisableBackups = false
	assert.False(t, c.DisableBackups())
}
