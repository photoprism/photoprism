package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestFolderAlbumPath(t *testing.T) {
	t.Run("Folder", func(t *testing.T) {
		assert.Equal(t, "2024/07/", folderAlbumPath(entity.Album{AlbumType: entity.AlbumFolder, AlbumPath: "2024/07"}))
		assert.Equal(t, "a|b*/", folderAlbumPath(entity.Album{AlbumType: entity.AlbumFolder, AlbumPath: "a|b*"}))
		assert.Equal(t, "2024/07/", folderAlbumPath(entity.Album{AlbumType: entity.AlbumFolder, AlbumPath: "/2024/07/"}))
		assert.Equal(t, "null/", folderAlbumPath(entity.Album{AlbumType: entity.AlbumFolder, AlbumPath: "null"}))
	})
	t.Run("Other", func(t *testing.T) {
		assert.Equal(t, "", folderAlbumPath(entity.Album{AlbumType: entity.AlbumFolder}))
		assert.Equal(t, "", folderAlbumPath(entity.Album{AlbumType: entity.AlbumFolder, AlbumPath: "//"}))
		assert.Equal(t, "", folderAlbumPath(entity.Album{AlbumType: entity.AlbumManual, AlbumPath: "2024/07"}))
		assert.Equal(t, "", folderAlbumPath(entity.Album{AlbumType: entity.AlbumMoment, AlbumPath: "2024/07"}))
	})
}

// folderAlbumTestCases returns folder album paths paired with a different folder whose pictures the
// album must not show.
func folderAlbumTestCases() []struct{ name, albumPath, otherPath string } {
	base := "zz-folder-album-" + rnd.Base36(6)
	top := "zz-folder-top-" + rnd.Base36(6)

	return []struct{ name, albumPath, otherPath string }{
		{"Underscore", base + "/u_a", base + "/uXa"},
		{"Percent", base + "/p%a", base + "/pXYa"},
		{"Star", base + "/s*", base + "/sAnything"},
		{"Pipe", base + "/p|" + top, top},
		{"Case", base + "/Trip", base + "/trip"},
		{"EmptyValue", "null", base + "/other"},
	}
}

// folderAlbumTestAlbum creates a folder album for albumPath as the folder sync does and removes it
// when the test ends.
func folderAlbumTestAlbum(t *testing.T, name, albumPath string) *entity.Album {
	t.Helper()

	f := form.SearchPhotos{Path: albumPath, Public: true}
	album := entity.NewFolderAlbum("Folder Album "+name, albumPath, f.Serialize())
	require.NotNil(t, album)
	require.NoError(t, album.Create())

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Delete(album).Error
		entity.FlushAlbumCache()
	})

	return album
}

func TestSharedPhotos_FolderAlbumPath(t *testing.T) {
	for _, tc := range folderAlbumTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			own := scopeBasePathPhoto(t, tc.albumPath, "own")
			other := scopeBasePathPhoto(t, tc.otherPath, "other")
			album := folderAlbumTestAlbum(t, tc.name, tc.albumPath)

			results, _, err := SharedPhotos(form.SearchPhotos{Scope: album.AlbumUID, Count: 100})
			require.NoError(t, err)

			var uids []string

			for _, r := range results {
				uids = append(uids, r.PhotoUID)
			}

			assert.Contains(t, uids, own.PhotoUID)
			assert.NotContains(t, uids, other.PhotoUID)
		})
	}
}

func TestUserPhotosGeo_FolderAlbumPath(t *testing.T) {
	for _, tc := range folderAlbumTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			own := scopeBasePathPhoto(t, tc.albumPath, "own")
			other := scopeBasePathPhoto(t, tc.otherPath, "other")
			album := folderAlbumTestAlbum(t, tc.name, tc.albumPath)

			results, err := UserPhotosGeo(form.SearchPhotosGeo{Scope: album.AlbumUID, Count: 1000}, nil)
			require.NoError(t, err)

			var uids []string

			for _, r := range results {
				uids = append(uids, r.PhotoUID)
			}

			assert.Contains(t, uids, own.PhotoUID)
			assert.NotContains(t, uids, other.PhotoUID)
		})
	}
}
