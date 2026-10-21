package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	// PurposeCode labels the authorization codes of the OAuth app flow.
	PurposeCode = "oauth-code"
	// PurposeRefresh labels the refresh tokens of the OAuth app flow.
	PurposeRefresh = "oauth-refresh"
	// valueMacLen is the number of HMAC bytes appended to a signed value, rendered as twice as many hex characters.
	valueMacLen = 16
)

// OAuth signs the authorization codes and refresh tokens of the OAuth app flow. Its key is set by
// config.Propagate; its signature path is not used.
var OAuth = &Signer{}

// SignCode returns the authorization code value with an appended signature, or "" if the signer is unconfigured.
func SignCode(value string) string {
	return OAuth.SignValue(PurposeCode, value)
}

// VerifyCode reports whether a signed authorization code carries a valid signature.
func VerifyCode(signed string) bool {
	_, ok := OAuth.VerifyValue(PurposeCode, signed)
	return ok
}

// SignRefresh returns the refresh token value with an appended signature, or "" if the signer is unconfigured.
func SignRefresh(value string) string {
	return OAuth.SignValue(PurposeRefresh, value)
}

// VerifyRefresh returns the value of a signed refresh token, and false if its signature is invalid.
func VerifyRefresh(signed string) (string, bool) {
	return OAuth.VerifyValue(PurposeRefresh, signed)
}

// SignValue returns "<value>.<mac>" with a MAC over the purpose and value, or "" if the signer is
// unconfigured or an argument is empty.
func (s *Signer) SignValue(purpose, value string) string {
	if !s.Configured() || purpose == "" || value == "" {
		return ""
	}

	return value + "." + s.valueMac(purpose, value)
}

// VerifyValue returns the value of a "<value>.<mac>" string signed for the purpose, and false if it has
// another shape or its MAC does not match.
func (s *Signer) VerifyValue(purpose, signed string) (string, bool) {
	if !s.Configured() || purpose == "" || len(signed) > maxTokenLen {
		return "", false
	}

	value, mac, ok := SplitValue(signed)

	if !ok || !hmac.Equal([]byte(mac), []byte(s.valueMac(purpose, value))) {
		return "", false
	}

	return value, true
}

// SplitValue splits a string with the shape "<value>.<mac>" of a signed value, without verifying the MAC.
func SplitValue(signed string) (value, mac string, ok bool) {
	i := strings.LastIndexByte(signed, '.')

	if i < 1 || len(signed)-i-1 != 2*valueMacLen {
		return "", "", false
	}

	return signed[:i], signed[i+1:], true
}

// valueMac returns the hex-encoded MAC of a value, keyed by a subkey derived for the purpose.
func (s *Signer) valueMac(purpose, value string) string {
	sub := hmac.New(sha256.New, s.Key)
	sub.Write([]byte("value:" + purpose))

	mac := hmac.New(sha256.New, sub.Sum(nil))
	mac.Write([]byte(value))

	return hex.EncodeToString(mac.Sum(nil)[:valueMacLen])
}
