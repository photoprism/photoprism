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
