package util

import (
	"crypto/rand"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func RandomToken(size int) string {
	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", b)
}

func UUID() string {
	return uuid.NewV4().String()
}
