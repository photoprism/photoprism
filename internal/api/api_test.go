package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/context"
)

// API test helper
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, ctx *context.Context) {
	ctx = context.TestContext()
	gin.SetMode(gin.TestMode)
	app = gin.New()

	router = app.Group("/api/v1")

	return app, router, ctx
}

// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
