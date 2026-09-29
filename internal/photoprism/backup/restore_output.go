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
	// restoreLinesKept is the number of other client output lines kept for diagnostics.
	restoreLinesKept = 50
	// restoreLineBytes is the number of bytes kept of a single client output line.
	restoreLineBytes = 1024
	// restoreEchoDelimiter encloses a failed statement that the client repeats in its output.
	restoreEchoDelimiter = "--------------"
)

// restoreErrorRegex matches the output line of a statement that failed while the client continued.
var restoreErrorRegex = regexp.MustCompile(`^ERROR (\d+) \([0-9A-Za-z]+\) at line (\d+):`)

// restoreSqliteErrorRegex matches the output line of a statement the SQLite client could not run.
var restoreSqliteErrorRegex = regexp.MustCompile(`^(Parse|Runtime) error near line (\d+):`)

// restoreOutput collects the error output of a restore client with bounded memory. Failed statements
// are counted and kept as error code and line number only, the statements the client repeats or quotes
// are skipped, and a limited number of other lines are kept for diagnostics.
type restoreOutput struct {
	line   []byte
	echo   bool
	failed int
	errors []string
	lines  []string
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

// endLine processes the current output line. The client repeats a failed statement after a delimiter and
// then reports it with an error line, so everything from the delimiter to that line is skipped. A client
// error, such as a lost connection, is also kept in full since it contains no statement or row data.
func (o *restoreOutput) endLine() {
	raw := strings.TrimSuffix(string(o.line), "\r")
	line := strings.TrimSpace(raw)
	o.line = o.line[:0]

	if raw == restoreEchoDelimiter {
		o.echo = true
		return
	}

	if m := restoreErrorRegex.FindStringSubmatch(line); m != nil {
		o.echo = false
		o.failed++

		if len(o.errors) < restoreErrorsKept {
			o.errors = append(o.errors, fmt.Sprintf("error %s at line %s", m[1], m[2]))
		}

		if strings.HasPrefix(m[1], "2") && len(m[1]) == 4 {
			o.keep(line)
		}

		return
	}

	switch m := restoreSqliteErrorRegex.FindStringSubmatch(line); {
	case m != nil:
		o.keep(fmt.Sprintf("%s error near line %s", m[1], m[2]))
	case o.echo || line == "" || line != raw && strings.TrimLeft(raw, " \t") != raw:
		// Skip repeated statements and indented lines, which quote statement text.
	default:
		o.keep(line)
	}
}

// keep stores an output line for diagnostics, up to restoreLinesKept lines.
func (o *restoreOutput) keep(line string) {
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

// String returns the other output lines that were kept.
func (o *restoreOutput) String() string {
	return strings.Join(o.lines, "\n")
}
