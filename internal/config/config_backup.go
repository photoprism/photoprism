package config

import (
	"path/filepath"

	"github.com/robfig/cron/v3"

	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	DefaultBackupSchedule = "0 12 * * *"
	DefaultBackupRetain   = 14
)

// BackupPath returns the backup storage path.
func (c *Config) BackupPath() string {
	if fs.PathWritable(c.options.BackupPath) {
		return fs.Abs(c.options.BackupPath)
	}

	return filepath.Join(c.StoragePath(), "backup")
}

// BackupIndex checks if SQL database dumps should be created based on the configured schedule.
func (c *Config) BackupIndex() bool {
	return c.options.BackupIndex
}

// BackupAlbums checks if album YAML file backups should be created based on the configured schedule.
func (c *Config) BackupAlbums() bool {
	return c.options.BackupAlbums
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

// BackupSchedule returns the backup schedule in cron format, e.g. "0 12 * * *" for daily at noon.
func (c *Config) BackupSchedule() string {
	if c.options.BackupSchedule == "" {
		return ""
	} else if _, err := cron.ParseStandard(c.options.BackupSchedule); err != nil {
		log.Tracef("config: invalid backup schedule (%s)", err)
		return ""
	}

	return c.options.BackupSchedule
}
