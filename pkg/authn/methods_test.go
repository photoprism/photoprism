package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodType_String(t *testing.T) {
	assert.Equal(t, "default", MethodDefault.String())
	assert.Equal(t, "personal", MethodPersonal.String())
	assert.Equal(t, "oauth2", MethodOAuth2.String())
	assert.Equal(t, "oidc", MethodOIDC.String())
	assert.Equal(t, "totp", MethodTOTP.String())
	assert.Equal(t, "default", MethodUnknown.String())
}

func TestMethodType_IsDefault(t *testing.T) {
	assert.Equal(t, true, MethodDefault.IsDefault())
	assert.Equal(t, false, MethodPersonal.IsDefault())
	assert.Equal(t, false, MethodOAuth2.IsDefault())
	assert.Equal(t, false, MethodOIDC.IsDefault())
	assert.Equal(t, false, MethodTOTP.IsDefault())
	assert.Equal(t, true, MethodUnknown.IsDefault())
}

func TestMethodType_Pretty(t *testing.T) {
	assert.Equal(t, "Default", MethodDefault.Pretty())
	assert.Equal(t, "Personal", MethodPersonal.Pretty())
	assert.Equal(t, "OAuth2", MethodOAuth2.Pretty())
	assert.Equal(t, "OIDC", MethodOIDC.Pretty())
	assert.Equal(t, "TOTP/2FA", MethodTOTP.Pretty())
	assert.Equal(t, "Default", MethodUnknown.Pretty())
}

func TestMethodType_Equal(t *testing.T) {
	assert.True(t, MethodTOTP.Equal("totp"))
	assert.False(t, MethodTOTP.Equal("2fa"))
}

func TestMethodType_NotEqual(t *testing.T) {
	assert.True(t, MethodTOTP.NotEqual("2fa"))
	assert.False(t, MethodTOTP.NotEqual("totp"))
}

func TestMethod(t *testing.T) {
	assert.Equal(t, MethodDefault, Method("default"))
	assert.Equal(t, MethodDefault, Method(""))
	assert.Equal(t, MethodDefault, Method("access_token"))
	assert.Equal(t, MethodOAuth2, Method("oauth2"))
	assert.Equal(t, MethodOIDC, Method("oidc"))
	assert.Equal(t, MethodOIDC, Method("sso"))
	assert.Equal(t, MethodTOTP, Method("2fa"))
	assert.Equal(t, MethodTOTP, Method("totp"))
	assert.Equal(t, MethodTOTP, Method("TOTP/2FA"))
}

func TestMethodType_IsUnknown(t *testing.T) {
	assert.True(t, MethodUnknown.IsUnknown())
	assert.False(t, MethodTOTP.IsUnknown())
}

func TestMethodType_IsSession(t *testing.T) {
	assert.True(t, MethodSession.IsSession())
	assert.False(t, MethodTOTP.IsSession())
}
