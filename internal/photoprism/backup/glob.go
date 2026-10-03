package backup

import (
	"path/filepath"
	"strings"
)

// globIn returns the names in dir that match pattern. The dir is made absolute and only the characters
// filepath.Match treats as special are escaped, so the dir matches itself without its ancestors being listed.
func globIn(dir, pattern string) ([]string, error) {
	absDir, err := filepath.Abs(dir)

	if err != nil {
		return nil, err
	}

	return filepath.Glob(filepath.Join(globEscape(absDir), pattern))
}

// globEscape escapes the characters filepath.Match treats as special.
func globEscape(s string) string {
	var b strings.Builder

	// Bytes rather than runes, so a name that is not valid UTF-8 keeps its bytes.
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '*', '?', '[':
			b.WriteByte('\\')
		}

		b.WriteByte(s[i])
	}

	return b.String()
}
