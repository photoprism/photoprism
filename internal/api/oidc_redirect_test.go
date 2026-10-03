package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestOidcReconcileHint(t *testing.T) {
	t.Run("LocalProvider", func(t *testing.T) {
		hint := oidcReconcileHint("alice", "local", "us12345")
		assert.Contains(t, hint, "account 'alice'")
		assert.Contains(t, hint, "'local' authentication")
		assert.Contains(t, hint, "photoprism users mod alice --auth oidc --auth-id us12345")
		assert.Contains(t, hint, "sign in locally")
	})
	t.Run("EmptyProviderDefaults", func(t *testing.T) {
		hint := oidcReconcileHint("admin", "", "us67890")
		assert.Contains(t, hint, "'default' authentication")
		assert.Contains(t, hint, "photoprism users mod admin --auth oidc --auth-id us67890")
	})
}

func TestOIDCRedirect(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		OIDCRedirect(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oidc/redirect")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	t.Run("OIDCNotEnabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		OIDCRedirect(router)

		r := AuthenticatedRequest(app, "GET", "/api/v1/oidc/redirect", "xxx")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	t.Run("AuthCodeRequired", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		require.True(t, conf.OIDCEnabled())
		app := newOidcTestApp(conf)

		for _, query := range []string{"", "?state=s1", "?code=c1"} {
			r := PerformRequest(app, http.MethodGet, "/api/v1/oidc/redirect"+query)
			assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
			assert.Equal(t, conf.LoginUri(), r.Header().Get("Location"))
		}
	})
	t.Run("AuthTemplatePreservesSessionStoragePreference", func(t *testing.T) {
		app := gin.New()
		conf := config.TestConfig()
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		app.GET("/oidc-auth-template", func(c *gin.Context) {
			c.HTML(http.StatusOK, "auth.gohtml", gin.H{
				"status":       StatusSuccess,
				"session_id":   "sess1example",
				"access_token": "token1example",
				"provider":     "oidc",
				"user": gin.H{
					"ID":          1,
					"Name":        "alice",
					"DisplayName": "Alice",
				},
				"config": conf.ClientPublic(),
			})
		})

		r := PerformRequest(app, http.MethodGet, "/oidc-auth-template")
		require.Equal(t, http.StatusOK, r.Code)

		body := r.Body.String()
		sessionDataKeys := extractTemplateList(t, body, "const sessionDataKeys = [", "];")
		assert.Contains(t, sessionDataKeys, `"session.token"`)
		assert.NotContains(t, sessionDataKeys, `"session"`)
		assert.Contains(t, body, `localStorage.getItem(namespacedKey("session")) === "true"`)
	})
	t.Run("ClearsEveryKeyTheAppAdopts", func(t *testing.T) {
		app := gin.New()
		conf := config.TestConfig()
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		app.GET("/oidc-auth-template", func(c *gin.Context) {
			c.HTML(http.StatusOK, "auth.gohtml", gin.H{
				"status":       StatusSuccess,
				"session_id":   "sess1example",
				"access_token": "token1example",
				"config":       conf.ClientPublic(),
			})
		})

		r := PerformRequest(app, http.MethodGet, "/oidc-auth-template")
		require.Equal(t, http.StatusOK, r.Code)

		cleared := extractTemplateList(t, r.Body.String(), "const sessionDataKeys = [", "];")

		// Read the names the frontend still adopts from storage, so the two cannot drift apart.
		for _, key := range adoptedStorageKeys(t) {
			assert.Contains(t, cleared, `"`+key+`"`, "the callback must clear every key the app adopts")
		}
	})
}

// adoptedStorageKeys returns the unprefixed storage keys session.js migrates onto the namespaced
// ones, read from the source so this test fails when a key is added there and not to the template.
// Adoption is a read, so the reads are what it derives from.
func adoptedStorageKeys(t *testing.T) []string {
	t.Helper()

	src, err := os.ReadFile("../../frontend/src/common/session.js")
	require.NoError(t, err)

	matches := regexp.MustCompile(`storage\.getItem\("([A-Za-z0-9_.]+)"\)`).FindAllStringSubmatch(string(src), -1)
	require.NotEmpty(t, matches, "no adopted storage keys found")

	seen := make(map[string]bool, len(matches))
	keys := make([]string, 0, len(matches))

	for _, m := range matches {
		// The namespaced names are built from storageKey at run time, so a literal starting with
		// "session." is the modern spelling and not one the migration adopts.
		if strings.HasPrefix(m[1], "session.") || seen[m[1]] {
			continue
		}

		seen[m[1]] = true
		keys = append(keys, m[1])
	}

	return keys
}

func TestOIDCRedirectErrorMessage(t *testing.T) {
	assert.Equal(t, i18n.ErrForbidden, oidcRedirectErrorMessage("access_denied"))
	assert.Equal(t, i18n.ErrUnauthorized, oidcRedirectErrorMessage("login_required"))
	assert.Equal(t, i18n.ErrUnexpected, oidcRedirectErrorMessage("server_error"))
	assert.Equal(t, i18n.ErrUnexpected, oidcRedirectErrorMessage("temporarily_unavailable"))
	assert.Equal(t, i18n.ErrInvalidCredentials, oidcRedirectErrorMessage(""))
}

// TestOIDCRedirect_ProviderError covers the RP callback receiving an OAuth error
// (no code) from the OP — it must render the instance's branded error page, not
// silently bounce to the login form.
func TestOIDCRedirect_ProviderError(t *testing.T) {
	_, _, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().OIDCUri = "https://dummy-oidc.example.com/"
	conf.Options().SiteUrl = "https://app.localssl.dev/"
	conf.Options().OIDCClient = "photoprism-develop"
	conf.Options().OIDCSecret = "9d8351a0-ca01-4556-9c37-85eb634869b9"
	t.Cleanup(func() {
		conf.SetAuthMode(config.AuthModePublic)
		conf.Options().OIDCUri = ""
		conf.Options().OIDCClient = ""
		conf.Options().OIDCSecret = ""
	})
	require.True(t, conf.OIDCEnabled())

	app := gin.New()
	app.LoadHTMLFiles(conf.TemplateFiles()...)
	router := app.Group("/api/v1")
	OIDCRedirect(router)

	r := PerformRequest(app, http.MethodGet, "/api/v1/oidc/redirect?error=access_denied&error_description=no+access+to+the+requested+instance&state=s1")
	require.Equal(t, http.StatusUnauthorized, r.Code, "an OAuth error redirect must render the branded error page; body=%s", r.Body.String())
	body := r.Body.String()
	// auth.gohtml renders the failed-status branch and stores the branded message.
	assert.Contains(t, body, `setItem("session.error"`)
	assert.Contains(t, body, i18n.Error(i18n.ErrForbidden).Error())
	// The page also carries the message key (messageId) so the Web UI can render
	// it in the current UI locale; notify.vue applies Tp to the source string.
	assert.Contains(t, body, `setItem("session.messageId"`)
	assert.Contains(t, body, i18n.Source(i18n.ErrForbidden))
}

// extractTemplateList returns the template source between the provided markers.
func extractTemplateList(t *testing.T, body string, start string, end string) string {
	t.Helper()

	startIndex := strings.Index(body, start)
	require.NotEqual(t, -1, startIndex)

	listStart := startIndex + len(start)
	endIndex := strings.Index(body[listStart:], end)
	require.NotEqual(t, -1, endIndex)

	return body[listStart : listStart+endIndex]
}

const (
	// oidcTestIssuer is the dummy-oidc address that Traefik serves over https.
	oidcTestIssuer = "https://dummy-oidc.localssl.dev"
	// oidcTestRedirectUri is the callback registered for the test client in dummy-oidc.
	oidcTestRedirectUri = "https://app.localssl.dev/api/v1/oidc/redirect"
)

// oidcTestSubjects lists the subjects of the dummy-oidc test identities, so cleanup finds the
// accounts they create even when registration renamed them.
var oidcTestSubjects = []string{"sub00000001", "sub00000002", "sub00000003", "sub00000004", "sub00000005", "sub00000006", "sub00000007", "sub00000008", "sub00000009"}

// oidcUnreachableUri is an https provider address that refuses connections, so discovery fails.
const oidcUnreachableUri = "https://127.0.0.1:1"

// useOidcTestConfig configures dummy-oidc as the identity provider for the duration of a test,
// and skips the test if the provider does not announce its https issuer.
func useOidcTestConfig(t *testing.T) *config.Config {
	t.Helper()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(oidcTestIssuer + "/.well-known/openid-configuration")

	if err != nil {
		t.Skipf("oidc: %s is not reachable, start the dummy-oidc service to run this test", oidcTestIssuer)
	}

	var discovery struct {
		Issuer string `json:"issuer"`
	}

	decodeErr := json.NewDecoder(resp.Body).Decode(&discovery)
	_ = resp.Body.Close()

	if decodeErr != nil || discovery.Issuer != oidcTestIssuer {
		t.Skipf("oidc: %s announces the issuer %q, restart dummy-oidc with image 260930 or later", oidcTestIssuer, discovery.Issuer)
	}

	conf := useOidcOptions(t, oidcTestIssuer)

	deleteOidcTestUsers(t)
	t.Cleanup(func() { deleteOidcTestUsers(t) })

	return conf
}

// useOidcOptions enables OIDC with the specified provider URI for the duration of a test.
func useOidcOptions(t *testing.T, providerUri string) *config.Config {
	t.Helper()

	conf := get.Config()
	opt := conf.Options()
	orig := *opt
	origMode := conf.AuthMode()

	t.Cleanup(func() {
		opt.SiteUrl = orig.SiteUrl
		opt.OIDCUri = orig.OIDCUri
		opt.OIDCClient = orig.OIDCClient
		opt.OIDCSecret = orig.OIDCSecret
		opt.OIDCRegister = orig.OIDCRegister
		opt.OIDCUsername = orig.OIDCUsername
		opt.OIDCGroup = orig.OIDCGroup
		opt.OIDCGroupRole = orig.OIDCGroupRole
		opt.OIDCDomain = orig.OIDCDomain
		opt.OIDCRole = orig.OIDCRole
		opt.OIDCWebDAV = orig.OIDCWebDAV
		opt.UsersQuota = orig.UsersQuota
		conf.SetAuthMode(origMode)
		config.FlushUsageCache()
		get.SetConfig(conf)
	})

	conf.SetAuthMode(config.AuthModePasswd)
	opt.SiteUrl = "https://app.localssl.dev/"
	opt.OIDCUri = providerUri
	opt.OIDCClient = "photoprism-develop"
	opt.OIDCSecret = "9d8351a0-ca01-4556-9c37-85eb634869b9"
	opt.OIDCRegister = true
	opt.OIDCUsername = ""
	opt.OIDCGroup = nil
	opt.OIDCGroupRole = nil
	opt.OIDCDomain = ""
	opt.OIDCRole = ""
	opt.OIDCWebDAV = false

	// Clear the cached client, so it is created for the options above.
	get.SetConfig(conf)

	return conf
}

// deleteOidcTestUsers permanently deletes the accounts created by dummy-oidc logins, including
// their details, settings and sessions.
func deleteOidcTestUsers(t *testing.T) {
	var users entity.Users

	db := entity.UnscopedDb()
	require.NoError(t, db.Where("auth_provider = ? AND auth_issuer = ? AND auth_id IN (?)", authn.ProviderOIDC.String(), oidcTestIssuer, oidcTestSubjects).Find(&users).Error)

	for _, user := range users {
		_, isFixture := entity.UserFixtures[user.UserName]
		require.False(t, isFixture, "refusing to delete the fixture account %s", user.UserName)

		assert.NoError(t, db.Delete(&entity.Session{}, "user_uid = ?", user.UserUID).Error)
		assert.NoError(t, db.Delete(&entity.UserDetails{}, "user_uid = ?", user.UserUID).Error)
		assert.NoError(t, db.Delete(&entity.UserSettings{}, "user_uid = ?", user.UserUID).Error)
		assert.NoError(t, db.Delete(&entity.User{}, "user_uid = ?", user.UserUID).Error)
		entity.FlushUserSessionCache(user.UserUID)
	}

	config.FlushUsageCache()
}

// findOidcTestUser returns the account a dummy-oidc login created for the subject, or nil.
func findOidcTestUser(subject string) *entity.User {
	return entity.FindUser(entity.User{AuthProvider: authn.ProviderOIDC.String(), AuthIssuer: oidcTestIssuer, AuthID: subject})
}

// oidcSessionCount returns the number of sessions of the specified user.
func oidcSessionCount(t *testing.T, user *entity.User) (count int) {
	require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("user_uid = ?", user.UserUID).Count(&count).Error)
	return count
}

// newOidcTestApp returns a router with the OIDC login and redirect endpoints.
func newOidcTestApp(conf *config.Config) *gin.Engine {
	app := gin.New()
	app.LoadHTMLFiles(conf.TemplateFiles()...)
	router := app.Group("/api/v1")
	OIDCLogin(router)
	OIDCRedirect(router)

	return app
}

// oidcLogin signs in through dummy-oidc as the identity selected by loginHint, following the
// provider's redirects, and returns the response of the redirect endpoint.
func oidcLogin(t *testing.T, app *gin.Engine, loginHint string) *httptest.ResponseRecorder {
	t.Helper()

	login := PerformRequest(app, http.MethodGet, "/api/v1/oidc/login")
	require.Equal(t, http.StatusFound, login.Code)

	cookies := login.Result().Cookies()
	require.NotEmpty(t, cookies)

	authUrl, err := url.Parse(login.Header().Get("Location"))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(authUrl.String(), oidcTestIssuer+"/authorize?"), authUrl.String())

	if loginHint != "" {
		q := authUrl.Query()
		q.Set("login_hint", loginHint)
		authUrl.RawQuery = q.Encode()
	}

	client := &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	var callback *url.URL

	for next := authUrl; callback == nil; {
		resp, reqErr := client.Get(next.String())
		require.NoError(t, reqErr)
		_ = resp.Body.Close()
		require.Contains(t, []int{http.StatusFound, http.StatusSeeOther}, resp.StatusCode, next.String())

		next, err = resp.Location()
		require.NoError(t, err)

		if strings.HasPrefix(next.String(), oidcTestRedirectUri+"?") {
			callback = next
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oidc/redirect?"+callback.RawQuery, nil)

	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	return w
}

// assertOidcLoginRefused checks that the redirect endpoint rendered the error page with the message.
func assertOidcLoginRefused(t *testing.T, r *httptest.ResponseRecorder, message i18n.Message) {
	t.Helper()
	assert.Equal(t, http.StatusUnauthorized, r.Code)
	assert.Contains(t, r.Body.String(), i18n.Error(message).Error())
	assert.NotContains(t, r.Body.String(), `setItem("session.token"`)
}

func TestOIDCRedirect_Flow(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Contains(t, r.Body.String(), `setItem("session.token"`)

		user := findOidcTestUser("sub00000001")
		require.NotNil(t, user)
		assert.Equal(t, "prefname", user.UserName)
		assert.Equal(t, authn.ProviderOIDC.String(), user.AuthProvider)
		assert.Equal(t, oidcTestIssuer, user.AuthIssuer)
		assert.Equal(t, "guest", user.UserRole)
		assert.Equal(t, "test@example.com", user.UserEmail)
		assert.True(t, user.EmailVerified())
		assert.Equal(t, "Test", user.DisplayName)
		assert.Equal(t, "testnick", user.Details().NickName)
		assert.True(t, user.CanLogin)
		assert.False(t, user.WebDAV)
		assert.Equal(t, 1, oidcSessionCount(t, user))
	})
	t.Run("RegisterWithWebDAV", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCWebDAV = true
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.True(t, findOidcTestUser("sub00000001").WebDAV)
	})
	t.Run("RegistrationDisabled", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCRegister = false
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "")

		assertOidcLoginRefused(t, r, i18n.ErrRegistrationDisabled)
		assert.Nil(t, findOidcTestUser("sub00000001"))
	})
	t.Run("UsersQuota", func(t *testing.T) {
		// The quota applies to the default role, and a guest is never counted.
		conf := useOidcTestConfig(t)
		conf.Options().UsersQuota = 1
		conf.Options().OIDCRole = "admin"
		config.FlushUsageCache()
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "")

		assertOidcLoginRefused(t, r, i18n.ErrQuotaExceeded)
		assert.Nil(t, findOidcTestUser("sub00000001"))

		conf.Options().OIDCRole = ""

		r = oidcLogin(t, app, "")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.NotNil(t, findOidcTestUser("sub00000001"))
	})
	t.Run("ExistingUser", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		require.Equal(t, http.StatusOK, oidcLogin(t, app, "").Code)
		created := findOidcTestUser("sub00000001")
		require.NotNil(t, created)
		require.NoError(t, created.Updates(entity.Values{"user_email": "local@example.com", "display_name": "Local Name"}))

		r := oidcLogin(t, app, "")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		user := findOidcTestUser("sub00000001")
		require.NotNil(t, user)
		assert.Equal(t, created.UserUID, user.UserUID)
		assert.Equal(t, "prefname", user.UserName)
		assert.Equal(t, "test@example.com", user.UserEmail)
		assert.Equal(t, "Test", user.DisplayName)
		assert.Equal(t, 2, oidcSessionCount(t, user))
	})
	t.Run("ExistingUserDisabled", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		require.Equal(t, http.StatusOK, oidcLogin(t, app, "").Code)
		created := findOidcTestUser("sub00000001")
		require.NotNil(t, created)
		require.NoError(t, created.Updates(entity.Values{"can_login": false, "display_name": "Local Name"}))

		r := oidcLogin(t, app, "")

		// The account is refused before the claims update its profile.
		assertOidcLoginRefused(t, r, i18n.ErrInvalidCredentials)
		assert.Equal(t, 1, oidcSessionCount(t, created))
		assert.Equal(t, "Local Name", findOidcTestUser("sub00000001").DisplayName)
	})
	t.Run("MissingState", func(t *testing.T) {
		// Without the cookies set by the login endpoint, the code exchange fails.
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oidc/redirect?state=s1&code=c1")

		assertOidcLoginRefused(t, r, i18n.ErrUnexpected)
	})
}

func TestOIDCRedirect_Groups(t *testing.T) {
	t.Run("GroupRole", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCGroupRole = []string{"photoprism-admin=admin"}
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "olivia")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		user := findOidcTestUser("sub00000002")
		require.NotNil(t, user)
		assert.Equal(t, "olivia", user.UserName)
		assert.Equal(t, "admin", user.UserRole)
	})
	t.Run("GroupRoleUpdate", func(t *testing.T) {
		// A group mapping added later updates the role of an existing account.
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		require.Equal(t, http.StatusOK, oidcLogin(t, app, "olivia").Code)
		require.Equal(t, "guest", findOidcTestUser("sub00000002").UserRole)

		conf.Options().OIDCGroupRole = []string{"photoprism-admin=admin"}

		r := oidcLogin(t, app, "olivia")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "admin", findOidcTestUser("sub00000002").UserRole)
	})
	t.Run("GroupRoleNotMember", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCGroupRole = []string{"photoprism-admin=admin"}
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "oscar")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "guest", findOidcTestUser("sub00000003").UserRole)
	})
	t.Run("RequiredGroup", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCGroup = []string{"photoprism-user"}
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "oscar")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.NotNil(t, findOidcTestUser("sub00000003"))
	})
	t.Run("RequiredGroupMissing", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCGroup = []string{"photoprism-user"}
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "olivia")

		assertOidcLoginRefused(t, r, i18n.ErrForbidden)
		assert.Nil(t, findOidcTestUser("sub00000002"))
	})
	t.Run("RequiredGroupOverage", func(t *testing.T) {
		// A token that omits the groups cannot satisfy a required group.
		conf := useOidcTestConfig(t)
		conf.Options().OIDCGroup = []string{"photoprism-user"}
		app := newOidcTestApp(conf)

		origAudit := event.AuditLog
		auditLogger, auditHook := logtest.NewNullLogger()
		event.AuditLog = auditLogger
		t.Cleanup(func() { event.AuditLog = origAudit })

		r := oidcLogin(t, app, "odette")

		assertOidcLoginRefused(t, r, i18n.ErrForbidden)
		assert.Nil(t, findOidcTestUser("sub00000006"))

		// The refusal names the omitted groups rather than a missing membership.
		var messages []string
		for _, entry := range auditHook.AllEntries() {
			messages = append(messages, entry.Message)
		}
		assert.Contains(t, strings.Join(messages, "\n"), "IdP omitted some or all groups")
	})
	t.Run("OverageWithoutRequiredGroup", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "odette")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.NotNil(t, findOidcTestUser("sub00000006"))
	})
	t.Run("PortalRoleNotInEdition", func(t *testing.T) {
		// The manager role that petra's Portal claim grants does not exist in this edition, so the
		// default role applies.
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "petra")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "guest", findOidcTestUser("sub00000007").UserRole)
	})
}

func TestOIDCRedirect_Email(t *testing.T) {
	t.Run("Domain", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCDomain = "example.com"
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.NotNil(t, findOidcTestUser("sub00000001"))
	})
	t.Run("DomainMismatch", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCDomain = "example.com"
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "ola")

		assertOidcLoginRefused(t, r, i18n.ErrForbidden)
		assert.Nil(t, findOidcTestUser("sub00000005"))
	})
	t.Run("DomainUnverified", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		conf.Options().OIDCDomain = "example.com"
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "otto")

		assertOidcLoginRefused(t, r, i18n.ErrVerifiedEmailRequired)
		assert.Nil(t, findOidcTestUser("sub00000004"))
	})
	t.Run("Unverified", func(t *testing.T) {
		// Without a domain requirement the account is created, and the email is kept unverified.
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "otto")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		user := findOidcTestUser("sub00000004")
		require.NotNil(t, user)
		assert.Equal(t, "otto@example.com", user.UserEmail)
		assert.False(t, user.EmailVerified())
	})
	t.Run("InUse", func(t *testing.T) {
		// An email another account holds is kept unverified and does not prevent the registration.
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		other := entity.NewUser()
		other.UserName = "zz-oidc-email"
		other.UserEmail = "oscar@example.com"
		other.SetProvider(authn.ProviderLocal)
		require.NoError(t, other.Create())
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", other.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", other.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(other).Error)
		})

		r := oidcLogin(t, app, "oscar")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		user := findOidcTestUser("sub00000003")
		require.NotNil(t, user)
		assert.Equal(t, "oscar@example.com", user.UserEmail)
		assert.False(t, user.EmailVerified())
	})
}

func TestOIDCRedirect_Username(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "nobody")

		assertOidcLoginRefused(t, r, i18n.ErrInvalidCredentials)
		assert.Nil(t, findOidcTestUser("sub00000009"))
	})
	t.Run("LocalAccountName", func(t *testing.T) {
		// An identity is matched by its subject, never by name, so a login whose name belongs to a
		// local account registers a separate account with a numeric suffix.
		conf := useOidcTestConfig(t)
		app := newOidcTestApp(conf)
		local := entity.FindUserByName("alice")
		require.NotNil(t, local)

		r := oidcLogin(t, app, "alice")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		user := findOidcTestUser("sub00000008")
		require.NotNil(t, user)
		assert.NotEqual(t, local.UserUID, user.UserUID)
		assert.Regexp(t, `^alice[0-9]{6}$`, user.UserName)

		alice := entity.FindUserByName("alice")
		require.NotNil(t, alice)
		assert.Equal(t, local.AuthProvider, alice.AuthProvider)
		assert.Equal(t, local.AuthID, alice.AuthID)
	})
}

func TestOIDCRedirect_Unavailable(t *testing.T) {
	t.Run("CdnRequest", func(t *testing.T) {
		conf := useOidcOptions(t, oidcUnreachableUri)
		app := newOidcTestApp(conf)

		for _, uri := range []string{"/api/v1/oidc/login", "/api/v1/oidc/redirect?state=s1&code=c1"} {
			req := httptest.NewRequest(http.MethodGet, uri, nil)
			req.Header.Set(header.CdnHost, "cdn.example.com")
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, uri)
		}
	})
	t.Run("ProviderUnreachable", func(t *testing.T) {
		conf := useOidcOptions(t, oidcUnreachableUri)
		require.True(t, conf.OIDCEnabled())
		app := newOidcTestApp(conf)

		login := PerformRequest(app, http.MethodGet, "/api/v1/oidc/login")

		assert.Equal(t, http.StatusInternalServerError, login.Code)
		assert.Contains(t, login.Body.String(), i18n.Error(i18n.ErrConnectionFailed).Error())
		assert.Empty(t, login.Header().Get("Location"))

		redirect := PerformRequest(app, http.MethodGet, "/api/v1/oidc/redirect?state=s1&code=c1")

		assertOidcLoginRefused(t, redirect, i18n.ErrInvalidCredentials)
	})
	t.Run("TooManyRequests", func(t *testing.T) {
		// Requests that do not complete a login consume the client's budget, and the next one is refused.
		conf := useOidcOptions(t, oidcUnreachableUri)
		app := newOidcTestApp(conf)

		origLogin := limiter.Login
		t.Cleanup(func() { limiter.Login = origLogin })

		cases := []struct {
			uri    string
			status int
		}{
			{"/api/v1/oidc/login", http.StatusInternalServerError},
			{"/api/v1/oidc/redirect", http.StatusTemporaryRedirect},
		}

		for _, c := range cases {
			limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), 3)

			for i := 0; i < 3; i++ {
				assert.Equal(t, c.status, PerformRequest(app, http.MethodGet, c.uri).Code, c.uri)
			}

			r := PerformRequest(app, http.MethodGet, c.uri)
			assert.Equal(t, http.StatusTooManyRequests, r.Code, c.uri)
			assert.Contains(t, r.Body.String(), i18n.Error(i18n.ErrTooManyRequests).Error())
		}
	})
}

func TestOIDCRedirect_Session(t *testing.T) {
	t.Run("DefaultRoleNotFederated", func(t *testing.T) {
		// A default role that federation cannot grant never yields a session.
		conf := useOidcTestConfig(t)
		conf.Options().OIDCRole = "visitor"
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "")

		assertOidcLoginRefused(t, r, i18n.ErrInvalidCredentials)

		if user := findOidcTestUser("sub00000001"); user != nil {
			assert.NotEqual(t, "visitor", user.UserRole)
			assert.Equal(t, 0, oidcSessionCount(t, user))
		}
	})
	t.Run("NoGroupsStored", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		require.False(t, conf.Portal())
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "olivia")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		sess := oidcTestSession(t, "sub00000002")
		assert.Empty(t, sess.GetData().Groups)
	})
	t.Run("PortalGroupsStored", func(t *testing.T) {
		conf := useOidcTestConfig(t)
		origEdition, origRole := conf.Options().Edition, conf.Options().NodeRole
		t.Cleanup(func() { conf.Options().Edition, conf.Options().NodeRole = origEdition, origRole })
		conf.Options().Edition = config.Portal
		require.True(t, conf.Portal())
		app := newOidcTestApp(conf)

		r := oidcLogin(t, app, "olivia")

		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		sess := oidcTestSession(t, "sub00000002")
		assert.ElementsMatch(t, []string{"photoprism-admin", "staff"}, sess.GetData().Groups)
	})
}

// oidcTestSession returns the only session of the account a dummy-oidc login created for the subject.
func oidcTestSession(t *testing.T, subject string) *entity.Session {
	t.Helper()

	user := findOidcTestUser(subject)
	require.NotNil(t, user)

	var sessions entity.Sessions
	require.NoError(t, entity.UnscopedDb().Where("user_uid = ?", user.UserUID).Find(&sessions).Error)
	require.Len(t, sessions, 1)

	return &sessions[0]
}
