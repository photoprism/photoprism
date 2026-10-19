package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// TestPhotos_StoredFilterNoMatch checks the album listing of a stored filter whose values have nothing usable left.
func TestPhotos_StoredFilterNoMatch(t *testing.T) {
	// PostgreSQL text columns cannot store invalid UTF-8, so such a filter only exists in MariaDB and SQLite.
	if entity.DbDialect() == dsn.DialectPostgreSQL {
		t.Skip("PostgreSQL cannot store a filter with invalid UTF-8")
	}

	// newAlbum creates an album that is removed after the test.
	newAlbum := func(t *testing.T, a *entity.Album) entity.Album {
		require.NotNil(t, a)
		require.NoError(t, a.Create())
		t.Cleanup(func() { _ = a.DeletePermanently() })
		return *a
	}

	t.Run("Moment", func(t *testing.T) {
		for _, filter := range []string{"\xff", "label:\xff", "public:true label:\xff*", "before:\xff"} {
			a := newAlbum(t, entity.NewMomentsAlbum("Stored Filter No Match", "stored-filter-no-match", filter))

			photos, count, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Count: 10})
			require.NoError(t, err, "%q", filter)
			assert.NotNil(t, photos)
			assert.Empty(t, photos, "%q", filter)
			assert.Zero(t, count)

			albumPhotos, err := AlbumPhotos(a, 10, false)
			require.NoError(t, err, "%q", filter)
			assert.Empty(t, albumPhotos, "%q", filter)

			results, err := PhotosGeo(form.SearchPhotosGeo{Scope: a.AlbumUID})
			require.NoError(t, err, "%q", filter)
			assert.NotNil(t, results)
			assert.Empty(t, results, "%q", filter)
		}
	})
	t.Run("Folder", func(t *testing.T) {
		// A folder album shows the pictures in its folder, also if its stored filter names it with invalid bytes.
		expected, _, err := Photos(form.SearchPhotos{Scope: "as6sg6bipogaaba1", Count: 100})
		require.NoError(t, err)
		require.NotEmpty(t, expected)

		a := newAlbum(t, entity.NewFolderAlbum("Stored Filter Folder", "1990/04", "path:\"\xc4\xd6\xdc\" public:true"))

		photos, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Count: 100})
		require.NoError(t, err)
		assert.Len(t, photos, len(expected))

		expectedGeo, err := PhotosGeo(form.SearchPhotosGeo{Scope: "as6sg6bipogaaba1"})
		require.NoError(t, err)
		results, err := PhotosGeo(form.SearchPhotosGeo{Scope: a.AlbumUID})
		require.NoError(t, err)
		assert.Len(t, results, len(expectedGeo))
	})
	t.Run("FolderWithoutPath", func(t *testing.T) {
		// A folder album without a path shows no pictures if its stored filter names a folder that is not valid UTF-8.
		for _, path := range []any{"", nil} {
			a := newAlbum(t, entity.NewFolderAlbum("Stored Filter No Path", "1990/04", "path:1990/04\xff public:true"))
			require.NoError(t, entity.UnscopedDb().Model(&entity.Album{}).Where("album_uid = ?", a.AlbumUID).UpdateColumn("album_path", path).Error)
			entity.FlushAlbumCache()

			photos, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Count: 100})
			require.NoError(t, err, "%v", path)
			assert.Empty(t, photos, "%v", path)

			results, err := PhotosGeo(form.SearchPhotosGeo{Scope: a.AlbumUID})
			require.NoError(t, err, "%v", path)
			assert.Empty(t, results, "%v", path)
		}
	})
	t.Run("FolderWithoutPathExact", func(t *testing.T) {
		// A folder album without a path compares the path of its stored filter exactly.
		expected, _, err := Photos(form.SearchPhotos{Scope: "as6sg6bipogaaba1", Count: 100})
		require.NoError(t, err)
		require.NotEmpty(t, expected)
		expectedGeo, err := PhotosGeo(form.SearchPhotosGeo{Scope: "as6sg6bipogaaba1"})
		require.NoError(t, err)
		require.NotEmpty(t, expectedGeo)

		for filter, found := range map[string]bool{
			"path:\"1990/04\" public:true":         true,
			"path:1990/04 public:true":             true,
			"path:1990* public:true":               false,
			"path:1990% public:true":               false,
			"path:1990/0% public:true":             false,
			"path:1990|1990/04 public:true":        false,
			"path:1990/* public:true":              false,
			"path:1990/04\u3000bridge public:true": false,
			"path:1990/04\u3000 public:true":       false,
			"path:1990/04\u00a0| public:true":      false,
			"path:\"1990/04":                       false,
			"path:1990/04":                         false,
			"path:1990/04 public:true label:x":     false,
			"public:true":                          false,
		} {
			for _, albumPath := range []any{"", nil} {
				a := newAlbum(t, entity.NewFolderAlbum("Stored Filter Exact", "1990/04", filter))
				require.NoError(t, entity.UnscopedDb().Model(&entity.Album{}).Where("album_uid = ?", a.AlbumUID).UpdateColumn("album_path", albumPath).Error)
				entity.FlushAlbumCache()

				photos, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Count: 100})
				require.NoError(t, err, "%q", filter)
				results, err := PhotosGeo(form.SearchPhotosGeo{Scope: a.AlbumUID})
				require.NoError(t, err, "%q", filter)

				if found {
					assert.Len(t, photos, len(expected), "%q", filter)
					assert.Len(t, results, len(expectedGeo), "%q", filter)
				} else {
					assert.Empty(t, photos, "%q", filter)
					assert.Empty(t, results, "%q", filter)
				}
			}
		}
	})
	t.Run("InvalidScope", func(t *testing.T) {
		// A value that can match nothing does not skip the check of the scope.
		_, _, err := Photos(form.SearchPhotos{Scope: "as6sg6bipogaab00", Query: "\xff ", Count: 10})
		assert.Equal(t, ErrInvalidId, err)
		_, err = PhotosGeo(form.SearchPhotosGeo{Scope: "as6sg6bipogaab00", Query: "\xff "})
		assert.Equal(t, ErrInvalidId, err)
	})
	t.Run("Valid", func(t *testing.T) {
		a := newAlbum(t, entity.NewMomentsAlbum("Stored Filter Valid", "stored-filter-valid", "public:true label:landscape"))

		photos, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Count: 10})
		require.NoError(t, err)
		assert.NotEmpty(t, photos)
	})
}
