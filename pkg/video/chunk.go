package video

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/sunfish-shogi/bufseekio"
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

// FileOffset returns the index of the chunk in the specified file, or -1 if it was not found.
func (c Chunk) FileOffset(fileName string) (int, error) {
	if !fs.FileExists(fileName) {
		return -1, errors.New("file not found")
	}

	file, err := os.Open(fileName)

	if err != nil {
		return -1, err
	}

	defer file.Close()

	index, err := c.DataOffset(file)

	return index, err
}

// DataOffset returns the index of the chunk in f, or -1 if it was not found.
func (c Chunk) DataOffset(file io.ReadSeeker) (int, error) {
	if file == nil {
		return -1, errors.New("file is nil")
	}

	data := c.Bytes()
	dataSize := len(data)
	blockSize := 128 * 1024
	buffer := make([]byte, blockSize)

	// Create buffered read seeker.
	r := bufseekio.NewReadSeeker(file, blockSize, dataSize)

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

		// Return data index, if found.
		if i := bytes.Index(buffer, data); i >= 0 {
			return offset + i, nil
		}

		offset += n
	}

	return -1, nil
}
