package api

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	testMetadataClientID = "https://claude.ai/oauth/mcp-client.json"
	testMetadataCallback = "https://callback.example.com/cb"
)

// stubAppMetadata serves a client metadata document for testMetadataClientID instead of fetching it.
func stubAppMetadata(t *testing.T) *int {
	t.Helper()

	calls := 0
	prev := oauth.FetchMetadata
	oauth.FetchMetadata = func(clientID string) ([]byte, http.Header, error) {
		calls++
		return []byte(`{"client_id":"` + clientID + `","client_name":"Claude","redirect_uris":["` + testAppCallback + `","` + testMetadataCallback + `"],"token_endpoint_auth_method":"none"}`), http.Header{"Content-Type": {"application/json"}}, nil
	}

	oauth.FlushMetadataCache()
	t.Cleanup(func() {
		oauth.FetchMetadata = prev
		oauth.FlushMetadataCache()
	})

	return &calls
}

func TestOAuthApp_MetadataClient(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`

	t.Run("RefusedByDefaultWithoutFetch", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		calls := stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsRegistered

		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+testAuthorizeQuery(conf, testMetadataClientID).Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
		assert.Equal(t, 0, *calls)
	})
	t.Run("FullFlow", func(t *testing.T) {
		app, router, conf, _ := prepareAppTokenTest(t)
		calls := stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny
		t.Cleanup(func() { entity.DeleteClientGrants(testMetadataClientID) })

		FindUserGrants(router)

		// The callback is on another origin, so the client ID origin must be shown.
		q := testAuthorizeQuery(conf, testMetadataClientID)
		q.Set("redirect_uri", testMetadataCallback)
		id := authorizeRequestID(t, app, q)
		adminToken := AuthenticateAdmin(app, router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "Claude", gjson.Get(r.Body.String(), "client.name").String())
		assert.Equal(t, "https://claude.ai", gjson.Get(r.Body.String(), "client.origin").String())

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)

		assert.Equal(t, "callback.example.com", loc.Host)

		tokenForm := codeTokenForm(testMetadataClientID, loc.Query().Get("code"))
		tokenForm.Set("redirect_uri", testMetadataCallback)
		r = postTokenForm(app, "/api/v1/oauth/token", tokenForm, "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		accessToken := gjson.Get(r.Body.String(), "access_token").String()
		refreshToken := gjson.Get(r.Body.String(), "refresh_token").String()

		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, accessToken, "").Code)

		r = postTokenForm(app, "/api/v1/oauth/token", refreshForm(testMetadataClientID, refreshToken), "", "")
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())

		// The document was fetched once and then served from the cache.
		assert.Equal(t, 1, *calls)

		grant := entity.FindUserClientGrant(entity.Admin.UserUID, testMetadataClientID)
		require.NotNil(t, grant)
		assert.Equal(t, "Claude", grant.ClientName)

		r = AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+entity.Admin.UserUID+"/grants", adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "https://claude.ai", gjson.Get(r.Body.String(), "#(ClientID==\""+testMetadataClientID+"\").Origin").String())

		// Issuing tokens does not add an auth_clients row for the client.
		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", testMetadataClientID).Count(&count).Error)
		assert.Zero(t, count)
	})
	t.Run("FetchChargesRateLimit", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		calls := stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny

		// The first request fetches the document and stores a request, the second only stores a request.
		before := limiter.Authorize.IP(header.UnknownIP).Tokens()
		authorizeRequestID(t, app, testAuthorizeQuery(conf, testMetadataClientID))
		after := limiter.Authorize.IP(header.UnknownIP).Tokens()
		assert.Less(t, after, before-1.5)

		authorizeRequestID(t, app, testAuthorizeQuery(conf, testMetadataClientID))
		assert.Greater(t, limiter.Authorize.IP(header.UnknownIP).Tokens(), after-1.5)
		assert.Equal(t, 1, *calls)

		// A token request of a cached client uses the cached document.
		after = limiter.Authorize.IP(header.UnknownIP).Tokens()
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(testMetadataClientID, tokens.SignCode(rnd.AuthToken(), time.Now().Add(time.Minute))), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())
		assert.Greater(t, limiter.Authorize.IP(header.UnknownIP).Tokens(), after-0.5)
	})
	t.Run("TokenFetchRateLimited", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		calls := stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny
		limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), 3)

		codes := make([]int, 0, 5)

		for i := 0; i < 5; i++ {
			code := tokens.SignCode(rnd.AuthToken(), time.Now().Add(time.Minute))
			codes = append(codes, postTokenForm(app, "/api/v1/oauth/token", codeTokenForm("https://claude.ai/oauth/client-"+strconv.Itoa(i)+".json", code), "", "").Code)
		}

		// Each request fetches a new document, and the code was not issued.
		assert.Equal(t, []int{http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest, http.StatusTooManyRequests, http.StatusTooManyRequests}, codes)
		assert.Equal(t, 3, *calls)
	})
	t.Run("InvalidValueNotFetched", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		calls := stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny
		installAuthorizeBudget(t, 3)

		// A code or refresh token that fails verification is refused before the document is fetched.
		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm("https://claude.ai/oauth/client-x.json", rnd.AuthToken()), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())
		r = postTokenForm(app, "/api/v1/oauth/token", refreshForm("https://claude.ai/oauth/client-y.json", "x"), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code, r.Body.String())
		assert.Equal(t, 0, *calls)
		assert.InDelta(t, 3, limiter.Authorize.IP(appTokenTestClientIP()).Tokens(), 0.01)
	})
	t.Run("ConsentFetchRateLimited", func(t *testing.T) {
		app, router, conf, _ := prepareAppTokenTest(t)
		calls := stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny

		id := authorizeRequestID(t, app, testAuthorizeQuery(conf, testMetadataClientID))
		adminToken := AuthenticateAdmin(app, router)

		oauth.FlushMetadataCache()
		limiter.Authorize = limiter.NewLimit(rate.Every(24*time.Hour), 3)
		limiter.Authorize.ReserveN(header.UnknownIP, 3)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent", `{"request":"`+id+`"}`, adminToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
		assert.Equal(t, 1, *calls)
	})
	t.Run("ErrorsAreRendered", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny

		q := testAuthorizeQuery(conf, testMetadataClientID)
		q.Set("code_challenge_method", "plain")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
	})
	t.Run("UnregisteredRedirectURI", func(t *testing.T) {
		app, _, conf, _ := prepareAppTokenTest(t)
		stubAppMetadata(t)
		conf.Options().OAuthApps = config.OAuthAppsAny

		q := testAuthorizeQuery(conf, testMetadataClientID)
		q.Set("redirect_uri", "https://evil.example.com/cb")
		r := PerformRequest(app, http.MethodGet, "/api/v1/oauth/authorize?"+q.Encode())
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Empty(t, r.Header().Get("Location"))
	})
}
