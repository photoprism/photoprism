package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// TestNoMatch checks that a filter value of only NUL bytes and invalid UTF-8 returns an empty result without
// an error, while the same search without it does not.
func TestNoMatch(t *testing.T) {
	t.Run("Photos", func(t *testing.T) {
		for _, f := range []form.SearchPhotos{{Path: "\xff"}, {Name: "\xff"}, {Filename: "\x00"}, {Query: "\x00"},
			{Query: "path:\xff"}, {Label: "\xfe\xff"}, {Query: "\x00 "}, {Query: "\xff|"}, {Path: "\xff "}, {Path: "\xff*"},
			{Label: "\xff "}, {Query: "before:\xff"}, {Query: "after:\xff"}, {Query: "label:\xff*"}} {
			f.Count, f.Merged = 10, true
			photos, count, err := Photos(f)
			require.NoError(t, err, "%+v", f)
			assert.NotNil(t, photos)
			assert.Empty(t, photos, "%+v", f)
			assert.Zero(t, count)
		}

		photos, _, err := Photos(form.SearchPhotos{Path: "2790*", Count: 10, Merged: true})
		require.NoError(t, err)
		assert.NotEmpty(t, photos)
		photos, _, err = Photos(form.SearchPhotos{Count: 10, Merged: true})
		require.NoError(t, err)
		assert.NotEmpty(t, photos)
	})
	t.Run("PhotosGeo", func(t *testing.T) {
		results, err := PhotosGeo(form.SearchPhotosGeo{Path: "\xff"})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
		results, err = PhotosGeo(form.SearchPhotosGeo{})
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
	t.Run("Albums", func(t *testing.T) {
		results, err := Albums(form.SearchAlbums{Query: "\xff", Count: 100})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
		results, err = Albums(form.SearchAlbums{Order: "\xff", Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
	t.Run("Labels", func(t *testing.T) {
		results, err := Labels(form.SearchLabels{Query: "\x00", Count: 100})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
		results, err = Labels(form.SearchLabels{Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
	t.Run("Subjects", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Name: "\xff", Count: 100})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
		results, err = Subjects(form.SearchSubjects{Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
	t.Run("Faces", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{UID: "\xff", Count: 100})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
		results, err = Faces(form.SearchFaces{Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, results)
	})
	t.Run("CamerasLenses", func(t *testing.T) {
		cameras, err := Cameras(form.SearchCameras{Query: "\xff", Count: 100})
		require.NoError(t, err)
		assert.NotNil(t, cameras)
		assert.Empty(t, cameras)
		cameras, err = Cameras(form.SearchCameras{Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, cameras)
		lenses, err := Lenses(form.SearchLenses{Query: "\x00", Count: 100})
		require.NoError(t, err)
		assert.NotNil(t, lenses)
		assert.Empty(t, lenses)
		lenses, err = Lenses(form.SearchLenses{Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, lenses)
	})
	t.Run("AlbumPhotos", func(t *testing.T) {
		// A stored filter whose value matches nothing returns an empty result.
		results, err := AlbumPhotos(entity.Album{AlbumUID: "as6sg6bipotaab19", AlbumType: entity.AlbumMoment, AlbumFilter: "\xff"}, 10, false)
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})
	t.Run("Users", func(t *testing.T) {
		for _, name := range []string{"\xff", "\xff ", "\x00*"} {
			users, err := Users(form.SearchUsers{Name: name, Count: 100})
			require.NoError(t, err)
			assert.NotNil(t, users)
			assert.Empty(t, users, "%q", name)
		}
		users, err := Users(form.SearchUsers{Name: "\xff", Count: 100})
		require.NoError(t, err)
		assert.Empty(t, users)
		users, err = Users(form.SearchUsers{Count: 100})
		require.NoError(t, err)
		assert.NotEmpty(t, users)
	})
}
