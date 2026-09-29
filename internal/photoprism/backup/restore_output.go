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

// restoreErrorRegex matches the output line of a statement that failed while the client continued,
// with or without error code, SQLSTATE, and source file, e.g. "ERROR 1062 (23000) at line 4: ...".
var restoreErrorRegex = regexp.MustCompile(`^ERROR(?: (\d+))?(?: \([0-9A-Za-z]+\))? at line (\d+)\b`)

// restoreClientErrorRegex matches a client error without a line number, e.g. "ERROR 2002 (HY000): ...".
var restoreClientErrorRegex = regexp.MustCompile(`^ERROR (\d+) \(([0-9A-Za-z]+)\):`)

// restoreSqliteErrorRegex matches the output line of a statement the SQLite client could not run, with
// the SQLite result code a runtime error ends with, e.g. "Runtime error near line 3: ... (19)".
var restoreSqliteErrorRegex = regexp.MustCompile(`^(?:(Parse|Runtime) error|Error:) near line (\d+):.*?(?:\((\d+)\))?$`)

// restoreSqliteOpenRegex matches the message of the SQLite client when it cannot open the database.
var restoreSqliteOpenRegex = regexp.MustCompile(`^Error: unable to open database`)

// restoreClientPrefixRegex matches a message the client writes under the path it was started with, e.g.
// "/usr/bin/mariadb: unknown option ...", and whether it is a warning.
var restoreClientPrefixRegex = regexp.MustCompile(`^(?:\S*/)?(?:mariadb|mysql|sqlite3|mariadb-dump|mysqldump): (\[Warning\])?`)

// restoreSqliteFatalCodes are SQLite result codes that fail every statement after them, e.g. because the
// database is read-only, locked, full, or damaged, so the restore failed rather than single statements.
var restoreSqliteFatalCodes = map[string]bool{
	"5": true, "6": true, "7": true, "8": true, "9": true, "10": true, "11": true, "13": true, "14": true, "26": true,
}

// restoreOutput collects the error output of a restore client with bounded memory. Failed statements
// are counted and kept as error code and line number only, and only known kinds of client messages are
// kept for diagnostics, so that statement text and row values are never kept.
type restoreOutput struct {
	line     []byte
	echo     bool
	started  bool
	failed   int
	errors   []string
	lines    []string
	problems int
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
// then reports it with an error line, so everything from the delimiter to that line is skipped. Client
// warnings are kept only before the first statement output, since the client writes them when connecting.
func (o *restoreOutput) endLine() {
	raw := strings.TrimSuffix(string(o.line), "\r")
	line := strings.TrimSpace(raw)
	o.line = o.line[:0]

	if raw == restoreEchoDelimiter {
		o.echo, o.started = true, true
		return
	}

	if m := restoreErrorRegex.FindStringSubmatch(line); m != nil {
		o.echo, o.started = false, true

		if isClientErrorCode(m[1]) {
			o.keep(line, true)
		}

		o.fail(m[1], m[2])
		return
	}

	if m := restoreSqliteErrorRegex.FindStringSubmatch(line); m != nil {
		o.started = true

		if restoreSqliteFatalCodes[m[3]] && m[1] == "" {
			o.keep(fmt.Sprintf("Error near line %s (%s)", m[2], m[3]), true)
		} else if restoreSqliteFatalCodes[m[3]] {
			o.keep(fmt.Sprintf("%s error near line %s (%s)", m[1], m[2], m[3]), true)
		}

		o.fail("", m[2])
		return
	}

	switch m, p := restoreClientErrorRegex.FindStringSubmatch(line), restoreClientPrefixRegex.FindStringSubmatch(line); {
	case o.echo || line == "":
	case m != nil && (isClientErrorCode(m[1]) || isAccessErrorCode(m[1])):
		o.keep(line, true)
	case m != nil:
		o.keep(fmt.Sprintf("ERROR %s (%s)", m[1], m[2]), true)
	case strings.HasPrefix(line, "WARNING:") || p != nil && p[1] != "":
		if !o.started {
			o.keep(line, false)
		}
	case p != nil || restoreSqliteOpenRegex.MatchString(line):
		o.keep(line, true)
	}
}

// fail counts a failed statement and keeps the first ones as error code and line number.
func (o *restoreOutput) fail(code, line string) {
	o.failed++

	if len(o.errors) >= restoreErrorsKept {
		return
	} else if code == "" {
		o.errors = append(o.errors, fmt.Sprintf("error at line %s", line))
	} else {
		o.errors = append(o.errors, fmt.Sprintf("error %s at line %s", code, line))
	}
}

// isClientErrorCode reports whether code is a MariaDB client error, such as a lost connection.
func isClientErrorCode(code string) bool {
	return len(code) == 4 && strings.HasPrefix(code, "2")
}

// isAccessErrorCode reports whether code is a server error about the account or database a client
// connects with, which contains no statement or row data.
func isAccessErrorCode(code string) bool {
	switch code {
	case "1044", "1045", "1049", "1698":
		return true
	}

	return false
}

// keep stores an output line for diagnostics, up to restoreLinesKept lines, and counts it as a problem
// unless it is a warning.
func (o *restoreOutput) keep(line string, problem bool) {
	if problem {
		o.problems++
	}

	if len(o.lines) < restoreLinesKept {
		o.lines = append(o.lines, line)
	}
}

// OnlyFailedStatements reports whether the client reported failed statements and no other problem.
func (o *restoreOutput) OnlyFailedStatements() bool {
	return o.failed > 0 && o.problems == 0
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
