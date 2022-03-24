package txt

import (
	"strconv"
	"strings"
)

// Int converts a string to a signed integer or 0 if invalid.
func Int(s string) int {
	if s == "" {
		return 0
	}

	result, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)

	if err != nil {
		return 0
	}

	return int(result)
}

// UInt converts a string to an unsigned integer or 0 if invalid.
func UInt(s string) uint {
	if s == "" {
		return 0
	}

	result, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32)

	if err != nil || result < 0 {
		return 0
	}

	return uint(result)
}

// IsUInt tests if a string represents an unsigned integer.
func IsUInt(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < 48 || r > 57 {
			return false
		}
	}

	return true
}

// IsPosInt checks if a string represents an integer greater than 0.
func IsPosInt(s string) bool {
	if s == "" || s == " " || s == "0" || s == "-1" {
		return false
	}

	for _, r := range s {
		if r < 48 || r > 57 {
			return false
		}
	}

	return true
}
