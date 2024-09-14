package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyType_String(t *testing.T) {
	assert.Equal(t, "totp", KeyTOTP.String())
	assert.Equal(t, "", KeyUnknown.String())
}

func TestKeyType_Pretty(t *testing.T) {
	assert.Equal(t, "TOTP", KeyTOTP.Pretty())
	assert.Equal(t, "Unknown", KeyUnknown.Pretty())
}

func TestKeyType_Equal(t *testing.T) {
	assert.True(t, KeyTOTP.Equal("totp"))
	assert.False(t, KeyTOTP.Equal("2fa"))
	assert.False(t, KeyTOTP.Equal(""))
	assert.True(t, KeyUnknown.Equal(""))
	assert.False(t, KeyUnknown.Equal("2fa"))
	assert.False(t, KeyUnknown.Equal("totp"))
}

func TestKeyType_NotEqual(t *testing.T) {
	assert.True(t, KeyTOTP.NotEqual("2fa"))
	assert.False(t, KeyTOTP.NotEqual("totp"))
	assert.False(t, KeyUnknown.NotEqual(""))
	assert.True(t, KeyUnknown.NotEqual("2fa"))
	assert.True(t, KeyUnknown.NotEqual("totp"))
}

func TestKey(t *testing.T) {
	assert.Equal(t, KeyTOTP, Key("totp"))
	assert.Equal(t, KeyTOTP, Key("otp"))
	assert.Equal(t, KeyUnknown, Key("false"))
	assert.NotEqual(t, "xxx", Key("xxx"))
}
