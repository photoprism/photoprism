package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_BackupPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Contains(t, c.BackupPath(), "/storage/testdata/backup")
}

func TestConfig_BackupIndex(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.BackupIndex())
}

func TestConfig_BackupAlbums(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.BackupAlbums())
}

func TestConfig_BackupRetain(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultBackupRetain, c.BackupRetain())
}

func TestConfig_BackupSchedule(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultBackupSchedule, c.BackupSchedule())
}
