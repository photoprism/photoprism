package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestPhotosAlbumQuery checks that user predicates intersect each album's membership.
func TestPhotosAlbumQuery(t *testing.T) {
	for _, tc := range []struct{ name, kind string }{{"Album", entity.AlbumManual}, {"Folder", entity.AlbumFolder}, {"Moment", entity.AlbumMoment}, {"Month", entity.AlbumMonth}, {"State", entity.AlbumState}} {
		t.Run(tc.name, func(t *testing.T) {
			kind := tc.kind
			base := "album-query-" + rnd.Base36(8)
			own := scopeBasePathPhoto(t, base, "zebra")
			other := scopeBasePathPhoto(t, base+"/other", "zebra")
			removed := scopeBasePathPhoto(t, base, "removed")
			require.NoError(t, entity.Db().Model(own).Update("photo_title", "Zebra").Error)
			require.NoError(t, entity.Db().Model(other).Update("photo_title", "Zebra").Error)
			own.PhotoTitle, other.PhotoTitle = "Zebra", "Zebra"
			require.NoError(t, own.IndexKeywords())
			require.NoError(t, other.IndexKeywords())
			t.Cleanup(func() {
				_ = entity.Db().Where("photo_id IN (?)", []uint{own.ID, other.ID}).Delete(&entity.PhotoKeyword{}).Error
			})
			a := entity.NewAlbum(base, kind)
			a.CreatedBy = entity.UserFixtures.Get("guest").UserUID
			a.AlbumPath = base
			if kind != entity.AlbumManual {
				a.AlbumFilter = "path:" + base + "/ public:true"
			}
			require.NoError(t, a.Create())
			t.Cleanup(func() {
				_ = entity.UnscopedDb().Where("album_uid = ?", a.AlbumUID).Delete(&entity.PhotoAlbum{}).Error
				_ = a.DeletePermanently()
				entity.FlushAlbumCache()
			})
			if kind == entity.AlbumManual {
				require.NoError(t, entity.Db().Create(&entity.PhotoAlbum{AlbumUID: a.AlbumUID, PhotoUID: own.PhotoUID}).Error)
			}
			require.NoError(t, entity.Db().Create(&entity.PhotoAlbum{AlbumUID: a.AlbumUID, PhotoUID: removed.PhotoUID, Hidden: true}).Error)

			for _, q := range []string{"zebra", "title:zebra", "type:image", "path:" + base + "/"} {
				results, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Query: q, Merged: true})
				require.NoError(t, err, q)
				require.Len(t, results, 1, q)
				assert.Equal(t, own.PhotoUID, results[0].PhotoUID, q)
			}
			for _, q := range []string{"zzqqxx", "path:" + base + "/other/", "private:true"} {
				results, _, err := SharedPhotos(form.SearchPhotos{Scope: a.AlbumUID, Query: q})
				require.NoError(t, err, q)
				if q != "private:true" {
					assert.Empty(t, results, q)
				}
			}
			results, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID, Query: "album:as6sg6bxpogaaba8 zebra"})
			require.NoError(t, err)
			require.Len(t, results, 1)
			assert.Equal(t, own.PhotoUID, results[0].PhotoUID)
			require.NoError(t, entity.Db().Model(own).Update("photo_private", true).Error)
			private, _, err := SharedPhotos(form.SearchPhotos{Scope: a.AlbumUID, Query: "private:true"})
			require.NoError(t, err)
			assert.Empty(t, private)
			private, _, err = UserPhotos(form.SearchPhotos{Scope: a.AlbumUID, Query: "private:true"}, scopeSession("guest"))
			require.NoError(t, err)
			assert.Empty(t, private)
			require.NoError(t, entity.Db().Model(own).Update("photo_private", false).Error)
			all, _, err := Photos(form.SearchPhotos{Scope: a.AlbumUID})
			require.NoError(t, err)
			require.Len(t, all, 1)
		})
	}
}
