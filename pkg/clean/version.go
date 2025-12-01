package clean

import (
	"fmt"
	"regexp"
	"strings"
)

// VersionRegexp finds semantic version-like strings.
var VersionRegexp = regexp.MustCompile(`(\d+\.)(\d+\.)(\*|\d+)`)
var PostgreSQLVersionRegexp = regexp.MustCompile("PostgreSQL \\d+\\.\\d+(\\.\\d)?")

// Version parses and returns a semantic version string.
func Version(s string) string {
	if s == "" {
		return ""
	}

	// Find version string with regular expression
	// and return it with "v" prefix if found.
	if v := PostgreSQLVersionRegexp.FindString(s); v != "" {
		return fmt.Sprintf("v%s", strings.Replace(v, "PostgreSQL ", "", -1))
	} else {
		if v := VersionRegexp.FindString(s); v != "" {
			return fmt.Sprintf("v%s", v)
		} else {
			return ""
		}
	}

	return ""
}
