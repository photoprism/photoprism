package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	// PurposeCode labels the authorization codes of the OAuth app flow.
	PurposeCode = "oauth-code"
	// PurposeRefresh labels the refresh tokens of the OAuth app flow.
	PurposeRefresh = "oauth-refresh"
	// valueMacLen is the number of HMAC bytes appended to a signed value, rendered as twice as many hex characters.
	valueMacLen = 16
)

var (
	// ErrInvalidValue is returned for a signed value with another shape or a MAC that does not match.
	ErrInvalidValue = errors.New("invalid signature")
	// ErrExpiredValue is returned for a signed value whose MAC matches and whose expiry has passed.
	ErrExpiredValue = errors.New("signature expired")
)

// OAuth signs the authorization codes and refresh tokens of the OAuth app flow. Its key is set by
// config.Propagate; its signature path is not used.
var OAuth = &Signer{}

// SignCode returns the authorization code value with an appended expiry and signature, or "" if it
// cannot be signed.
func SignCode(value string, expires time.Time) string {
	return OAuth.SignValue(PurposeCode, value, expires.Unix())
}

// VerifyCode returns ErrInvalidValue or ErrExpiredValue if a signed authorization code is not valid.
func VerifyCode(signed string) error {
	_, err := OAuth.VerifyValue(PurposeCode, signed, time.Now().Unix())
	return err
}

// SignRefresh returns the refresh token value with an appended expiry and signature, or "" if it cannot
// be signed.
func SignRefresh(value string, expires time.Time) string {
	return OAuth.SignValue(PurposeRefresh, value, expires.Unix())
}

// VerifyRefresh returns the value of a signed refresh token. It returns ErrInvalidValue if the token is not
// valid, and the value with ErrExpiredValue if it has expired.
func VerifyRefresh(signed string) (string, error) {
	return OAuth.VerifyValue(PurposeRefresh, signed, time.Now().Unix())
}

// SignValue returns "<value>.<expires>.<mac>" with a MAC over the purpose, value and expiry, or "" if the
// signer is unconfigured, an argument is empty or the expiry is not positive.
func (s *Signer) SignValue(purpose, value string, expires int64) string {
	if !s.Configured() || purpose == "" || value == "" || expires <= 0 {
		return ""
	}

	body := value + "." + strconv.FormatInt(expires, 10)

	return body + "." + s.valueMac(purpose, body)
}

// VerifyValue returns the value of a "<value>.<expires>.<mac>" string signed for the purpose. It returns
// ErrInvalidValue if the string has another shape or its MAC does not match, and the value with
// ErrExpiredValue if its expiry is before now, so a value is valid through the second it expires.
func (s *Signer) VerifyValue(purpose, signed string, now int64) (string, error) {
	if !s.Configured() || purpose == "" || len(signed) > maxTokenLen {
		return "", ErrInvalidValue
	}

	value, expires, mac, ok := SplitValue(signed)

	if !ok || !hmac.Equal([]byte(mac), []byte(s.valueMac(purpose, signed[:len(signed)-len(mac)-1]))) {
		return "", ErrInvalidValue
	} else if expires < now {
		return value, ErrExpiredValue
	}

	return value, nil
}

// SplitValue splits a string with the shape "<value>.<expires>.<mac>" of a signed value, without
// verifying the MAC or the expiry.
func SplitValue(signed string) (value string, expires int64, mac string, ok bool) {
	i := strings.LastIndexByte(signed, '.')

	if i < 1 || len(signed)-i-1 != 2*valueMacLen {
		return "", 0, "", false
	}

	body, mac := signed[:i], signed[i+1:]

	if j := strings.LastIndexByte(body, '.'); j < 1 {
		return "", 0, "", false
	} else if expires, ok = parseExpires(body[j+1:]); !ok {
		return "", 0, "", false
	} else {
		return body[:j], expires, mac, true
	}
}

// parseExpires parses a Unix time in its canonical decimal form, which is positive and has no sign or
// leading zeros, so each expiry has exactly one encoding.
func parseExpires(s string) (int64, bool) {
	expires, err := strconv.ParseInt(s, 10, 64)

	if err != nil || expires <= 0 || strconv.FormatInt(expires, 10) != s {
		return 0, false
	}

	return expires, true
}

// valueMac returns the hex-encoded MAC of a value, keyed by a subkey derived for the purpose.
func (s *Signer) valueMac(purpose, value string) string {
	sub := hmac.New(sha256.New, s.Key)
	sub.Write([]byte("value:" + purpose))

	mac := hmac.New(sha256.New, sub.Sum(nil))
	mac.Write([]byte(value))

	return hex.EncodeToString(mac.Sum(nil)[:valueMacLen])
}
