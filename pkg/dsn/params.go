package dsn

import (
	"fmt"
	"regexp"
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

// paramNameRegex matches a DSN parameter name that may be logged.
var paramNameRegex = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

// LoggableParamNames returns the names that look like DSN parameter names, so a query segment without
// a name, such as a stray value, is counted but never logged.
func LoggableParamNames(names []string) []string {
	result := make([]string, 0, len(names))

	for _, name := range names {
		if paramNameRegex.MatchString(name) {
			result = append(result, name)
		}
	}

	return result
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

// Query returns the parameters of a MySQL/MariaDB DSN as the driver reads them, i.e. the part after the
// first "?" that follows the last "/".
func Query(s string) string {
	_, query, _ := strings.Cut(s[strings.LastIndex(s, "/")+1:], "?")
	return query
}
