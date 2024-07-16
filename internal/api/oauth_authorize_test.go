package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthAuthorize(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthAuthorize(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
