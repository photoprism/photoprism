package entity

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt/clip"
)

const (
	// TypeBytes is the byte budget for sanitized type and identifier strings.
	TypeBytes = 64

	// PathBytes is the byte budget shared by the album_path, photo_path, and
	// folders.path columns (all VARBINARY(1024)). Path values are clipped to it
	// on write so the byte-exact comparisons between these columns hold.
	PathBytes = 1024
)

// ToASCII removes all non-ASCII runes from the string.
func ToASCII(s string) string {
	result := make([]rune, 0, len(s))

	for _, r := range s {
		if r <= 127 {
			result = append(result, r)
		}
	}

	return string(result)
}

// Clip trims leading/trailing whitespace and limits the string to maxLen bytes
// without splitting a multi-byte UTF-8 rune.
func Clip(s string, maxLen int) string {
	return clip.Bytes(s, maxLen)
}

// ClipPath limits a filesystem path to the PathBytes byte budget on a UTF-8
// rune boundary, so a multi-byte path cannot overflow the album_path,
// photo_path, or folders.path columns or break the byte-exact comparisons
// between them.
func ClipPath(p string) string {
	return clip.Bytes(p, PathBytes)
}

// ClipType normalizes identifier-like strings by stripping non-ASCII runes and clipping to 32 characters.
func ClipType(s string) string {
	return Clip(ToASCII(s), TypeBytes)
}

// ClipTypeLower lowercases the string before applying ClipType.
func ClipTypeLower(s string) string {
	return ClipType(strings.ToLower(s))
}
