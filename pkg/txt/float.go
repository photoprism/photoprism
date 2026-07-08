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

// Float64 converts a string to a 64-bit floating point number or 0 if invalid.
func Float64(s string) float64 {
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
	return float32(Float64(s))
}

// FloatRange parses a string as floating point number range and returns an error if it's not a valid range.
func FloatRange(s string, min, max float64) (start float64, end float64, err error) {
	if s == "" || len(s) > 40 {
		return start, end, errors.New("invalid range")
	}

	valid := false

	p := 0
	startValue := make([]byte, 0, 20)
	endValue := make([]byte, 0, 20)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' {
			if i == 0 || p == 1 {
				if p == 0 {
					startValue = append(startValue, c)
				} else {
					endValue = append(endValue, c)
				}
			} else {
				p = 1
			}
		}
		if c == '.' || c >= '0' && c <= '9' {
			valid = true
			if p == 0 {
				startValue = append(startValue, c)
			} else {
				endValue = append(endValue, c)
			}
		}
	}

	if !valid {
		return start, end, errors.New("invalid range")
	}

	if p == 0 {
		start = Float64(string(startValue))
		end = start
	} else {
		start = Float64(string(startValue))
		end = Float64(string(endValue))
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
