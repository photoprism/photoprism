package entity

import (
	"reflect"
	"strings"
)

const (
	TrimTypeString = 32
)

// Values is a shortcut for map[string]interface{}
type Values map[string]interface{}

// GetValues extracts entity Values.
func GetValues(m interface{}, omit ...string) (result Values) {
	skip := func(name string) bool {
		for _, s := range omit {
			if name == s {
				return true
			}
		}

		return false
	}

	result = make(map[string]interface{})

	elem := reflect.ValueOf(m).Elem()
	relType := elem.Type()

	for i := 0; i < relType.NumField(); i++ {
		name := relType.Field(i).Name

		if skip(name) {
			continue
		}

		result[name] = elem.Field(i).Interface()
	}

	return result
}

// ToASCII removes all non-ascii characters from a string and returns it.
func ToASCII(s string) string {
	result := make([]rune, 0, len(s))

	for _, r := range s {
		if r <= 127 {
			result = append(result, r)
		}
	}

	return string(result)
}

// Trim shortens a string to the given number of characters, and removes all leading and trailing white space.
func Trim(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	l := len(s)

	if l <= maxLen {
		return s
	} else {
		return s[:l-1]
	}
}

// SanitizeTypeString converts a type string to lowercase, omits invalid runes, and shortens it if needed.
func SanitizeTypeString(s string) string {
	return Trim(ToASCII(strings.ToLower(s)), TrimTypeString)
}

// TypeString returns an entity type string for logging.
func TypeString(entityType string) string {
	if entityType == "" {
		return "unknown"
	}

	return entityType
}
