package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetShares(t *testing.T) {
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Shares(router)
		r := PerformRequest(app, "GET", "/api/v1/1jxf3jfn2k/st9lxuqxpogaaba7")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	//TODO Why does it panic?
	/*t.Run("valid token and share", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Shares(router)
		r := PerformRequest(app, "GET", "/api/v1/4jxf3jfn2k/at9lxuqxpogaaba7")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})*/
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Shares(router)
		r := PerformRequest(app, "GET", "/api/v1/xxx")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	//TODO Why does it panic?
	/*t.Run("valid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Shares(router)
		r := PerformRequest(app, "GET", "/api/v1/4jxf3jfn2k")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})*/
}
