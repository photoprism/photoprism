package video

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Chunk represents a fixed length file chunk.
type Chunk [4]byte

// Get returns the chunk as byte array.
func (c Chunk) Get() [4]byte {
	return c
}

// Hex returns the chunk as hex formatted string.
func (c Chunk) Hex() string {
	return fmt.Sprintf("0x%x", c[:])
}

// String returns the chunk as string.
func (c Chunk) String() string {
	return string(c[:])
}

// Bytes returns the chunk as byte slice.
func (c Chunk) Bytes() []byte {
	return c[:]
}

// Uint32 returns the chunk as unsigned integer.
func (c Chunk) Uint32() uint32 {
	return binary.BigEndian.Uint32(c.Bytes())
}

// Equal compares the chunk with a byte slice.
func (c Chunk) Equal(b []byte) bool {
	return bytes.Equal(c.Bytes(), b)
}

// FileOffset returns the index of the chunk, or -1 if it was not found.
func (c Chunk) FileOffset(fileName string) (index int, err error) {
	if !fs.FileExists(fileName) {
		return -1, errors.New("file not found")
	}

	file, err := os.Open(fileName) //nolint:gosec // fileName validated by caller

	if err != nil {
		return -1, err
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	return c.DataOffset(file, 0, -1)
}

// DataOffset returns the index of the chunk in file, or -1 if it was not found.
// Delegates to Chunks.DataOffset so the single-needle and multi-needle scans
// share one buffered-read implementation.
func (c Chunk) DataOffset(file io.ReadSeeker, offset, maxOffset int) (int, error) {
	pos, _, err := Chunks{c}.DataOffset(file, offset, maxOffset)
	return pos, err
}
