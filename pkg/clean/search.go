package clean

import (
	"strings"
)

// spaced returns the string padded with a space left and right.
func spaced(s string) string {
	return Space + s + Space
}

// replace performs a case-insensitive string replacement.
// replaceFoldASCII replaces all case-insensitive ASCII matches of needle
// in s with repl. It avoids regex compilation and extra allocations.
func replaceFoldASCII(s, needle, repl string) string {
	if s == "" || needle == "" {
		return s
	}

	// Quick check to see if there's any possible match using a lowercased scan.
	// We implement a simple ASCII case-insensitive search.
	toLower := func(b byte) byte {
		if b >= 'A' && b <= 'Z' {
			return b + 32
		}
		return b
	}

	nl := len(needle)
	// Precompute lower-case needle bytes.
	nb := make([]byte, nl)
	for i := 0; i < nl; i++ {
		nb[i] = toLower(needle[i])
	}

	// First pass: find if any match exists; if not, return s unchanged.
	// Second pass: build result with replacements.
	// Implement both in one pass by building only when the first match is seen.
	var out []byte
	i := 0
	last := 0
	for i <= len(s)-nl {
		// Compare at position i.
		j := 0
		for ; j < nl; j++ {
			if toLower(s[i+j]) != nb[j] {
				break
			}
		}
		if j == nl {
			// Match found.
			if out == nil {
				// Allocate with an estimate: original len.
				out = make([]byte, 0, len(s))
			}
			out = append(out, s[last:i]...)
			out = append(out, repl...)
			i += nl
			last = i
			continue
		}
		i++
	}
	if out == nil {
		return s
	}
	// Append the tail.
	out = append(out, s[last:]...)
	return string(out)
}

// SearchString replaces search operator with default symbols.
func SearchString(s string) string {
	if s == "" || reject(s, LengthLimit) {
		return Empty
	}

	// Normalize.
	s = strings.ReplaceAll(s, "%%", "%")
	s = strings.ReplaceAll(s, "%", "*")
	s = strings.ReplaceAll(s, "**", "*")

	// Trim.
	return strings.Trim(s, "|\\<>\n\r\t")
}

// SearchQuery replaces search operator with default symbols.
func SearchQuery(s string) string {
	if s == "" || reject(s, LengthLimit) {
		return Empty
	}

	// Normalize.
	s = replaceFoldASCII(s, spaced(EnOr), Or)
	s = replaceFoldASCII(s, spaced(EnOr), Or)
	s = replaceFoldASCII(s, spaced(EnAnd), And)
	s = replaceFoldASCII(s, spaced(EnWith), And)
	s = replaceFoldASCII(s, spaced(EnIn), And)
	s = replaceFoldASCII(s, spaced(EnAt), And)
	s = strings.ReplaceAll(s, SpacedPlus, And)
	s = strings.ReplaceAll(s, "%%", "%")
	s = strings.ReplaceAll(s, "%", "*")
	s = strings.ReplaceAll(s, "**", "*")

	// Trim.
	return strings.Trim(s, "|${}\\<>: \n\r\t")
}
