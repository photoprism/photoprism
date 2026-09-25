package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	assert.Equal(t, 36, count)

	count, err = Albums(backupPath, false)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, count)

	if err = os.RemoveAll(backupPath); err != nil {
		t.Fatal(err)
	}
}

func TestAlbums_FailedRunIsRetried(t *testing.T) {
	prev := backupAlbumsTime
	t.Cleanup(func() { backupAlbumsTime = prev })
	backupAlbumsTime = time.Time{}

	// A file in place of the manual album dir makes only those saves fail, so the run is partial.
	backupPath := t.TempDir()
	blocker := filepath.Join(backupPath, "album")
	require.NoError(t, os.WriteFile(blocker, nil, fs.ModeFile))

	count, err := Albums(backupPath, false)

	require.Error(t, err)
	assert.Greater(t, count, 0)
	assert.Less(t, count, 36)
	assert.True(t, backupAlbumsTime.IsZero())

	require.NoError(t, os.Remove(blocker))

	count, err = Albums(backupPath, false)

	require.NoError(t, err)
	assert.Equal(t, 36, count)
	assert.False(t, backupAlbumsTime.IsZero())
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
