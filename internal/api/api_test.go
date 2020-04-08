package api

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

// NewApiTest returns new API test helper
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	conf = config.TestConfig()
	gin.SetMode(gin.TestMode)
	app = gin.New()
	router = app.Group("/api/v1")
	return app, router, conf
}

// Performs API request with empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Performs API request including request body as string.
func PerformRequestWithBody(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
