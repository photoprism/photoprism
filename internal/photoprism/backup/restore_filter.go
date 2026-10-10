package backup

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/photoprism/photoprism/pkg/dsn"
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
// dump is read through uniqueChecksReader; other dumps are returned as they are.
func restoreReader(driver string, r io.Reader) io.Reader {
	switch driver {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		return &uniqueChecksReader{r: bufio.NewReaderSize(r, uniqueChecksLineBytes)}
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
