package oidc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestVerifiedEmail(t *testing.T) {
	t.Run("Verified", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Email = "Jane@Example.com"
		info.EmailVerified = true
		assert.Equal(t, "jane@example.com", VerifiedEmail(info))
	})
	t.Run("Unverified", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Email = "jane@example.com"
		assert.Equal(t, "", VerifiedEmail(info))
	})
	t.Run("VerifiedWithoutEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.EmailVerified = true
		assert.Equal(t, "", VerifiedEmail(info))
	})
	t.Run("VerifiedInvalid", func(t *testing.T) {
		for _, email := range []string{"not-an-email", strings.Repeat("a", 250) + "@example.com"} {
			info := &oidc.UserInfo{}
			info.Email = email
			info.EmailVerified = true
			assert.Equal(t, "", VerifiedEmail(info), email)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail(nil))
	})
}
