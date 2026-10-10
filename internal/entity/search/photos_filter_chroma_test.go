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
	t.Run("EarlierMono", func(t *testing.T) {
		// Earlier versions stored 0 for monochrome files.
		earlier := entity.PhotoFixtures.Get("Photo12").PhotoUID
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", earlier).UpdateColumn("file_chroma", 0).Error)
		t.Cleanup(func() {
			_ = entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", earlier).UpdateColumn("file_chroma", 1).Error
		})

		assert.Contains(t, find(t, "mono:true"), earlier)
		assert.NotContains(t, find(t, "chroma:4"), earlier)
		assert.NotContains(t, find(t, "chroma:20"), earlier)
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
			assert.LessOrEqual(t, chroma[0], int16(1), "chroma of %s", r.PhotoUID)
			assert.GreaterOrEqual(t, chroma[0], int16(0), "chroma of %s", r.PhotoUID)
		}

		// A monochrome value stored by earlier versions is found too.
		earlier := results[0].PhotoUID
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", earlier).UpdateColumn("file_chroma", 0).Error)
		t.Cleanup(func() {
			_ = entity.UnscopedDb().Model(&entity.File{}).Where("photo_uid = ? AND file_primary = TRUE", earlier).UpdateColumn("file_chroma", 1).Error
		})

		results, err = PhotosGeo(f)
		require.NoError(t, err)

		uids := make([]string, 0, len(results))

		for _, r := range results {
			uids = append(uids, r.PhotoUID)
		}

		assert.Contains(t, uids, earlier)

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
			assert.NotEqual(t, earlier, r.PhotoUID)
		}
	})
	t.Run("HighChroma", func(t *testing.T) {
		uids := find(t, "chroma:20")
		assert.NotContains(t, uids, mono)
		assert.NotContains(t, uids, low)
		assert.NotEmpty(t, uids)
	})
}
