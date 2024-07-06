package oidc

import (
	"net/url"
	"testing"

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
			"openid email profile",
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
			"openid email profile",
			"https://app.localssl.dev/",
			true,
		)

		assert.NoError(t, err)
		assert.IsType(t, &Client{}, client)
	})
}
