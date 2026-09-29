package dsn

import (
	"fmt"
	"strings"
)

// ParamRules maps the DSN query parameters a caller accepts to a check of their values.
type ParamRules map[string]func(string) bool

// FilterParams returns the parameters of a MySQL/MariaDB DSN query that rules accept, in their original
// order, and the names of those it dropped, with "" for a segment without "=". An accepted parameter with
// an invalid value or more than one occurrence is an error, so that a setting is never silently weakened.
func FilterParams(query string, rules ParamRules) (params string, dropped []string, err error) {
	kept := make([]string, 0, len(rules))
	seen := make(map[string]bool, len(rules))

	for param := range strings.SplitSeq(query, "&") {
		if param == "" {
			continue
		}

		key, value, found := strings.Cut(param, "=")
		valid, ok := rules[key]

		switch {
		case !ok && !found:
			// A segment without a name is reported with an empty one, so its text is never logged.
			dropped = append(dropped, "")
		case !ok:
			dropped = append(dropped, key)
		case seen[key]:
			return "", nil, fmt.Errorf("duplicate dsn parameter %s", key)
		case !valid(value):
			return "", nil, fmt.Errorf("invalid dsn parameter %s", key)
		default:
			seen[key] = true
			kept = append(kept, param)
		}
	}

	return strings.Join(kept, "&"), dropped, nil
}

// MergeParams appends the parameters in defaults whose names params does not contain.
func MergeParams(params, defaults string) string {
	names := make(map[string]bool)

	for param := range strings.SplitSeq(params, "&") {
		key, _, _ := strings.Cut(param, "=")
		names[key] = true
	}

	merged := params

	for param := range strings.SplitSeq(defaults, "&") {
		key, _, _ := strings.Cut(param, "=")

		if param == "" || names[key] {
			continue
		}

		if merged != "" {
			merged += "&"
		}

		merged += param
		names[key] = true
	}

	return merged
}

// HasParam reports whether a DSN query contains the parameter with the specified name.
func HasParam(params, name string) bool {
	for param := range strings.SplitSeq(params, "&") {
		if key, _, _ := strings.Cut(param, "="); key == name {
			return true
		}
	}

	return false
}
