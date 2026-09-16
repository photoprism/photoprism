package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUri(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		result := Uri("https://docs.photoprism.app/getting-started/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started/config-options/#file-converters", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := Uri("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
		assert.Equal(t, "", result)
	})
	t.Run("Emoji", func(t *testing.T) {
		result := Uri("Hello 👍")
		assert.Equal(t, "Hello%20%F0%9F%91%8D", result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := Uri("")
		assert.Equal(t, "", result)
	})
}

func TestUriRedacted(t *testing.T) {
	t.Run("WithCredentials", func(t *testing.T) {
		result := UriRedacted("https://user:secret@example.com/path?q=1")
		assert.Equal(t, "https://user:***@example.com/path?q=1", result)
	})
	t.Run("WithoutCredentials", func(t *testing.T) {
		result := UriRedacted("https://docs.photoprism.app/getting-started/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started/config-options/#file-converters", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := UriRedacted("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
		assert.Equal(t, "", result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := UriRedacted("")
		assert.Equal(t, "", result)
	})
}

func BenchmarkUri(b *testing.B) {
	for b.Loop() {
		Uri("https://docs.photoprism.app/getting-started/config-options/#file-converters")
	}
}

func BenchmarkUriRedacted(b *testing.B) {
	for b.Loop() {
		UriRedacted("https://user:secret@docs.photoprism.app/getting-started/config-options/#file-converters")
	}
}

func BenchmarkUriEmpty(b *testing.B) {
	for b.Loop() {
		Uri("")
	}
}

func TestUriCredentialParam(t *testing.T) {
	t.Run("Credential", func(t *testing.T) {
		for _, name := range []string{
			"key", "api_key", "X-Api-Key", "apikey", "token", "access_token", "AccessToken",
			"secret", "client_secret", "password", "passwd", "pwd", "auth", "authorization",
			"credential", "credentials", "sig", "signature",
		} {
			assert.Truef(t, UriCredentialParam(name), "%s must be treated as a credential", name)
		}
	})
	t.Run("NotACredential", func(t *testing.T) {
		for _, name := range []string{"tier", "model", "format", "stream", "temperature", "n", ""} {
			assert.Falsef(t, UriCredentialParam(name), "%s must be shown", name)
		}
	})
}

func TestUriRedactedQuery(t *testing.T) {
	t.Run("Unparsable", func(t *testing.T) {
		assert.Equal(t, "", UriRedacted("://nope"))
	})
	t.Run("NoQuery", func(t *testing.T) {
		assert.Equal(t, "https://api.example.com/v1", UriRedacted("https://api.example.com/v1"))
	})
	t.Run("QueryOrderIsNotRelevant", func(t *testing.T) {
		// Re-encoding sorts the parameters, so the assertion is on what each one holds.
		result := UriRedacted("https://api.example.com/v1?z=1&api_key=notreal&a=2")
		assert.Contains(t, result, "api_key=***")
		assert.Contains(t, result, "z=1")
		assert.Contains(t, result, "a=2")
		assert.NotContains(t, result, "notreal")
	})
	t.Run("MalformedQuery", func(t *testing.T) {
		for _, s := range []string{
			"https://api.example.com/v1?token=notreal;tail",
			"https://api.example.com/v1?token=notreal%zz",
			"https://api.example.com/v1?other=ok&token=notreal;tail",
			"https://api.example.com/v1?notreal;tail",
		} {
			result := UriRedacted(s)
			assert.NotContainsf(t, result, "notreal", "%s must not keep its query", s)
			assert.Containsf(t, result, "https://api.example.com/v1?***", "%s must report the removal", s)
		}
	})
	t.Run("MalformedQueryKeepsUserinfoRedaction", func(t *testing.T) {
		result := UriRedacted("https://user:pass@api.example.com/v1?token=notreal;tail")
		assert.Equal(t, "https://user:***@api.example.com/v1?***", result)
	})
	t.Run("UserinfoAndQuery", func(t *testing.T) {
		// A value that reads as redacted must not sit beside one that is not.
		result := UriRedacted("https://user:pass@api.example.com/v1?access_token=notreal")
		assert.NotContains(t, result, "pass@")
		assert.NotContains(t, result, "notreal")
		assert.Contains(t, result, "user:***@")
		assert.Contains(t, result, "access_token=***")
	})
}

func TestUriCredentials(t *testing.T) {
	// The split follows net/url: the first colon bounds the name, the last at sign bounds the
	// userinfo. A name with no password beside it is treated as the secret, because nothing
	// distinguishes an access token in that position from an account name.
	t.Run("Password", func(t *testing.T) {
		assert.Equal(t, "https://user:"+UriRedactedValue+"@example.com/a",
			UriCredentials("https://user:notreal@example.com/a"))
	})
	t.Run("PasswordHoldingAColon", func(t *testing.T) {
		assert.Equal(t, "https://user:"+UriRedactedValue+"@example.com/a",
			UriCredentials("https://user:not:real@example.com/a"))
	})
	t.Run("PasswordHoldingAnAtSign", func(t *testing.T) {
		assert.Equal(t, "https://user:"+UriRedactedValue+"@example.com/a",
			UriCredentials("https://user:not@real@example.com/a"))
	})
	t.Run("NameHoldingAnAtSign", func(t *testing.T) {
		assert.Equal(t, "smtps://noreply@example.com:"+UriRedactedValue+"@mail.example.com/a",
			UriCredentials("smtps://noreply@example.com:notreal@mail.example.com/a"))
	})
	t.Run("NameWithoutPassword", func(t *testing.T) {
		assert.Equal(t, "https://"+UriRedactedValue+"@example.com/a",
			UriCredentials("https://ghp000000000000000000@example.com/a"))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		// Nothing follows the colon, so the name is what is left to hide.
		assert.Equal(t, "https://"+UriRedactedValue+"@example.com/a",
			UriCredentials("https://user:@example.com/a"))
	})
	t.Run("NothingToHide", func(t *testing.T) {
		// A marker here would report a removal that did not happen.
		assert.Equal(t, "https://@example.com/a", UriCredentials("https://@example.com/a"))
		assert.Equal(t, "https://:@example.com/a", UriCredentials("https://:@example.com/a"))
	})
	t.Run("PasswordWithoutAName", func(t *testing.T) {
		assert.Equal(t, "https://:"+UriRedactedValue+"@example.com/a",
			UriCredentials("https://:notreal@example.com/a"))
	})
	t.Run("InLongerText", func(t *testing.T) {
		out := UriCredentials("failed to download https://user:notreal@example.com/a (timeout)")
		assert.NotContains(t, out, "notreal")
		assert.Contains(t, out, "example.com/a")
	})
	t.Run("TwoUrls", func(t *testing.T) {
		out := UriCredentials("from https://a:1234@x.example to https://b:5678@y.example")
		assert.NotContains(t, out, "1234")
		assert.NotContains(t, out, "5678")
	})
	t.Run("NotAUrl", func(t *testing.T) {
		for _, s := range []string{
			"invalid key=value pair at a:b@c",
			"user@example.com",
			"dial https://example.com:8080 failed",
			"https://example.com/a?q=1:2@3",
			`{"endpoint":"https://api.example.com","user":"bob@example.com"}`,
			"https://[2001:db8::1]:8443/a",
			// An AD principal name is an identifier rather than a credential, and every shape a
			// directory renders it in keeps it out of the userinfo position.
			"ldap: bind as jdoe@example.com failed",
			"ldap://dc.example.com/dc=example,dc=com??sub?(userprincipalname=jdoe@example.com)",
			"ldaps://dc.example.com:636/CN=Users,DC=example,DC=com?(mail=jdoe@example.com)",
			"ldap: search ldap://dc.example.com/dc=example,dc=com failed for (userprincipalname=jdoe@example.com)",
		} {
			assert.Equal(t, s, UriCredentials(s))
		}
	})
	t.Run("Idempotent", func(t *testing.T) {
		out := UriCredentials("https://user:notreal@example.com/a")
		assert.Equal(t, out, UriCredentials(out))
	})
}

func TestUriRedactedName(t *testing.T) {
	// A name with no password beside it carries the secret in some conventions, and the two are
	// indistinguishable from the value alone.
	t.Run("NameWithoutPassword", func(t *testing.T) {
		assert.Equal(t, "https://"+UriRedactedValue+"@github.example.com/org/repo.git",
			UriRedacted("https://ghp000000000000000000@github.example.com/org/repo.git"))
	})
	t.Run("NameWithPasswordIsKept", func(t *testing.T) {
		assert.Equal(t, "https://proxy-user:"+UriRedactedValue+"@proxy.example.com:3128",
			UriRedacted("https://proxy-user:notreal@proxy.example.com:3128"))
	})
	t.Run("NoUserinfo", func(t *testing.T) {
		assert.Equal(t, "https://proxy.example.com:3128", UriRedacted("https://proxy.example.com:3128"))
	})
	t.Run("NothingToHide", func(t *testing.T) {
		// Redacted marks a password that is set and empty, which would report a removal that did
		// not happen, so the component is dropped instead.
		assert.Equal(t, "https://@proxy.example.com", UriRedacted("https://@proxy.example.com"))
		assert.Equal(t, "https://:@proxy.example.com", UriRedacted("https://:@proxy.example.com"))
	})
}
