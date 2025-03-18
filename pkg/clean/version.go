package clean

import (
	"fmt"
	"regexp"
)

var VersionRegexp = regexp.MustCompile("(\\d+\\.)(\\d+\\.)(\\*|\\d+)")

// Version parses and returns a semantic version string.
func Version(s string) string {
	if s == "" {
		return ""
	}

	// Find version string with regular expression
	// and return it with "v" prefix if found.
	if v := VersionRegexp.FindString(s); v != "" {
		return fmt.Sprintf("v%s", v)
	} else {
		return ""
	}
}
