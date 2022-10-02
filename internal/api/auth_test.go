package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/session"
)

// AuthenticateAdmin Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateAdmin(app *gin.Engine, router *gin.RouterGroup) (sessId string) {
	return AuthenticateUser(app, router, "admin", "photoprism")
}

// AuthenticateUser Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateUser(app *gin.Engine, router *gin.RouterGroup, name string, password string) (sessId string) {
	CreateSession(router)

	f := form.Login{
		UserName: name,
		Password: password,
	}

	loginStr, err := json.Marshal(f)

	if err != nil {
		log.Fatal(err)
	}

	r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", string(loginStr))
	sessId = r.Header().Get(session.Header)

	return
}

// Performs authenticated API request with empty request body.
func AuthenticatedRequest(r http.Handler, method, path, sess string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)

	if sess != "" {
		req.Header.Add(session.Header, sess)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// Performs an authenticated API request containing the request body as a string.
func AuthenticatedRequestWithBody(r http.Handler, method, path, body string, sess string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)

	if sess != "" {
		req.Header.Add(session.Header, sess)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}
