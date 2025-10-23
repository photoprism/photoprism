package rnd

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/photoprism/photoprism/pkg/checksum"
)

const (
	SessionIdLength   = 64
	AuthTokenLength   = 48
	JoinTokenLength   = 24
	AppPasswordLength = 27
	Separator         = '-'
)

// joinTokenSeparators determines where token separators (hyphens) appear.
var joinTokenSeparators = [...]int{7, 16}

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

// IsAuthToken checks if the string represents a valid auth token.
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
			b = append(b, Separator)
		} else if i == AppPasswordLength-1 {
			b = append(b, checksum.Char(b))
			return string(b)
		} else if r, err := rand.Int(rand.Reader, m); err == nil {
			b = append(b, CharsetBase62[r.Int64()])
		}
	}

	return string(b)
}

// IsAppPassword checks if the string represents a valid app password.
func IsAppPassword(s string, verifyChecksum bool) bool {
	// Verify token length.
	if len(s) != AppPasswordLength {
		return false
	}

	// Check characters.
	sep := 0
	for _, r := range s {
		if r == Separator {
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
	return s[AppPasswordLength-1] == checksum.Char([]byte(s[:AppPasswordLength-1]))
}

// JoinToken generates a random, human-friendly cluster join token.
// The token has a length of 24 characters and is separated by 2 dashes.
//
// Example: pGVplw8-eISgkdQN-Mep62nQ
func JoinToken() string {
	m := big.NewInt(int64(len(CharsetBase62)))
	token := make([]byte, 0, JoinTokenLength)

	for i := 0; i < JoinTokenLength-1; i++ {
		if isJoinTokenSeparatorIndex(i) {
			token = append(token, Separator)
			continue
		}

		ch := CharsetBase62[0]
		if r, err := rand.Int(rand.Reader, m); err == nil {
			ch = CharsetBase62[r.Int64()]
		}

		token = append(token, ch)
	}

	token = append(token, checksum.Char(token))

	return string(token)
}

func isJoinTokenSeparatorIndex(i int) bool {
	for _, pos := range joinTokenSeparators {
		if i == pos {
			return true
		}
	}

	return false
}

// IsJoinToken checks if the string represents a join token.
func IsJoinToken(s string, strict bool) bool {
	// Basic mode: No token, not valid.
	if s == "" {
		return false
	}

	// Non-strict mode: only enforce minimum length so legacy tokens that were
	// longer than the auto-generated format continue to work.
	if !strict {
		return len(s) >= JoinTokenLength
	}

	// Strict validation enforces canonical formatting and checksum.
	if len(s) != JoinTokenLength {
		return false
	}

	sep := 0
	for idx, r := range s {
		if r == Separator {
			if !isJoinTokenSeparatorIndex(idx) {
				return false
			}
			sep++
			continue
		}

		if (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}

	if sep != len(joinTokenSeparators) {
		return false
	}

	return s[JoinTokenLength-1] == checksum.Char([]byte(s[:JoinTokenLength-1]))
}

// IsAuthAny checks if the string represents a valid auth token or app password.
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

// IsSessionID checks if the string represents a valid session id.
func IsSessionID(id string) bool {
	if len(id) != SessionIdLength {
		return false
	}

	return IsHex(id)
}
