package fs

import (
	"errors"
	"path/filepath"
)

// Resolve returns the absolute file path, with all symlinks resolved.
func Resolve(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("no such file or directory")
	}

	if target, err := filepath.EvalSymlinks(filePath); err != nil {
		return "", errors.New("no such file or directory")
	} else if target, err = filepath.Abs(target); target != "" {
		return target, err
	} else {
		return filepath.Abs(filePath)
	}
}
