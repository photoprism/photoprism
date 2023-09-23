package video

import (
	"bytes"
	"errors"
	"io"

	"github.com/sunfish-shogi/bufseekio"
)

// Chunks represents a list of file chunks.
type Chunks []Chunk

// Contains tests if the chunk is contained in this list.
func (c Chunks) Contains(s [4]byte) bool {
	if len(c) == 0 {
		return false
	}

	// Find matches.
	for i := range c {
		if s == c[i] {
			return true
		}
	}

	return false
}

// ContainsAny checks if at least one common chunk exists in this list.
func (c Chunks) ContainsAny(b [][4]byte) bool {
	if len(c) == 0 || len(b) == 0 {
		return false
	}

	// Find matches.
	for i := range c {
		for j := range b {
			if b[j] == c[i] {
				return true
			}
		}
	}

	// Not found.
	return false
}

// FileTypeOffset returns the file type start offset in f, or -1 if it was not found.
func (c Chunks) FileTypeOffset(file io.ReadSeeker) (int, error) {
	if file == nil {
		return -1, errors.New("file is nil")
	}

	ftyp := ChunkFTYP.Bytes()
	blockSize := 128 * 1024
	buffer := make([]byte, blockSize)

	// Create buffered read seeker.
	r := bufseekio.NewReadSeeker(file, blockSize, 8)

	// Index offset.
	var offset int

	// Search in batches.
	for {
		n, err := r.Read(buffer)
		buffer = buffer[:n]

		if err != nil {
			if err != io.EOF {
				return -1, err
			}

			break
		} else if n == 0 {
			break
		}

		// Find ftyp chunk.
		if i := bytes.Index(buffer, ftyp); i < 0 {
			// Not found.
		} else if j := i + 4; j < 8 || len(buffer) < j+4 {
			// Skip.
		} else if k := j + 4; c.Contains(*(*[4]byte)(buffer[j:k])) {
			return offset + i - 4, nil
		}

		offset += n
	}

	return -1, nil
}
