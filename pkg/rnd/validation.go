package rnd

import (
	"strconv"
	"strings"
	"time"
)

// IsUnique checks if the string is a valid unique ID.
func IsUnique(s string, prefix byte) bool {
	if IsUUID(s) {
		// Standard UUID.
		return true
	}

	return IsUID(s, prefix)
}

// IsUID checks if the string is a valid entity UID.
func IsUID(s string, prefix byte) bool {
	// Check length.
	if len(s) != 16 {
		return false
	}

	// Check prefix.
	if prefix != 0 && s[0] != prefix {
		return false
	}

	// Check character range.
	if !IsAlnum(s) {
		return false
	}

	// Check timestamp.
	if ts := s[1:7]; ts == "000000" {
		return true
	} else if t, err := strconv.ParseInt(ts, 36, 64); err != nil {
		return false
	} else if t < 1483228800 || t > time.Now().UTC().Unix() {
		return false
	}

	// Valid.
	return true
}

// InvalidUID checks if the UID is empty or invalid.
func InvalidUID(s string, prefix byte) bool {
	return !IsUID(s, prefix)
}

// IsUUID tests if the string looks like a standard UUID.
func IsUUID(s string) bool {
	return len(s) == 36 && IsHex(s)
}

// IsCanonicalUUID tests if the string is a lowercase UUID in the canonical 8-4-4-4-12 layout.
// Use it for a value that identifies a record, so the separator positions are fixed and storage
// cannot reinterpret the value. IsUUID is the lenient variant, for UUIDs read from XMP or Exif.
func IsCanonicalUUID(s string) bool {
	if len(s) != 36 {
		return false
	}

	for i, r := range s {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
				return false
			}
		}
	}

	return true
}

// SanitizeUUID normalizes UUIDs found in XMP or Exif metadata.
func SanitizeUUID(s string) string {
	if s == "" {
		return ""
	}

	s = strings.ReplaceAll(strings.TrimSpace(s), "\"", "")

	if start := strings.LastIndex(s, ":"); start != -1 {
		s = s[start+1:]
	}

	if !IsUUID(s) {
		return ""
	}

	return strings.ToLower(s)
}

// IsAlnum returns true if the string only contains alphanumeric ascii chars without whitespace.
func IsAlnum(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}

	return true
}

// IsHex returns true if the string only contains hex numbers, dashes and letters without whitespace.
func IsHex(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') && r != '-' {
			return false
		}
	}

	return true
}
