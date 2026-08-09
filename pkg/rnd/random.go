package rnd

import (
	"crypto/rand"
)

// maxCharsetLen is the largest charset randomChars can sample without bias,
// since it draws one byte per character.
const maxCharsetLen = 256

// randomFill fills b with cryptographically secure random bytes, and panics if the system entropy
// source is unavailable, so that callers always receive either a fully random buffer or no result
// at all. As of Go 1.24 crypto/rand.Read never returns an error and aborts the process itself, so
// the panic is unreachable in practice and keeps the guarantee independent of the Go version.
func randomFill(b []byte) {
	if _, err := rand.Read(b); err != nil {
		panic("rnd: random number generator unavailable (" + err.Error() + ")")
	}
}

// randomChars fills b with characters drawn uniformly from charset, which must hold between 1 and
// maxCharsetLen bytes. It panics on an out-of-range charset, which is a programming error rather
// than a runtime condition.
func randomChars(b []byte, charset string) {
	n := len(charset)

	if n < 1 || n > maxCharsetLen {
		panic("rnd: charset must contain 1 to 256 characters")
	}

	// Byte values in the incomplete final block are rejected and redrawn, so that every character
	// stays equally likely. Reducing with a plain modulo would favor the first 256%n characters.
	limit := maxCharsetLen - maxCharsetLen%n

	randomFill(b)

	var redraw [1]byte

	for i, c := range b {
		for int(c) >= limit {
			randomFill(redraw[:])
			c = redraw[0]
		}

		// #nosec G602 -- i is a range index over b, and c%n is bounded by len(charset).
		b[i] = charset[int(c)%n]
	}
}
