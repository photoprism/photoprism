package oidc

import (
	"net/url"
	"testing"

	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("Prod", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"https://app.localssl.dev/",
			false,
		)

		assert.Error(t, err)
		assert.Nil(t, client)
	})
	t.Run("Debug", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"https://app.localssl.dev/",
			true,
		)

		assert.NoError(t, err)
		assert.IsType(t, &Client{}, client)
	})
	t.Run("EmptyScopes", func(t *testing.T) {
		uri, err := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			"",
			"https://app.localssl.dev/",
			true,
		)

		assert.NoError(t, err)
		assert.IsType(t, &Client{}, client)
	})
	t.Run("IssuerUriMissing", func(t *testing.T) {
		client, err := NewClient(
			nil,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"https://app.localssl.dev/",
			true,
		)

		assert.Error(t, err)
		assert.Nil(t, client)
	})
	t.Run("EmptyRedirectUrl", func(t *testing.T) {
		uri, parseErr := url.Parse("http://dummy-oidc:9998")

		assert.NoError(t, parseErr)

		client, _ := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"",
			true,
		)

		assert.Nil(t, client)
	})
	t.Run("ServiceDiscoveryFails", func(t *testing.T) {
		uri, err := url.Parse("https://dummy-oidc:9998")

		assert.NoError(t, err)

		client, err := NewClient(
			uri,
			"csg6yqvykh0780f9",
			"nd09wkee0ElsMvzLGkgWS9wJAttHwF2h",
			authn.OidcDefaultScopes,
			"https://app.localssl.dev/",
			true,
		)

		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
