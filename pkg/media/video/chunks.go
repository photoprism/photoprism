package video

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/sunfish-shogi/bufseekio"
)

const (
	// sampleEntryHeaderLen is the number of leading bytes of an ISO BMFF visual
	// sample entry inspected during validation: box size (4) + coding name (4) +
	// reserved (6) + data_reference_index (2).
	sampleEntryHeaderLen = 16
	// minVisualSampleEntrySize is the smallest spec-compliant VisualSampleEntry
	// box (8-byte box header + 8-byte SampleEntry base + 70-byte VisualSampleEntry
	// fixed fields), and maxVisualSampleEntrySize caps the plausible size so a
	// random four-byte value preceding a colliding coding name is rejected.
	minVisualSampleEntrySize = 86
	maxVisualSampleEntrySize = 1 << 20
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
	const cachedBlocks = 4 // Number of blocks bufseekio keeps cached; reads do not overlap.

	buffer := make([]byte, blockSize)
	r := bufseekio.NewReadSeeker(file, blockSize, cachedBlocks)

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

// isVisualSampleEntry reports whether box begins with a valid ISO BMFF visual
// sample entry header: a big-endian box size within plausible bounds, the
// four-byte coding name, six reserved zero bytes, and a nonzero
// data_reference_index. The reserved zero bytes and nonzero index are mandated
// by ISO/IEC 14496-12, so a coding name that merely collides with random
// payload bytes is rejected. box must hold at least sampleEntryHeaderLen bytes
// starting at the box size field; the coding name (box[4:8]) is matched by the
// caller.
func isVisualSampleEntry(box []byte) bool {
	if len(box) < sampleEntryHeaderLen {
		return false
	}

	if size := binary.BigEndian.Uint32(box[0:4]); size < minVisualSampleEntrySize || size > maxVisualSampleEntrySize {
		return false
	}

	// The six bytes following the coding name are a const-zero reserved field.
	for _, b := range box[8:14] {
		if b != 0 {
			return false
		}
	}

	// data_reference_index is a 1-based index and is never zero in practice.
	return binary.BigEndian.Uint16(box[14:16]) != 0
}

// SampleEntryOffset scans the head of file (up to maxOffset, or to EOF when
// maxOffset <= 0) for the first chunk in c that is framed as a valid ISO BMFF
// visual sample entry, returning its offset and the matching chunk. Unlike a
// raw byte search, each candidate coding name is validated with
// isVisualSampleEntry so that a four-byte code colliding with random payload
// bytes — common in raw video elementary streams such as DV — is not mistaken
// for a codec. Returns -1 and a zero Chunk when no valid sample entry is found.
func (c Chunks) SampleEntryOffset(file io.ReadSeeker, maxOffset int) (int, Chunk, error) {
	if file == nil {
		return -1, Chunk{}, errors.New("file is nil")
	} else if len(c) == 0 {
		return -1, Chunk{}, nil
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return -1, Chunk{}, err
	}

	const blockSize = 128 * 1024
	// carry retains the trailing bytes of each block so a coding name straddling
	// a block boundary, and the validation window of a candidate near the end of
	// a block, stay visible within a single buffer on the next read.
	const carry = sampleEntryHeaderLen

	buffer := make([]byte, carry+blockSize)

	base := 0 // Absolute file offset of buffer[0].
	have := 0 // Number of valid bytes currently in buffer.

	for {
		n, err := io.ReadFull(file, buffer[have:])

		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return -1, Chunk{}, err
		}

		have += n
		atEOF := err == io.EOF || err == io.ErrUnexpectedEOF

		// A sample entry's coding name sits at box offset +4, so a candidate at
		// index i needs the window buffer[i-4 : i+12]; i therefore starts at 4.
		best, bestChunk := -1, Chunk{}
		for _, chunk := range c {
			needle := chunk.Bytes()
			for from := 4; from+len(needle) <= have; {
				rel := bytes.Index(buffer[from:have], needle)
				if rel < 0 {
					break
				}

				i := from + rel

				// Defer candidates whose validation window is not fully buffered yet;
				// the carry-over re-surfaces them with complete context next read.
				if i-4+sampleEntryHeaderLen > have {
					break
				}

				if isVisualSampleEntry(buffer[i-4 : i-4+sampleEntryHeaderLen]) {
					if best < 0 || i < best {
						best, bestChunk = i, chunk
					}
					break
				}

				from = i + 1
			}
		}

		if best >= 0 {
			return base + best, bestChunk, nil
		}

		if atEOF {
			return -1, Chunk{}, nil
		}

		if maxOffset > 0 && base+have >= maxOffset {
			return -1, Chunk{}, nil
		}

		// Carry the trailing bytes to the front and refill the remainder.
		if have > carry {
			copy(buffer, buffer[have-carry:have])
			base += have - carry
			have = carry
		}
	}
}
