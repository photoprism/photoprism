package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestOAuthToken_AuthorizationCode covers the authorization_code grant branch
// that OAuthToken delegates to OAuthAuthorizationCodeHandler (the Portal OIDC OP
// token handler in production). The branch runs before the CE form binding, so
// it must also work with client_secret_basic credentials.
func TestOAuthToken_AuthorizationCode(t *testing.T) {
	t.Run("UnsupportedWhenNoHook", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		prev := OAuthAuthorizationCodeHandler
		OAuthAuthorizationCodeHandler = nil
		defer func() { OAuthAuthorizationCodeHandler = prev }()

		OAuthToken(router)

		data := url.Values{
			"grant_type":    {authn.GrantAuthorizationCode.String()},
			"code":          {"raw-code"},
			"redirect_uri":  {"https://photos.example.com/cb"},
			"code_verifier": {"verifier"},
			"client_id":     {"cs5cpu17n6gj2qo5"},
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "unsupported_grant_type")
	})
	t.Run("DelegatesToHook", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		prev := OAuthAuthorizationCodeHandler
		OAuthAuthorizationCodeHandler = func(c *gin.Context) {
			c.String(http.StatusOK, "delegated:"+c.PostForm("grant_type"))
		}
		defer func() { OAuthAuthorizationCodeHandler = prev }()

		OAuthToken(router)

		data := url.Values{
			"grant_type": {authn.GrantAuthorizationCode.String()},
			"code":       {"raw-code"},
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "delegated:authorization_code", w.Body.String())
	})
	t.Run("DelegatesWithBasicAuthCredentials", func(t *testing.T) {
		// client_secret_basic puts the credentials in the Authorization header
		// while grant_type stays in the body. The delegation must trigger off the
		// body grant_type (not be hijacked into the client_credentials path) and
		// leave the Basic credentials readable by the delegated handler.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		var gotID, gotSecret string
		var gotBasic bool
		prev := OAuthAuthorizationCodeHandler
		OAuthAuthorizationCodeHandler = func(c *gin.Context) {
			gotID, gotSecret, gotBasic = c.Request.BasicAuth()
			c.String(http.StatusOK, "ok")
		}
		defer func() { OAuthAuthorizationCodeHandler = prev }()

		OAuthToken(router)

		data := url.Values{
			"grant_type":    {authn.GrantAuthorizationCode.String()},
			"code":          {"raw-code"},
			"code_verifier": {"verifier"},
			"redirect_uri":  {"https://photos.example.com/cb"},
		}
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		req.SetBasicAuth("cs5cpu17n6gj2qo5", "client-secret")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, gotBasic, "BasicAuth must stay readable by the delegated handler")
		assert.Equal(t, "cs5cpu17n6gj2qo5", gotID)
		assert.Equal(t, "client-secret", gotSecret)
	})
}

func TestOAuthToken(t *testing.T) {
	t.Run("ClientSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("InvalidClientID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"123"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("NotExistingClient", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"
		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2yy6"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("InvalidClientSecret", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0f"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("ClientAuthNotEnabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5gfsvbd7ejzn8m"},
			"client_secret": {"aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("ClientUnknownAuthMethod", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs7pvt5h8rw9he34"},
			"client_secret": {"1831986451da7acf34690b703ff528f67bcf255e005270e9"},
			"scope":         {"*"},
		}
		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UserNoSession", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordAlice"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UserSuccess", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordAlice"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("AppPasswordsDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		conf.Settings().Features.AppPasswords = false
		defer func() { conf.Settings().Features.AppPasswords = true }()

		OAuthToken(router)

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordAlice"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest("POST", "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("UnregisteredUser", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"Visitor"},
			"username":    {"visitor"},
			"password":    {"69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("UsersDontMatch", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordBob"},
			"username":    {"bob"},
			"password":    {"Bobbob123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("DeletedUser", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "deleted", "Deleted123!")

		OAuthToken(router)

		var method = "POST"
		var path = "/api/v1/oauth/token"

		data := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"AppPasswordDeleted"},
			"username":    {"deleted"},
			"password":    {"Deleted123!"},
			"scope":       {"*"},
		}

		req, _ := http.NewRequest(method, path, strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)
		req.Header.Add(header.XAuthToken, sessId)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		t.Logf("Header: %s", w.Header())
		t.Logf("BODY: %s", w.Body.String())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestOAuthToken_DeletedClient covers the client lifecycle gate on the token endpoint.
// The same credentials are presented before and after the client is deleted, so the
// refusal is attributable to the deletion rather than to the request or the secret.
func TestOAuthToken_DeletedClient(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	OAuthToken(router)

	client := entity.NewClient().SetName("Token Lifecycle").SetRole(acl.RoleClient.String())
	client.AuthScope = "metrics"

	if err := client.Create(); err != nil {
		t.Fatal(err)
	}

	secret, err := client.NewSecret()

	if err != nil {
		t.Fatal(err)
	}

	tokenRequest := func() *httptest.ResponseRecorder {
		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {client.ClientUID},
			"client_secret": {secret},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		return w
	}

	var sessId string

	t.Run("LiveClient", func(t *testing.T) {
		w := tokenRequest()

		if !assert.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String()) {
			return
		}

		sessId = gjson.Get(w.Body.String(), "session_id").String()
		assert.NotEmpty(t, sessId)
	})
	t.Run("DeletedClient", func(t *testing.T) {
		if err = client.Delete(); err != nil {
			t.Fatal(err)
		}

		if deleted := entity.FindClientByUID(client.ClientUID); assert.NotNil(t, deleted) {
			assert.True(t, deleted.Deleted())
		}

		w := tokenRequest()
		assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
	})
	t.Run("SessionRevoked", func(t *testing.T) {
		// Without a session id from the live request there is nothing to prove here.
		if !assert.NotEmpty(t, sessId) {
			return
		}

		_, findErr := entity.FindSession(sessId)
		assert.Error(t, findErr)
	})
}

// TestOAuthToken_InactiveUser covers clients that belong to a user account: the account must still be
// active for the client credentials grant to issue a token.
func TestOAuthToken_InactiveUser(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	OAuthToken(router)

	// newUserClient creates a user and a client that belongs to it, and returns both with the secret.
	newUserClient := func(t *testing.T) (*entity.User, *entity.Client, string) {
		t.Helper()

		user := entity.NewUser()
		user.UserName = "client-owner-" + rnd.Base36(6)
		user.UserRole = acl.RoleAdmin.String()
		user.CanLogin = true
		require.NoError(t, user.Create())

		client := entity.NewClient().SetName("Owned Client").SetRole(acl.RoleClient.String()).SetUser(user)
		client.AuthScope = "metrics"
		require.NoError(t, client.Create())

		secret, err := client.NewSecret()
		require.NoError(t, err)

		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.Session{}, "client_uid = ?", client.ClientUID)
			entity.UnscopedDb().Unscoped().Delete(client)
			entity.UnscopedDb().Unscoped().Delete(&entity.Password{}, "uid = ?", client.ClientUID)
			entity.UnscopedDb().Unscoped().Delete(&entity.UserDetails{}, "user_uid = ?", user.UserUID)
			entity.UnscopedDb().Unscoped().Delete(&entity.UserSettings{}, "user_uid = ?", user.UserUID)
			entity.UnscopedDb().Unscoped().Delete(user)
		})

		return user, client, secret
	}

	tokenRequest := func(client *entity.Client, secret string) *httptest.ResponseRecorder {
		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {client.ClientUID},
			"client_secret": {secret},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
		req.Header.Add(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		return w
	}

	t.Run("ActiveUser", func(t *testing.T) {
		_, client, secret := newUserClient(t)

		w := tokenRequest(client, secret)
		assert.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	})
	t.Run("DeletedUser", func(t *testing.T) {
		user, client, secret := newUserClient(t)
		require.NoError(t, user.Delete())

		w := tokenRequest(client, secret)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
	})
	t.Run("MissingUser", func(t *testing.T) {
		user, client, secret := newUserClient(t)
		require.NoError(t, entity.UnscopedDb().Unscoped().Delete(user).Error)

		w := tokenRequest(client, secret)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
	})
	t.Run("ExpiredUser", func(t *testing.T) {
		user, client, secret := newUserClient(t)
		expired := time.Now().Add(-time.Hour)
		require.NoError(t, entity.Db().Model(user).UpdateColumn("expires_at", &expired).Error)

		w := tokenRequest(client, secret)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "body=%s", w.Body.String())
	})
	t.Run("ExpiredSuperAdmin", func(t *testing.T) {
		user, client, secret := newUserClient(t)
		expired := time.Now().Add(-time.Hour)
		require.NoError(t, entity.Db().Model(user).UpdateColumns(entity.Values{"expires_at": &expired, "super_admin": true}).Error)

		w := tokenRequest(client, secret)
		assert.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	})
}
