package rnd

import (
	"crypto/rand"
	"fmt"
	"log"
)

// SessionID returns a new session id.
func SessionID() string {
	b := make([]byte, 24)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", b)
}

// IsSessionID checks if the string is a session id.
func IsSessionID(s string) bool {
	if len(s) != 48 {
		return false
	}

	return IsHex(s)
}
