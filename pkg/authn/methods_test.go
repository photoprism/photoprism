package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodType_String(t *testing.T) {
	assert.Equal(t, "default", MethodDefault.String())
	assert.Equal(t, "oauth2", MethodOAuth2.String())
	assert.Equal(t, "2fa", Method2FA.String())
	assert.Equal(t, "default", MethodUndefined.String())
}

func TestMethodType_Is(t *testing.T) {
	assert.Equal(t, true, MethodDefault.Is(MethodDefault))
	assert.Equal(t, true, MethodOAuth2.Is(MethodOAuth2))
	assert.Equal(t, true, Method2FA.Is(Method2FA))
	assert.Equal(t, true, MethodUndefined.Is(MethodUndefined))
}

func TestMethodType_IsNot(t *testing.T) {
	assert.Equal(t, true, MethodDefault.IsNot(MethodUndefined))
	assert.Equal(t, false, MethodDefault.IsNot(MethodDefault))
	assert.Equal(t, false, MethodOAuth2.IsNot(MethodOAuth2))
	assert.Equal(t, false, Method2FA.IsNot(Method2FA))
	assert.Equal(t, true, Method2FA.IsNot(MethodOAuth2))
	assert.Equal(t, true, MethodUndefined.IsNot(MethodDefault))
}

func TestMethodType_IsUndefined(t *testing.T) {
	assert.True(t, MethodUndefined.IsUndefined())
	assert.False(t, Method2FA.IsUndefined())
}

func TestMethodType_IsDefault(t *testing.T) {
	assert.Equal(t, true, MethodDefault.IsDefault())
	assert.Equal(t, false, MethodOAuth2.IsDefault())
	assert.Equal(t, false, Method2FA.IsDefault())
	assert.Equal(t, true, MethodUndefined.IsDefault())
}

func TestMethodType_Pretty(t *testing.T) {
	assert.Equal(t, "Default", MethodDefault.Pretty())
	assert.Equal(t, "OAuth2", MethodOAuth2.Pretty())
	assert.Equal(t, "2FA", Method2FA.Pretty())
	assert.Equal(t, "Default", MethodUndefined.Pretty())
}

func TestMethodType_Equal(t *testing.T) {
	assert.False(t, Method2FA.Equal("totp"))
	assert.True(t, Method2FA.Equal("2fa"))
}

func TestMethodType_NotEqual(t *testing.T) {
	assert.False(t, Method2FA.NotEqual("2fa"))
	assert.True(t, Method2FA.NotEqual("totp"))
}

func TestMethod(t *testing.T) {
	assert.Equal(t, MethodUndefined, Method(""))
	assert.Equal(t, MethodDefault, Method("default"))
	assert.Equal(t, MethodDefault, Method("access_token"))
	assert.Equal(t, MethodDefault, Method("false"))
	assert.Equal(t, MethodOAuth2, Method("oauth2"))
	assert.Equal(t, Method2FA, Method("2fa"))
	assert.Equal(t, Method2FA, Method("totp"))
	assert.Equal(t, Method2FA, Method("2FA"))
}

func TestMethods(t *testing.T) {
	types := Methods("oauth2, totp")
	assert.Equal(t, MethodOAuth2, types[0])
	assert.Equal(t, Method2FA, types[1])
}

func TestMethodType_IsSession(t *testing.T) {
	assert.True(t, MethodSession.IsSession())
	assert.False(t, Method2FA.IsSession())
}
