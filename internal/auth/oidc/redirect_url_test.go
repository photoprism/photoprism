package oidc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedirectURL verifies callback URL generation for root and path-prefixed site URLs.
func TestRedirectURL(t *testing.T) {
	t.Run("RootPath", func(t *testing.T) {
		redirectURL, err := RedirectURL("https://app.localssl.dev/")

		assert.NoError(t, err)
		assert.Equal(t, "https://app.localssl.dev/api/v1/oidc/redirect", redirectURL)
	})
	t.Run("PortalTenantPath", func(t *testing.T) {
		redirectURL, err := RedirectURL("https://app.localssl.dev/i/pro-1/")

		assert.NoError(t, err)
		assert.Equal(t, "https://app.localssl.dev/i/pro-1/api/v1/oidc/redirect", redirectURL)
	})
	t.Run("MissingSiteURL", func(t *testing.T) {
		redirectURL, err := RedirectURL("")

		assert.Error(t, err)
		assert.Empty(t, redirectURL)
	})
}

// TestCookiePath verifies the RP cookie path covers both OIDC endpoints for root
// and path-prefixed site URLs, so the state and PKCE cookies survive to the callback.
func TestCookiePath(t *testing.T) {
	t.Run("RootPath", func(t *testing.T) {
		assert.Equal(t, "/api/v1/oidc", CookiePath("https://app.localssl.dev/"))
	})
	t.Run("PortalTenantPath", func(t *testing.T) {
		cookiePath := CookiePath("https://app.localssl.dev/i/pro-1/")
		assert.Equal(t, "/i/pro-1/api/v1/oidc", cookiePath)

		// The cookie path must be a prefix of both the login leg and the callback so
		// the browser returns the state/PKCE cookies on the redirect back.
		redirectURL, err := RedirectURL("https://app.localssl.dev/i/pro-1/")
		assert.NoError(t, err)
		assert.Contains(t, redirectURL, cookiePath+"/redirect")
		assert.True(t, strings.HasPrefix("/i/pro-1/api/v1/oidc/login", cookiePath+"/"))
	})
	t.Run("NoTrailingSlash", func(t *testing.T) {
		assert.Equal(t, "/i/pro-1/api/v1/oidc", CookiePath("https://app.localssl.dev/i/pro-1"))
	})
}
