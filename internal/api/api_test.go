package api

import (
	"fmt"
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

// NewApiTest returns new API test helper.
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	gin.SetMode(gin.TestMode)
	app = gin.New()
	router = app.Group("/api/v1")
	return app, router, service.Config()
}

// NewApiTest returns new API test helper with authenticated admin session.
func NewAdminApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config, sessId string) {
	return NewAuthenticatedApiTest("admin", "photoprism")
}

// NewApiTest returns new API test helper with authenticated admin session.
func NewAuthenticatedApiTest(username string, password string) (app *gin.Engine, router *gin.RouterGroup, conf *config.Config, sessId string) {
	app = gin.New()
	router = app.Group("/api/v1")
	CreateSession(router)
	reader := strings.NewReader(fmt.Sprintf(`{"username": %s, "password": "%s"}`, username, password))
	req, _ := http.NewRequest("POST", "/api/v1/session", reader)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	sessId = w.Header().Get("X-Session-ID")
	gin.SetMode(gin.TestMode)
	return app, router, service.Config(), sessId
}

// Performs API request with empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Performs authenticated API request with empty request body.
func AuthenticatedRequest(r http.Handler, method, path, sess string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Add("X-Session-ID", sess)
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
