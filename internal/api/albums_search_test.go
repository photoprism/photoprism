package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestSearchAlbums(t *testing.T) {
	entity.ValidateFixtures(t)
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?count=10&uid=as6sg6bxpogaaba7")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(1), count.Int())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("SearchByType", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?count=10&type=album")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.LessOrEqual(t, int64(1), gjson.Get(r.Body.String(), "#").Int())
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		SearchAlbums(router)
		r := PerformRequest(app, "GET", "/api/v1/albums?count=10")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AdminWithoutType", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		SearchAlbums(router)
		sessId := AuthenticateAdmin(t, app, router)
		// An admin has full access, so a type-less listing is permitted.
		r := AuthenticatedRequest(app, "GET", "/api/v1/albums?count=10", sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("GuestWithoutTypeOrUidDenied", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		SearchAlbums(router)
		sessId := AuthenticateUser(t, app, router, "gandalf", "Gandalf123!")
		// Without a type or UID filter, a non-admin role is denied (admin-only default resource).
		r := AuthenticatedRequest(app, "GET", "/api/v1/albums?count=10", sessId)
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, "Permission denied", gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("GuestUidLookupAllowed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		SearchAlbums(router)
		sessId := AuthenticateUser(t, app, router, "gandalf", "Gandalf123!")
		t.Cleanup(func() {
			require.NoError(t, entity.UnscopedDb().Save(entity.UserFixtures.Pointer("gandalf")).Error)
		})
		// A lookup by album UID is authorized via the albums resource (the share-link dialog
		// case), so it must not return "Permission denied".
		r := AuthenticatedRequest(app, "GET", "/api/v1/albums?count=1&uid=as6sg6bxpogaaba8", sessId)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.NotEqual(t, "Permission denied", gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("GuestSearchByTypeAllowed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		SearchAlbums(router)
		sessId := AuthenticateUser(t, app, router, "gandalf", "Gandalf123!")
		t.Cleanup(func() { require.NoError(t, entity.UnscopedDb().Save(entity.UserFixtures.Pointer("gandalf")).Error) })
		// A typed album search is authorized via the albums resource and scoped to shared albums.
		r := AuthenticatedRequest(app, "GET", "/api/v1/albums?count=10&type=album", sessId)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
