package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthUserinfo(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthUserinfo(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/userinfo")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
