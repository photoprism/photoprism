package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// TestSearch_NoMatch checks that list endpoints answer a filter value whose removed NUL bytes and invalid
// UTF-8 leave nothing usable with an empty array, and the same request without it with results.
func TestSearch_NoMatch(t *testing.T) {
	values := []string{"%00", "%FF%FE", "%00%20", "%FF%7C", "%FF*"}

	t.Run("AlbumsLabels", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchAlbums(router)
		SearchLabels(router)

		for _, uri := range []string{"/api/v1/albums?count=10", "/api/v1/labels?count=10"} {
			r := PerformRequest(app, "GET", uri)
			require.Equal(t, http.StatusOK, r.Code, uri)
			require.NotEqual(t, "[]", r.Body.String(), uri)

			for _, v := range values {
				r = PerformRequest(app, "GET", uri+"&q="+v)
				assert.Equal(t, http.StatusOK, r.Code, uri+v)
				assert.Equal(t, "[]", r.Body.String(), uri+v)
			}
		}
	})
	t.Run("Services", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchServices(router)
		sess := AuthenticateAdmin(app, router)

		r := AuthenticatedRequest(app, "GET", "/api/v1/services?count=10", sess)
		require.Equal(t, http.StatusOK, r.Code)
		require.NotEqual(t, "[]", r.Body.String())

		for _, v := range values {
			r = AuthenticatedRequest(app, "GET", "/api/v1/services?count=10&q="+v, sess)
			assert.Equal(t, http.StatusOK, r.Code, v)
			assert.Equal(t, "[]", r.Body.String(), v)
		}
	})
	t.Run("UserSessions", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		FindUserSessions(router)
		sess := AuthenticateUser(app, router, "alice", "Alice123!")
		uri := "/api/v1/users/uqxetse3cy5eo9z2/sessions?count=100"

		r := AuthenticatedRequest(app, "GET", uri, sess)
		require.Equal(t, http.StatusOK, r.Code)
		require.NotEqual(t, "[]", r.Body.String())

		for _, v := range values {
			r = AuthenticatedRequest(app, "GET", uri+"&q="+v, sess)
			assert.Equal(t, http.StatusOK, r.Code, v)
			assert.Equal(t, "[]", r.Body.String(), v)
		}
	})
}
