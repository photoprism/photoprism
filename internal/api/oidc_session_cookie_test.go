package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// findCookie returns the named cookie from the recorder, or nil if absent.
func findCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, ck := range w.Result().Cookies() {
		if ck.Name == name {
			return ck
		}
	}
	return nil
}

func TestOIDCSessionCookiePath(t *testing.T) {
	t.Run("FromConfig", func(t *testing.T) {
		// APIv1 is mounted at conf.BaseUri(config.ApiUri); the cookie must be
		// scoped to that path + "/oauth" so the browser sends it to the moved
		// /api/v1/oauth/authorize endpoint.
		assert.Equal(t, get.Config().BaseUri(config.ApiUri)+"/oauth", OIDCSessionCookiePath(get.Config()))
	})
	t.Run("NilConfigFallsBackToBareApiUri", func(t *testing.T) {
		assert.Equal(t, config.ApiUri+"/oauth", OIDCSessionCookiePath(nil))
	})
}

func TestOIDCSessionCookieClearPath(t *testing.T) {
	// mk builds an isolated instance config (NodeRole defaults to instance) with
	// the given OIDC issuer and SiteUrl so the derivation can be exercised.
	mk := func(issuer, siteURL string) *config.Config {
		c := config.NewConfig(config.CliTestContext())
		c.Options().OIDCUri = issuer
		c.Options().SiteUrl = siteURL
		return c
	}
	t.Run("NilConfig", func(t *testing.T) {
		assert.Equal(t, "", OIDCSessionCookieClearPath(nil))
	})
	t.Run("SharedDomainIssuerAtRoot", func(t *testing.T) {
		// Shared-domain Portal at the origin root: the cookie lives at /api/v1/oauth,
		// not under the instance's /i/<tenant> base.
		c := mk("https://app.localssl.dev/", "https://app.localssl.dev/i/pro-1/")
		assert.Equal(t, config.ApiUri+"/oauth", OIDCSessionCookieClearPath(c))
	})
	t.Run("SharedDomainIssuerWithBasePath", func(t *testing.T) {
		c := mk("https://app.localssl.dev/portal/", "https://app.localssl.dev/i/pro-1/")
		assert.Equal(t, "/portal"+config.ApiUri+"/oauth", OIDCSessionCookieClearPath(c))
	})
	t.Run("SubdomainIsolatedReturnsEmpty", func(t *testing.T) {
		// Issuer on a different host than the instance: the host-only OP cookie is
		// unreachable, so the instance must not attempt to clear it.
		c := mk("https://portal.example.com/", "https://node1.example.com/")
		assert.Equal(t, "", OIDCSessionCookieClearPath(c))
	})
	t.Run("ExternalIdPReturnsEmpty", func(t *testing.T) {
		c := mk("https://keycloak.example.com/realms/main", "https://app.localssl.dev/i/pro-1/")
		assert.Equal(t, "", OIDCSessionCookieClearPath(c))
	})
	t.Run("NoOIDCReturnsEmpty", func(t *testing.T) {
		c := mk("", "https://app.localssl.dev/i/pro-1/")
		assert.Equal(t, "", OIDCSessionCookieClearPath(c))
	})
}

func TestSignParseOIDCSession(t *testing.T) {
	id := rnd.SessionID(rnd.AuthToken())
	t.Run("RoundTrip", func(t *testing.T) {
		v := signOIDCSession(id, time.Now().Add(time.Minute))
		got, ok := parseOIDCSession(v)
		assert.True(t, ok)
		assert.Equal(t, id, got)
	})
	t.Run("Expired", func(t *testing.T) {
		v := signOIDCSession(id, time.Now().Add(-time.Minute))
		_, ok := parseOIDCSession(v)
		assert.False(t, ok)
	})
	t.Run("Tampered", func(t *testing.T) {
		v := signOIDCSession(id, time.Now().Add(time.Minute))
		_, ok := parseOIDCSession(v + "x")
		assert.False(t, ok, "a tampered signature must not verify")
	})
	t.Run("Malformed", func(t *testing.T) {
		for _, in := range []string{"", "no-dot", "a.b", "tooshort.sig"} {
			_, ok := parseOIDCSession(in)
			assert.False(t, ok, "malformed value %q must not verify", in)
		}
	})
	t.Run("NonSessionIDPayloadRejected", func(t *testing.T) {
		// A correctly-signed value whose payload id is not a session id must
		// still be rejected, so the cookie can only ever carry a real session ref.
		v := signOIDCSession("not-a-session-id", time.Now().Add(time.Minute))
		_, ok := parseOIDCSession(v)
		assert.False(t, ok)
	})
}

func TestSetOIDCSessionCookie(t *testing.T) {
	t.Run("SignsSessionReferenceNotToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		sess := &entity.Session{ID: rnd.SessionID(rnd.AuthToken()), UserUID: "uqxetse3cy5eo9z2"}
		SetOIDCSessionCookie(c, sess, "/api/v1/oauth", true)
		ck := findCookie(w, OIDCSessionCookie)
		if assert.NotNil(t, ck) {
			assert.False(t, rnd.IsAuthToken(ck.Value), "the cookie must not store a usable bearer token")
			id, ok := parseOIDCSession(ck.Value)
			assert.True(t, ok)
			assert.Equal(t, sess.ID, id, "the cookie value must resolve back to the session id")
			assert.Equal(t, "/api/v1/oauth", ck.Path)
			assert.True(t, ck.HttpOnly)
			assert.True(t, ck.Secure)
			assert.Equal(t, http.SameSiteLaxMode, ck.SameSite)
			assert.Equal(t, int(OIDCSessionCookieTTL.Seconds()), ck.MaxAge)
		}
	})
	t.Run("EmptyPathFallsBackToBareApiUri", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetOIDCSessionCookie(c, &entity.Session{ID: rnd.SessionID(rnd.AuthToken()), UserUID: "uqxetse3cy5eo9z2"}, "", true)
		ck := findCookie(w, OIDCSessionCookie)
		if assert.NotNil(t, ck) {
			assert.Equal(t, config.ApiUri+"/oauth", ck.Path)
		}
	})
	t.Run("InsecureOmitsSecureFlag", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetOIDCSessionCookie(c, &entity.Session{ID: rnd.SessionID(rnd.AuthToken()), UserUID: "uqxetse3cy5eo9z2"}, "/api/v1/oauth", false)
		ck := findCookie(w, OIDCSessionCookie)
		if assert.NotNil(t, ck) {
			assert.False(t, ck.Secure)
		}
	})
	t.Run("UserlessSessionSetsNothing", func(t *testing.T) {
		// Client/service sessions have no user; the OP cookie would only emit a
		// signal the reader rejects via NoUser(), so none is written.
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetOIDCSessionCookie(c, &entity.Session{ID: rnd.SessionID(rnd.AuthToken())}, "/api/v1/oauth", true)
		assert.Nil(t, findCookie(w, OIDCSessionCookie))
	})
	t.Run("InvalidSessionSetsNothing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetOIDCSessionCookie(c, &entity.Session{ID: "tooshort"}, "/api/v1/oauth", true)
		assert.Nil(t, findCookie(w, OIDCSessionCookie))
	})
	t.Run("NilSessionSetsNothing", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetOIDCSessionCookie(c, nil, "/api/v1/oauth", true)
		assert.Nil(t, findCookie(w, OIDCSessionCookie))
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SetOIDCSessionCookie(nil, &entity.Session{ID: rnd.SessionID(rnd.AuthToken())}, "/api/v1/oauth", true)
		})
	})
}

func TestClearOIDCSessionCookie(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		ClearOIDCSessionCookie(c, "/api/v1/oauth", false)
		ck := findCookie(w, OIDCSessionCookie)
		if assert.NotNil(t, ck) {
			assert.Equal(t, "", ck.Value)
			assert.Equal(t, "/api/v1/oauth", ck.Path)
			assert.True(t, ck.MaxAge < 0)
		}
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.NotPanics(t, func() { ClearOIDCSessionCookie(nil, "/api/v1/oauth", false) })
	})
}

func TestOIDCSessionCookieSession(t *testing.T) {
	newCtx := func(ck *http.Cookie) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		if ck != nil {
			req.AddCookie(ck)
		}
		c.Request = req
		return c
	}
	t.Run("AbsentReturnsNil", func(t *testing.T) {
		assert.Nil(t, OIDCSessionCookieSession(newCtx(nil)))
	})
	t.Run("MalformedReturnsNil", func(t *testing.T) {
		c := newCtx(&http.Cookie{Name: OIDCSessionCookie, Value: "garbage"}) //nolint:gosec // test builds a request cookie; transport attributes are irrelevant
		assert.Nil(t, OIDCSessionCookieSession(c))
	})
	t.Run("UnknownSessionReturnsNil", func(t *testing.T) {
		// A correctly-signed reference to a session that does not exist resolves
		// to nil (and the stale cookie is cleared).
		v := signOIDCSession(rnd.SessionID(rnd.AuthToken()), time.Now().Add(time.Minute))
		c := newCtx(&http.Cookie{Name: OIDCSessionCookie, Value: v}) //nolint:gosec // test builds a request cookie; transport attributes are irrelevant
		assert.Nil(t, OIDCSessionCookieSession(c))
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.Nil(t, OIDCSessionCookieSession(nil))
	})
}

func TestLoadOrCreateOIDCSessionKey(t *testing.T) {
	// Use a DB-backed isolated config and restore the global afterwards, matching
	// newPortalJWTFixture; a DB-less config would leave the global entity DB unusable
	// for later tests in this package.
	withTempConfig := func(t *testing.T, suffix string) *config.Config {
		conf := config.NewMinimalTestConfigWithDb("oidc-session-key-"+suffix, t.TempDir())
		orig := get.Config()
		get.SetConfig(conf)
		t.Cleanup(func() { get.SetConfig(orig) })
		return conf
	}
	t.Run("PersistsAndReloads", func(t *testing.T) {
		conf := withTempConfig(t, "persist")

		k1 := loadOrCreateOIDCSessionKey()
		require.Len(t, k1, oidcSessionKeyLen)

		// A second call returns the same persisted key, kept at 0600.
		k2 := loadOrCreateOIDCSessionKey()
		assert.Equal(t, k1, k2, "the key must be stable across calls")

		path := filepath.Join(conf.PortalConfigPath(), "keys", oidcSessionKeyFile)
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})
	t.Run("UnwritableKeysDirReturnsNil", func(t *testing.T) {
		conf := withTempConfig(t, "unwritable")

		// Place a regular file where the keys directory must be created so the
		// MkdirAll inside loadOrCreateOIDCSessionKey fails and it returns nil.
		require.NoError(t, fs.MkdirAll(conf.PortalConfigPath()))
		require.NoError(t, os.WriteFile(filepath.Join(conf.PortalConfigPath(), "keys"), []byte("x"), fs.ModeSecretFile))
		assert.Nil(t, loadOrCreateOIDCSessionKey())
	})
}

func TestOIDCSessionCookieAdmission(t *testing.T) {
	conf := get.Config()
	user := entity.FindLocalUser("alice")
	require.NotNil(t, user)

	setCookie := func(sess *entity.Session) *http.Cookie {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		SetOIDCSessionCookie(c, sess, "/api/v1/oauth", false)
		return findCookie(w, OIDCSessionCookie)
	}

	t.Run("AppPasswordGetsNoCookie", func(t *testing.T) {
		sess, err := entity.AddClientSession("cookie-app-pw", conf.SessionMaxAge(), "photos:read", authn.GrantPassword, user)
		require.NoError(t, err)
		assert.Nil(t, setCookie(sess))
	})
	t.Run("ClientCredentialsGetNoCookie", func(t *testing.T) {
		sess, err := entity.AddClientSession("cookie-client", conf.SessionMaxAge(), "cluster", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		assert.Nil(t, setCookie(sess))
	})
	t.Run("CookieSignedForAnAppPasswordIsRejected", func(t *testing.T) {
		// A signed, unexpired cookie whose session the OP does not admit: the reader
		// must reject it rather than trust the signature alone.
		sess, err := entity.AddClientSession("cookie-legacy-app", conf.SessionMaxAge(), "*", authn.GrantPassword, user)
		require.NoError(t, err)

		value := signOIDCSession(sess.ID, time.Now().Add(OIDCSessionCookieTTL))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		c.Request.AddCookie(&http.Cookie{Name: OIDCSessionCookie, Value: value}) //nolint:gosec // test request cookie

		assert.Nil(t, OIDCSessionCookieSession(c))

		cleared := findCookie(w, OIDCSessionCookie)
		require.NotNil(t, cleared, "an unresolvable cookie must be cleared")
		assert.Equal(t, "", cleared.Value)
	})
	t.Run("IneligibleAfterIssueIsRejectedAndCleared", func(t *testing.T) {
		sess := entity.NewSession(conf.SessionMaxAge(), 0)
		sess.SetProvider(authn.ProviderLocal)
		sess.SetMethod(authn.MethodDefault)
		sess.SetUser(user)
		require.NoError(t, sess.Create())

		ck := setCookie(sess)
		require.NotNil(t, ck, "an interactive session must receive a cookie")

		// Withdraw web access after the cookie was written, restoring what was there.
		canLogin, superAdmin := user.CanLogin, user.SuperAdmin

		require.NoError(t, entity.Db().Model(&entity.User{}).Where("user_uid = ?", user.UserUID).
			Updates(entity.Values{"can_login": false, "super_admin": false}).Error)
		t.Cleanup(func() {
			require.NoError(t, entity.Db().Model(&entity.User{}).Where("user_uid = ?", user.UserUID).
				Updates(entity.Values{"can_login": canLogin, "super_admin": superAdmin}).Error)
			entity.FlushSessionCache()
		})
		entity.FlushSessionCache()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize", nil)
		c.Request.AddCookie(&http.Cookie{Name: OIDCSessionCookie, Value: ck.Value}) //nolint:gosec // test request cookie

		assert.Nil(t, OIDCSessionCookieSession(c))

		cleared := findCookie(w, OIDCSessionCookie)
		require.NotNil(t, cleared, "an unresolvable cookie must be cleared")
		assert.Equal(t, "", cleared.Value)
	})
}

func TestOIDCSessionCookieEndSession(t *testing.T) {
	conf := get.Config()
	user := entity.FindLocalUser("alice")
	require.NotNil(t, user)

	request := func(value string) (*httptest.ResponseRecorder, *gin.Context) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/oauth/logout", nil)
		if value != "" {
			c.Request.AddCookie(&http.Cookie{Name: OIDCSessionCookie, Value: value}) //nolint:gosec // test request cookie
		}
		return w, c
	}

	t.Run("NoCookie", func(t *testing.T) {
		_, c := request("")
		assert.Nil(t, OIDCSessionCookieEndSession(c))
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.Nil(t, OIDCSessionCookieEndSession(nil))
	})
	t.Run("Tampered", func(t *testing.T) {
		w, c := request("not-a-signed-value")
		assert.Nil(t, OIDCSessionCookieEndSession(c))

		cleared := findCookie(w, OIDCSessionCookie)
		require.NotNil(t, cleared)
		assert.Equal(t, "", cleared.Value)
	})
	t.Run("ResolvesASessionAdmissionDeclines", func(t *testing.T) {
		// Ending a session must not be narrower than starting one, so this reader
		// returns what the admission-gated reader refuses.
		sess, err := entity.AddClientSession("end-session-app-pw", conf.SessionMaxAge(), "*", authn.GrantPassword, user)
		require.NoError(t, err)

		value := signOIDCSession(sess.ID, time.Now().Add(OIDCSessionCookieTTL))

		_, c := request(value)
		found := OIDCSessionCookieEndSession(c)
		require.NotNil(t, found)
		assert.Equal(t, sess.ID, found.ID)

		_, gated := request(value)
		assert.Nil(t, OIDCSessionCookieSession(gated))
	})
}
