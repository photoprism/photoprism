package session

import (
	"crypto/rand"
	"fmt"
)

func NewID() string {
	b := make([]byte, 24)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", b)
}
