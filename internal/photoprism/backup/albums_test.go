package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestAlbums(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "testdata", "albums")
	require.NoError(t, os.MkdirAll(backupPath, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(backupPath) })

	count, err := Albums(backupPath, true)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 35, count)

	count, err = Albums(backupPath, false)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, count)
}

func TestRestoreAlbums(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "testdata", "albums")
	require.NoError(t, os.MkdirAll(backupPath, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(backupPath) })

	count, err := RestoreAlbums(backupPath, true)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, count)
}
