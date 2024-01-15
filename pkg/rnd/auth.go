package rnd

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

const (
	SessionIdLength     = 64
	AuthTokenLength     = 48
	AuthSecretLength    = 23
	AuthSecretSeparator = '-'
)

// AuthToken returns a random hexadecimal character string that can be used for authentication purposes.
//
// Examples: 9fa8e562564dac91b96881040e98f6719212a1a364e0bb25
func AuthToken() string {
	b := make([]byte, 24)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", b)
}

// IsAuthToken checks if the string might be a valid auth token.
func IsAuthToken(s string) bool {
	if l := len(s); l == AuthTokenLength {
		return IsHex(s)
	}

	return false
}

// AuthSecret returns a random, human-friendly string that can be used instead of a regular auth token.
// It is separated by 3 dashes for better readability and has a total length of 23 characters.
//
// Example: iXrDz-aY16n-4IUWM-otkM3
func AuthSecret() string {
	m := big.NewInt(int64(len(CharsetBase62)))
	b := make([]byte, AuthSecretLength)

	for i := range b {
		if r, err := rand.Int(rand.Reader, m); err == nil {
			if (i+1)%6 == 0 {
				b[i] = AuthSecretSeparator
			} else {
				b[i] = CharsetBase62[r.Int64()]
			}
		}
	}

	return string(b)
}

// IsAuthSecret checks if the string might be a valid auth secret.
func IsAuthSecret(s string) bool {
	if len(s) != AuthSecretLength {
		return false
	}

	sep := 0

	for _, r := range s {
		if r == AuthSecretSeparator {
			sep++
		} else if (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}

	return sep == AuthSecretLength/6
}

// IsAuthAny checks if the string might be a valid auth token or secret.
func IsAuthAny(s string) bool {
	// Check if string might be a regular auth token.
	if IsAuthToken(s) {
		return true
	}

	// Check if string might be a human-friendly auth secret.
	if IsAuthSecret(s) {
		return true
	}

	return false
}

// SessionID returns the hashed session id string.
func SessionID(token string) string {
	return Sha256([]byte(token))
}

// IsSessionID checks if the string is a session id string.
func IsSessionID(id string) bool {
	if len(id) != SessionIdLength {
		return false
	}

	return IsHex(id)
}
