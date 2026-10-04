package vision

import (
	"os"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Environment variable name suffixes that the logged service fields expand.
var (
	modelEnvSuffixes = []string{"_MODEL"}
	uriEnvSuffixes   = []string{"_URL", "_URI", "_HOST"}
)

// refusedEnvWarned holds the field, value and variable combinations that were logged.
var refusedEnvWarned sync.Map

// expandEnvSuffix expands the environment variables in s whose names end in one of the suffixes. Other
// variables expand to an empty string, and their names are returned as refused.
func expandEnvSuffix(s string, suffixes []string) (expanded string, refused []string) {
	if !strings.Contains(s, "$") {
		return s, nil
	}

	ensureEnv()

	expanded = os.Expand(s, func(name string) string {
		if envNameAllowed(name, suffixes) {
			return os.Getenv(name)
		}

		refused = append(refused, name)

		return ""
	})

	return expanded, refused
}

// envNameAllowed reports whether the name is a valid variable name that ends in one of the suffixes.
func envNameAllowed(name string, suffixes []string) bool {
	if !envNameValid(name) {
		return false
	}

	for _, suffix := range suffixes {
		if len(name) > len(suffix) && strings.HasSuffix(name, suffix) {
			return true
		}
	}

	return false
}

// envNameValid reports whether the name consists of ASCII letters, digits, and underscores and does
// not start with a digit. os.Expand passes anything between "${" and "}" as a name, e.g. "VAR:-default".
func envNameValid(name string) bool {
	if name == "" || name[0] >= '0' && name[0] <= '9' {
		return false
	}

	for i := 0; i < len(name); i++ {
		if c := name[i]; c != '_' && (c < '0' || c > '9') && (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}

	return true
}

// forgetRefusedEnv removes the logged refusals of a field and configured value, so they are logged again.
func forgetRefusedEnv(field, value string) {
	prefix := field + "\x00" + value + "\x00"

	refusedEnvWarned.Range(func(key, _ any) bool {
		if k, ok := key.(string); ok && strings.HasPrefix(k, prefix) {
			refusedEnvWarned.Delete(key)
		}

		return true
	})
}

// warnRefusedEnv logs a warning once per field, configured value, and refused variable name. Only the
// name of a variable that is set is quoted, as other text after a "$" may be part of a literal value.
func warnRefusedEnv(field, value string, refused []string, suffixes []string) {
	for _, name := range refused {
		if _, warned := refusedEnvWarned.LoadOrStore(field+"\x00"+value+"\x00"+name, struct{}{}); warned {
			continue
		}

		if _, set := os.LookupEnv(name); set && envNameValid(name) {
			log.Warnf("vision: %s does not expand %s, as only variables ending in %s are expanded there",
				field, clean.Log(name), txt.JoinOr(suffixes))
		} else {
			log.Warnf("vision: %s contains an unset or invalid variable reference, which is not expanded", field)
		}
	}
}
