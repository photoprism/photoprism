package rnd

import "strings"

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
	if len(s) != 16 {
		return false
	}

	if !IsAlnum(s) {
		return false
	}

	return prefix == 0 || s[0] == prefix
}

// InvalidUID checks if the UID is empty or invalid.
func InvalidUID(s string, prefix byte) bool {
	return !IsUID(s, prefix)
}

// IsUUID tests if the string looks like a standard UUID.
func IsUUID(s string) bool {
	return len(s) == 36 && IsHex(s)
}

// SanitizeUUID normalizes UUIDs found in XMP or Exif metadata.
func SanitizeUUID(s string) string {
	if s == "" {
		return ""
	}

	s = strings.Replace(strings.TrimSpace(s), "\"", "", -1)

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
		if (r < 48 || r > 57) && (r < 97 || r > 102) && (r < 65 || r > 70) && r != 45 {
			return false
		}
	}

	return true
}
