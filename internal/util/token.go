package util

import (
	"crypto/rand"
	"fmt"
)

func RandomToken(size int) string {
	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", b)
}
