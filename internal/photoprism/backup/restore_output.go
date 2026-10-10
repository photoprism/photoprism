package backup

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const (
	// restoreErrorsKept is the number of failed statements reported with their error code and line.
	restoreErrorsKept = 5
	// restoreLinesKept is the number of client output lines kept.
	restoreLinesKept = 5
	// restoreLineBytes is the number of bytes kept of a single client output line.
	restoreLineBytes = 1024
)

// restoreErrorRegex matches the line a client writes for a statement that failed while it continued,
// e.g. "ERROR 1062 (23000) at line 4: ...".
var restoreErrorRegex = regexp.MustCompile(`^ERROR(?: (\d+))?.*? at line (\d+)`)

// restoreOutput collects the error output of a restore client with bounded memory: it counts the failed
// statements, and keeps the first lines for the error of a failed restore and the warnings before them.
type restoreOutput struct {
	line     []byte
	lines    []string
	warnings []string
	failed   int
	errors   []string
}

// Write processes client output line by line and never fails.
func (o *restoreOutput) Write(p []byte) (int, error) {
	n := len(p)

	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		chunk := p

		if i >= 0 {
			chunk = p[:i]
		}

		if room := restoreLineBytes - len(o.line); room > 0 {
			o.line = append(o.line, chunk[:min(room, len(chunk))]...)
		}

		if i < 0 {
			break
		}

		o.endLine()
		p = p[i+1:]
	}

	return n, nil
}

// endLine counts the current output line if it reports a failed statement, and keeps it if it is one of
// the first lines. Warnings are kept only before any other output, as the client writes them when connecting.
func (o *restoreOutput) endLine() {
	line := strings.TrimSpace(string(o.line))
	o.line = o.line[:0]

	if line == "" {
		return
	}

	if m := restoreErrorRegex.FindStringSubmatch(line); m == nil {
		// Not a failed statement.
	} else if o.failed++; len(o.errors) >= restoreErrorsKept {
		// Only the first failed statements are reported.
	} else if m[1] == "" {
		o.errors = append(o.errors, fmt.Sprintf("error at line %s", m[2]))
	} else {
		o.errors = append(o.errors, fmt.Sprintf("error %s at line %s", m[1], m[2]))
	}

	if len(o.warnings) == len(o.lines) && len(o.lines) < restoreLinesKept && strings.HasPrefix(line, "WARNING:") {
		o.warnings = append(o.warnings, line)
	}

	if len(o.lines) < restoreLinesKept {
		o.lines = append(o.lines, line)
	}
}

// Failures returns the number of failed statements and the first ones as error code and line number.
func (o *restoreOutput) Failures() restoreFailures {
	return restoreFailures{Count: o.failed, Errors: o.errors}
}

// Close processes a last line that has no line break.
func (o *restoreOutput) Close() {
	if len(o.line) > 0 {
		o.endLine()
	}
}

// String returns the first output lines.
func (o *restoreOutput) String() string {
	return strings.Join(o.lines, "\n")
}

// Warnings returns the warnings written before any other output.
func (o *restoreOutput) Warnings() string {
	return strings.Join(o.warnings, "\n")
}
