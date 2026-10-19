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
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// approvedCode runs the authorization request and approves it as admin, and returns the authorization code.
func approvedCode(t *testing.T, app *gin.Engine, router *gin.RouterGroup, conf *config.Config, client *entity.Client) string {
	t.Helper()

	id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
	authToken := AuthenticateAdmin(app, router)

	r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())

	loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
	require.NoError(t, err)

	return loc.Query().Get("code")
}

// postTokenForm sends a form-encoded token request with optional HTTP Basic credentials.
func postTokenForm(app http.Handler, path string, values url.Values, basicID, basicSecret string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)

	if basicID != "" {
		req.SetBasicAuth(basicID, basicSecret)
	}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	return w
}

// codeTokenForm returns the form values of an authorization_code token request.
func codeTokenForm(clientID, code string) url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {testAppCallback},
		"code_verifier": {testAppVerifier},
	}
}

// prepareAppTokenTest prepares the app flow with the token, revoke and MCP endpoints.
func prepareAppTokenTest(t *testing.T) (*gin.Engine, *gin.RouterGroup, *config.Config, *entity.Client) {
	t.Helper()

	app, router, conf, client := prepareAppFlowTest(t, "mcp")
	OAuthToken(router)
	OAuthRevoke(router)
	ServeMCP(router)
	SearchPhotos(router)

	prevHandler := entity.GrantDeletedHandler
	t.Cleanup(func() {
		entity.DeleteClientGrants(client.ClientUID)
		entity.GrantDeletedHandler = prevHandler
	})

	return app, router, conf, client
}

func TestOAuthAppToken(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`

	t.Run("FullFlow", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, header.CacheControlNoStore, r.Header().Get(header.CacheControl))
		accessToken := gjson.Get(r.Body.String(), "access_token").String()
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()
		assert.Equal(t, "Bearer", gjson.Get(r.Body.String(), "token_type").String())
		assert.Equal(t, int64(3600), gjson.Get(r.Body.String(), "expires_in").Int())
		assert.Equal(t, "mcp", gjson.Get(r.Body.String(), "scope").String())
		require.NotEmpty(t, accessToken)
		require.NotEmpty(t, refreshToken)

		// The code is single-use.
		r = postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())

		// The access token works with the MCP endpoint only.
		w := mcpPost(app, initBody, accessToken, "")
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

		r = AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos?count=1", accessToken)
		assert.Equal(t, http.StatusUnauthorized, r.Code)

		// Refresh rotates the refresh token.
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		nextRefresh := gjson.Get(r.Body.String(), "refresh_token").String()
		nextAccess := gjson.Get(r.Body.String(), "access_token").String()
		assert.NotEqual(t, refreshToken, nextRefresh)
		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, nextAccess, "").Code)

		// Reusing the old refresh token revokes the grant and its access tokens.
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())
		assert.Equal(t, http.StatusUnauthorized, mcpPost(app, initBody, nextAccess, "").Code)

		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {nextRefresh}}, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("WrongVerifier", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		form := codeTokenForm(client.ClientUID, code)
		form.Set("code_verifier", testAppVerifier+"x")

		r := postTokenForm(app, "/api/v1/oauth/token", form, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("WrongResource", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		form := codeTokenForm(client.ClientUID, code)
		form.Set("resource", "https://other.example.com/api/v1/mcp")

		r := postTokenForm(app, "/api/v1/oauth/token", form, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidTarget, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("ClientSecret", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm("", code), client.ClientUID, "secret")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidClient, gjson.Get(r.Body.String(), "error").String())

		form := codeTokenForm(client.ClientUID, code)
		form.Set("client_secret", "secret")
		r = postTokenForm(app, "/api/v1/oauth/token", form, "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("BasicClientID", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm("", code), client.ClientUID, "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
	})
	t.Run("ClientDisabled", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		client.AuthEnabled = false
		require.NoError(t, client.Save())

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AppClientCannotUseClientCredentials", func(t *testing.T) {
		app, _, _, client := prepareAppTokenTest(t)
		r := postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"client_credentials"}, "client_id": {client.ClientUID}, "client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u"}}, "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AppClientWithSecret", func(t *testing.T) {
		// The auth method alone refuses the grant, also if a secret is stored for the client.
		app, _, _, client := prepareAppTokenTest(t)
		secret := "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u"
		pw := entity.NewPassword(client.ClientUID, secret, false)
		require.NoError(t, pw.Save())
		t.Cleanup(func() { _ = pw.Delete() })
		r := postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"client_credentials"}, "client_id": {client.ClientUID}, "client_secret": {secret}}, "", "")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("AppsDisabled", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		conf.Options().OAuthApps = config.OAuthAppsNone

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, "unsupported_grant_type", gjson.Get(r.Body.String(), "error").String())
	})
}

func TestOAuthAppRefresh(t *testing.T) {
	t.Run("OtherClient", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code)
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		other, err := entity.AddClient(entityAppForm("Other App"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = other.Delete() })

		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {other.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)

		// The grant was not revoked.
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("UserCannotSignIn", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code)
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		superAdmin := admin.SuperAdmin
		require.NoError(t, admin.Updates(entity.Values{"CanLogin": false, "SuperAdmin": false}))
		t.Cleanup(func() { _ = admin.Updates(entity.Values{"CanLogin": true, "SuperAdmin": superAdmin}) })

		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("BasicCredentials", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code)
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}, client.ClientUID, "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
	})
}

func TestOAuthRevoke_AppTokens(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`

	t.Run("RefreshToken", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code)
		accessToken := gjson.Get(r.Body.String(), "access_token").String()
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		r = postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {refreshToken}, "token_type_hint": {"refresh_token"}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, http.StatusUnauthorized, mcpPost(app, initBody, accessToken, "").Code)

		// Revoking an unknown refresh token also succeeds.
		r = postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("AccessToken", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code)
		accessToken := gjson.Get(r.Body.String(), "access_token").String()
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		r = postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {accessToken}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, http.StatusUnauthorized, mcpPost(app, initBody, accessToken, "").Code)

		// The grant remains, so the refresh token still works.
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestMCP_AppTokenAudience(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	code := approvedCode(t, app, router, conf, client)
	r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
	require.Equal(t, http.StatusOK, r.Code)
	accessToken := gjson.Get(r.Body.String(), "access_token").String()

	// A token issued for another resource is refused, e.g. after the site URL changed.
	sess, err := entity.FindAnySession(entityRnd(accessToken))
	require.NoError(t, err)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("id = ?", sess.ID).UpdateColumn("auth_audience", "https://other.example.com/api/v1/mcp").Error)
	entity.FlushSessionCache()

	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
	w := mcpPost(app, initBody, accessToken, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Header().Get(header.WWWAuthenticate), "invalid_token")
	assert.Equal(t, authn.ProviderOAuthApp.String(), sess.AuthProvider)
}

// entityAppForm returns the form of an app client with the test callback.
func entityAppForm(name string) form.Client {
	return form.Client{ClientName: name, AuthMethod: authn.MethodPKCE.String(), AuthScope: "mcp", RedirectURIs: []string{testAppCallback}}
}

// entityRnd returns the session ID of an auth token.
func entityRnd(token string) string {
	return rnd.SessionID(token)
}

// tokenPair redeems an approved code and returns the access and refresh tokens.
func tokenPair(t *testing.T, app *gin.Engine, router *gin.RouterGroup, conf *config.Config, client *entity.Client) (string, string) {
	t.Helper()

	code := approvedCode(t, app, router, conf, client)
	r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())

	return gjson.Get(r.Body.String(), "access_token").String(), gjson.Get(r.Body.String(), "refresh_token").String()
}

func TestOAuthAppToken_Rechecks(t *testing.T) {
	t.Run("CallbackRemoved", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)

		data := client.GetData()
		data.RedirectURIs = []string{"http://127.0.0.1/callback"}
		require.NoError(t, client.SetData(data).Save())

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("UserCannotSignIn", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		superAdmin := admin.SuperAdmin
		require.NoError(t, admin.Updates(entity.Values{"CanLogin": false, "SuperAdmin": false}))
		t.Cleanup(func() { _ = admin.Updates(entity.Values{"CanLogin": true, "SuperAdmin": superAdmin}) })

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("AccessTokenOfUserWhoCannotSignIn", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, _ := tokenPair(t, app, router, conf, client)
		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		require.Equal(t, http.StatusOK, mcpPost(app, initBody, accessToken, "").Code)

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		superAdmin := admin.SuperAdmin
		require.NoError(t, admin.Updates(entity.Values{"CanLogin": false, "SuperAdmin": false}))
		t.Cleanup(func() { _ = admin.Updates(entity.Values{"CanLogin": true, "SuperAdmin": superAdmin}) })
		entity.FlushSessionCache()

		assert.NotEqual(t, http.StatusOK, mcpPost(app, initBody, accessToken, "").Code)
	})
	t.Run("ScopeNotWidened", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		require.NoError(t, client.SetScope("mcp photos").Save())
		accessToken, _ := tokenPair(t, app, router, conf, client)
		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`

		for i := 0; i < 2; i++ {
			assert.Equal(t, http.StatusOK, mcpPost(app, initBody, accessToken, "").Code)
			assert.Equal(t, http.StatusUnauthorized, AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos?count=1", accessToken).Code)
		}

		sess, err := entity.FindAnySession(rnd.SessionID(accessToken))
		require.NoError(t, err)
		assert.Equal(t, "mcp", sess.AuthScope)
		assert.Empty(t, sess.ClientUID)
	})
	t.Run("RefreshReplacesAccessToken", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, refreshToken := tokenPair(t, app, router, conf, client)

		r := postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		require.Equal(t, http.StatusOK, r.Code)

		// The previous access token remains valid for requests in flight.
		_, err := entity.FindAnySession(rnd.SessionID(accessToken))
		assert.NoError(t, err)

		// Tokens issued in the same second have no defined order.
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("id = ?", rnd.SessionID(accessToken)).UpdateColumn("created_at", time.Now().Add(-time.Hour)).Error)

		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {gjson.Get(r.Body.String(), "refresh_token").String()}}, "", "")
		require.Equal(t, http.StatusOK, r.Code)

		_, err = entity.FindAnySession(rnd.SessionID(accessToken))
		assert.Error(t, err)

		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("auth_provider = ? AND user_uid = ? AND auth_id <> ''", "oauth_app", entity.Admin.UserUID).Count(&count).Error)
		assert.LessOrEqual(t, count, int64(2))
	})
	t.Run("RefreshNarrowsScope", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		require.NoError(t, client.SetScope("mcp photos").Save())
		id := authorizeRequestID(t, app, func() url.Values {
			q := testAuthorizeQuery(conf, client.ClientUID)
			q.Set("scope", "mcp photos")
			return q
		}())
		authToken := AuthenticateAdmin(app, router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","photos"]}`, authToken)
		require.Equal(t, http.StatusOK, r.Code)
		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)
		r = postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, loc.Query().Get("code")), "", "")
		require.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "mcp photos", gjson.Get(r.Body.String(), "scope").String())
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		require.NoError(t, client.SetScope("mcp").Save())
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "mcp", gjson.Get(r.Body.String(), "scope").String())
	})
	t.Run("RefreshRefusedUserKeepsGrant", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		_, refreshToken := tokenPair(t, app, router, conf, client)

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		superAdmin := admin.SuperAdmin
		require.NoError(t, admin.Updates(entity.Values{"CanLogin": false, "SuperAdmin": false}))

		r := postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)

		require.NoError(t, admin.Updates(entity.Values{"CanLogin": true, "SuperAdmin": superAdmin}))

		// The refresh token was not consumed.
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
	})
	t.Run("AppsDisabledChargesLimit", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		conf.Options().OAuthApps = config.OAuthAppsNone
		limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)

		limited := false

		for i := 0; i < 10 && !limited; i++ {
			limited = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "refresh_token": {"abc"}}, "", "").Code == http.StatusTooManyRequests
		}

		assert.True(t, limited)
	})
}

func TestOAuthAppTokenResponse(t *testing.T) {
	t.Run("GrantRevoked", func(t *testing.T) {
		_, _, _, client := prepareAppTokenTest(t)
		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)

		grant, _, err := oauth.NewGrant(&entity.AuthCode{UserUID: admin.UserUID, ClientID: client.ClientUID, Scope: "mcp", Resource: "https://photos.example.com/api/v1/mcp"}, admin, "Test App")
		require.NoError(t, err)
		require.NoError(t, grant.Delete())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
		oauthAppTokenResponse(c, grant, admin, "refresh", authn.GrantRefreshToken, "client test", "refresh token")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(w.Body.String(), "error").String())

		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("auth_id = ?", grant.GrantUID).Count(&count).Error)
		assert.Zero(t, count)
	})
}

func TestOAuthAppTokenClient(t *testing.T) {
	_, _, _, client := prepareAppTokenTest(t)

	request := func(form url.Values, basicID, basicSecret string) (*oauth.Client, bool, int) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		c.Request.Header.Set(header.ContentType, header.ContentTypeForm)

		if basicID != "" {
			c.Request.SetBasicAuth(basicID, basicSecret)
		}

		result, ok := oauthAppTokenClient(c, "test", false)

		return result, ok, w.Code
	}

	t.Run("Form", func(t *testing.T) {
		result, ok, _ := request(url.Values{"client_id": {client.ClientUID}}, "", "")
		require.True(t, ok)
		assert.Equal(t, client.ClientUID, result.ID)
	})
	t.Run("Basic", func(t *testing.T) {
		result, ok, _ := request(url.Values{}, client.ClientUID, "")
		require.True(t, ok)
		assert.Equal(t, client.ClientUID, result.ID)
	})
	t.Run("Mismatch", func(t *testing.T) {
		_, ok, code := request(url.Values{"client_id": {"cs5gfen1bgxz7s9i"}}, client.ClientUID, "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, code)
	})
	t.Run("Secret", func(t *testing.T) {
		_, ok, code := request(url.Values{"client_id": {client.ClientUID}, "client_secret": {"x"}}, "", "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, code)
	})
	t.Run("NotAnApp", func(t *testing.T) {
		_, ok, code := request(url.Values{"client_id": {"cs5gfen1bgxz7s9i"}}, "", "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, code)
	})
}

func TestOAuthAppTokenUser(t *testing.T) {
	newContext := func() (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
		return c, w
	}

	t.Run("Active", func(t *testing.T) {
		c, _ := newContext()
		admin := entity.FindUserByName("admin")
		user, ok := oauthAppTokenUser(c, admin.UserUID, "client test", "test")
		require.True(t, ok)
		assert.Equal(t, admin.UserUID, user.UserUID)
	})
	t.Run("Unknown", func(t *testing.T) {
		c, w := newContext()
		_, ok := oauthAppTokenUser(c, "uqxetse3cy5eo9z9", "client test", "test")
		assert.False(t, ok)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOAuthTokenError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))
	assert.Equal(t, `{"error":"invalid_grant","error_description":"invalid"}`, w.Body.String())
}

func TestAuthAudience(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	accessToken, _ := tokenPair(t, app, router, conf, client)

	newContext := func(token string) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/mcp", nil)
		header.SetAuthorization(c.Request, token)
		return c
	}

	t.Run("MatchingAudience", func(t *testing.T) {
		s := AuthAudience(newContext(accessToken), conf.OAuthResourceUrl(), "mcp", "view")
		assert.Equal(t, http.StatusOK, s.HttpStatus())
	})
	t.Run("NoAudience", func(t *testing.T) {
		s := Auth(newContext(accessToken), "mcp", "view")
		assert.Equal(t, http.StatusUnauthorized, s.HttpStatus())
	})
	t.Run("OtherAudience", func(t *testing.T) {
		s := AuthAudience(newContext(accessToken), "https://other.example.com/api/v1/mcp", "mcp", "view")
		assert.Equal(t, http.StatusUnauthorized, s.HttpStatus())
	})
}

func TestLookupSession_Audience(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	accessToken, _ := tokenPair(t, app, router, conf, client)

	t.Run("Refused", func(t *testing.T) {
		s, err := lookupSession("192.0.2.10", accessToken, "")
		assert.Nil(t, s)
		assert.Error(t, err)
	})
	t.Run("Accepted", func(t *testing.T) {
		s, err := lookupSession("192.0.2.10", accessToken, conf.OAuthResourceUrl())
		require.NoError(t, err)
		assert.True(t, s.IsOAuthApp())
	})
}

func TestOAuthAppToken_RateLimit(t *testing.T) {
	t.Run("InvalidCodesKeepSessions", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, _ := tokenPair(t, app, router, conf, client)
		limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)

		for i := 0; i < 10; i++ {
			r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, rnd.AuthToken()), "", "")
			require.Equal(t, http.StatusBadRequest, r.Code)
			r = postTokenForm(app, "/api/v1/oauth/token", codeTokenForm("csunknown0000001", rnd.AuthToken()), "", "")
			require.Equal(t, http.StatusUnauthorized, r.Code)
		}

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, accessToken, "").Code)
	})
}

func TestOAuthRevoke_AppTokenHeader(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	accessToken, _ := tokenPair(t, app, router, conf, client)

	alice := entity.FindUserByName("admin")
	require.NotNil(t, alice)
	other, err := entity.AddClientSession("revoke-app-header", 3600, "*", authn.GrantCLI, alice)
	require.NoError(t, err)
	t.Cleanup(func() { _ = other.Delete() })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/revoke", strings.NewReader(url.Values{"token": {other.RefID}, "token_type_hint": {"ref_id"}}.Encode()))
	req.Header.Set(header.ContentType, header.ContentTypeForm)
	header.SetAuthorization(req, accessToken)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	_, err = entity.FindAnySession(other.ID)
	assert.NoError(t, err)
}

func TestOAuthAppRefresh_ClientScope(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	_, refreshToken := tokenPair(t, app, router, conf, client)

	require.NoError(t, client.SetScope("photos").Save())
	r := postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
	assert.Equal(t, http.StatusBadRequest, r.Code)

	// The refresh token was not consumed.
	require.NoError(t, client.SetScope("mcp").Save())
	r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
	assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
}

// refreshForm returns the form values of a refresh_token request.
func refreshForm(clientID, refreshToken string) url.Values {
	return url.Values{"grant_type": {"refresh_token"}, "client_id": {clientID}, "refresh_token": {refreshToken}}
}

// signIn returns an auth token for the user from the session route, which must already be registered.
func signIn(app *gin.Engine, username, password string) string {
	r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", form.AsJson(form.Login{Username: username, Password: password}))
	return gjson.Get(r.Body.String(), "access_token").String()
}
