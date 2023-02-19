package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	tmpDir := os.TempDir()

	linkName := filepath.Join(tmpDir, uuid.NewString()+"-link.tmp")
	targetName := filepath.Join(tmpDir, uuid.NewString()+".tmp")

	// Delete files after test.
	defer func(link, target string) {
		_ = os.Remove(link)
		_ = os.Remove(target)
	}(linkName, targetName)

	// Create empty test target file.
	if targetFile, err := os.OpenFile(targetName, os.O_RDONLY|os.O_CREATE, ModeFile); err != nil {
		t.Fatal(err)
	} else if err = targetFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(targetName, linkName); err != nil {
		t.Fatal(err)
	}

	if result, err := Resolve(linkName); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, targetName, result)
	}
}
