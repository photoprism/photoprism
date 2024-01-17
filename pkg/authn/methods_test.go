package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodType_String(t *testing.T) {
	assert.Equal(t, "default", MethodDefault.String())
	assert.Equal(t, "access_token", MethodAccessToken.String())
	assert.Equal(t, "oauth2", MethodOAuth2.String())
	assert.Equal(t, "oidc", MethodOIDC.String())
	assert.Equal(t, "2fa", Method2FA.String())
	assert.Equal(t, "default", MethodUnknown.String())
}

func TestMethodType_IsDefault(t *testing.T) {
	assert.Equal(t, true, MethodDefault.IsDefault())
	assert.Equal(t, false, MethodAccessToken.IsDefault())
	assert.Equal(t, false, MethodOAuth2.IsDefault())
	assert.Equal(t, false, MethodOIDC.IsDefault())
	assert.Equal(t, false, Method2FA.IsDefault())
	assert.Equal(t, true, MethodUnknown.IsDefault())
}

func TestMethodType_Pretty(t *testing.T) {
	assert.Equal(t, "Default", MethodDefault.Pretty())
	assert.Equal(t, "Access Token", MethodAccessToken.Pretty())
	assert.Equal(t, "OAuth2", MethodOAuth2.Pretty())
	assert.Equal(t, "OIDC", MethodOIDC.Pretty())
	assert.Equal(t, "2FA", Method2FA.Pretty())
	assert.Equal(t, "Default", MethodUnknown.Pretty())
}

func TestMethodType_Equal(t *testing.T) {
	assert.True(t, Method2FA.Equal("2fa"))
	assert.False(t, MethodAccessToken.Equal("2fa"))
}

func TestMethodType_NotEqual(t *testing.T) {
	assert.False(t, Method2FA.NotEqual("2fa"))
	assert.True(t, MethodAccessToken.NotEqual("2fa"))
}

func TestMethod(t *testing.T) {
	assert.Equal(t, MethodDefault, Method("default"))
	assert.Equal(t, MethodDefault, Method(""))
	assert.Equal(t, MethodAccessToken, Method("access_token"))
	assert.Equal(t, MethodOAuth2, Method("oauth2"))
	assert.Equal(t, MethodOIDC, Method("oidc"))
	assert.Equal(t, MethodOIDC, Method("sso"))
	assert.Equal(t, Method2FA, Method("2fa"))
	assert.Equal(t, Method2FA, Method("totp"))
}
