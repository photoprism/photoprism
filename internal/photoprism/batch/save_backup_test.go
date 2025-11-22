package batch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestUpdateAlbumBackups(t *testing.T) {
	conf := get.Config()
	require.NotNil(t, conf)
	album := entity.AlbumFixtures.Get("christmas2030")
	require.True(t, album.HasID())

	t.Run("WritesFile", func(t *testing.T) {
		original := conf.BackupAlbums()
		conf.Options().BackupAlbums = true
		t.Cleanup(func() { conf.Options().BackupAlbums = original })

		backupFile, _, err := album.YamlFileName(conf.BackupAlbumsPath())
		require.NoError(t, err)
		_ = os.Remove(backupFile)

		values := &PhotosForm{
			Albums: Items{
				Action: ActionUpdate,
				Items: []Item{
					{Value: album.AlbumUID, Action: ActionAdd},
					{Value: album.AlbumUID, Action: ActionAdd},
					{Value: "invalid", Action: ActionAdd},
				},
			},
		}

		updateAlbumBackups(values)
		require.FileExists(t, backupFile)

		t.Cleanup(func() { _ = os.Remove(backupFile) })
	})
	t.Run("SkipsWhenDisabled", func(t *testing.T) {
		original := conf.BackupAlbums()
		conf.Options().BackupAlbums = false
		t.Cleanup(func() { conf.Options().BackupAlbums = original })

		backupFile := filepath.Join(conf.BackupAlbumsPath(), album.AlbumType, album.AlbumUID+".yml")
		_ = os.Remove(backupFile)

		values := &PhotosForm{
			Albums: Items{
				Action: ActionUpdate,
				Items:  []Item{{Value: album.AlbumUID, Action: ActionAdd}},
			},
		}

		updateAlbumBackups(values)
		_, err := os.Stat(backupFile)
		require.True(t, os.IsNotExist(err))
	})
}
