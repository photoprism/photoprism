package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidRedirectURI(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, ValidRedirectURI("https://claude.ai/api/mcp/auth_callback"))
		assert.True(t, ValidRedirectURI("https://chatgpt.com/connector/oauth/abc?x=1"))
		assert.True(t, ValidRedirectURI("http://127.0.0.1:33418/callback"))
		assert.True(t, ValidRedirectURI("http://[::1]/callback"))
		assert.True(t, ValidRedirectURI("http://localhost:6274/oauth/callback"))
		assert.True(t, ValidRedirectURI("app.photoprism.mobile:/oauth/callback"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, ValidRedirectURI(""))
		assert.False(t, ValidRedirectURI("/relative"))
		assert.False(t, ValidRedirectURI("http://photos.example.com/callback"))
		assert.False(t, ValidRedirectURI("https://example.com/callback#fragment"))
		assert.False(t, ValidRedirectURI("https://example.com/callback#"))
		assert.False(t, ValidRedirectURI("https://user:pass@example.com/callback"))
		assert.False(t, ValidRedirectURI("javascript:alert(1)"))
		assert.False(t, ValidRedirectURI("javascript:/alert(1)"))
		assert.False(t, ValidRedirectURI("data:/text/html,hello"))
		assert.False(t, ValidRedirectURI("app.photoprism.mobile://callback"))
		assert.False(t, ValidRedirectURI("app.photoprism.mobile:callback"))
		assert.False(t, ValidRedirectURI("https:///callback"))
		assert.False(t, ValidRedirectURI("https://:443/callback"))
		assert.False(t, ValidRedirectURI("app.photoprism.mobile:///callback"))
		assert.False(t, ValidRedirectURI("a.:/callback"))
		assert.False(t, ValidRedirectURI(".a:/callback"))
		assert.False(t, ValidRedirectURI("JaVaScRiPt:alert(1)"))
		assert.False(t, ValidRedirectURI("vbscript:/msgbox"))
		assert.False(t, ValidRedirectURI(`https:\\evil.com`))
		assert.False(t, ValidRedirectURI(`https:/\evil.com`))
		assert.False(t, ValidRedirectURI(`http://127.0.0.1\@evil.com/`))
		assert.False(t, ValidRedirectURI("http://localhost./callback"))
		assert.False(t, ValidRedirectURI("http://0.0.0.0/callback"))
		assert.False(t, ValidRedirectURI("http://127.0.0.1.nip.io/callback"))
	})
}

func TestMatchRedirectURI(t *testing.T) {
	registered := []string{"https://claude.ai/api/mcp/auth_callback", "http://127.0.0.1/callback", "app.photoprism.mobile:/oauth"}

	t.Run("Exact", func(t *testing.T) {
		assert.True(t, MatchRedirectURI(registered, "https://claude.ai/api/mcp/auth_callback"))
		assert.True(t, MatchRedirectURI(registered, "app.photoprism.mobile:/oauth"))
	})
	t.Run("LoopbackPort", func(t *testing.T) {
		assert.True(t, MatchRedirectURI(registered, "http://127.0.0.1:49152/callback"))
		assert.False(t, MatchRedirectURI(registered, "http://127.0.0.1:49152/other"))
		assert.False(t, MatchRedirectURI(registered, "http://localhost:49152/callback"))
		assert.False(t, MatchRedirectURI(registered, "http://127.0.0.1:49152/callback?x=1"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.False(t, MatchRedirectURI(registered, "https://claude.ai/api/mcp/auth_callback/"))
		assert.False(t, MatchRedirectURI(registered, "https://claude.ai/api/mcp/auth_callback?x=1"))
		assert.False(t, MatchRedirectURI(registered, "https://CLAUDE.ai/api/mcp/auth_callback"))
		assert.False(t, MatchRedirectURI(registered, ""))
		assert.False(t, MatchRedirectURI(nil, "https://claude.ai/api/mcp/auth_callback"))
	})
}

func TestPrivateUseScheme(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, privateUseScheme("app.photoprism.mobile"))
		assert.True(t, privateUseScheme("com.example"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, privateUseScheme("javascript"))
		assert.False(t, privateUseScheme("a."))
		assert.False(t, privateUseScheme("a..b"))
		assert.False(t, privateUseScheme(""))
	})
}

func TestMatchLoopbackURI(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		assert.True(t, matchLoopbackURI("http://127.0.0.1/cb", "http://127.0.0.1:8080/cb"))
		assert.True(t, matchLoopbackURI("http://[::1]:1/cb", "http://[::1]:2/cb"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.False(t, matchLoopbackURI("https://127.0.0.1/cb", "https://127.0.0.1:8080/cb"))
		assert.False(t, matchLoopbackURI("http://127.0.0.1/cb", "http://127.0.0.2:8080/cb"))
		assert.False(t, matchLoopbackURI("http://127.0.0.1/cb", "http://127.0.0.1:8080/%63b"))
		assert.False(t, matchLoopbackURI("http://example.com/cb", "http://example.com:8080/cb"))
		assert.False(t, matchLoopbackURI("%zz", "http://127.0.0.1/cb"))
	})
}

func TestRedirectURL(t *testing.T) {
	t.Run("InvalidScheme", func(t *testing.T) {
		_, err := RedirectURL("javascript:alert(document.domain)//", map[string][]string{"code": {"abc"}})
		assert.Error(t, err)
	})
	t.Run("OverridesRegisteredParams", func(t *testing.T) {
		result, err := CodeRedirectURL("https://app.example.com/cb?state=old&iss=evil", "abc", "new", "https://photos.example.com/")
		assert.NoError(t, err)
		assert.Equal(t, "https://app.example.com/cb?code=abc&iss=https%3A%2F%2Fphotos.example.com%2F&state=new", result)
	})
	t.Run("KeepsQuery", func(t *testing.T) {
		result, err := RedirectURL("https://app.example.com/cb?x=1", map[string][]string{"code": {"abc"}, "state": {""}})
		assert.NoError(t, err)
		assert.Equal(t, "https://app.example.com/cb?code=abc&x=1", result)
		assert.NotContains(t, result, "state")
	})
	t.Run("InvalidURI", func(t *testing.T) {
		_, err := RedirectURL("https://exa mple.com/%zz", nil)
		assert.Error(t, err)
	})
}

func TestCodeRedirectURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := CodeRedirectURL("https://app.example.com/cb", "abc", "xyz", "https://photos.example.com/")
		assert.NoError(t, err)
		assert.Equal(t, "https://app.example.com/cb?code=abc&iss=https%3A%2F%2Fphotos.example.com%2F&state=xyz", result)
	})
	t.Run("NoIssuer", func(t *testing.T) {
		_, err := CodeRedirectURL("https://app.example.com/cb", "abc", "xyz", "")
		assert.Error(t, err)
	})
	t.Run("NoCode", func(t *testing.T) {
		_, err := CodeRedirectURL("https://app.example.com/cb", "", "xyz", "https://photos.example.com/")
		assert.Error(t, err)
	})
	t.Run("JavaScript", func(t *testing.T) {
		_, err := CodeRedirectURL("javascript:alert(1)", "abc", "xyz", "https://photos.example.com/")
		assert.Error(t, err)
	})
	t.Run("NoState", func(t *testing.T) {
		result, err := CodeRedirectURL("https://app.example.com/cb?state=old", "abc", "", "https://photos.example.com/")
		assert.NoError(t, err)
		assert.NotContains(t, result, "state=")
	})
}

func TestErrorRedirectURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := ErrorRedirectURL("https://app.example.com/cb", ErrCodeAccessDenied, "denied", "xyz", "https://photos.example.com/")
		assert.NoError(t, err)
		assert.Equal(t, "https://app.example.com/cb?error=access_denied&error_description=denied&iss=https%3A%2F%2Fphotos.example.com%2F&state=xyz", result)
	})
	t.Run("NoIssuer", func(t *testing.T) {
		_, err := ErrorRedirectURL("https://app.example.com/cb", ErrCodeAccessDenied, "denied", "xyz", "")
		assert.Error(t, err)
	})
	t.Run("JavaScript", func(t *testing.T) {
		_, err := ErrorRedirectURL("javascript:alert(document.domain)//", ErrCodeAccessDenied, "denied", "", "https://photos.example.com/")
		assert.Error(t, err)
	})
}
