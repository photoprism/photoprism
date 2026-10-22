package tokens

import (
	"strconv"
	"strings"
	"testing"
	"time"

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

// inOneHour returns a time one hour from now.
func inOneHour() time.Time {
	return time.Now().Add(time.Hour)
}

func TestSignCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.AuthToken()
		expires := inOneHour()
		signed := SignCode(value, expires)
		assert.True(t, strings.HasPrefix(signed, value+"."+strconv.FormatInt(expires.Unix(), 10)+"."))
		assert.Len(t, signed, len(value)+1+10+1+2*valueMacLen)
		assert.Equal(t, signed, SignCode(value, expires))
		assert.NoError(t, VerifyCode(signed))
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withOAuthKey(t, nil)
		assert.Equal(t, "", SignCode(rnd.AuthToken(), inOneHour()))
	})
}

func TestVerifyCode(t *testing.T) {
	t.Run("Forged", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.AuthToken()
		assert.ErrorIs(t, VerifyCode(value), ErrInvalidValue)
		assert.ErrorIs(t, VerifyCode(value+"."+strconv.FormatInt(inOneHour().Unix(), 10)+"."+strings.Repeat("0", 2*valueMacLen)), ErrInvalidValue)
		assert.ErrorIs(t, VerifyCode(rnd.AuthToken()+SignCode(value, inOneHour())[len(value):]), ErrInvalidValue)
	})
	t.Run("Expired", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		assert.ErrorIs(t, VerifyCode(SignCode(rnd.AuthToken(), time.Now().Add(-time.Second))), ErrExpiredValue)
		assert.NoError(t, VerifyCode(SignCode(rnd.AuthToken(), time.Now().Add(time.Second))))
	})
	t.Run("ExpiryChanged", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.AuthToken()
		past := strconv.FormatInt(time.Now().Add(-time.Minute).Unix(), 10)
		signed := SignCode(value, time.Now().Add(-time.Minute))
		_, _, mac, ok := SplitValue(signed)
		assert.True(t, ok)
		assert.Equal(t, value+"."+past+"."+mac, signed)
		assert.ErrorIs(t, VerifyCode(value+"."+strconv.FormatInt(inOneHour().Unix(), 10)+"."+mac), ErrInvalidValue)
	})
	t.Run("OtherKey", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		signed := SignCode(rnd.AuthToken(), inOneHour())
		OAuth.Key = []byte(strings.Repeat("k", KeyLen))
		assert.ErrorIs(t, VerifyCode(signed), ErrInvalidValue)
	})
	t.Run("RefreshTokenRefused", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		assert.ErrorIs(t, VerifyCode(SignRefresh(rnd.AuthToken(), inOneHour())), ErrInvalidValue)
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		signed := SignCode(rnd.AuthToken(), inOneHour())
		OAuth.Key = nil
		assert.ErrorIs(t, VerifyCode(signed), ErrInvalidValue)
	})
}

func TestSignRefresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.RefreshToken(rnd.GenerateUID('g'))
		signed := SignRefresh(value, inOneHour())
		assert.True(t, strings.HasPrefix(signed, value+"."))
		got, err := VerifyRefresh(signed)
		assert.NoError(t, err)
		assert.Equal(t, value, got)
	})
	t.Run("Unconfigured", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen-1))
		assert.Equal(t, "", SignRefresh(rnd.RefreshToken(rnd.GenerateUID('g')), inOneHour()))
	})
}

func TestVerifyRefresh(t *testing.T) {
	t.Run("Forged", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.RefreshToken(rnd.GenerateUID('g'))
		got, err := VerifyRefresh(value)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.Empty(t, got)
		signed := SignRefresh(value, inOneHour())
		got, err = VerifyRefresh(rnd.RefreshToken(rnd.GenerateUID('g')) + signed[len(value):])
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.Empty(t, got)
	})
	t.Run("Expired", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		value := rnd.RefreshToken(rnd.GenerateUID('g'))
		got, err := VerifyRefresh(SignRefresh(value, time.Now().Add(-time.Second)))
		assert.ErrorIs(t, err, ErrExpiredValue)
		assert.Equal(t, value, got)
	})
	t.Run("CodeRefused", func(t *testing.T) {
		withOAuthKey(t, make([]byte, KeyLen))
		_, err := VerifyRefresh(SignCode(rnd.RefreshToken(rnd.GenerateUID('g')), inOneHour()))
		assert.ErrorIs(t, err, ErrInvalidValue)
	})
}

func TestSigner_SignValue(t *testing.T) {
	s := &Signer{Key: make([]byte, KeyLen)}

	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "value.1.", s.SignValue("a", "value", 1)[:len("value.1.")])
		assert.NotEqual(t, s.SignValue("a", "value", 1), s.SignValue("b", "value", 1))
		assert.NotEqual(t, s.SignValue("a", "value", 1)[len("value.1."):], s.SignValue("a", "value", 2)[len("value.2."):])
	})
	t.Run("InvalidArgs", func(t *testing.T) {
		assert.Equal(t, "", s.SignValue("", "value", 1))
		assert.Equal(t, "", s.SignValue("a", "", 1))
		assert.Equal(t, "", s.SignValue("a", "value", 0))
		assert.Equal(t, "", s.SignValue("a", "value", -1))
	})
}

func TestSigner_VerifyValue(t *testing.T) {
	s := &Signer{Key: make([]byte, KeyLen)}
	signed := s.SignValue("a", "value", 100)

	t.Run("Success", func(t *testing.T) {
		value, err := s.VerifyValue("a", signed, 99)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
		value, err = s.VerifyValue("a", signed, 100)
		assert.NoError(t, err)
		assert.Equal(t, "value", value)
	})
	t.Run("Expired", func(t *testing.T) {
		value, err := s.VerifyValue("a", signed, 101)
		assert.ErrorIs(t, err, ErrExpiredValue)
		assert.Equal(t, "value", value)
	})
	t.Run("ExpiryChanged", func(t *testing.T) {
		_, err := s.VerifyValue("a", "value.200"+signed[len("value.100"):], 99)
		assert.ErrorIs(t, err, ErrInvalidValue)
		_, err = s.VerifyValue("a", "value.0100"+signed[len("value.100"):], 99)
		assert.ErrorIs(t, err, ErrInvalidValue)
	})
	t.Run("InvalidShape", func(t *testing.T) {
		for _, v := range []string{"", ".", "value", "value.", "." + signed[len("value."):], signed + "0", signed[:len(signed)-1], strings.Repeat("a", maxTokenLen+1)} {
			value, err := s.VerifyValue("a", v, 99)
			assert.ErrorIs(t, err, ErrInvalidValue, v)
			assert.Empty(t, value, v)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		long := s.SignValue("a", strings.Repeat("v", maxTokenLen), 100)
		_, err := s.VerifyValue("a", long, 99)
		assert.ErrorIs(t, err, ErrInvalidValue)
	})
	t.Run("OtherPurpose", func(t *testing.T) {
		_, err := s.VerifyValue("b", signed, 99)
		assert.ErrorIs(t, err, ErrInvalidValue)
		_, err = s.VerifyValue("", signed, 99)
		assert.ErrorIs(t, err, ErrInvalidValue)
	})
	t.Run("ExpiredOtherPurpose", func(t *testing.T) {
		_, err := s.VerifyValue("b", signed, 101)
		assert.ErrorIs(t, err, ErrInvalidValue)
	})
}

func TestSplitValue(t *testing.T) {
	mac := strings.Repeat("a", 2*valueMacLen)

	t.Run("Success", func(t *testing.T) {
		value, expires, gotMac, ok := SplitValue("g1.abc.1700000000." + mac)
		assert.True(t, ok)
		assert.Equal(t, "g1.abc", value)
		assert.Equal(t, int64(1700000000), expires)
		assert.Equal(t, mac, gotMac)
	})
	t.Run("InvalidShape", func(t *testing.T) {
		for _, v := range []string{"", "value", "." + mac, "value." + mac, ".1." + mac, "value.1." + mac[1:], "value.1." + mac + "a", "value.01." + mac, "value.+1." + mac, "value.0." + mac, "value.-1." + mac, "value.x." + mac, "value.99999999999999999999." + mac} {
			_, _, _, ok := SplitValue(v)
			assert.False(t, ok, v)
		}
	})
}

func TestParseExpires(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expires, ok := parseExpires("1700000000")
		assert.True(t, ok)
		assert.Equal(t, int64(1700000000), expires)
	})
	t.Run("NotCanonical", func(t *testing.T) {
		for _, v := range []string{"", "0", "-1", "+1", "01", " 1", "1.0", "9223372036854775808"} {
			_, ok := parseExpires(v)
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
