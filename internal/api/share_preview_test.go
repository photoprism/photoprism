package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPreview(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SharePreview(router)
		r := PerformRequest(app, "GET", "/a/1jxf3jfn2k/st9lxuqxpogaaba7/preview")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("invalid token", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SharePreview(router)
		r := PerformRequest(app, "GET", "/a/xxx/st9lxuqxpogaaba7/preview")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
