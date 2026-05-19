package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

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
