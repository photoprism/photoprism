package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
)

type CloseableResponseRecorder struct {
	*httptest.ResponseRecorder
	closeCh chan bool
}

func (r *CloseableResponseRecorder) CloseNotify() <-chan bool {
	return r.closeCh
}

func (r *CloseableResponseRecorder) closeClient() {
	r.closeCh <- true
}

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	c := config.TestConfig()
	get.SetConfig(c)

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}

// NewApiTest returns new API test helper.
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	gin.SetMode(gin.TestMode)

	app = gin.New()
	router = app.Group("/api/v1")

	return app, router, get.Config()
}

// Executes an API request with an empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// Executes an API request with the request body as a string.
func PerformRequestWithBody(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}

// Executes an API request with a stream response.
func PerformRequestWithStream(r http.Handler, method, path string) *CloseableResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := &CloseableResponseRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	r.ServeHTTP(w, req)

	return w
}
