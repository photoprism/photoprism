package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
)

func TestUsername(t *testing.T) {
	t.Run("ClaimPreferredUsername", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d"
		info.PreferredUsername = "Jane Doe"
		result := Username(info, authn.ClaimPreferredUsername)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("ClaimPreferredUsernameMissing", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		result := Username(info, authn.ClaimPreferredUsername)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("ClaimName", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.ClaimName)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("ClaimNickname", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.ClaimNickname)
		assert.Equal(t, "jens.mander", result)
	})
	t.Run("ClaimEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.ClaimEmail)
		assert.Equal(t, "jane@doe.com", result)
	})
}
