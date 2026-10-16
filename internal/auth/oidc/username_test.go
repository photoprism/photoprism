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
	t.Run("NotPortableSkipped", func(t *testing.T) {
		// Claims that would give a username with "%" or a hidden handle fall back to the next claim.
		for _, preferred := range []string{"50%off", "__svc", "_.svc"} {
			info := &oidc.UserInfo{}
			info.Name = "Jane Doe"
			info.Email = "jane@doe.com"
			info.EmailVerified = true
			info.PreferredUsername = preferred
			assert.Equal(t, "jane.doe", Username(info, authn.OidcClaimPreferredUsername), preferred)
			assert.False(t, SkipsClaim(info, authn.OidcClaimPreferredUsername), preferred)
		}
	})
	t.Run("CollapsedDots", func(t *testing.T) {
		// Display names whose handles would have ".." get one dot instead.
		for name, expected := range map[string]string{"John F. Kennedy": "john.f.kennedy", "Smith, John": "smith.john",
			"Dr. Who": "dr.who", "Jane (Doe)": "jane.doe", "Jane  Smith": "jane.smith", "...Bob!": "bob", " _Ann_ ": "ann",
			"Mr. Smith Jr.": "mr.smith.jr", "Bob .\x01": "bob"} {
			info := &oidc.UserInfo{}
			info.Name = name
			info.Email = "jane@doe.com"
			info.EmailVerified = true
			assert.Equal(t, expected, Username(info, authn.OidcClaimName), name)
			assert.True(t, authn.PortableUsername(expected), expected)
		}
	})
	t.Run("PreferredUsernameNotTrimmed", func(t *testing.T) {
		// A preferred username is not display text, so its dots are kept.
		info := &oidc.UserInfo{}
		info.PreferredUsername = "bob."
		info.Name = "Bob"
		assert.Equal(t, "bob.", Username(info, authn.OidcClaimPreferredUsername))
	})
	t.Run("NicknameTrimmed", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Nickname = "-Jens-"
		assert.Equal(t, "jens", Username(info, authn.OidcClaimNickname))
	})
	t.Run("NotPortableEmailHandle", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "50%off"
		info.Email = "__svc@doe.com"
		info.EmailVerified = true
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
	})
	t.Run("NotPortableEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Email = "__svc@doe.com"
		info.EmailVerified = true
		assert.Equal(t, "", Username(info, authn.OidcClaimEmail))
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
		assert.Equal(t, "jane", result)
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
		assert.Equal(t, "jane", result)
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
		assert.Equal(t, "jane", result)
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
		info.Name = "Jose Garcia"
		info.Nickname = "Pepe"
		info.Email = "jose@example.com"
		info.EmailVerified = true
		assert.Equal(t, "jose.garcia", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "jose.garcia", Username(info, authn.OidcClaimName))
		assert.Equal(t, "pepe", Username(info, authn.OidcClaimNickname))
		assert.Equal(t, "jose@example.com", Username(info, authn.OidcClaimEmail))
	})
	t.Run("NotASCIIEmailHandle", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "josé"
		info.Name = "José García"
		info.Email = "Jose.Garcia@example.com"
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
		info.EmailVerified = true
		assert.Equal(t, "jose.garcia", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "jose.garcia", Username(info, authn.OidcClaimName))
		assert.Equal(t, "jose.garcia", Username(info, authn.OidcClaimNickname))
		assert.Equal(t, "jose.garcia@example.com", Username(info, authn.OidcClaimEmail))
	})
	t.Run("NotASCIIEmailWithoutHandle", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "josé"
		info.Email = ".jose@example.com"
		info.EmailVerified = true
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
	})
	t.Run("NotASCIIBeforeSanitizing", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "\u212Aate"
		info.Name = "Kate Doe"
		assert.Equal(t, "kate.doe", Username(info, authn.OidcClaimPreferredUsername))
		info.PreferredUsername = "kate@exämple.com"
		assert.Equal(t, "kate.doe", Username(info, authn.OidcClaimPreferredUsername))
		info.PreferredUsername = "jane\u200b"
		assert.Equal(t, "kate.doe", Username(info, authn.OidcClaimPreferredUsername))
		info.Name = "Kåte"
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
	})
	t.Run("NotASCIIEmail", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "josé"
		info.Email = "\u212Aate@example.com"
		info.EmailVerified = true
		assert.Equal(t, "", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "", Username(info, authn.OidcClaimEmail))
		info.Nickname = "kate"
		assert.Equal(t, "kate", Username(info, authn.OidcClaimPreferredUsername))
		assert.Equal(t, "", Username(info, authn.OidcClaimEmail))
		info.EmailVerified = false
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
	t.Run("EmailClaim", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Email = "\u212Aate@example.com"
		info.EmailVerified = true
		info.Nickname = "kate"
		assert.True(t, SkipsClaim(info, authn.OidcClaimEmail))
		info.Nickname = ""
		assert.True(t, SkipsClaim(info, authn.OidcClaimEmail))
		info.Email = "kate@example.com"
		assert.False(t, SkipsClaim(info, authn.OidcClaimEmail))
	})
	t.Run("EmailHandle", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.PreferredUsername = "josé"
		info.Email = "jose@example.com"
		info.EmailVerified = true
		assert.True(t, SkipsClaim(info, authn.OidcClaimPreferredUsername))
	})
}

func TestClaimOrder(t *testing.T) {
	assert.Equal(t, []string{"preferred_username", "name", "nickname"}, claimOrder(""))
	assert.Equal(t, []string{"preferred_username", "name", "nickname"}, claimOrder(authn.OidcClaimPreferredUsername))
	assert.Equal(t, []string{"name", "preferred_username", "nickname"}, claimOrder(authn.OidcClaimName))
	assert.Equal(t, []string{"nickname", "preferred_username", "name"}, claimOrder(authn.OidcClaimNickname))
	assert.Equal(t, []string{"preferred_username", "name", "nickname"}, claimOrder(authn.OidcClaimEmail))
}

func TestClaimValue(t *testing.T) {
	info := &oidc.UserInfo{}
	info.PreferredUsername = "pref"
	info.Name = "Name"
	info.Nickname = "nick"
	assert.Equal(t, "pref", claimValue(info, authn.OidcClaimPreferredUsername))
	assert.Equal(t, "Name", claimValue(info, authn.OidcClaimName))
	assert.Equal(t, "nick", claimValue(info, authn.OidcClaimNickname))
	assert.Equal(t, "pref", claimValue(info, ""))
}

func TestVerifiedEmailUsername(t *testing.T) {
	info := &oidc.UserInfo{}
	info.Email = "Jane@Example.com"
	assert.Equal(t, "", verifiedEmailUsername(info))
	info.EmailVerified = true
	assert.Equal(t, "jane@example.com", verifiedEmailUsername(info))
	info.Email = ".jane@example.com"
	assert.Equal(t, "", verifiedEmailUsername(info))
	info.Email = "a@b"
	assert.Equal(t, "", verifiedEmailUsername(info))
}

func TestHandle(t *testing.T) {
	t.Run("Collapsed", func(t *testing.T) {
		assert.Equal(t, "john.f.kennedy", handle("John F. Kennedy"))
		assert.Equal(t, "a.b", handle("a ... b"))
	})
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "jane.doe", handle("Jane Doe"))
		assert.Equal(t, "", handle(""))
		assert.Equal(t, "", handle(".hidden"))
	})
}

func TestTrimText(t *testing.T) {
	t.Run("Trimmed", func(t *testing.T) {
		assert.Equal(t, "jane.doe", trimText("jane.doe."))
		assert.Equal(t, "bob", trimText("..._bob!^"))
		assert.Equal(t, "", trimText("..."))
	})
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "jane.doe", trimText("jane.doe"))
		assert.Equal(t, "a_b", trimText("a_b"))
	})
}
