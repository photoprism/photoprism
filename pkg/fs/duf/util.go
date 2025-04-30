package duf

import (
	"fmt"
	"slices"
	"strings"

	"github.com/IGLOU-EU/go-wildcard"
)

// parseCommaSeparatedValues parses comma separated string into a map.
func parseCommaSeparatedValues(values string) FilterValues {
	m := make(FilterValues)
	for _, v := range strings.Split(values, ",") {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}

		v = strings.ToLower(v)
		m[v] = struct{}{}
	}

	return m
}

// validateGroups validates the parsed group maps.
func validateGroups(m FilterValues) error {
	for k := range m {
		found := slices.Contains(groups, k)

		if !found {
			return fmt.Errorf("unknown device group: %s", k)
		}
	}

	return nil
}

// findInKey parse a slice of pattern to match the given key.
func findInKey(str string, km FilterValues) bool {
	for p := range km {
		if wildcard.Match(p, str) {
			return true
		}
	}

	return false
}
