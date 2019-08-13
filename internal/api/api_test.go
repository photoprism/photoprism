package api

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

// API test helper
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	conf = config.TestConfig()
	gin.SetMode(gin.TestMode)
	app = gin.New()
	router = app.Group("/api/v1")
	return app, router, conf
}

// API test helper with mock database
func NewApiTestMockDB(t *testing.T) (app *gin.Engine, router *gin.RouterGroup, conf *config.Config, mock sqlmock.Sqlmock) {
	conf, mock = config.TestConfigMockDB(t)
	gin.SetMode(gin.TestMode)
	app = gin.New()
	router = app.Group("/api/v1")
	return app, router, conf, mock
}

// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func AssertJSON(actual []byte, data interface{}, t *testing.T) {
	expected, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when marshaling expected json data", err)
	}

	if bytes.Compare(expected, actual) != 0 {
		t.Errorf("the expected json: %s is different from actual %s", expected, actual)
	}
}
