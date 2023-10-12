package txt

import (
	"errors"
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

// IntVal converts a string to a validated integer or a default if invalid.
func IntVal(s string, min, max, def int) (i int) {
	if s == "" {
		return def
	} else if s[0] == ' ' {
		s = strings.TrimSpace(s)
	}

	result, err := strconv.ParseInt(s, 10, 32)

	if err != nil {
		return def
	}

	i = int(result)

	if i < min {
		return def
	} else if max != 0 && i > max {
		return def
	}

	return i
}

// IntRange parses a string as integer range and returns an error if it's not a valid range.
func IntRange(s string, min, max int) (start int, end int, err error) {
	if s == "" || len(s) > 40 {
		return start, end, errors.New("invalid range")
	}

	valid := false

	p := 0
	v := [][]byte{make([]byte, 0, 20), make([]byte, 0, 20)}

	for i, r := range s {
		if r == 45 {
			if i == 0 || p == 1 {
				v[p] = append(v[p], byte(r))
			} else if p == 0 {
				p = 1
			}
		}
		if r == 46 || r >= 48 && r <= 57 {
			valid = true
			v[p] = append(v[p], byte(r))
		}
	}

	if !valid {
		return start, end, errors.New("invalid range")
	}

	if p == 0 {
		start = Int(string(v[0]))
		end = start
	} else {
		start = Int(string(v[0]))
		end = Int(string(v[1]))
	}

	if start > max {
		start = max
	} else if start < min {
		start = min
	}

	if end > max {
		end = max
	} else if end < min {
		end = min
	}

	return start, end, nil
}

// UInt converts a string to an unsigned integer or 0 if invalid.
func UInt(s string) uint {
	if s == "" {
		return 0
	} else if s[0] == ' ' {
		s = strings.TrimSpace(s)
	}

	result, err := strconv.ParseInt(s, 10, 32)

	if err != nil || result < 0 {
		return 0
	}

	return uint(result)
}

// Int64 converts a string to a signed 64-bit integer or 0 if invalid.
func Int64(s string) int64 {
	if s == "" {
		return 0
	}

	i := strings.SplitN(Numeric(s), ".", 2)

	result, err := strconv.ParseInt(i[0], 10, 64)

	if err != nil {
		return 0
	}

	return result
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
