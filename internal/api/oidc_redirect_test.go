package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
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
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		conf.Options().OIDCUri = "http://dummy-oidc:9998"
		conf.Options().SiteUrl = "https://app.localssl.dev/"
		conf.Options().OIDCClient = "photoprism-develop"
		conf.Options().OIDCSecret = "9d8351a0-ca01-4556-9c37-85eb634869b9"

		OIDCRedirect(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/oidc/redirect")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
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
