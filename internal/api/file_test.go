package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	t.Run("search for existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/files/123xxx")
		assert.Equal(t, http.StatusOK, result.Code)
		assert.Contains(t, result.Body.String(), "\"FileName\":\"exampleFileName.jpg\"")
	})
	t.Run("search for not existing file", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		GetFile(router, ctx)
		result := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
