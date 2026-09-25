package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
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

// writeAlbumYaml writes a minimal album backup file with the given type and returns the album UID.
func writeAlbumYaml(t *testing.T, backupPath, albumType, title string) string {
	t.Helper()

	uid := rnd.GenerateUID(entity.AlbumUID)
	data := "UID: " + uid + "\nSlug: " + txt.Slug(title) + "\nType: " + albumType + "\nTitle: " + title + "\nFilter: public:true\n"
	require.NoError(t, os.MkdirAll(filepath.Join(backupPath, entity.AlbumManual), fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(backupPath, entity.AlbumManual, uid+".yml"), []byte(data), fs.ModeFile))

	return uid
}

func TestWriteAlbumYaml(t *testing.T) {
	backupPath := t.TempDir()
	uid := writeAlbumYaml(t, backupPath, entity.AlbumManual, "Yaml Helper Test")

	a := entity.Album{}
	require.NoError(t, a.LoadFromYaml(filepath.Join(backupPath, entity.AlbumManual, uid+".yml")))
	assert.Equal(t, uid, a.AlbumUID)
	assert.Equal(t, entity.AlbumManual, a.AlbumType)
	assert.Equal(t, "public:true", a.AlbumFilter)
}

func TestRestoreAlbums_Types(t *testing.T) {
	t.Run("KnownType", func(t *testing.T) {
		backupPath := t.TempDir()
		uid := writeAlbumYaml(t, backupPath, entity.AlbumManual, "Restore Known Type Test")

		count, err := RestoreAlbums(backupPath, true)

		require.NoError(t, err)
		assert.Equal(t, 1, count)
		found := entity.FindAlbum(entity.Album{AlbumUID: uid})
		require.NotNil(t, found)
		t.Cleanup(func() { _ = found.DeletePermanently() })
		assert.Equal(t, entity.AlbumManual, found.AlbumType)
	})
	t.Run("UnknownType", func(t *testing.T) {
		backupPath := t.TempDir()
		uid := writeAlbumYaml(t, backupPath, "calendar", "Restore Unknown Type Test")

		count, err := RestoreAlbums(backupPath, true)

		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Nil(t, entity.FindAlbum(entity.Album{AlbumUID: uid}))
	})
}

func TestAlbums_UnknownTypeSkipped(t *testing.T) {
	prev := backupAlbumsTime
	t.Cleanup(func() { backupAlbumsTime = prev })

	album := entity.NewAlbum("Backup Unknown Type Test", entity.AlbumManual)
	require.NoError(t, album.Save())
	t.Cleanup(func() { _ = album.DeletePermanently() })
	require.NoError(t, album.Update("album_type", "calendar"))

	backupPath := t.TempDir()

	count, err := Albums(backupPath, true)

	require.NoError(t, err)
	assert.Equal(t, 36, count)
	matches, err := filepath.Glob(filepath.Join(backupPath, "*", album.AlbumUID+".yml"))
	require.NoError(t, err)
	assert.Empty(t, matches)
}
