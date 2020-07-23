package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/sirupsen/logrus"
)

// NewApiTest returns new API test helper
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	gin.SetMode(gin.TestMode)
	app = gin.New()
	router = app.Group("/api/v1")
	return app, router, service.Config()
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

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	c := config.TestConfig()
	service.SetConfig(c)

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}
