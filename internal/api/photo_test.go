package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetPhotos(t *testing.T) {
	conf := test.NewConfig()
	app := gin.Default()

	v1 := app.Group("/api/v1")
	{
		GetPhotos(v1, conf)
	}

	r := performRequest(app, "GET", "/api/v1/photos?count=10")

	assert.Equal(t, http.StatusOK, r.Code)
}
