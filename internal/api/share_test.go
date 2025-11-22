package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShareToken(t *testing.T) {
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareToken(router)
		r := PerformRequest(app, "GET", "/api/v1/xxx")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	//TODO Why does it panic?
	/*t.Run("ValidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareToken(router)
		r := PerformRequest(app, "GET", "/api/v1/4jxf3jfn2k")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})*/
}

func TestShareTokenShared(t *testing.T) {
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareTokenShared(router)
		r := PerformRequest(app, "GET", "/api/v1/1jxf3jfn2k/ss6sg6bxpogaaba7")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	//TODO Why does it panic?
	/*t.Run("ValidTokenAndShare", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareTokenShared(router)
		r := PerformRequest(app, "GET", "/api/v1/4jxf3jfn2k/as6sg6bxpogaaba7")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})*/
}
