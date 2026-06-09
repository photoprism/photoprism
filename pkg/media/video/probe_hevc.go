package video

import (
	"io"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
)

// IsHEVC reports whether the reader contains an HEVC video stream by scanning
// the head of the file (up to HeadScanLimit) for a valid HEVC sample entry in
// a single pass. Candidates are validated as real visual sample entries so a
// four-byte code colliding with random payload bytes is not mistaken for HEVC.
// Returns false on read errors or empty input.
func IsHEVC(file io.ReadSeeker) bool {
	if file == nil {
		return false
	}

	pos, _, err := HevcChunks.SampleEntryOffset(file, HeadScanLimit)

	return err == nil && pos > 0
}

// IsHEVCFile reports whether the named file contains an HEVC video stream.
// Returns false if the file is missing, unreadable, or contains no HEVC chunk.
func IsHEVCFile(fileName string) bool {
	if fileName == "" || !fs.FileExists(fileName) {
		return false
	}

	file, err := os.Open(fileName) //nolint:gosec // fileName validated by caller
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()

	return IsHEVC(file)
}
