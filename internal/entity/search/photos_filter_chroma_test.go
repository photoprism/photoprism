package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotos_ChromaFilters(t *testing.T) {
	// find returns the UIDs of the pictures that a query finds.
	find := func(t *testing.T, query string) []string {
		f := form.SearchPhotos{Query: query, Count: 1000, Merged: true}
		require.NoError(t, f.ParseQueryString())

		photos, _, err := Photos(f)
		require.NoError(t, err)

		uids := make([]string, 0, len(photos))

		for _, p := range photos {
			uids = append(uids, p.PhotoUID)
		}

		return uids
	}

	// Primary files with a chroma of 1 are monochrome; one of them is given a chroma of 4, i.e. few colors.
	mono := entity.PhotoFixtures.Get("Photo08").PhotoUID
	low := entity.PhotoFixtures.Get("Photo11").PhotoUID

	stmt := entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", low)
	require.NoError(t, stmt.UpdateColumn("file_chroma", 4).Error)
	t.Cleanup(func() {
		_ = entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", low).UpdateColumn("file_chroma", 1).Error
	})

	t.Run("Mono", func(t *testing.T) {
		uids := find(t, "mono:true")
		assert.Contains(t, uids, mono)
		assert.NotContains(t, uids, low)
	})
	t.Run("LowChroma", func(t *testing.T) {
		uids := find(t, "chroma:4")
		assert.NotContains(t, uids, mono)
		assert.Contains(t, uids, low)
	})
	t.Run("Geo", func(t *testing.T) {
		f := form.SearchPhotosGeo{Query: "mono:true", Count: 1000}
		require.NoError(t, f.ParseQueryString())

		results, err := PhotosGeo(f)
		require.NoError(t, err)
		require.NotEmpty(t, results)

		for _, r := range results {
			var chroma []int16
			require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", r.PhotoUID).Pluck("file_chroma", &chroma).Error)
			require.NotEmpty(t, chroma)
			assert.Equal(t, int16(1), chroma[0], "chroma of %s", r.PhotoUID)
		}

		// Low chroma values exclude monochrome files.
		lowChroma := form.SearchPhotosGeo{Query: "chroma:4", Count: 1000}
		require.NoError(t, lowChroma.ParseQueryString())

		results, err = PhotosGeo(lowChroma)
		require.NoError(t, err)

		for _, r := range results {
			var chroma []int16
			require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", r.PhotoUID).Pluck("file_chroma", &chroma).Error)
			require.NotEmpty(t, chroma)
			assert.Greater(t, chroma[0], int16(1), "chroma of %s", r.PhotoUID)
		}
	})
	t.Run("HighChroma", func(t *testing.T) {
		uids := find(t, "chroma:20")
		assert.NotContains(t, uids, mono)
		assert.NotContains(t, uids, low)
		assert.NotEmpty(t, uids)
	})
	t.Run("Percent", func(t *testing.T) {
		// Values are stored as percent plus one, and the filters refer to the percentage.
		boundary := entity.PhotoFixtures.Get("Photo12").PhotoUID
		set := func(chroma int16) {
			require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", boundary).UpdateColumn("file_chroma", chroma).Error)
		}
		t.Cleanup(func() { set(1) })

		set(2)
		assert.NotContains(t, find(t, "mono:true"), boundary, "1%")
		assert.Contains(t, find(t, "chroma:4"), boundary, "1%")
		set(5)
		assert.Contains(t, find(t, "chroma:4"), boundary, "4%")
		set(6)
		assert.NotContains(t, find(t, "chroma:4"), boundary, "5%")
		set(21)
		assert.NotContains(t, find(t, "chroma:20"), boundary, "20%")
		set(22)
		assert.Contains(t, find(t, "chroma:20"), boundary, "21%")
		set(100)
		assert.Contains(t, find(t, "chroma:98"), boundary, "99%")
		assert.NotContains(t, find(t, "chroma:99"), boundary, "99%")
		assert.Empty(t, find(t, "chroma:100"))
		assert.Empty(t, find(t, "chroma:32767"))

		geo := form.SearchPhotosGeo{Query: "chroma:32767", Count: 1000}
		require.NoError(t, geo.ParseQueryString())
		results, err := PhotosGeo(geo)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
