package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestAlbums(t *testing.T) {
	backupPath, err := filepath.Abs("./testdata/albums")

	if err != nil {
		t.Fatal(err)
	}

	if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	count, err := Albums(backupPath, true)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 31, count)

	count, err = Albums(backupPath, false)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, count)

	if err = os.RemoveAll(backupPath); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreAlbums(t *testing.T) {
	backupPath, err := filepath.Abs("./testdata/albums")

	if err != nil {
		t.Fatal(err)
	}

	if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	count, err := RestoreAlbums(backupPath, true)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, count)

	if err = os.RemoveAll(backupPath); err != nil {
		t.Fatal(err)
	}
}
