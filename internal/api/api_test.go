package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/header"
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
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Init test config.
	c := config.TestConfig()
	get.SetConfig(c)
	defer c.CloseDb()

	// Increase login rate limit for testing.
	limiter.Login = limiter.NewLimit(1, 10000)

	// Run unit tests.
	code := m.Run()

	os.Exit(code)
}

// NewApiTest returns new API test helper.
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	gin.SetMode(gin.TestMode)

	app = gin.New()
	router = app.Group("/api/v1")

	return app, router, get.Config()
}

// PerformRequest runs an API request with an empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// PerformRequestWithBody runs an API request with the request body as a string.
func PerformRequestWithBody(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}

// PerformRequestWithStream runs an API request with a stream response.
func PerformRequestWithStream(r http.Handler, method, path string) *CloseableResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := &CloseableResponseRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	r.ServeHTTP(w, req)

	return w
}

// AuthenticateAdmin Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateAdmin(app *gin.Engine, router *gin.RouterGroup) (authToken string) {
	return AuthenticateUser(app, router, "admin", "photoprism")
}

// AuthenticateUser Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateUser(app *gin.Engine, router *gin.RouterGroup, username string, password string) (authToken string) {
	CreateSession(router)

	r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", form.AsJson(form.Login{
		Username: username,
		Password: password,
	}))

	authToken = gjson.Get(r.Body.String(), "access_token").String()

	return
}

// Performs authenticated API request with empty request body.
func AuthenticatedRequest(r http.Handler, method, path, authToken string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)

	header.SetAuthorization(req, authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// Performs an authenticated API request containing the request body as a string.
func AuthenticatedRequestWithBody(r http.Handler, method, path, body string, authToken string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)

	header.SetAuthorization(req, authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}
