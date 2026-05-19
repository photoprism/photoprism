package oidc

import (
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
