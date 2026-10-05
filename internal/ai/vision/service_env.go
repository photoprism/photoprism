package vision

import (
	"os"
	"strconv"
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
	return expandEnvValues(s, suffixes, nil)
}

// expandEnvValues expands the environment variables in s whose names end in one of the suffixes, passing
// each value through fn, if set. Other variables expand to an empty string, and their names are returned.
func expandEnvValues(s string, suffixes []string, fn func(string) string) (expanded string, refused []string) {
	if !strings.Contains(s, "$") {
		return s, nil
	}

	ensureEnv()

	expanded = os.Expand(s, func(name string) string {
		if !envNameAllowed(name, suffixes) {
			refused = append(refused, name)
			return ""
		} else if fn != nil {
			return fn(os.Getenv(name))
		}

		return os.Getenv(name)
	})

	return expanded, refused
}

// expandUriEnv expands the variables in a service URI like expandEnvSuffix. A value that ends the path part
// of the URI has its fragment dropped and its query moved to the end, after the URI's own query, so the text
// after a base URL variable stays in the path. Other values are expanded as written.
func expandUriEnv(s string) (expanded string, refused []string) {
	if strings.Contains(s, "\x00") {
		return expandEnvSuffix(s, uriEnvSuffixes)
	}

	var suffixes []string

	// The query and fragment of each value are replaced by numbered markers, which are resolved once
	// their position is known.
	expanded, refused = expandEnvValues(s, uriEnvSuffixes, func(value string) string {
		if i := strings.IndexAny(value, "?#"); i >= 0 {
			suffixes = append(suffixes, value[i:])
			return value[:i] + "\x00" + strconv.Itoa(len(suffixes)-1) + "\x00"
		}

		return value
	})

	if len(suffixes) == 0 {
		return expanded, refused
	}

	var moved []string
	var b strings.Builder

	for rest := expanded; rest != ""; {
		before, marker, found := strings.Cut(rest, "\x00")
		b.WriteString(before)

		if !found {
			break
		}

		index, after, _ := strings.Cut(marker, "\x00")
		suffix := suffixes[txt.Int(index)]
		rest = after

		if !uriPathEnd(b.String(), after) {
			b.WriteString(suffix)
		} else if query, _, _ := strings.Cut(suffix, "#"); len(query) > 1 {
			moved = append(moved, query[1:])
		}
	}

	if expanded = b.String(); len(moved) == 0 {
		return expanded, refused
	}

	expanded, fragment, hasFragment := strings.Cut(expanded, "#")

	switch {
	case !strings.Contains(expanded, "?"):
		expanded += "?"
	case !strings.HasSuffix(expanded, "?") && !strings.HasSuffix(expanded, "&"):
		expanded += "&"
	}

	expanded += strings.Join(moved, "&")

	if hasFragment {
		expanded += "#" + fragment
	}

	return expanded, refused
}

// uriPathEnd reports whether the position between before and after ends the path part of a URI, i.e. its
// authority started in before, which has no query or fragment, and after continues with a path, query, or
// fragment, if anything.
func uriPathEnd(before, after string) bool {
	if !strings.Contains(before, "//") || strings.ContainsAny(before, "?#") {
		return false
	}

	return after == "" || strings.ContainsAny(after[:1], "/?#")
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

// warnRefusedEnv reports a warning once per field, configured value, and refused variable name. Only the
// name of a variable that is set is quoted, as other text after a "$" may be part of a literal value.
func warnRefusedEnv(field, value string, refused []string, suffixes []string) {
	for _, name := range refused {
		if _, warned := refusedEnvWarned.LoadOrStore(field+"\x00"+value+"\x00"+name, struct{}{}); warned {
			continue
		}

		if _, set := os.LookupEnv(name); set && envNameValid(name) {
			warnModel("%s does not expand %s, as only variables ending in %s are expanded there",
				field, clean.Log(name), txt.JoinOr(suffixes))
		} else {
			warnModel("%s contains an unset or invalid variable reference, which is not expanded", field)
		}
	}
}
