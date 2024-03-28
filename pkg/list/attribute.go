package list

import (
	"fmt"
	"strings"
)

// KeyValue represents a key-value attribute.
type KeyValue struct {
	Key   string
	Value string
}

// ParseKeyValue creates a new key-value attribute from a string.
func ParseKeyValue(s string) *KeyValue {
	f := KeyValue{}

	return f.Parse(s)
}

// Key returns the sanitized attribute key.
func Key(s string) string {
	var i = 0

	// Remove non-alphanumeric characters.
	result := strings.Map(func(r rune) rune {
		switch r {
		case '.', '@', '-', '+', '_', '#':
			return r
		case '*':
			i++
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		} else if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') || r == 127 {
			return -1
		}
		i++
		return r
	}, s)

	if i == 0 {
		return ""
	}

	return result
}

// Value returns the sanitized attribute value.
func Value(s string) string {
	// Remove control characters.
	return strings.Map(func(r rune) rune {
		if r <= 32 || r == 127 {
			return -1
		}
		switch r {
		case '(', ')', '<', '>', '\'', '"', '*':
			return r
		}
		return r
	}, s)
}

// Parse parses a string attribute into key and value.
func (f *KeyValue) Parse(s string) *KeyValue {
	k, v, _ := strings.Cut(s, ":")

	// Set key.
	if k = Key(k); k == "" {
		return nil
	} else {
		f.Key = k
	}

	// Default?
	if f.Key == All {
		return f
	} else if v = Value(v); v == "" {
		f.Value = True
		return f
	}

	// Set value.
	if b := Bool[strings.ToLower(v)]; b != "" {
		f.Value = b
	} else {
		f.Value = v
	}

	return f
}

// String return the attribute as string.
func (f *KeyValue) String() string {
	if f == nil {
		return ""
	}

	if f.Key == All {
		return All
	}

	if Bool[strings.ToLower(f.Value)] == True {
		return f.Key
	}

	if s := fmt.Sprintf("%s:%s", f.Key, f.Value); len(s) < StringLengthLimit {
		return s
	}

	return ""
}
