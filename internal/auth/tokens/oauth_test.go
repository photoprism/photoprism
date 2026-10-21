package tokens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// withOAuthKey configures the OAuth signer for a test and restores it afterward.
func withOAuthKey(t *testing.T, key []byte) {
	t.Helper()
	orig := *OAuth
	OAuth.Key = key
	t.Cleanup(func() { *OAuth = orig })
}

func TestSignCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.AuthToken()
		signed := SignCode(value)
		assert.True(t, strings.HasPrefix(signed, value+"."))
		assert.Len(t, signed, len(value)+1+2*valueMacLen)
		assert.Equal(t, signed, SignCode(value))
		assert.True(t, VerifyCode(signed))
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withOAuthKey(t, nil)
		assert.Equal(t, "", SignCode(rnd.AuthToken()))
	})
}

func TestVerifyCode(t *testing.T) {
	t.Run("Forged", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.AuthToken()
		assert.False(t, VerifyCode(value))
		assert.False(t, VerifyCode(value+"."+strings.Repeat("0", 2*valueMacLen)))
		assert.False(t, VerifyCode(rnd.AuthToken()+SignCode(value)[len(value):]))
	})
	t.Run("OtherKey", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		signed := SignCode(rnd.AuthToken())
		OAuth.Key = []byte(strings.Repeat("k", KeyLen))
		assert.False(t, VerifyCode(signed))
	})
	t.Run("RefreshTokenRefused", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		assert.False(t, VerifyCode(SignRefresh(rnd.AuthToken())))
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		signed := SignCode(rnd.AuthToken())
		OAuth.Key = nil
		assert.False(t, VerifyCode(signed))
	})
}

func TestSignRefresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.RefreshToken(rnd.GenerateUID('g'))
		signed := SignRefresh(value)
		assert.True(t, strings.HasPrefix(signed, value+"."))
		got, ok := VerifyRefresh(signed)
		assert.True(t, ok)
		assert.Equal(t, value, got)
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen-1))
		assert.Equal(t, "", SignRefresh(rnd.RefreshToken(rnd.GenerateUID('g'))))
	})
}

func TestVerifyRefresh(t *testing.T) {
	t.Run("Forged", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.RefreshToken(rnd.GenerateUID('g'))
		_, ok := VerifyRefresh(value)
		assert.False(t, ok)
		signed := SignRefresh(value)
		_, ok = VerifyRefresh(rnd.RefreshToken(rnd.GenerateUID('g')) + signed[len(value):])
		assert.False(t, ok)
	})
	t.Run("CodeRefused", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		_, ok := VerifyRefresh(SignCode(rnd.RefreshToken(rnd.GenerateUID('g'))))
		assert.False(t, ok)
	})
}

func TestSigner_SignValue(t *testing.T) {
	s := &Signer{Key: make([]byte, KeyLen)}

	t.Run("Success", func(t *testing.T) {
		assert.NotEqual(t, s.SignValue("a", "value"), s.SignValue("b", "value"))
	})
	t.Run("EmptyArgs", func(t *testing.T) {
		assert.Equal(t, "", s.SignValue("", "value"))
		assert.Equal(t, "", s.SignValue("a", ""))
	})
}

func TestSigner_VerifyValue(t *testing.T) {
	s := &Signer{Key: make([]byte, KeyLen)}
	signed := s.SignValue("a", "value")

	t.Run("Success", func(t *testing.T) {
		value, ok := s.VerifyValue("a", signed)
		assert.True(t, ok)
		assert.Equal(t, "value", value)
	})
	t.Run("InvalidShape", func(t *testing.T) {
		for _, v := range []string{"", ".", "value", "value.", "." + signed[len("value."):], signed + "0", signed[:len(signed)-1], strings.Repeat("a", maxTokenLen+1)} {
			_, ok := s.VerifyValue("a", v)
			assert.False(t, ok, v)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		long := s.SignValue("a", strings.Repeat("v", maxTokenLen))
		_, ok := s.VerifyValue("a", long)
		assert.False(t, ok)
	})
	t.Run("OtherPurpose", func(t *testing.T) {
		_, ok := s.VerifyValue("b", signed)
		assert.False(t, ok)
		_, ok = s.VerifyValue("", signed)
		assert.False(t, ok)
	})
}

func TestSplitValue(t *testing.T) {
	mac := strings.Repeat("a", 2*valueMacLen)

	t.Run("Success", func(t *testing.T) {
		value, gotMac, ok := SplitValue("g1.abc." + mac)
		assert.True(t, ok)
		assert.Equal(t, "g1.abc", value)
		assert.Equal(t, mac, gotMac)
	})
	t.Run("InvalidShape", func(t *testing.T) {
		for _, v := range []string{"", "value", "." + mac, "value." + mac[1:], "value." + mac + "a"} {
			_, _, ok := SplitValue(v)
			assert.False(t, ok, v)
		}
	})
}

func TestSigner_ValueMac(t *testing.T) {
	t.Run("KnownAnswer", func(t *testing.T) {
		s := &Signer{Key: make([]byte, KeyLen)}
		assert.Equal(t, "f557c9722f3e37562c2fb8eef83d089b", s.valueMac(PurposeCode, "value"))
	})
	t.Run("PurposeAndValue", func(t *testing.T) {
		s := &Signer{Key: make([]byte, KeyLen)}
		assert.NotEqual(t, s.valueMac(PurposeCode, "value"), s.valueMac(PurposeRefresh, "value"))
		assert.NotEqual(t, s.valueMac(PurposeCode, "value"), s.valueMac(PurposeCode, "other"))
	})
}
