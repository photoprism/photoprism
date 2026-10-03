package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// installAuthBudget replaces the shared authentication limiters with a burst of n and restores
// them afterwards. The default burst of 60, combined with one shared bucket for an unknown
// client IP, never exhausts inside a test, so a route measured against it reads as unthrottled
// whether or not it enforces anything. Note that limiter.NewLimit floors the burst at 3.
func installAuthBudget(t *testing.T, n int) {
	oldLogin, oldAuth := limiter.Login, limiter.Auth

	t.Cleanup(func() { limiter.Login, limiter.Auth = oldLogin, oldAuth })

	limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), n)
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), n)
}

// deadAuthToken returns a well-formed auth token that resolves to no session, so a request
// carrying it passes form validation and is refused by the lookup rather than before it.
func deadAuthToken() string {
	return rnd.AuthToken()
}

// TestOAuthRevoke_RateLimit covers the failure budget on the revoke endpoint. Every case names
// the bucket it measures: both limiters are installed at the same burst, so a subtest that only
// asserts a 429 cannot tell limiter.Auth from limiter.Login, and the charging rule is the thing
// under test.
func TestOAuthRevoke_RateLimit(t *testing.T) {
	const revokePath = "/api/v1/oauth/revoke"
	const tokenPath = "/api/v1/oauth/token" // #nosec G101 test constant, not a credential

	// The handler resolves an httptest request with no RemoteAddr to this address.
	const testIP = header.UnknownIP

	revokeForm := func(app http.Handler, values url.Values, authToken string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest(http.MethodPost, revokePath, strings.NewReader(values.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)

		if authToken != "" {
			req.Header.Set(header.XAuthToken, authToken)
		}

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		return w
	}

	t.Run("RepeatedUnresolvedTokens", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthRevoke(router)
		installAuthBudget(t, 3)

		// Well-formed tokens, so each request passes validation and is refused by the
		// lookup rather than before it.
		for i := range 3 {
			w := revokeForm(app, url.Values{
				"token":           {deadAuthToken()},
				"token_type_hint": {"access_token"},
			}, "")

			assert.Equal(t, http.StatusUnauthorized, w.Code, "attempt %d", i+1)
		}

		w := revokeForm(app, url.Values{
			"token":           {deadAuthToken()},
			"token_type_hint": {"access_token"},
		}, "")

		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// The budget spent is the authentication one, which is a separate bucket from the
		// interactive sign-in budget. Both carry the same burst in this harness, so only
		// this assertion tells them apart.
		assert.False(t, limiter.Login.Reject(testIP), "revoke must not charge the login budget")
	})
	t.Run("MalformedTokensAreAlsoCharged", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthRevoke(router)
		installAuthBudget(t, 3)

		// A token that fails validation is refused earlier, at a different charge site.
		for i := range 3 {
			w := revokeForm(app, url.Values{
				"token":           {"550b1b7b5b9a8e5e0d8b5c0e3a9f6d2c1b4a7e8f"},
				"token_type_hint": {"access_token"},
			}, "")

			assert.Equal(t, http.StatusUnauthorized, w.Code, "attempt %d", i+1)
		}

		w := revokeForm(app, url.Values{
			"token":           {"550b1b7b5b9a8e5e0d8b5c0e3a9f6d2c1b4a7e8f"},
			"token_type_hint": {"access_token"},
		}, "")

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.False(t, limiter.Login.Reject(testIP), "revoke must not charge the login budget")
	})
	t.Run("ControlTokenEndpointChargesTheSameBudget", func(t *testing.T) {
		// The positive control. The client secret is rejected by form validation, which is
		// the token endpoint's own limiter.Auth charge site, so this measures the same
		// bucket the revoke cases do rather than the interactive sign-in one.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthToken(router)
		installAuthBudget(t, 3)

		token := func() *httptest.ResponseRecorder {
			data := url.Values{
				"grant_type":    {authn.GrantClientCredentials.String()},
				"client_id":     {"cs5cpu17n6gj2qo5"},
				"client_secret": {"not a valid secret!"},
				"scope":         {"metrics"},
			}

			req, _ := http.NewRequest(http.MethodPost, tokenPath, strings.NewReader(data.Encode()))
			req.Header.Set(header.ContentType, header.ContentTypeForm)

			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			return w
		}

		for i := range 3 {
			assert.Equal(t, http.StatusUnauthorized, token().Code, "attempt %d", i+1)
		}

		assert.Equal(t, http.StatusTooManyRequests, token().Code)
		assert.True(t, limiter.Auth.Reject(testIP), "the authentication budget must be the one spent")
	})
	t.Run("SharedBudgetAppliesToBothRoutes", func(t *testing.T) {
		// The budget the revoke endpoint spends is the one the token endpoint enforces,
		// so exhausting it here must throttle there too.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthRevoke(router)
		OAuthToken(router)
		installAuthBudget(t, 3)

		for range 3 {
			w := revokeForm(app, url.Values{
				"token":           {deadAuthToken()},
				"token_type_hint": {"access_token"},
			}, "")

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		}

		assert.True(t, limiter.Auth.Reject(testIP))
		assert.False(t, limiter.Login.Reject(testIP))

		data := url.Values{
			"grant_type":    {authn.GrantClientCredentials.String()},
			"client_id":     {"cs5cpu17n6gj2qo5"},
			"client_secret": {"xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
			"scope":         {"metrics"},
		}

		req, _ := http.NewRequest(http.MethodPost, tokenPath, strings.NewReader(data.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})
	t.Run("OneRejectionCostsOneToken", func(t *testing.T) {
		// A request carrying a dead header token and a dead form token fails two lookups.
		// api.Session charges the budget for the first one internally, so the handler must
		// not charge again: one rejection costs one token.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OAuthRevoke(router)
		installAuthBudget(t, 3)

		for i := range 3 {
			w := revokeForm(app, url.Values{
				"token":           {deadAuthToken()},
				"token_type_hint": {"access_token"},
			}, deadAuthToken())

			assert.Equal(t, http.StatusUnauthorized, w.Code, "attempt %d", i+1)
		}

		// Three requests, three tokens: the budget is spent, not overspent.
		assert.True(t, limiter.Auth.Reject(testIP))
	})
	t.Run("ProvenPossessionIsNotCharged", func(t *testing.T) {
		// A user session token is a valid credential, and revoking it through this endpoint
		// is refused because the endpoint revokes client sessions. That refusal is an
		// authorization outcome, so it does not spend an authentication failure budget.
		// OAuth discovery advertises the endpoint, so a conforming client reaches this
		// shape by following it.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		OAuthRevoke(router)
		installAuthBudget(t, 3)

		for i := range 5 {
			w := revokeForm(app, url.Values{
				"token":           {sessId},
				"token_type_hint": {"access_token"},
			}, "")

			assert.Equal(t, http.StatusForbidden, w.Code, "attempt %d: %s", i+1, w.Body.String())
		}

		assert.False(t, limiter.Auth.Reject(testIP), "a refusal of a valid token must not be charged")
	})
	t.Run("WebUIAppPasswordRevoke", func(t *testing.T) {
		// The Web UI deletes an app password by posting its ref id together with the
		// signed-in user's own auth token. That is the most common revoke in the product,
		// so neither the deletion nor a stale entry that no longer resolves may be charged:
		// a user tidying up a list of app passwords must not throttle themselves out.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		OAuthToken(router)
		OAuthRevoke(router)

		appPassword := url.Values{
			"grant_type":  {authn.GrantPassword.String()},
			"client_name": {"RevokeRateLimitApp"},
			"username":    {"alice"},
			"password":    {"Alice123!"},
			"scope":       {"*"},
		}

		createReq, _ := http.NewRequest(http.MethodPost, tokenPath, strings.NewReader(appPassword.Encode()))
		createReq.Header.Set(header.ContentType, header.ContentTypeForm)
		createReq.Header.Set(header.XAuthToken, sessId)

		createResp := httptest.NewRecorder()
		app.ServeHTTP(createResp, createReq)

		if !assert.Equal(t, http.StatusOK, createResp.Code, "%s", createResp.Body.String()) {
			return
		}

		created, err := entity.FindSession(gjson.Get(createResp.Body.String(), "session_id").String())

		if err != nil {
			t.Fatal(err)
		}

		refID := created.RefID
		assert.True(t, rnd.IsRefID(refID), "the web ui posts the ref id it was given as ID")

		installAuthBudget(t, 3)

		// The ref id is an identifier rather than a secret, so the header is what
		// authorizes this. It resolves and the app password is deleted.
		w := revokeForm(app, url.Values{"token": {refID}}, sessId)
		assert.Equal(t, http.StatusOK, w.Code, "%s", w.Body.String())

		// Deleting the same entry again finds nothing, which a stale list produces. It is
		// refused, but the caller proved possession, so it is not an authentication failure.
		for i := range 4 {
			w = revokeForm(app, url.Values{"token": {refID}}, sessId)
			assert.Equal(t, http.StatusUnauthorized, w.Code, "retry %d: %s", i+1, w.Body.String())
		}

		assert.False(t, limiter.Auth.Reject(testIP), "the web ui revoke flow must not spend the budget")
	})
	t.Run("LegitimateRevokesAreNotCharged", func(t *testing.T) {
		// The behavior to preserve: a caller revoking its own tokens must not throttle
		// itself, so a success costs nothing from the budget.
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		dropIsolatedClientSessions(t)

		OAuthToken(router)
		OAuthRevoke(router)
		installAuthBudget(t, 3)

		for i := range 4 {
			data := url.Values{
				"grant_type":    {"client_credentials"},
				"client_id":     {isolatedClientID},
				"client_secret": {isolatedClientSecret},
				"scope":         {"metrics"},
			}

			createReq, _ := http.NewRequest(http.MethodPost, tokenPath, strings.NewReader(data.Encode()))
			createReq.Header.Set(header.ContentType, header.ContentTypeForm)

			createResp := httptest.NewRecorder()
			app.ServeHTTP(createResp, createReq)

			if !assert.Equal(t, http.StatusOK, createResp.Code, "mint %d: %s", i+1, createResp.Body.String()) {
				return
			}

			authToken := gjson.Get(createResp.Body.String(), "access_token").String()

			// Possession-based revocation: the token is the credential, so this carries
			// no session header and must keep working.
			w := revokeForm(app, url.Values{
				"token":           {authToken},
				"token_type_hint": {"access_token"},
			}, "")

			assert.Equal(t, http.StatusOK, w.Code, "revoke %d: %s", i+1, w.Body.String())
		}

		assert.False(t, limiter.Auth.Reject(testIP))
	})
}

// TestOAuthRevoke_ChargesOnlyUnprovenRequests pins the charging rule itself, independently of
// any status code: only a request that proved possession of nothing is counted.
func TestOAuthRevoke_ChargesOnlyUnprovenRequests(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	dropIsolatedClientSessions(t)

	OAuthRevoke(router)

	post := func(values url.Values, authToken string) int {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/revoke", strings.NewReader(values.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)

		if authToken != "" {
			req.Header.Set(header.XAuthToken, authToken)
		}

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		return w.Code
	}

	// spent reports how many tokens a single request took from a fresh budget.
	spent := func(values url.Values, authToken string) (int, int) {
		installAuthBudget(t, 3)

		before := limiter.Auth.IP(header.UnknownIP).Tokens()
		code := post(values, authToken)
		after := limiter.Auth.IP(header.UnknownIP).Tokens()

		return code, int(before - after + 0.5)
	}

	t.Run("UnresolvedToken", func(t *testing.T) {
		code, cost := spent(url.Values{
			"token":           {rnd.AuthToken()},
			"token_type_hint": {"access_token"},
		}, "")

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, 1, cost)
	})
	t.Run("UnresolvedTokenWithDeadHeader", func(t *testing.T) {
		code, cost := spent(url.Values{
			"token":           {rnd.AuthToken()},
			"token_type_hint": {"access_token"},
		}, rnd.AuthToken())

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, 1, cost, "two failed lookups in one request still cost one token")
	})
	t.Run("MalformedHeaderToken", func(t *testing.T) {
		// Session() charges the budget only for a header token it recognizes as well
		// formed, so a token it rejects outright has to be counted here instead.
		code, cost := spent(url.Values{
			"token":           {rnd.AuthToken()},
			"token_type_hint": {"access_token"},
		}, "not-a-token")

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, 1, cost)
	})
	t.Run("ValidHeaderAndUnresolvedToken", func(t *testing.T) {
		// The header proves possession, so the refusal below it is not an authentication
		// failure and costs nothing.
		sess, err := entity.AddClientSession(isolatedClientID, conf.SessionMaxAge(), "metrics", authn.GrantClientCredentials, nil)

		if err != nil {
			t.Fatal(err)
		}

		code, cost := spent(url.Values{
			"token":           {rnd.AuthToken()},
			"token_type_hint": {"access_token"},
		}, sess.AuthToken())

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, 0, cost)
	})
	t.Run("MalformedToken", func(t *testing.T) {
		code, cost := spent(url.Values{
			"token":           {"tooshort"},
			"token_type_hint": {"access_token"},
		}, "")

		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, 1, cost)
	})
	t.Run("SuccessfulRevoke", func(t *testing.T) {
		sess, err := entity.AddClientSession(isolatedClientID, conf.SessionMaxAge(), "metrics", authn.GrantClientCredentials, nil)

		if err != nil {
			t.Fatal(err)
		}

		code, cost := spent(url.Values{
			"token":           {sess.AuthToken()},
			"token_type_hint": {"access_token"},
		}, "")

		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, 0, cost)
	})
}

// TestOAuthRevoke_StaleHeaderWithValidToken covers the combination where a request carries a
// header that no longer resolves together with a token it does hold. The token proves
// possession, so the request is not an authentication failure and costs nothing - a client that
// attaches a default auth header to every call must not throttle its own revocations.
func TestOAuthRevoke_StaleHeaderWithValidToken(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	dropIsolatedClientSessions(t)

	OAuthToken(router)
	OAuthRevoke(router)
	installAuthBudget(t, 3)

	// One more session than the budget has tokens, so a charge per success would throttle.
	for i := range 4 {
		sess, err := entity.AddClientSession(isolatedClientID, conf.SessionMaxAge(), "metrics", authn.GrantClientCredentials, nil)

		if err != nil {
			t.Fatal(err)
		}

		before := limiter.Auth.IP(header.UnknownIP).Tokens()

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/revoke", strings.NewReader(url.Values{
			"token":           {sess.AuthToken()},
			"token_type_hint": {"access_token"},
		}.Encode()))
		req.Header.Set(header.ContentType, header.ContentTypeForm)
		req.Header.Set(header.XAuthToken, rnd.AuthToken())

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "revoke %d: %s", i+1, w.Body.String())
		assert.Equal(t, 0, int(before-limiter.Auth.IP(header.UnknownIP).Tokens()+0.5), "revoke %d must cost nothing", i+1)
	}

	// The shared budget is untouched, so the routes that enforce it still answer normally.
	assert.False(t, limiter.Auth.Reject(header.UnknownIP))
}
