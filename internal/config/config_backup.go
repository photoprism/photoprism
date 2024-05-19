package config

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	DefaultBackupSchedule = "daily"
	DefaultBackupRetain   = 3
)

// BackupPath returns the backup storage path based on the specified type, or the base path if none is specified.
func (c *Config) BackupPath(backupType string) string {
	if s := clean.TypeLowerUnderscore(backupType); s == "" {
		return c.BackupBasePath()
	} else {
		return filepath.Join(c.BackupBasePath(), s)
	}
}

// BackupBasePath returns the backup storage base path.
func (c *Config) BackupBasePath() string {
	if fs.PathWritable(c.options.BackupPath) {
		return fs.Abs(c.options.BackupPath)
	}

	return filepath.Join(c.StoragePath(), "backup")
}

// BackupSchedule returns the backup schedule in cron format, e.g. "0 12 * * *" for daily at noon.
func (c *Config) BackupSchedule() string {
	return Schedule(c.options.BackupSchedule)
}

// BackupRetain returns the maximum number of SQL database dumps to keep, or -1 to keep all.
func (c *Config) BackupRetain() int {
	if c.options.BackupRetain == 0 {
		return DefaultBackupRetain
	} else if c.options.BackupRetain < -1 {
		return -1
	}

	return c.options.BackupRetain
}

// BackupDatabase checks if index database backups should be created based on the configured schedule.
func (c *Config) BackupDatabase() bool {
	return c.options.BackupDatabase
}

// BackupDatabasePath returns the backup path for index database dumps.
func (c *Config) BackupDatabasePath() string {
	if driver := c.DatabaseDriver(); driver != "" {
		return c.BackupPath(driver)
	}

	return c.BackupPath("index")
}

// BackupAlbums checks if album YAML file backups should be created based on the configured schedule.
func (c *Config) BackupAlbums() bool {
	return c.options.BackupAlbums
}

// BackupAlbumsPath returns the backup path for album YAML files.
func (c *Config) BackupAlbumsPath() string {
	if dir := filepath.Join(c.StoragePath(), "albums"); fs.PathExists(dir) {
		return dir
	}

	return c.BackupPath("albums")
}

// DisableBackups checks if creating and updating sidecar YAML files should be disabled.
func (c *Config) DisableBackups() bool {
	if !c.SidecarWritable() {
		return true
	}

	return c.options.DisableBackups
}
