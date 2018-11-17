package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/test"
)

func NewTest() (app *gin.Engine, router *gin.RouterGroup, conf photoprism.Config) {
	conf = test.NewConfig()
	gin.SetMode(gin.TestMode)
	app = gin.New()

	router = app.Group("/api/v1")

	return app, router, conf
}

func TestRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
