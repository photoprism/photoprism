package rnd

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

// Sha224 returns the SHA224 checksum of the byte slice as a hex string.
func Sha224(b []byte) string {
	return fmt.Sprintf("%x", sha256.Sum224(b))
}

// Sha256 returns the SHA256 checksum of the byte slice as a hex string.
func Sha256(b []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

// Sha512 returns the SHA512 checksum of the byte slice as a hex string.
func Sha512(b []byte) string {
	return fmt.Sprintf("%x", sha512.Sum512(b))
}
