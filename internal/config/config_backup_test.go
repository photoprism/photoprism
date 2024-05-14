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

func TestConfig_BackupSchedule(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultBackupSchedule, c.BackupSchedule())
}

func TestConfig_BackupRetain(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultBackupRetain, c.BackupRetain())
}

func TestConfig_BackupDatabase(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.BackupDatabase())
	c.options.BackupDatabase = false
	assert.False(t, c.BackupDatabase())
	c.options.BackupDatabase = true
	assert.True(t, c.BackupDatabase())
}

func TestConfig_BackupDatabasePath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupDatabasePath(), "/storage/testdata/backup/sqlite")
}

func TestConfig_BackupAlbums(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.True(t, c.BackupAlbums())
	c.options.BackupAlbums = false
	assert.False(t, c.BackupAlbums())
	c.options.BackupAlbums = true
	assert.True(t, c.BackupAlbums())

}

func TestConfig_BackupAlbumsPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupAlbumsPath(), "/albums")
}

func TestConfig_DisableBackups(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.DisableBackups())
	c.options.DisableBackups = true
	assert.True(t, c.DisableBackups())
	c.options.DisableBackups = false
	assert.False(t, c.DisableBackups())
}
