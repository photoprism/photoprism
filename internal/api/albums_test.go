package api

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestGetAlbums(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, ctx, mock := NewApiTestMockDB(t)

		albs := []photoprism.AlbumSearchResult{
			{1, time.Now(), time.Now(), time.Time{},
				"b6a090d3-761c-4df5-a77c-8e97d5731fe9", "new-album",
				"New Album", 0, false, "", "",},
		}
		alb := albs[0]
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "album_uuid",
			"album_slug", "album_name", "album_count", "album_favorite", "album_description", "album_notes"}).
			AddRow(alb.ID, alb.CreatedAt, alb.UpdatedAt, alb.DeletedAt,
				alb.AlbumUUID, alb.AlbumSlug, alb.AlbumName, alb.AlbumCount,
				alb.AlbumFavorite, alb.AlbumDescription, alb.AlbumNotes,)

		mock.ExpectQuery("^SELECT albums(.+)").WillReturnRows(rows)

		GetAlbums(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/albums?count=10")

		assert.Equal(t, http.StatusOK, result.Code)
		AssertJSON(result.Body.Bytes(), albs, t)
		config.CleanTestConfig()
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetAlbums(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusBadRequest, result.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		t.Log(router)
		t.Log(ctx)
		result := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		t.Log(result.Body)

		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}

func TestLikeAlbum(t *testing.T) {
	t.Run("like not existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeAlbum(router, ctx)

		result := PerformRequest(app, "POST", "/api/v1/albums/98789876/like")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})

}

func TestDislikeAlbum(t *testing.T) {
	t.Run("dislike not existing album", func(t *testing.T) {
		app, router, ctx := NewApiTest()

		LikeAlbum(router, ctx)

		result := PerformRequest(app, "DELETE", "/api/v1/albums/5678/like")
		t.Log(result.Body)
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
