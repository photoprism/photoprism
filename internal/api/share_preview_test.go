package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestGetPreview(t *testing.T) {
	// preview registers GET /:token/:shared/preview on the /api/v1 group, so the request path
	// must carry exactly those three segments for the handler to run at all.
	preview := func(t *testing.T, path string) int {
		t.Helper()
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		SharePreview(router)

		return PerformRequest(app, "GET", path).Code
	}

	t.Run("UnknownShare", func(t *testing.T) {
		assert.Equal(t, http.StatusTemporaryRedirect, preview(t, "/api/v1/1jxf3jfn2k/ss6sg6bxpogaaba7/preview"))
	})
	t.Run("InvalidToken", func(t *testing.T) {
		assert.Equal(t, http.StatusTemporaryRedirect, preview(t, "/api/v1/xxx/as6sg6bxpogaaba8/preview"))
	})
	t.Run("RejectsOversizedToken", func(t *testing.T) {
		assert.Equal(t, http.StatusTemporaryRedirect, preview(t, "/api/v1/"+strings.Repeat("a", 161)+"/as6sg6bxpogaaba8/preview"))
	})
	t.Run("RejectsUnusableToken", func(t *testing.T) {
		assert.Equal(t, http.StatusTemporaryRedirect, preview(t, "/api/v1/..../as6sg6bxpogaaba8/preview"))
	})
	t.Run("RouteMismatch", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, preview(t, "/api/v1/s/1jxf3jfn2k/as6sg6bxpogaaba8/preview"))
	})
}
