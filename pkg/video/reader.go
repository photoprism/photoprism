package video

import (
	"errors"
	"io"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Reader reads an existing file from an offset until the end.
type Reader struct {
	fileName string
	file     *os.File
	offset   int64
}

// NewReader creates a new Reader.
func NewReader(fileName string, offset int64) (*Reader, error) {
	// Check if the file exists.
	if !fs.FileExists(fileName) {
		return nil, errors.New("file not found")
	}

	// Open file for reading.
	file, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	// Ensure that the offset is positive and we are starting at the right position.
	if offset < 0 {
		offset = 0
	} else if offset > 0 {
		if _, seekErr := file.Seek(offset, io.SeekStart); seekErr != nil {
			_ = file.Close()
			return nil, err
		}
	}

	// Create new reader and return it.
	return &Reader{fileName: fileName, file: file, offset: int64(offset)}, nil
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read and any error encountered.
func (r *Reader) Read(p []byte) (n int, err error) {
	return r.file.Read(p)
}

// Close closes the file after reading.
func (r *Reader) Close() error {
	r.fileName = ""
	r.offset = 0
	return r.file.Close()
}

// ReadSeeker reads an existing file from an offset until the end.
type ReadSeeker struct {
	file   io.ReadSeeker
	offset int64
}

// NewReadSeeker creates a new ReadSeeker.
func NewReadSeeker(file io.ReadSeeker, offset int64) *ReadSeeker {
	// Ensure that the offset is positive.
	if offset < 0 {
		offset = 0
	}

	return &ReadSeeker{
		file:   file,
		offset: offset,
	}
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read and any error encountered.
func (r *ReadSeeker) Read(p []byte) (n int, err error) {
	return r.file.Read(p)
}

// Seek sets the offset for the next Read or Write to offset, interpreted according to whence:
// - SeekStart means relative to the start of the file
// - SeekCurrent means relative to the current offset
// - SeekEnd means relative to the end (for example, offset = -2 specifies the penultimate byte of the file)
//
// Seek returns the new offset relative to the start of the file or an error, if any.
// Seeking to an offset before the start of the file is an error.
func (r *ReadSeeker) Seek(offset int64, whence int) (pos int64, err error) {
	switch whence {
	case io.SeekStart:
		pos, err = r.file.Seek(offset+r.offset, whence)
	case io.SeekCurrent, io.SeekEnd:
		pos, err = r.file.Seek(offset, whence)
	default:
		return 0, errors.New("unknown whence")
	}

	return pos - r.offset, err
}
