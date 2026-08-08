package rnd

import (
	"github.com/photoprism/photoprism/pkg/checksum"
)

// CharsetBase10 contains digits 0-9.
const CharsetBase10 = checksum.CharsetBase10

// CharsetBase36 contains lowercase letters and digits.
const CharsetBase36 = checksum.CharsetBase36

// CharsetBase62 contains upper/lowercase letters and digits.
const CharsetBase62 = checksum.CharsetBase62

// Base10 generates a random token containing numbers only.
func Base10(length int) string {
	return Charset(length, CharsetBase10)
}

// Base36 generates a random token containing lowercase letters and numbers.
func Base36(length int) string {
	return Charset(length, CharsetBase36)
}

// Base62 generates a random token containing upper and lower case letters as well as numbers.
func Base62(length int) string {
	return Charset(length, CharsetBase62)
}

// Charset generates a random token with the specified length and charset.
// It returns an empty string for a non-positive length or an unusable charset.
func Charset(length int, charset string) string {
	if length < 1 || len(charset) < 1 || len(charset) > maxCharsetLen {
		return ""
	} else if length > 4096 {
		length = 4096
	}

	b := make([]byte, length)
	randomChars(b, charset)

	return string(b)
}
