package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestClientTokenInfo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "1h0m0s; up to 5 tokens", clientTokenInfo(&entity.Client{AuthMethod: authn.MethodOAuth2.String(), AuthExpires: 3600, AuthTokens: 5}))
		assert.Equal(t, "1h0m0s", clientTokenInfo(&entity.Client{AuthMethod: authn.MethodOAuth2.String(), AuthExpires: 3600, AuthTokens: -1}))
		assert.Equal(t, "up to 3 tokens", clientTokenInfo(&entity.Client{AuthMethod: authn.MethodOAuth2.String(), AuthTokens: 3}))
	})
	t.Run("AppClient", func(t *testing.T) {
		assert.Equal(t, "", clientTokenInfo(&entity.Client{AuthMethod: authn.MethodPKCE.String(), AuthExpires: 86400, AuthTokens: 10}))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", clientTokenInfo(nil))
	})
}

func TestClientLastActive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "2026-10-23 08:28:49", clientLastActive(&entity.Client{LastActive: 1792744129}))
	})
	t.Run("NeverActive", func(t *testing.T) {
		assert.Equal(t, "", clientLastActive(&entity.Client{}))
		assert.Equal(t, "", clientLastActive(nil))
	})
}
