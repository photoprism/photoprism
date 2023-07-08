package config

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// createError returns a new directory create error.
func createError(path string, err error) (result error) {
	if fs.FileExists(path) {
		result = fmt.Errorf("directory path %s is a file, please check your configuration", clean.Log(path))
	} else {
		result = fmt.Errorf("failed to create the directory %s, check configuration and permissions", clean.Log(path))
	}

	log.Debug(err)

	return result
}

// notFoundError returns a new directory not found error.
func notFoundError(name string) error {
	return fmt.Errorf("invalid %s path, check configuration and permissions", clean.Log(name))
}
