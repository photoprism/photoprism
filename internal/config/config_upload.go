package config

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// UploadNSFW checks if NSFW photos can be uploaded.
func (c *Config) UploadNSFW() bool {
	return c.options.UploadNSFW
}

// UploadAllow returns the file extensions that users are allowed to upload.
func (c *Config) UploadAllow() fs.ExtList {
	return fs.NewExtList(c.options.UploadAllow)
}

// UploadArchives checks if zip and tar.gz archives are allowed to be uploaded.
func (c *Config) UploadArchives() bool {
	return c.options.UploadArchives
}

// UploadLimit returns the maximum aggregated size of uploaded files in MB, or -1 if there is none.
// A larger value is clamped to MaxSizeLimit.
func (c *Config) UploadLimit() int {
	if c.options.UploadLimit <= 0 {
		return -1
	} else if c.options.UploadLimit > MaxSizeLimit {
		return MaxSizeLimit
	}

	return c.options.UploadLimit
}

// UploadLimitBytes returns the maximum aggregated size of uploaded files in bytes.
func (c *Config) UploadLimitBytes() int64 {
	if result := c.UploadLimit(); result <= 0 {
		return -1
	} else {
		return int64(result) * 1024 * 1024
	}
}
