package rnd

import (
	"crypto/rand"
	"math/big"
)

const CharsetBase36 = "abcdefghijklmnopqrstuvwxyz0123456789"
const CharsetBase62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Base36 generates a random token containing lowercase letters and numbers.
func Base36(length int) string {
	return Charset(length, CharsetBase36)
}

// Base62 generates a random token containing upper and lower case letters as well as numbers.
func Base62(length int) string {
	return Charset(length, CharsetBase62)
}

// Charset generates a random token with the specified length and charset.
func Charset(length int, charset string) string {
	if length < 1 {
		return ""
	} else if length > 4096 {
		length = 4096
	}

	m := big.NewInt(int64(len(charset)))
	b := make([]byte, length)

	for i := range b {
		if r, err := rand.Int(rand.Reader, m); err == nil {
			b[i] = charset[r.Int64()]
		}
	}

	return string(b)
}
