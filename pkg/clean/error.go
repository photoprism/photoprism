package clean

import (
	iofs "io/fs"
	"net"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	// errorPathPlaceholder replaces the locations removed by Error.
	errorPathPlaceholder = txt.Masked
	// errorPathSeparators are the characters a value must contain to count as a location.
	errorPathSeparators = `/\`
	// errorPathTrivial are the characters a location must consist of more than.
	errorPathTrivial = `/\.`
	// errorUnwrapNodes bounds how many errors in a chain are inspected.
	errorUnwrapNodes = 256
	// errorOmitted replaces a message whose chain is too large to inspect completely.
	errorOmitted = "error details omitted"
)

// errorNil reports whether an error is absent, treating a non-nil interface that holds a nil
// pointer as absent.
func errorNil(err error) bool {
	if err == nil {
		return true
	}

	v := reflect.ValueOf(err)

	return v.Kind() == reflect.Pointer && v.IsNil()
}

// Error sanitizes an error message so that it can be safely logged or displayed, replacing the
// file paths the errors in its chain carry with a placeholder. A path that a wrapper rendered
// into message text stays, so producers wrap with %w. Use ErrorFull where the reader is an operator.
func Error(err error) string {
	if errorNil(err) {
		return "no error"
	}

	paths, complete := errorPaths(err)

	// A chain too large to inspect completely may name a path this did not collect, so the
	// message is dropped rather than rendered as though it had been. The caller's own text
	// still identifies the site, and the oversized message is never materialized.
	if !complete {
		return errorOmitted
	}

	// Both sides are scrubbed before the replacement, so the copy a producer rendered into the
	// message and the copy the chain carries reduce to the same text and one pass removes both.
	s := UriCredentials(err.Error())

	for _, p := range paths {
		s = strings.ReplaceAll(s, UriCredentials(p), errorPathPlaceholder)
	}

	return errorText(s)
}

// ErrorFull sanitizes an error message and keeps the file paths it names. Reserve it for the
// console and the CLI, where the path is the detail an operator acts on.
func ErrorFull(err error) string {
	if errorNil(err) {
		return "no error"
	}

	return errorText(err.Error())
}

// errorText renders an error message for a reader: it removes the credential of any URL, bounds
// the length, and maps the problematic characters, the field separator among them. That order
// matters, since the character map would otherwise hide a percent-encoded credential from the scrub. A message is a sentence rather than a value, so it is folded into one
// field instead of being quoted the way Log bounds the values it renders.
func errorText(s string) string {
	if s = strings.TrimSpace(s); s == "" {
		return "unknown error"
	}

	// Applied to the whole message, since a URL can appear in one in more than one spelling.
	s = UriCredentials(s)

	// Limit error message length.
	if len(s) > LengthLimit {
		s = s[:LengthLimit]
	}

	// Remove non-printable and other potentially problematic characters.
	s = strings.Map(func(r rune) rune {
		switch {
		case unsafeSpaceRune(r):
			return unsafeSpace
		case unsafeDropRune(r):
			return -1
		case unsafeRune(r):
			return unsafeMarker
		}

		switch r {
		case '`', '"':
			return '\''
		case '%', '\\', '$', '<', '>', '{', '}', FieldSep:
			return '?'
		default:
			return r
		}
	}, s)

	// A message of only dropped characters empties here, leaving a failure with no cause.
	if strings.TrimSpace(s) == "" {
		return "unknown error"
	}

	return s
}

// quotedInner returns a value as strconv.Quote renders it, without the enclosing quotes, which is
// the spelling an error using %q puts in its message.
func quotedInner(s string) string {
	q := strconv.Quote(s)

	return q[1 : len(q)-1]
}

// addrString renders a network address, which is absent on most operating errors. A nil pointer
// behind the interface is checked the way the chain walk checks a nil error, since String is
// called while a failure is already being handled.
func addrString(a net.Addr) string {
	if a == nil {
		return ""
	} else if v := reflect.ValueOf(a); v.Kind() == reflect.Pointer && v.IsNil() {
		return ""
	}

	return a.String()
}

// errorLocation reports whether a value names a location rather than a bare name. Replacement
// is by substring, so a value without a separator would also match ordinary words, and one
// built only from separators and dots would match a path fragment of every message.
func errorLocation(s string) bool {
	return strings.ContainsAny(s, errorPathSeparators) && strings.Trim(s, errorPathTrivial) != ""
}

// errorPaths returns the locations named by err and the errors it wraps, deduplicated and longest
// first so that replacing one cannot leave a shorter one's remainder behind. A URL, a command name
// and a socket address are collected alongside a file path, since each is a location and each is
// rendered by the error that carries it. It reports complete unless the chain exceeded the budget.
func errorPaths(err error) (out []string, complete bool) {
	seen := make(map[string]struct{})
	nodes := 0
	complete = true

	var walk func(error)

	walk = func(e error) {
		if e == nil {
			return
		} else if nodes >= errorUnwrapNodes {
			complete = false
			return
		}

		// A non-nil interface can hold a nil pointer, which neither branch below survives.
		if v := reflect.ValueOf(e); v.Kind() == reflect.Pointer && v.IsNil() {
			return
		}

		nodes++

		var found []string

		switch t := e.(type) {
		case *iofs.PathError:
			found = []string{t.Path}
		case *os.LinkError:
			found = []string{t.Old, t.New}
		case *url.Error:
			// Rendered with %q, so the escaped spelling is what the message carries.
			found = []string{t.URL, quotedInner(t.URL)}
		case *exec.Error:
			// Quoted the same way, so a Windows path with backslashes needs the escaped spelling.
			found = []string{t.Name, quotedInner(t.Name)}
		case *net.OpError:
			found = []string{addrString(t.Source), addrString(t.Addr)}
		}

		for _, p := range found {
			if _, dup := seen[p]; !dup && errorLocation(p) {
				seen[p] = struct{}{}
				out = append(out, p)
			}
		}

		switch u := e.(type) {
		case interface{ Unwrap() error }:
			walk(u.Unwrap())
		case interface{ Unwrap() []error }:
			for _, w := range u.Unwrap() {
				walk(w)
			}
		}
	}

	walk(err)

	sort.SliceStable(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })

	return out, complete
}
