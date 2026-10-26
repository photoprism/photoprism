package api

import (
	"crypto/sha256"
	"encoding/base64"
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
)

const (
	testAppVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	testAppCallback = "https://claude.ai/api/mcp/auth_callback"
)

// testAppChallenge returns the S256 code challenge of testAppVerifier.
func testAppChallenge() string {
	sum := sha256.Sum256([]byte(testAppVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// prepareAppFlowTest enables the app flow, registers the authorize and consent routes, and adds an app client
// that is deleted after the test.
func prepareAppFlowTest(t *testing.T, scope string) (app *gin.Engine, router *gin.RouterGroup, conf *config.Config, client *entity.Client) {
	t.Helper()

	app, router, conf = NewApiTest()

	// Use separate rate limiters, so these requests do not deplete the ones other tests share.
	origAuth, origAuthorize := limiter.Auth, limiter.Authorize
	limiter.Auth = limiter.NewLimit(rate.Every(10*time.Second), 60)
	limiter.Authorize = limiter.NewLimit(rate.Every(10*time.Second), 60)
	t.Cleanup(func() { limiter.Auth, limiter.Authorize = origAuth, origAuthorize })

	opt := conf.Options()
	apps, public, authMode := opt.OAuthApps, opt.Public, opt.AuthMode
	t.Cleanup(func() { opt.OAuthApps, opt.Public, opt.AuthMode = apps, public, authMode })
	opt.OAuthApps, opt.Public, opt.AuthMode = config.OAuthAppsRegistered, false, config.AuthModePasswd

	OAuthAuthorize(router)
	OAuthConsent(router)
	OAuthConsentApprove(router)
	OAuthConsentDeny(router)

	var err error
	client, err = entity.AddClient(form.Client{ClientName: "Claude", AuthMethod: authn.MethodPKCE.String(), AuthScope: scope,
		RedirectURIs: []string{testAppCallback, "http://127.0.0.1/callback"}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Delete() })

	return app, router, conf, client
}

// testAuthorizeQuery returns the query of a valid authorization request for the client.
func testAuthorizeQuery(conf *config.Config, clientID string) url.Values {
	return url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {testAppCallback},
		"scope":                 {"mcp"},
		"state":                 {"xyz"},
		"code_challenge":        {testAppChallenge()},
		"code_challenge_method": {"S256"},
		"resource":              {conf.OAuthResourceUrl()},
	}
}

// authorizeRequestID sends an authorization request and returns the pending request ID from the consent redirect.
func authorizeRequestID(t *testing.T, app http.Handler, q url.Values) string {
	t.Helper()

	r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
	require.Equal(t, http.StatusFound, r.Code, r.Body.String())

	loc, err := url.Parse(r.Header().Get("Location"))
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(loc.Path, OAuthConsentUri), loc.Path)

	id := loc.Query().Get("request")
	require.NotEmpty(t, id)

	return id
}

// newConsentSession stores a web session of the user with the optional scope and deletes it after the test.
func newConsentSession(t *testing.T, user *entity.User, scope string) string {
	t.Helper()

	sess := entity.NewSession(3600, 0)
	sess.SetUser(user)
	sess.AuthScope = scope
	require.NoError(t, sess.Create())
	t.Cleanup(func() { _ = sess.Delete() })

	return sess.AuthToken()
}

// newConsentUser stores an admin account with the specified options and deletes it after the test.
func newConsentUser(t *testing.T, name, scope string, canLogin bool) *entity.User {
	t.Helper()

	u := entity.NewUser()
	u.UserName = name
	u.UserRole = "admin"
	u.UserScope = scope
	u.CanLogin = canLogin
	require.NoError(t, u.Create())
	t.Cleanup(func() { _ = u.Delete() })

	return u
}

// errorRedirect asserts a redirect to the client callback with the OAuth error and the issuer.
func errorRedirect(t *testing.T, conf *config.Config, r interface{ Header() http.Header }, errCode string) {
	t.Helper()

	loc, err := url.Parse(r.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "claude.ai", loc.Host)
	assert.Equal(t, errCode, loc.Query().Get("error"))
	assert.Equal(t, conf.SiteUrl(), loc.Query().Get("iss"))
	assert.Equal(t, "xyz", loc.Query().Get("state"))
	assert.Empty(t, loc.Query().Get("code"))
}

func TestOAuthAppAuthorize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp photos")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))

		m, err := oauth.FindRequest(id)
		require.NoError(t, err)
		assert.Equal(t, client.ClientUID, m.ClientID)
		assert.Equal(t, "mcp", m.Scope)
		assert.Equal(t, conf.OAuthResourceUrl(), m.Resource)
	})
	t.Run("DefaultResourceAndScope", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Del("resource")
		q.Set("scope", "photos")
		id := authorizeRequestID(t, app, q)

		m, err := oauth.FindRequest(id)
		require.NoError(t, err)
		assert.Equal(t, "mcp photos", m.Scope)
		assert.Equal(t, conf.OAuthResourceUrl(), m.Resource)
	})
	t.Run("LoopbackPort", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("redirect_uri", "http://127.0.0.1:49152/callback")
		authorizeRequestID(t, app, q)
	})
	t.Run("NoPKCE", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Del("code_challenge")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		assert.Equal(t, http.StatusFound, r.Code)
		errorRedirect(t, conf, r, oauth.ErrCodeInvalidRequest)
	})
	t.Run("PlainPKCE", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("code_challenge_method", "plain")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		errorRedirect(t, conf, r, oauth.ErrCodeInvalidRequest)
	})
	t.Run("UnregisteredRedirectURI", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("redirect_uri", "https://evil.example.com/callback")
		q.Del("code_challenge")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
		assert.Equal(t, oauth.ErrCodeInvalidRequest, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("WrongResource", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("resource", "https://other.example.com/api/v1/mcp")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		errorRedirect(t, conf, r, oauth.ErrCodeInvalidTarget)
	})
	t.Run("UnknownScope", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "*")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp password")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		errorRedirect(t, conf, r, oauth.ErrCodeInvalidScope)
	})
	t.Run("ScopeNotAllowedForClient", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		errorRedirect(t, conf, r, oauth.ErrCodeInvalidScope)
	})
	t.Run("UnknownClient", func(t *testing.T) {
		app, _, conf, _ := prepareAppFlowTest(t, "mcp")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+testAuthorizeQuery(conf, "cs5gfen1bgxz7s9i").Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
		assert.Equal(t, oauth.ErrCodeInvalidClient, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("DisabledClientMatchesUnknown", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		client.AuthEnabled = false
		require.NoError(t, client.Save())

		perform := func(clientID, accept string) *httptest.ResponseRecorder {
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?"+testAuthorizeQuery(conf, clientID).Encode(), nil)
			req.Header.Set(header.Accept, accept)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			return w
		}

		disabledHTML := perform(client.ClientUID, "text/html")
		unknownHTML := perform("cs5gfen1bgxz7s9i", "text/html")
		assert.Equal(t, http.StatusBadRequest, disabledHTML.Code)
		assert.Contains(t, disabledHTML.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, disabledHTML.Body.String(), "<p>Unknown client.</p>")
		assert.Equal(t, unknownHTML.Code, disabledHTML.Code)
		assert.Equal(t, unknownHTML.Body.String(), disabledHTML.Body.String())

		disabledJSON := perform(client.ClientUID, "application/json")
		unknownJSON := perform("cs5gfen1bgxz7s9i", "application/json")
		assert.Equal(t, http.StatusBadRequest, disabledJSON.Code)
		assert.JSONEq(t, `{"error":"invalid_client","error_description":"unknown client"}`, disabledJSON.Body.String())
		assert.Equal(t, unknownJSON.Code, disabledJSON.Code)
		assert.Equal(t, unknownJSON.Body.String(), disabledJSON.Body.String())
	})
	t.Run("UnregisteredRedirectURIPage", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("redirect_uri", "https://evil.example.com/callback")
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode(), nil)
		req.Header.Set(header.FetchMode, "navigate")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Empty(t, w.Header().Get("Location"))
		assert.Contains(t, w.Body.String(), "<p>redirect_uri not registered.</p>")
		assert.Contains(t, w.Body.String(), "<code>invalid_request</code>")
		assert.NotContains(t, w.Body.String(), "evil.example.com")
		assert.NotContains(t, w.Body.String(), client.ClientUID)
	})
	t.Run("MetadataClientRefusedByDefault", func(t *testing.T) {
		app, _, conf, _ := prepareAppFlowTest(t, "mcp")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+testAuthorizeQuery(conf, "https://claude.ai/oauth/client.json").Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, oauth.ErrCodeInvalidClient, gjson.Get(r.Body.String(), "error").String())
	})
	t.Run("ChargesRateLimit", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		before := limiter.Authorize.IP(header.UnknownIP).Tokens()
		authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		assert.Less(t, limiter.Authorize.IP(header.UnknownIP).Tokens(), before-0.5)
	})
	t.Run("RateLimitKeepsSessions", func(t *testing.T) {
		app, router, _, _ := prepareAppFlowTest(t, "mcp")
		authToken := AuthenticateAdmin(app, router)
		limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)
		limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), 3)

		limited := false

		for i := 0; i < 10 && !limited; i++ {
			limited = PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize").Code == http.StatusTooManyRequests
		}

		assert.True(t, limited)

		// Existing sessions keep working.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"0123456789abcdef0123456789abcdef0123456789abcdef"}`, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("RepeatedRedirectURI", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Add("redirect_uri", "https://evil.example.com/cb")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
	})
}

func TestOAuthConsent(t *testing.T) {
	t.Run("ApproveAdmin", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		id := authorizeRequestID(t, app, q)
		authToken := AuthenticateAdmin(app, router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "Claude", gjson.Get(r.Body.String(), "client.name").String())
		assert.Equal(t, "https://claude.ai", gjson.Get(r.Body.String(), "client.origin").String())
		assert.Equal(t, "admin", gjson.Get(r.Body.String(), "account.name").String())
		assert.True(t, gjson.Get(r.Body.String(), "available").Bool())
		assert.Equal(t, `[{"name":"mcp","available":true},{"name":"photos","available":true}]`, gjson.Get(r.Body.String(), "scopes").Raw)

		// GET cannot approve.
		r = AuthenticatedRequest(app, http.MethodGet, "/api/v1/oauth/consent/approve?request="+id, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)

		// Partial approval.
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())

		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)
		assert.Equal(t, "claude.ai", loc.Host)
		assert.Equal(t, "xyz", loc.Query().Get("state"))
		assert.Equal(t, conf.SiteUrl(), loc.Query().Get("iss"))

		code, err := oauth.RedeemCode(loc.Query().Get("code"), client.ClientUID, testAppCallback, testAppVerifier)
		require.NoError(t, err)
		assert.Equal(t, "mcp", code.Scope)
		assert.Equal(t, entity.Admin.UserUID, code.UserUID)

		// The request is completed only once.
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("ApproveWithoutMCP", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		id := authorizeRequestID(t, app, q)
		authToken := AuthenticateAdmin(app, router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["photos"]}`, authToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","albums"]}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("AccountWithoutMCP", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		id := authorizeRequestID(t, app, q)
		authToken := AuthenticateAdmin(app, router)

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		admin.MCP = -1
		require.NoError(t, admin.Save())
		t.Cleanup(func() {
			admin.MCP = 0
			_ = admin.Save()
		})

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.False(t, gjson.Get(r.Body.String(), "available").Bool())
		assert.Equal(t, `[{"name":"mcp","available":false},{"name":"photos","available":false}]`, gjson.Get(r.Body.String(), "scopes").Raw)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","photos"]}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("RoleWithoutMCP", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		id := authorizeRequestID(t, app, q)
		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.False(t, gjson.Get(r.Body.String(), "available").Bool())
		assert.Equal(t, `[{"name":"mcp","available":false},{"name":"photos","available":false}]`, gjson.Get(r.Body.String(), "scopes").Raw)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","photos"]}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)

		// No code was issued.
		var codes int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.AuthCode{}).Where("client_id = ? AND code_type = ?", client.ClientUID, entity.AuthCodeCode).Count(&codes).Error)
		assert.Zero(t, codes)

		// The link back to the app passes access_denied with the issuer.
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/deny", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)
		assert.Equal(t, oauth.ErrCodeAccessDenied, loc.Query().Get("error"))
		assert.Equal(t, conf.SiteUrl(), loc.Query().Get("iss"))
		assert.Equal(t, "xyz", loc.Query().Get("state"))

		_, err = oauth.FindRequest(id)
		assert.ErrorIs(t, err, oauth.ErrRequestNotFound)
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("ClientSession", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))

		alice := entity.FindUserByName("alice")
		require.NotNil(t, alice)
		sess, err := entity.AddClientSession("consent-app-password", 3600, "*", authn.GrantCLI, alice)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, sess.AuthToken())
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("UnknownRequest", func(t *testing.T) {
		app, router, _, _ := prepareAppFlowTest(t, "mcp")
		authToken := AuthenticateAdmin(app, router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"0123456789abcdef0123456789abcdef0123456789abcdef"}`, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("Deny", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		authToken := AuthenticateAdmin(app, router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/deny", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code)
		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)
		assert.Equal(t, oauth.ErrCodeAccessDenied, loc.Query().Get("error"))

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}

func TestOAuthConsentScopes(t *testing.T) {
	t.Run("SessionScopeWithoutMCP", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		id := authorizeRequestID(t, app, q)
		authToken := newConsentSession(t, newConsentUser(t, "consent-scoped-photos", "photos", true), "photos")

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.False(t, gjson.Get(r.Body.String(), "available").Bool())

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","photos"]}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("SessionScopeDropsScope", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp photos")
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", "mcp photos")
		id := authorizeRequestID(t, app, q)
		authToken := newConsentSession(t, newConsentUser(t, "consent-scoped-mcp", "mcp", true), "mcp")

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.True(t, gjson.Get(r.Body.String(), "available").Bool())
		assert.Equal(t, `[{"name":"mcp","available":true},{"name":"photos","available":false}]`, gjson.Get(r.Body.String(), "scopes").Raw)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","photos"]}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp","mcp"]}`, authToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)
		code, err := oauth.RedeemCode(loc.Query().Get("code"), client.ClientUID, testAppCallback, testAppVerifier)
		require.NoError(t, err)
		assert.Equal(t, "mcp", code.Scope)
	})
}

func TestOAuthConsentRequest(t *testing.T) {
	t.Run("LoginDisabled", func(t *testing.T) {
		app, _, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		u := newConsentUser(t, "consent-no-login", "*", true)
		authToken := newConsentSession(t, u, "")
		require.NoError(t, u.Updates(entity.Values{"CanLogin": false}))

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ClientDisabled", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		authToken := AuthenticateAdmin(app, router)

		client.AuthEnabled = false
		require.NoError(t, client.Save())

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("CallbackRemoved", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		authToken := AuthenticateAdmin(app, router)

		data := client.GetData()
		data.RedirectURIs = []string{"http://127.0.0.1/callback"}
		require.NoError(t, client.SetData(data).Save())

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("AppsDisabled", func(t *testing.T) {
		app, router, conf, client := prepareAppFlowTest(t, "mcp")
		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, client.ClientUID))
		authToken := AuthenticateAdmin(app, router)
		conf.Options().OAuthApps = config.OAuthAppsNone

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, _, conf, _ := prepareAppFlowTest(t, "mcp")
		conf.Options().Public = true
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"abc"}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("InvalidBody", func(t *testing.T) {
		app, router, _, _ := prepareAppFlowTest(t, "mcp")
		authToken := AuthenticateAdmin(app, router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{}`, authToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}

func TestOAuthConsentUsableScopes(t *testing.T) {
	admin := entity.FindUserByName("alice")
	guest := entity.FindUserByName("gandalf")
	require.NotNil(t, admin)
	require.NotNil(t, guest)

	t.Run("Unrestricted", func(t *testing.T) {
		s := &entity.Session{}
		assert.Equal(t, []string{"mcp", "photos"}, oauthConsentUsableScopes(s, admin, []string{"mcp", "photos"}))
	})
	t.Run("SessionScope", func(t *testing.T) {
		s := &entity.Session{AuthScope: "mcp"}
		assert.Equal(t, []string{"mcp"}, oauthConsentUsableScopes(s, admin, []string{"mcp", "photos"}))
		s.AuthScope = "photos"
		assert.Empty(t, oauthConsentUsableScopes(s, admin, []string{"mcp", "photos"}))
	})
	t.Run("RoleWithoutMCP", func(t *testing.T) {
		assert.Empty(t, oauthConsentUsableScopes(&entity.Session{}, guest, []string{"mcp", "photos"}))
	})
	t.Run("AccountSetting", func(t *testing.T) {
		u := *admin
		u.MCP = 1
		assert.Equal(t, []string{"mcp", "photos"}, oauthConsentUsableScopes(&entity.Session{}, &u, []string{"mcp", "photos"}))
		u.MCP = -1
		assert.Empty(t, oauthConsentUsableScopes(&entity.Session{}, &u, []string{"mcp", "photos"}))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Empty(t, oauthConsentUsableScopes(nil, admin, []string{"mcp"}))
		assert.Empty(t, oauthConsentUsableScopes(&entity.Session{}, nil, []string{"mcp"}))
	})
}

func TestOAuthErrorRedirect(t *testing.T) {
	t.Run("Redirect", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		oauthErrorRedirect(c, testAppCallback, oauth.ErrCodeInvalidScope, "unsupported scope", "xyz")
		assert.Equal(t, http.StatusFound, w.Code)
		loc, err := url.Parse(w.Header().Get("Location"))
		require.NoError(t, err)
		assert.Equal(t, oauth.ErrCodeInvalidScope, loc.Query().Get("error"))
		assert.NotEmpty(t, loc.Query().Get("iss"))
	})
	t.Run("InvalidRedirectURI", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		oauthErrorRedirect(c, "javascript:alert(1)", oauth.ErrCodeInvalidScope, "unsupported scope", "xyz")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Empty(t, w.Header().Get("Location"))
	})
}
