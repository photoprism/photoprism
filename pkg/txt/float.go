package txt

import (
	"errors"
	"strconv"
	"strings"
)

// IsFloat checks if the string represents a floating point number.
func IsFloat(s string) bool {
	if s == "" {
		return false
	}

	s = strings.TrimSpace(s)

	for _, r := range s {
		if r != '.' && r != ',' && (r < '0' || r > '9') {
			return false
		}
	}

	return true
}

// Float converts a string to a 64-bit floating point number or 0 if invalid.
func Float(s string) float64 {
	if s == "" {
		return 0
	}

	f, err := strconv.ParseFloat(Numeric(s), 64)

	if err != nil {
		return 0
	}

	return f
}

// Float32 converts a string to a 32-bit floating point number or 0 if invalid.
func Float32(s string) float32 {
	return float32(Float(s))
}

// FloatRange parses a string as floating point number range and returns an error if it's not a valid range.
func FloatRange(s string, min, max float64) (start float64, end float64, err error) {
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
		start = Float(string(v[0]))
		end = start
	} else {
		start = Float(string(v[0]))
		end = Float(string(v[1]))
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
