package clean

import (
	"path"
	"strings"
	"unicode/utf8"
)

// pathReserved lists the characters Path removes from a path, in addition to the unsafe ones.
const pathReserved = `\|"?*%<>{}`

// portableReserved lists the characters a portable file or folder name must not contain: those Path removes,
// ":", which several file systems and command-line tools reserve, and "~", which stands for a home directory.
const portableReserved = pathReserved + ":~"

// Path removes reserved and unsafe characters from a path string and trims ASCII spaces, keeping the
// other characters as they are. It returns an empty string if the result contains ".." or "//".
func Path(s string) string {
	if s == "" || reject(s, 512) {
		return ""
	}

	s = strings.Trim(pathString(s, pathReserved), " ")

	if reject(s, 0) || strings.Contains(s, "..") || strings.Contains(s, "//") {
		return ""
	}

	return s
}

// pathString removes the runes reported by unsafePathRune and those in reserved from a file or path name,
// and keeps every other character and invalid byte as it is. It repeats until nothing is removed, since
// invalid bytes around a removed rune can form a new one. reserved must not contain U+FFFD.
func pathString(s, reserved string) string {
	for {
		var b []byte

		for i := 0; i < len(s); {
			r, n := utf8.DecodeRuneInString(s[i:])

			if unsafePathRune(r) || strings.ContainsRune(reserved, r) {
				if b == nil {
					b = append(make([]byte, 0, len(s)), s[:i]...)
				}
			} else if b != nil {
				b = append(b, s[i:i+n]...)
			}

			i += n
		}

		if b == nil {
			return s
		}

		s = string(b)
	}
}

// UserPath sanitizes and normalizes user-provided paths for use as input.
func UserPath(dir string) string {
	if dir == "" {
		return dir
	}

	// The following characters must never be used in a filepath: \ > < | : &
	dir = strings.Trim(path.Clean(Path(strings.ReplaceAll(dir, "\\", "/"))), " ~./\\|?*%<>")

	if strings.Contains(dir, "/.") || strings.Contains(dir, "..") || strings.Contains(dir, "//") {
		return ""
	}

	return dir
}

// SlashPath normalizes a slash-based relative path by converting backslashes
// to slashes, trimming whitespace, and removing leading and trailing slashes.
func SlashPath(dir string) string {
	dir = strings.TrimSpace(strings.ReplaceAll(dir, "\\", "/"))
	return strings.Trim(dir, "/")
}

// PortableName reports whether a single file or folder name is valid UTF-8 and contains none of the characters
// PhotoPrism reserves in file and folder names. It does not check for "/", ".", "..", or an empty name.
func PortableName(name string) bool {
	if !utf8.ValidString(name) || strings.ContainsAny(name, portableReserved) {
		return false
	}

	return !strings.ContainsFunc(name, unsafePathRune)
}
