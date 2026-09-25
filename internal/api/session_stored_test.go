package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// removeStoredSession removes a session row directly, leaving the session cache as it is.
func removeStoredSession(t *testing.T, authToken string) string {
	t.Helper()

	id := rnd.SessionID(authToken)
	require.NoError(t, entity.UnscopedDb().Exec("DELETE FROM auth_sessions WHERE id = ?", id).Error)

	return id
}

// storedSessionCount returns how many rows the sessions table holds for the given id.
func storedSessionCount(t *testing.T, id string) (n int) {
	t.Helper()
	require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("id = ?", id).Count(&n).Error)
	return n
}

// TestCreateSession_RemovedSession checks that a session whose row was removed is not reused.
func TestCreateSession_RemovedSession(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	authToken := AuthenticateUser(app, router, "alice", "Alice123!")
	require.NotEmpty(t, authToken)

	id := removeStoredSession(t, authToken)

	r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"token": "1jxf3jfn2k"}`, authToken)

	assert.Equal(t, 0, storedSessionCount(t, id), "the session row must stay deleted")
	assert.NotEqual(t, authToken, gjson.Get(r.Body.String(), "access_token").String())
}

// TestOAuthToken_RemovedSession checks that an app password cannot be created from a removed session.
func TestOAuthToken_RemovedSession(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	authToken := AuthenticateUser(app, router, "alice", "Alice123!")
	require.NotEmpty(t, authToken)

	OAuthToken(router)

	id := removeStoredSession(t, authToken)

	data := url.Values{
		"grant_type":  {authn.GrantPassword.String()},
		"client_name": {"AppPasswordRemoved"},
		"username":    {"alice"},
		"password":    {"Alice123!"},
		"scope":       {"*"},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add(header.ContentType, header.ContentTypeForm)
	req.Header.Add(header.XAuthToken, authToken)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, storedSessionCount(t, id), "the session row must stay deleted")
}
