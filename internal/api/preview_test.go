package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPreview(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetPreview(router, ctx)
		r := PerformRequest(app, "GET", "/api/v1/preview")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
