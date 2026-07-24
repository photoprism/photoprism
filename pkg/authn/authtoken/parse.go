package authtoken

import (
	"crypto/hmac"
	"net/url"
)

// Verify checks a bunny.net Advanced token against the message rebuilt from the request. It returns
// nil when the signature matches and the token has not expired (now and expires are Unix seconds),
// ErrMalformed when the token is empty, ErrExpired when the deadline has passed, and ErrSignature when
// the token is forged, tampered, or signed with a different key. The comparison is constant-time.
func Verify(key []byte, signaturePath string, expires int64, params url.Values, clientIP, token string, now int64) error {
	if token == "" {
		return ErrMalformed
	}

	if now > expires {
		return ErrExpired
	}

	want := Sign(key, signaturePath, expires, params, clientIP)

	if !hmac.Equal([]byte(token), []byte(want)) {
		return ErrSignature
	}

	return nil
}

// Valid reports whether token is a valid, unexpired signature for the request.
func Valid(key []byte, signaturePath string, expires int64, params url.Values, clientIP, token string, now int64) bool {
	return Verify(key, signaturePath, expires, params, clientIP, token, now) == nil
}
