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
	t.Run("ClaimWithoutName", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Subject = "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d"
		info.PreferredUsername = "./"
		info.Name = "Jane Doe"
		assert.Equal(t, "jane.doe", Username(info, authn.OidcClaimPreferredUsername))
		info.Name = ".."
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
		info.Email = ".bob@example.com"
		info.EmailVerified = true
		assert.Equal(t, "", Username(info, authn.OidcClaimEmail))
		info.Email = "bob@example.com"
		assert.Equal(t, "bob@example.com", Username(info, authn.OidcClaimEmail))
	})
	t.Run("NotASCII", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Subject = "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d"
		info.PreferredUsername = "José"
		info.Name = "José García"
		info.Nickname = "Pepe"
		info.Email = "jose@example.com"
		assert.Equal(t, "pepe", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "pepe", Username(info, authn.OidcClaimName))
		assert.Equal(t, "pepe", Username(info, authn.OidcClaimNickname))
		assert.Equal(t, "pepe", Username(info, authn.OidcClaimEmail))
		info.Nickname = "Pépé"
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
		info.EmailVerified = true
		assert.Equal(t, "jose@example.com", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "jose@example.com", Username(info, authn.OidcClaimName))
		assert.Equal(t, "jose@example.com", Username(info, authn.OidcClaimNickname))
	})
	t.Run("NotASCIIBeforeSanitizing", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "\u212Aate"
		info.Name = "Kate Doe"
		assert.Equal(t, "kate.doe", Username(info, authn.OidcClaimPreferredUsername))
		info.PreferredUsername = "kate@exämple.com"
		assert.Equal(t, "kate.doe", Username(info, authn.OidcClaimPreferredUsername))
	})
	t.Run("NotASCIIEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "josé"
		info.Email = "\u212Aate@example.com"
		info.EmailVerified = true
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "", Username(info, authn.OidcClaimEmail))
		info.Nickname = "kate"
		assert.Equal(t, "kate", Username(info, authn.OidcClaimEmail))
	})
}

func TestSkipsClaim(t *testing.T) {
	t.Run("Skipped", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "José"
		info.Name = "Jose Garcia"
		assert.True(t, SkipsClaim(info, authn.OidcClaimPreferredUsername))
		assert.False(t, SkipsClaim(info, authn.OidcClaimName))
	})
	t.Run("NothingLeft", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "José"
		assert.True(t, SkipsClaim(info, authn.OidcClaimName))
	})
	t.Run("ASCII", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "jane"
		info.Name = "José García"
		assert.False(t, SkipsClaim(info, authn.OidcClaimPreferredUsername))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, SkipsClaim(&oidc.UserInfo{}, authn.OidcClaimPreferredUsername))
	})
}

func TestAsciiHandle(t *testing.T) {
	assert.Equal(t, "jane.doe", asciiHandle("Jane Doe"))
	assert.Equal(t, "jane", asciiHandle("jane@example.com"))
	assert.Equal(t, "", asciiHandle("José"))
	assert.Equal(t, "", asciiHandle("\u212Aate"))
	assert.Equal(t, "", asciiHandle(""))
}

func TestAsciiEmail(t *testing.T) {
	assert.Equal(t, "jane@example.com", asciiEmail("Jane@Example.com"))
	assert.Equal(t, "", asciiEmail("\u212Aate@example.com"))
	assert.Equal(t, "", asciiEmail("jose@exämple.com"))
	assert.Equal(t, "", asciiEmail(""))
}
