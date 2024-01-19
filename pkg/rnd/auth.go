package rnd

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"log"
	"math/big"
)

const (
	SessionIdLength      = 64
	AuthTokenLength      = 48
	AppPasswordLength    = 27
	AppPasswordSeparator = '-'
)

// AuthToken generates a random hexadecimal character token for authenticating client applications.
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

// AppPassword generates a random, human-friendly authentication token that can also be used as
// password replacement for client applications. It is separated by 3 dashes for better readability
// and has a total length of 27 characters.
//
// Example: OXiV72-wTtiL9-d04jO7-X7XP4p
func AppPassword() string {
	m := big.NewInt(int64(len(CharsetBase62)))
	b := make([]byte, 0, AppPasswordLength)

	for i := 0; i < AppPasswordLength; i++ {
		if (i+1)%7 == 0 {
			b = append(b, AppPasswordSeparator)
		} else if i == AppPasswordLength-1 {
			b = append(b, CharsetBase62[crc32.ChecksumIEEE(b)%62])
			return string(b)
		} else if r, err := rand.Int(rand.Reader, m); err == nil {
			b = append(b, CharsetBase62[r.Int64()])
		}
	}

	return string(b)
}

// IsAppPassword checks if the string might be a valid app password.
func IsAppPassword(s string, verifyChecksum bool) bool {
	// Verify token length.
	if len(s) != AppPasswordLength {
		return false
	}

	// Check characters.
	sep := 0
	for _, r := range s {
		if r == AppPasswordSeparator {
			sep++
		} else if (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}

	// Check number of separators.
	if sep != AppPasswordLength/7 {
		return false
	} else if !verifyChecksum {
		return true
	}

	// Verify token checksum.
	return s[AppPasswordLength-1] == CharsetBase62[crc32.ChecksumIEEE([]byte(s[:AppPasswordLength-1]))%62]
}

// IsAuthAny checks if the string might be a valid auth token or app password.
func IsAuthAny(s string) bool {
	// Check if string might be a regular auth token.
	if IsAuthToken(s) {
		return true
	}

	// Check if string might be a human-friendly app password.
	if IsAppPassword(s, false) {
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
