package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/auth/tokens"
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

// appTokenTestClientIP returns the client IP that requests sent with postTokenForm come from.
func appTokenTestClientIP() string {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
	return ClientIP(c)
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

	prevDeleted, prevSuspended := entity.GrantDeletedHandler, entity.GrantSuspendedHandler
	t.Cleanup(func() {
		entity.DeleteClientGrants(client.ClientUID)
		entity.GrantDeletedHandler, entity.GrantSuspendedHandler = prevDeleted, prevSuspended
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
	t.Run("MismatchedHint", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, refreshToken := tokenPair(t, app, router, conf, client)
		installAuthorizeBudget(t, 3)
		installAuthBudget(t, 3)
		ip := appTokenTestClientIP()

		// A hint that does not match the token is ignored, and the token is revoked as the type it has.
		r := postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {accessToken}, "token_type_hint": {"refresh_token"}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, http.StatusUnauthorized, mcpPost(app, initBody, accessToken, "").Code)

		grantUID, _, _ := strings.Cut(refreshToken, ".")
		require.NotNil(t, entity.FindGrant(grantUID))
		r = postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {refreshToken}, "token_type_hint": {"access_token"}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Nil(t, entity.FindGrant(grantUID))

		assert.InDelta(t, 3, limiter.Auth.IP(ip).Tokens(), 0.01)
		assert.InDelta(t, 3, limiter.Authorize.IP(ip).Tokens(), 0.01)
	})
	t.Run("Charges", func(t *testing.T) {
		app, _, _, _ := prepareAppTokenTest(t)
		ip := appTokenTestClientIP()

		// revoke sends a revoke request and returns the charges to limiter.Auth and limiter.Authorize.
		revoke := func(t *testing.T, values url.Values) (int, float64, float64) {
			installAuthorizeBudget(t, 3)
			installAuthBudget(t, 3)
			r := postTokenForm(app, "/api/v1/oauth/revoke", values, "", "")
			return r.Code, 3 - limiter.Auth.IP(ip).Tokens(), 3 - limiter.Authorize.IP(ip).Tokens()
		}

		// An unknown or malformed access token counts towards limiter.Auth.
		code, auth, authorize := revoke(t, url.Values{"token": {rnd.AuthToken()}, "token_type_hint": {"access_token"}})
		assert.Equal(t, http.StatusUnauthorized, code)
		assert.InDelta(t, 1, auth, 0.01)
		assert.InDelta(t, 0, authorize, 0.01)
		code, auth, authorize = revoke(t, url.Values{"token": {"abc"}})
		assert.Equal(t, http.StatusUnauthorized, code)
		assert.InDelta(t, 1, auth, 0.01)
		assert.InDelta(t, 0, authorize, 0.01)

		// A refresh token that fails verification is answered without a lookup, also with another hint.
		forged := rnd.RefreshToken(rnd.GenerateUID(entity.GrantUID))
		code, auth, authorize = revoke(t, url.Values{"token": {forged}, "token_type_hint": {"access_token"}})
		assert.Equal(t, http.StatusOK, code)
		assert.InDelta(t, 0, auth, 0.01)
		assert.InDelta(t, 0, authorize, 0.01)
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
		limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), 3)

		codes := make([]int, 0, 6)

		for i := 0; i < 6; i++ {
			codes = append(codes, postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "refresh_token": {"abc"}}, "", "").Code)
		}

		assert.Equal(t, []int{400, 400, 400, 429, 429, 429}, codes)
		assert.True(t, limiter.Authorize.Reject(appTokenTestClientIP()))
		assert.False(t, limiter.Auth.Reject(appTokenTestClientIP()))
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
	t.Run("GrantReadError", func(t *testing.T) {
		_, _, _, client := prepareAppTokenTest(t)
		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)

		grant, _, err := oauth.NewGrant(&entity.AuthCode{UserUID: admin.UserUID, ClientID: client.ClientUID, Scope: "mcp", Resource: "https://photos.example.com/api/v1/mcp"}, admin, "Test App")
		require.NoError(t, err)
		name := "test:fail_grant_reads_" + rnd.Base36(8)
		require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (entity.Grant{}).TableName() {
				_ = db.AddError(errors.New("grant lookup failed"))
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
		oauthAppTokenResponse(c, grant, admin, "refresh", authn.GrantRefreshToken, "client test", "refresh token")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, oauth.ErrCodeServerError, gjson.Get(w.Body.String(), "error").String())

		// The access token stored before the check is deleted.
		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("auth_id = ?", grant.GrantUID).Count(&count).Error)
		assert.Zero(t, count)
	})
}

func TestOAuthAppTokenClient(t *testing.T) {
	_, _, _, client := prepareAppTokenTest(t)

	// clientIp is the address the requests below come from.
	ipCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ipCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", nil)
	clientIp := ClientIP(ipCtx)

	request := func(form url.Values, basicID, basicSecret string) (*oauth.Client, bool, int) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(form.Encode()))
		c.Request.Header.Set(header.ContentType, header.ContentTypeForm)

		if basicID != "" {
			c.Request.SetBasicAuth(basicID, basicSecret)
		}

		result, ok := oauthAppTokenClient(c, "test", false, nil)

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
	t.Run("InvalidValue", func(t *testing.T) {
		installAuthorizeBudget(t, 3)
		queries := countTableQueries(t, entity.Client{}.TableName())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(url.Values{"client_id": {rnd.GenerateUID(entity.ClientUID)}}.Encode()))
		c.Request.Header.Set(header.ContentType, header.ContentTypeForm)
		_, ok := oauthAppTokenClient(c, "test", false, func() error { return oauth.ErrInvalidToken })
		assert.False(t, ok)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(w.Body.String(), "error").String())
		assert.Equal(t, int64(0), queries.Load())
		assert.InDelta(t, 3, limiter.Authorize.IP(clientIp).Tokens(), 0.01)
	})
	t.Run("NotAnApp", func(t *testing.T) {
		_, ok, code := request(url.Values{"client_id": {"cs5gfen1bgxz7s9i"}}, "", "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, code)
	})
	t.Run("UnknownClientCharged", func(t *testing.T) {
		installAuthorizeBudget(t, 3)
		installAuthBudget(t, 3)
		queries := countClientQueries(t)

		for range 3 {
			_, ok, code := request(url.Values{"client_id": {rnd.GenerateUID(entity.ClientUID)}}, "", "")
			assert.False(t, ok)
			assert.Equal(t, http.StatusUnauthorized, code)
		}

		// Once the budget is spent, unknown client IDs are refused without a lookup.
		before := queries.Load()
		_, ok, code := request(url.Values{"client_id": {rnd.GenerateUID(entity.ClientUID)}}, "", "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusTooManyRequests, code)
		assert.Equal(t, before, queries.Load())
		assert.False(t, limiter.Auth.Reject(clientIp))
	})
	t.Run("MetadataClientNotCharged", func(t *testing.T) {
		installAuthorizeBudget(t, 3)

		for range 5 {
			_, ok, code := request(url.Values{"client_id": {"https://unregistered.example.com/oauth/client.json"}}, "", "")
			assert.False(t, ok)
			assert.Equal(t, http.StatusUnauthorized, code)
		}

		assert.False(t, limiter.Authorize.Reject(clientIp))
	})
	t.Run("KnownClientNotCharged", func(t *testing.T) {
		installAuthorizeBudget(t, 3)

		for range 5 {
			_, ok, _ := request(url.Values{"client_id": {client.ClientUID}}, "", "")
			require.True(t, ok)
		}

		assert.False(t, limiter.Authorize.Reject(clientIp))
	})
	t.Run("OverBudget", func(t *testing.T) {
		installAuthorizeBudget(t, 3)

		for !limiter.Authorize.Reject(clientIp) {
			limiter.Authorize.Reserve(clientIp)
		}

		// Registered client IDs are refused before the lookup, metadata client IDs are not.
		_, ok, code := request(url.Values{"client_id": {client.ClientUID}}, "", "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusTooManyRequests, code)

		_, ok, code = request(url.Values{"client_id": {"https://unregistered.example.com/oauth/client.json"}}, "", "")
		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, code)
	})
}

// installAuthorizeBudget replaces limiter.Authorize with a budget of n failures, and restores it after the test.
func installAuthorizeBudget(t *testing.T, n int) {
	t.Helper()
	prev := limiter.Authorize
	t.Cleanup(func() { limiter.Authorize = prev })
	limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), n)
}

func TestOAuthChargeMissingGrant(t *testing.T) {
	ip := "203.0.113.71"

	t.Run("Charged", func(t *testing.T) {
		installAuthorizeBudget(t, 3)
		oauthChargeMissingGrant(ip, oauth.ErrGrantNotFound)
		assert.InDelta(t, 2, limiter.Authorize.IP(ip).Tokens(), 0.01)
	})
	t.Run("OverBudget", func(t *testing.T) {
		installAuthorizeBudget(t, 3)
		for range 5 {
			oauthChargeMissingGrant(ip, oauth.ErrGrantNotFound)
		}
		assert.InDelta(t, 0, limiter.Authorize.IP(ip).Tokens(), 0.01)
	})
	t.Run("OtherErrors", func(t *testing.T) {
		installAuthorizeBudget(t, 3)
		oauthChargeMissingGrant(ip, oauth.ErrInvalidToken)
		oauthChargeMissingGrant(ip, oauth.ErrTokenExpired)
		oauthChargeMissingGrant(ip, oauth.ErrInvalidGrant)
		oauthChargeMissingGrant(ip, errors.New("database error"))
		oauthChargeMissingGrant(ip, oauth.ErrRefreshTokenReused)
		oauthChargeMissingGrant(ip, nil)
		assert.InDelta(t, 3, limiter.Authorize.IP(ip).Tokens(), 0.01)
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
		}

		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, accessToken, "").Code)
	})
	t.Run("UnknownClientsCharged", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, _ := tokenPair(t, app, router, conf, client)
		installAuthorizeBudget(t, 3)
		installAuthBudget(t, 3)
		queries := countClientQueries(t)

		codes := make([]int, 0, 6)

		for range 6 {
			before := queries.Load()
			r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(rnd.GenerateUID(entity.ClientUID), tokens.SignCode(rnd.AuthToken(), time.Now().Add(time.Minute))), "", "")
			codes = append(codes, r.Code)

			if r.Code == http.StatusTooManyRequests {
				assert.Equal(t, before, queries.Load(), "a refused request looked up the client")
			}
		}

		assert.Equal(t, []int{401, 401, 401, 429, 429, 429}, codes)

		// Bearer access from the same address is not affected.
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

func TestOAuthAppToken_DisabledClient(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	call := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`

	r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, approvedCode(t, app, router, conf, client)), "", "")
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	accessToken := gjson.Get(r.Body.String(), "access_token").String()
	refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()
	mcpSession := mcpInitSession(t, app, accessToken)
	require.Equal(t, http.StatusOK, mcpPost(app, call, accessToken, mcpSession).Code)

	client.AuthEnabled = false
	require.NoError(t, client.Save())

	// The access token ends at once, the grant stays, and a refresh is refused while the client is disabled.
	assert.Equal(t, http.StatusUnauthorized, mcpPost(app, call, accessToken, mcpSession).Code)
	grants, err := entity.FindClientGrants(client.ClientUID)
	require.NoError(t, err)
	assert.Len(t, grants, 1)

	r = postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
	assert.Equal(t, http.StatusUnauthorized, r.Code)
	assert.Equal(t, oauth.ErrCodeInvalidClient, gjson.Get(r.Body.String(), "error").String())

	// Once enabled again, the refresh succeeds without a new approval, and the MCP session that was open has ended.
	client.AuthEnabled = true
	require.NoError(t, client.Save())

	r = postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	nextToken := gjson.Get(r.Body.String(), "access_token").String()
	assert.Equal(t, http.StatusNotFound, mcpPost(app, call, nextToken, mcpSession).Code)
	assert.Equal(t, http.StatusOK, mcpPost(app, call, nextToken, mcpInitSession(t, app, nextToken)).Code)
}

// countTableQueries counts the queries of a table until the test ends.
func countTableQueries(t *testing.T, table string) *atomic.Int64 {
	t.Helper()

	name := "test:count_table_queries_" + rnd.Base36(8)
	count := &atomic.Int64{}

	require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == table {
			count.Add(1)
		}
	}))
	t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

	return count
}

func TestOAuthAppToken_SignedValues(t *testing.T) {
	ip := appTokenTestClientIP()

	// refused sends a token request that must be refused as invalid_grant, and checks that it charges
	// limiter.Authorize the specified number of times and leaves limiter.Auth untouched.
	refused := func(t *testing.T, app http.Handler, path string, values url.Values, charges float64) {
		t.Helper()
		prevAuth, prevAuthorize := limiter.Auth, limiter.Authorize
		t.Cleanup(func() { limiter.Auth, limiter.Authorize = prevAuth, prevAuthorize })
		limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)
		limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), 3)
		r := postTokenForm(app, path, values, "", "")
		if path == "/api/v1/oauth/token" {
			assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())
			assert.Equal(t, oauth.ErrCodeInvalidGrant, gjson.Get(r.Body.String(), "error").String())
		} else {
			assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		}
		assert.InDelta(t, 3, limiter.Auth.IP(ip).Tokens(), 0.01)
		assert.InDelta(t, 3-charges, limiter.Authorize.IP(ip).Tokens(), 0.01)
	}

	t.Run("ForgedCode", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		queries := countTableQueries(t, entity.AuthCode{}.TableName())
		clientQueries := countTableQueries(t, entity.Client{}.TableName())
		entity.FlushClientCache()

		value, _, _, ok := tokens.SplitValue(code)
		require.True(t, ok)
		expired := tokens.SignCode(value, time.Now().Add(-time.Second))

		for _, forged := range []string{rnd.AuthToken(), tokens.SignRefresh(rnd.AuthToken(), time.Now().Add(time.Minute)), code[:len(code)-1] + "x", expired} {
			refused(t, app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, forged), 0)
		}

		assert.Equal(t, int64(0), queries.Load())
		assert.Equal(t, int64(0), clientQueries.Load())

		// The genuine code was not consumed, and its lookup is counted.
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Positive(t, queries.Load())

		// A second redemption of the genuine code is refused.
		refused(t, app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), 0)
	})
	t.Run("CachedClient", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		queries := countTableQueries(t, entity.Client{}.TableName())

		// The authorization request read the client, so the token endpoint finds it in the cache.
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()
		refused(t, app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, rnd.AuthToken()), 0)
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, int64(0), queries.Load())
	})
	t.Run("ForgedRefreshToken", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()
		grantUID, _, _ := strings.Cut(refreshToken, ".")
		queries := countTableQueries(t, entity.Grant{}.TableName())
		clientQueries := countTableQueries(t, entity.Client{}.TableName())
		entity.FlushClientCache()

		value, _, _, ok := tokens.SplitValue(refreshToken)
		require.True(t, ok)
		expired := tokens.SignRefresh(value, time.Now().Add(-time.Second))

		for _, forged := range []string{rnd.RefreshToken(grantUID), tokens.SignCode(rnd.RefreshToken(grantUID), time.Now().Add(time.Minute)), refreshToken[:len(refreshToken)-1] + "x", expired} {
			refused(t, app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {forged}}, 0)
			refused(t, app, "/api/v1/oauth/revoke", url.Values{"token": {forged}, "token_type_hint": {"refresh_token"}}, 0)
		}

		assert.Equal(t, int64(0), queries.Load())
		assert.Equal(t, int64(0), clientQueries.Load())
		assert.NotNil(t, entity.FindGrant(grantUID))

		// The genuine refresh token rotates into another signed one.
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		next := gjson.Get(r.Body.String(), "refresh_token").String()
		_, err := tokens.VerifyRefresh(next)
		assert.NoError(t, err)
		assert.NotEqual(t, refreshToken, next)
		assert.Positive(t, queries.Load())
	})
	t.Run("RevokeWithUnknownBearer", func(t *testing.T) {
		app, _, _, _ := prepareAppTokenTest(t)
		limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)
		limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), 3)

		// A header token that resolves to no session counts towards limiter.Auth once.
		req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/revoke", strings.NewReader(url.Values{"token": {rnd.RefreshToken(rnd.GenerateUID(entity.GrantUID))}}.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		req.Header.Set(header.Auth, header.AuthBearer+" "+rnd.AuthToken())
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.InDelta(t, 2, limiter.Auth.IP(ip).Tokens(), 0.01)
		assert.InDelta(t, 3, limiter.Authorize.IP(ip).Tokens(), 0.01)
	})
	t.Run("StaleRefreshTokenRevokeNotCharged", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		first := gjson.Get(r.Body.String(), "refresh_token").String()
		r = postTokenForm(app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {first}}, "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())

		// Revoking a refresh token that is no longer current succeeds without effect.
		refused(t, app, "/api/v1/oauth/revoke", url.Values{"token": {first}, "token_type_hint": {"refresh_token"}}, 0)
	})
	t.Run("GrantReadError", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		_, refreshToken := tokenPair(t, app, router, conf, client)
		installAuthorizeBudget(t, 3)
		name := "test:fail_grant_reads_" + rnd.Base36(8)
		require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (entity.Grant{}).TableName() {
				_ = db.AddError(errors.New("grant lookup failed"))
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

		// A grant that cannot be read is answered with a server error.
		r := postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
		assert.Equal(t, http.StatusInternalServerError, r.Code, r.Body.String())
		r = postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {refreshToken}, "token_type_hint": {"refresh_token"}}, "", "")
		assert.Equal(t, http.StatusInternalServerError, r.Code, r.Body.String())
		assert.InDelta(t, 3, limiter.Authorize.IP(ip).Tokens(), 0.01)
	})
	t.Run("GrantGone", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()
		grantUID, _, _ := strings.Cut(refreshToken, ".")
		grant := entity.FindGrant(grantUID)
		require.NotNil(t, grant)
		require.NoError(t, grant.Delete())

		// A refresh token whose grant is gone counts towards limiter.Authorize at both endpoints.
		refused(t, app, "/api/v1/oauth/token", url.Values{"grant_type": {"refresh_token"}, "client_id": {client.ClientUID}, "refresh_token": {refreshToken}}, 1)
		refused(t, app, "/api/v1/oauth/revoke", url.Values{"token": {refreshToken}, "token_type_hint": {"refresh_token"}}, 1)
	})
}

// auditMessagesOf returns the audit messages logged by each request, one per request.
func auditMessagesOf(t *testing.T, app http.Handler, requests ...func() *httptest.ResponseRecorder) []string {
	t.Helper()

	hook := captureAuditLog(t)
	result := make([]string, 0, len(requests))

	for _, request := range requests {
		hook.Reset()
		request()
		require.NotEmpty(t, hook.AllEntries())
		result = append(result, hook.LastEntry().Message)
	}

	return result
}

func TestOAuthAppTokenClient_ConstantAuditMessages(t *testing.T) {
	ip := appTokenTestClientIP()

	// tokenRequest returns a function that sends a token request with the form values.
	tokenRequest := func(app http.Handler, values url.Values) func() *httptest.ResponseRecorder {
		return func() *httptest.ResponseRecorder { return postTokenForm(app, "/api/v1/oauth/token", values, "", "") }
	}

	t.Run("ClientSecret", func(t *testing.T) {
		app, _, _, _ := prepareAppTokenTest(t)
		installAuthorizeBudget(t, 3)
		form := func(id string) url.Values {
			v := codeTokenForm(id, rnd.AuthToken())
			v.Set("client_secret", "secret")
			return v
		}
		messages := auditMessagesOf(t, app, tokenRequest(app, form(rnd.GenerateUID(entity.ClientUID))), tokenRequest(app, form(rnd.GenerateUID(entity.ClientUID))))
		assert.Equal(t, messages[0], messages[1])
	})
	t.Run("InvalidValue", func(t *testing.T) {
		app, _, _, _ := prepareAppTokenTest(t)
		installAuthorizeBudget(t, 3)
		messages := auditMessagesOf(t, app,
			tokenRequest(app, codeTokenForm(rnd.GenerateUID(entity.ClientUID), rnd.AuthToken())),
			tokenRequest(app, codeTokenForm(rnd.GenerateUID(entity.ClientUID), rnd.AuthToken())))
		assert.Equal(t, messages[0], messages[1])
	})
	t.Run("RateLimited", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny
		installAuthorizeBudget(t, 3)
		for !limiter.Authorize.Reject(ip) {
			limiter.Authorize.Reserve(ip)
		}

		// Requests refused over the limit write no audit entry.
		hook := captureAuditLog(t)
		code := tokens.SignCode(rnd.AuthToken(), time.Now().Add(time.Minute))
		assert.Equal(t, http.StatusTooManyRequests, tokenRequest(app, codeTokenForm("https://claude.ai/oauth/audit-a.json", code))().Code)
		assert.Equal(t, http.StatusTooManyRequests, tokenRequest(app, codeTokenForm(rnd.GenerateUID(entity.ClientUID), code))().Code)
		conf.Options().OAuthApps = config.OAuthAppsNone
		assert.Equal(t, http.StatusTooManyRequests, tokenRequest(app, codeTokenForm(rnd.GenerateUID(entity.ClientUID), code))().Code)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("MetadataNotAllowed", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		conf.Options().OAuthApps = config.OAuthAppsRegistered
		installAuthorizeBudget(t, 3)
		code := tokens.SignCode(rnd.AuthToken(), time.Now().Add(time.Minute))
		messages := auditMessagesOf(t, app,
			tokenRequest(app, codeTokenForm("https://example.com/audit-a.json", code)),
			tokenRequest(app, codeTokenForm("https://example.com/audit-b.json", code)))
		assert.Equal(t, messages[0], messages[1])
		assert.InDelta(t, 3, limiter.Authorize.IP(ip).Tokens(), 0.01)
	})
}

func TestOAuthAppRefresh_RefusedBeforeGrantLookup(t *testing.T) {
	ip := appTokenTestClientIP()

	// exhaust uses up limiter.Authorize for the test client address.
	exhaust := func() {
		for !limiter.Authorize.Reject(ip) {
			limiter.Authorize.Reserve(ip)
		}
	}

	t.Run("MetadataClient", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny
		installAuthorizeBudget(t, 3)
		queries := countTableQueries(t, entity.Grant{}.TableName())
		gone := tokens.SignRefresh(rnd.RefreshToken(rnd.GenerateUID(entity.GrantUID)), time.Now().Add(time.Hour))

		// Within the budget, the grant is looked up and its absence is charged.
		r := postTokenForm(app, "/api/v1/oauth/token", refreshForm(testMetadataClientID, gone), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())
		assert.Equal(t, int64(1), queries.Load())

		// Over the budget, a valid refresh token is refused before the lookup.
		exhaust()
		r = postTokenForm(app, "/api/v1/oauth/token", refreshForm(testMetadataClientID, gone), "", "")
		assert.Equal(t, http.StatusTooManyRequests, r.Code, r.Body.String())
		assert.Equal(t, int64(1), queries.Load())
	})
	t.Run("Revoke", func(t *testing.T) {
		app, _, _, _ := prepareAppTokenTest(t)
		installAuthorizeBudget(t, 3)
		installAuthBudget(t, 3)
		queries := countTableQueries(t, entity.Grant{}.TableName())
		gone := tokens.SignRefresh(rnd.RefreshToken(rnd.GenerateUID(entity.GrantUID)), time.Now().Add(time.Hour))
		form := url.Values{"token": {gone}, "token_type_hint": {"refresh_token"}}

		r := postTokenForm(app, "/api/v1/oauth/revoke", form, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, int64(1), queries.Load())
		assert.InDelta(t, 2, limiter.Authorize.IP(ip).Tokens(), 0.01)

		exhaust()
		r = postTokenForm(app, "/api/v1/oauth/revoke", form, "", "")
		assert.Equal(t, http.StatusTooManyRequests, r.Code, r.Body.String())
		assert.Equal(t, int64(1), queries.Load())

		// A value that fails verification is answered without a lookup and is not refused.
		r = postTokenForm(app, "/api/v1/oauth/revoke", url.Values{"token": {rnd.RefreshToken(rnd.GenerateUID(entity.GrantUID))}, "token_type_hint": {"refresh_token"}}, "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, int64(1), queries.Load())
		assert.InDelta(t, 3, limiter.Auth.IP(ip).Tokens(), 0.01)
	})
}
