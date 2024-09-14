package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
)

func TestUsername(t *testing.T) {
	t.Run("PreferredUsername", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d"
		info.PreferredUsername = "Jane Doe"
		result := Username(info, authn.OidcClaimPreferredUsername)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("PreferredUsernameFallbackName", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		result := Username(info, authn.OidcClaimPreferredUsername)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("PreferredUsernameFallbackNickname", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		result := Username(info, authn.OidcClaimPreferredUsername)
		assert.Equal(t, "jens.mander", result)
	})
	t.Run("PreferredUsernameFallbackEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		result := Username(info, authn.OidcClaimPreferredUsername)
		assert.Equal(t, "jane@doe.com", result)
	})
	t.Run("Name", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimName)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("NameFallbackPreferredUsername", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = ""
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		info.PreferredUsername = "Jane Doe"
		result := Username(info, authn.OidcClaimName)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("NameFallbackNickname", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = ""
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimName)
		assert.Equal(t, "jens.mander", result)
	})
	t.Run("NameFallbackEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = ""
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimName)
		assert.Equal(t, "jane@doe.com", result)
	})
	t.Run("Nickname", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimNickname)
		assert.Equal(t, "jens.mander", result)
	})
	t.Run("NicknameFallbackPreferredUsername", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = ""
		info.PreferredUsername = "Jane Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimNickname)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("NicknameFallbackName", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = ""
		info.PreferredUsername = ""
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimNickname)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("NicknameFallbackEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = ""
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = ""
		info.PreferredUsername = ""
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimNickname)
		assert.Equal(t, "jane@doe.com", result)
	})
	t.Run("Email", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimEmail)
		assert.Equal(t, "jane@doe.com", result)
	})
	t.Run("EmailFallbackPreferredUsername", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = ""
		info.PreferredUsername = "Jane Doe"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimEmail)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("EmailFallbackName", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = ""
		info.PreferredUsername = ""
		info.Nickname = "Jens Mander"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimEmail)
		assert.Equal(t, "jane.doe", result)
	})
	t.Run("EmailFallbackNickname", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = ""
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = ""
		info.PreferredUsername = ""
		info.Nickname = "Jens Mander"
		info.EmailVerified = true
		info.Subject = "abcd123"
		result := Username(info, authn.OidcClaimEmail)
		assert.Equal(t, "jens.mander", result)
	})
}
