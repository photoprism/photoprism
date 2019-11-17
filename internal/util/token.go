package util

import (
	"crypto/rand"
	"fmt"
)

func RandomToken(size int) (string, error) {
	b := make([]byte, size)

	_, err := rand.Read(b)

	return fmt.Sprintf("%x", b), err
}
