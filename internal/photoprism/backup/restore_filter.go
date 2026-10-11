package backup

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var (
	// uniqueChecksOff is the dump header line that turns off unique checks for the restore session.
	uniqueChecksOff = []byte("/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;")
	// uniqueChecksOn is the line uniqueChecksOff is replaced with.
	uniqueChecksOn = []byte("/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=1 */;")
)

const (
	// uniqueChecksScanBytes is the size of the dump prefix in which the header line must start to be replaced.
	uniqueChecksScanBytes = 64 << 10
	// uniqueChecksLineBytes is the buffer size for reading lines; longer lines are passed on in parts.
	uniqueChecksLineBytes = 4096
)

// restoreReader returns the input a dump is restored from with the specified driver. A MariaDB or MySQL
// dump is read through uniqueChecksReader, a PostgreSQL dump through postgresRestrictedReader, and other
// dumps are returned as they are.
func restoreReader(driver string, r io.Reader) io.Reader {
	switch driver {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		return &uniqueChecksReader{r: bufio.NewReaderSize(r, uniqueChecksLineBytes)}
	case dsn.DriverPostgres:
		return newPostgresRestrictedReader(r)
	default:
		return r
	}
}

// uniqueChecksReader passes a dump on unchanged, except that the first line in its header that turns off
// unique checks is replaced with one that keeps them on, so InnoDB rolls back only a statement that fails.
// Foreign key checks stay as the dump sets them, since either check suffices.
type uniqueChecksReader struct {
	r       *bufio.Reader
	pending []byte
	err     error
	read    int
	partial bool
	done    bool
}

// Read reads from the dump, checking each whole line within the scan prefix until the header is replaced.
func (u *uniqueChecksReader) Read(p []byte) (int, error) {
	for len(u.pending) == 0 {
		if u.err != nil {
			return 0, u.err
		} else if u.done {
			return u.r.Read(p)
		}

		// The line stays valid until the next read, which happens only once it has been passed on.
		line, err := u.r.ReadSlice('\n')
		complete := err == nil

		if !u.partial && complete && isUniqueChecksOff(line) {
			line = append(append([]byte{}, uniqueChecksOn...), line[len(uniqueChecksOff):]...)
			u.done = true
		}

		u.read += len(line)
		u.partial = !complete
		u.pending = line

		if err != nil && err != bufio.ErrBufferFull {
			u.err = err
		}

		if !u.done && (u.read >= uniqueChecksScanBytes || errors.Is(u.err, io.EOF)) {
			log.Debugf("restore: found no unique checks header line, reading the dump as it is")
			u.done = true
		}
	}

	n := copy(p, u.pending)
	u.pending = u.pending[n:]

	return n, nil
}

// isUniqueChecksOff reports whether line is the header line that turns off unique checks, ending with
// a line break.
func isUniqueChecksOff(line []byte) bool {
	rest, ok := bytes.CutPrefix(line, uniqueChecksOff)
	return ok && (bytes.Equal(rest, []byte("\n")) || bytes.Equal(rest, []byte("\r\n")))
}

// postgresHeldBytes bounds the blank lines held back after an "\\unrestrict" line; beyond it, they are passed on.
const postgresHeldBytes = 64 << 10

// postgresRestrictLine matches the "\restrict" and "\unrestrict" lines pg_dump writes, with an alphanumeric key.
var postgresRestrictLine = regexp.MustCompile(`^\\(un)?restrict [0-9A-Za-z]+\r?\n?$`)

// postgresRestrictedReader passes a PostgreSQL dump to psql in restricted mode: it starts with a "\restrict"
// line holding a random key, so psql rejects every other backslash command, and drops the "\restrict" line in
// the header of the dump and the "\unrestrict" line at its end, which would otherwise fail or lift the
// restriction. Other lines are passed on unchanged, as they may be part of COPY data or quoted text.
type postgresRestrictedReader struct {
	r       *bufio.Reader
	pending []byte
	held    []byte
	err     error
	header  bool
	copying bool
	partial bool
}

// newPostgresRestrictedReader returns a postgresRestrictedReader for a dump, restricted with a new random key.
func newPostgresRestrictedReader(r io.Reader) *postgresRestrictedReader {
	return &postgresRestrictedReader{
		r:       bufio.NewReaderSize(r, uniqueChecksLineBytes),
		pending: []byte(`\restrict ` + rnd.Base62(32) + "\n"),
		header:  true,
	}
}

// Read reads from the dump line by line. An "\unrestrict" line is held back with the blank lines after it
// until a line that is not blank, and dropped with them if the dump ends before it.
func (p *postgresRestrictedReader) Read(b []byte) (int, error) {
	for len(p.pending) == 0 {
		if p.err != nil {
			return 0, p.err
		}

		// The line stays valid until the next read, which happens only once it has been passed on.
		line, err := p.r.ReadSlice('\n')
		start := !p.partial
		p.partial = errors.Is(err, bufio.ErrBufferFull)

		if err != nil && !p.partial {
			p.err = err
		}

		trimmed := bytes.TrimRight(line, "\r\n")
		blank := len(bytes.TrimSpace(trimmed)) == 0

		switch {
		case !start:
			// The rest of a long line.
		case p.copying:
			p.copying = !bytes.Equal(trimmed, []byte(`\.`))
		case p.header && postgresRestrictLine.Match(line) && bytes.HasPrefix(line, []byte(`\restrict`)):
			p.header, line = false, nil
			continue
		case postgresRestrictLine.Match(line) && bytes.HasPrefix(line, []byte(`\unrestrict`)):
			p.pending, p.held = p.held, append([]byte{}, line...)
			continue
		case bytes.HasPrefix(trimmed, []byte("COPY ")) && bytes.HasSuffix(trimmed, []byte(" FROM stdin;")):
			p.copying = true
		}

		if start && !blank && !bytes.HasPrefix(trimmed, []byte("--")) {
			p.header = false
		}

		switch {
		case p.held != nil && start && blank && len(p.held) < postgresHeldBytes:
			p.held = append(p.held, line...)
		case p.held != nil:
			p.held = append(p.held, line...)
			p.pending, p.held = p.held, nil
		default:
			p.pending = line
		}
	}

	n := copy(b, p.pending)
	p.pending = p.pending[n:]

	return n, nil
}
