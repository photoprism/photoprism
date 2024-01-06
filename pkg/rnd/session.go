package rnd

import (
	"crypto/rand"
	"fmt"
	"log"
)

// AuthToken returns a new session id.
func AuthToken() string {
	b := make([]byte, 24)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", b)
}

// IsAuthToken checks if the string is a session id.
func IsAuthToken(s string) bool {
	if len(s) != 48 {
		return false
	}

	return IsHex(s)
}

// SessionID returns the hashed session id string.
func SessionID(s string) string {
	return Sha256([]byte(s))
}

// IsSessionID checks if the string is a session id string.
func IsSessionID(s string) bool {
	if len(s) != 64 {
		return false
	}

	return IsHex(s)
}
