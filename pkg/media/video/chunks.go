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
		} else if k := j + 4; c.Contains([4]byte(buffer[j:k])) {
			return offset + i - 4, nil
		}

		offset += n
	}

	return -1, nil
}

// DataOffset scans file for the first occurrence of any chunk in c and returns
// the matching offset together with the chunk that matched. The search starts
// at offset and stops at maxOffset (or EOF when maxOffset < 0). A single pass
// is made: each buffered block is searched for every chunk at once and the
// earliest match within the block wins. Returns -1 and a zero Chunk if no
// chunk in c is found before the cap or EOF.
func (c Chunks) DataOffset(file io.ReadSeeker, offset, maxOffset int) (int, Chunk, error) {
	if file == nil {
		return -1, Chunk{}, errors.New("file is nil")
	} else if len(c) == 0 {
		return -1, Chunk{}, nil
	}

	const blockSize = 128 * 1024
	const chunkLen = 4 // Chunk is [4]byte; bufseekio uses this as the inter-block overlap.

	buffer := make([]byte, blockSize)
	r := bufseekio.NewReadSeeker(file, blockSize, chunkLen)

	if seekOffset, seekErr := r.Seek(int64(offset), io.SeekStart); seekErr != nil {
		return -1, Chunk{}, seekErr
	} else {
		offset = int(seekOffset)
	}

	// Search in batches.
	for {
		n, err := r.Read(buffer)
		buffer = buffer[:n]

		if err != nil {
			if err != io.EOF {
				return -1, Chunk{}, err
			}

			break
		} else if n == 0 {
			break
		}

		// Pick the earliest match across all chunks within this buffer.
		bestIdx := -1
		var bestChunk Chunk
		for i := range c {
			if idx := bytes.Index(buffer, c[i].Bytes()); idx >= 0 && (bestIdx < 0 || idx < bestIdx) {
				bestIdx = idx
				bestChunk = c[i]
			}
		}

		if bestIdx >= 0 {
			return offset + bestIdx, bestChunk, nil
		}

		offset += n

		// Return if the chunk was not found up to the maximum offset.
		if maxOffset > 0 && maxOffset <= offset {
			return -1, Chunk{}, nil
		}
	}

	return -1, Chunk{}, nil
}
