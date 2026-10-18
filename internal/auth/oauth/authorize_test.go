package oauth

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testAuthorizeValues returns the query values of a valid authorization request.
func testAuthorizeValues() url.Values {
	return url.Values{
		"response_type":         {"code"},
		"client_id":             {"cs5gfen1bgxz7s9i"},
		"redirect_uri":          {"https://claude.ai/api/mcp/auth_callback"},
		"scope":                 {" mcp  photos "},
		"state":                 {"xyz"},
		"code_challenge":        {testChallenge(testVerifier)},
		"code_challenge_method": {"S256"},
		"resource":              {"https://photos.example.com/api/v1/mcp"},
	}
}

func TestParseAuthorizeRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := ParseAuthorizeRequest(testAuthorizeValues())
		assert.Equal(t, "code", r.ResponseType)
		assert.Equal(t, "cs5gfen1bgxz7s9i", r.ClientID)
		assert.Equal(t, "mcp photos", r.Scope)
		assert.Equal(t, []string{"mcp", "photos"}, r.Scopes())
		assert.Equal(t, "https://photos.example.com/api/v1/mcp", r.Resource)
		assert.Nil(t, r.Validate())
	})
	t.Run("Repeated", func(t *testing.T) {
		v := testAuthorizeValues()
		v.Add("response_type", "code")
		v.Add("client_id", "other")
		r := ParseAuthorizeRequest(v)
		assert.True(t, r.repeated("response_type"))
		assert.True(t, r.repeated("client_id"))
		assert.False(t, r.repeated("redirect_uri"))
	})
}

func TestAuthorizeRequest_ValidateClient(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var r *AuthorizeRequest
		require.NotNil(t, r.ValidateClient())
	})
	cases := map[string]func(v url.Values){
		"NoClientID":        func(v url.Values) { v.Del("client_id") },
		"ClientIDSpace":     func(v url.Values) { v.Set("client_id", "a b") },
		"ClientIDTooLong":   func(v url.Values) { v.Set("client_id", strings.Repeat("a", MaxClientIDLength+1)) },
		"NoRedirectURI":     func(v url.Values) { v.Del("redirect_uri") },
		"RedirectURIScheme": func(v url.Values) { v.Set("redirect_uri", "javascript:alert(1)") },
		"RedirectRepeated":  func(v url.Values) { v.Add("redirect_uri", "https://evil.example.com/cb") },
		"ClientIDRepeated":  func(v url.Values) { v.Add("client_id", "other") },
		"MaskedRepeated": func(v url.Values) {
			v.Add("response_type", "code")
			v.Add("redirect_uri", "https://evil.example.com/cb")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			v := testAuthorizeValues()
			mutate(v)
			err := ParseAuthorizeRequest(v).ValidateClient()
			require.NotNil(t, err)
			assert.Equal(t, ErrCodeInvalidRequest, err.Code)
		})
	}
}

func TestAuthorizeRequest_Validate(t *testing.T) {
	cases := map[string]struct {
		mutate func(v url.Values)
		code   string
	}{
		"NoResponseType":    {func(v url.Values) { v.Del("response_type") }, ErrCodeInvalidRequest},
		"TokenResponseType": {func(v url.Values) { v.Set("response_type", "token") }, ErrCodeUnsupportedResponse},
		"NoChallenge":       {func(v url.Values) { v.Del("code_challenge") }, ErrCodeInvalidRequest},
		"PlainMethod":       {func(v url.Values) { v.Set("code_challenge_method", "plain") }, ErrCodeInvalidRequest},
		"NoMethod":          {func(v url.Values) { v.Del("code_challenge_method") }, ErrCodeInvalidRequest},
		"ShortChallenge":    {func(v url.Values) { v.Set("code_challenge", "abc") }, ErrCodeInvalidRequest},
		"ScopeTooLong":      {func(v url.Values) { v.Set("scope", strings.Repeat("a", MaxScopeLength+1)) }, ErrCodeInvalidScope},
		"StateControlChar":  {func(v url.Values) { v.Set("state", "a\nb") }, ErrCodeInvalidRequest},
		"StateTooLong":      {func(v url.Values) { v.Set("state", strings.Repeat("a", MaxStateLength+1)) }, ErrCodeInvalidRequest},
		"ResourceRelative":  {func(v url.Values) { v.Set("resource", "/api/v1/mcp") }, ErrCodeInvalidTarget},
		"ResourceFragment":  {func(v url.Values) { v.Set("resource", "https://photos.example.com/api/v1/mcp#x") }, ErrCodeInvalidTarget},
		"ScopeRepeated":     {func(v url.Values) { v.Add("scope", "admin") }, ErrCodeInvalidRequest},
		"ResourceRepeated":  {func(v url.Values) { v.Add("resource", "https://other.example.com/") }, ErrCodeInvalidRequest},
		"InvalidClient":     {func(v url.Values) { v.Del("client_id") }, ErrCodeInvalidRequest},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			v := testAuthorizeValues()
			c.mutate(v)
			err := ParseAuthorizeRequest(v).Validate()
			require.NotNil(t, err)
			assert.Equal(t, c.code, err.Code)
			assert.NotEmpty(t, err.Error())
		})
	}
	t.Run("NoResourceNoScopeNoState", func(t *testing.T) {
		v := testAuthorizeValues()
		v.Del("resource")
		v.Del("scope")
		v.Del("state")
		assert.Nil(t, ParseAuthorizeRequest(v).Validate())
	})
}

func TestAuthorizeRequest_Scopes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, (&AuthorizeRequest{Scope: "mcp photos"}).Scopes())
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, (&AuthorizeRequest{}).Scopes())
	})
	t.Run("Nil", func(t *testing.T) {
		var r *AuthorizeRequest
		assert.Nil(t, r.Scopes())
	})
}

func TestNewRequestError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		err := NewRequestError(ErrCodeInvalidTarget, "resource invalid")
		assert.Equal(t, ErrCodeInvalidTarget, err.Code)
		assert.Equal(t, "resource invalid", err.Description)
	})
}

func TestRequestError(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		err := NewRequestError(ErrCodeInvalidScope, "scope invalid")
		assert.Equal(t, "scope invalid", err.Error())
		assert.Equal(t, ErrCodeInvalidScope, err.Code)
	})
	t.Run("Nil", func(t *testing.T) {
		var err *RequestError
		assert.Equal(t, "", err.Error())
	})
}

func TestValidValue(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, ValidValue("", 10))
		assert.True(t, ValidValue("abc äöü", 10))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, ValidValue("abcdef", 5))
		assert.False(t, ValidValue("a\x00b", 10))
		assert.False(t, ValidValue(string([]byte{0xff}), 10))
		assert.False(t, ValidValue("a\u202eb", 10))
		assert.False(t, ValidValue("a\u2028b", 10))
	})
}

func TestValidAbsoluteURL(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, ValidAbsoluteURL("https://photos.example.com/api/v1/mcp"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, ValidAbsoluteURL("/api/v1/mcp"))
		assert.False(t, ValidAbsoluteURL("https://photos.example.com/#"))
		assert.False(t, ValidAbsoluteURL("https://u@photos.example.com/"))
		assert.False(t, ValidAbsoluteURL("%zz"))
	})
}
