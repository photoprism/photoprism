package fs

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// SymlinksSupported tests if a storage path supports symlinks.
func SymlinksSupported(storagePath string) (bool, error) {
	linkName := filepath.Join(storagePath, uuid.NewString()+"-link.tmp")
	targetName := filepath.Join(storagePath, uuid.NewString()+".tmp")

	// Delete files after test.
	defer func(link, target string) {
		_ = os.Remove(link)
		_ = os.Remove(target)
	}(linkName, targetName)

	// Create empty test target file.
	if targetFile, err := os.OpenFile(targetName, os.O_RDONLY|os.O_CREATE, ModeFile); err != nil {
		return false, err
	} else if err = targetFile.Close(); err != nil {
		return false, err
	}

	// Create test link.
	if err := os.Symlink(filepath.Base(targetName), linkName); err != nil {
		return false, err
	}

	// Resolve and compare test target.
	if linkTarget, err := Resolve(linkName); err != nil {
		return false, err
	} else {
		return linkTarget == targetName, nil
	}
}
