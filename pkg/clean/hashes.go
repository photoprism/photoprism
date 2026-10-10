package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	// HashHexMinLen is the length of the shortest hex-encoded checksum MaskHashes masks, an MD5 digest.
	HashHexMinLen = 32
	// HashHexPrefixLen is the number of leading hex digits MaskHashes keeps, as the cache folder names do.
	HashHexPrefixLen = 3
)

// MaskHashes masks every run of at least HashHexMinLen hex digits in s except for its first
// HashHexPrefixLen digits, e.g. "2ca***", which covers checksums such as MD5, SHA1, and SHA256 digests.
func MaskHashes(s string) string {
	if len(s) < HashHexMinLen {
		return s
	}

	var b strings.Builder

	last := 0

	for i := 0; i < len(s); {
		if !isHexByte(s[i]) {
			i++
			continue
		}

		j := i

		for j < len(s) && isHexByte(s[j]) {
			j++
		}

		if j-i >= HashHexMinLen {
			if last == 0 {
				b.Grow(len(s))
			}

			b.WriteString(s[last : i+HashHexPrefixLen])
			b.WriteString(txt.Masked)
			last = j
		}

		i = j
	}

	if last == 0 {
		return s
	}

	b.WriteString(s[last:])

	return b.String()
}

// isHexByte reports whether b is a hex digit in either case.
func isHexByte(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}
