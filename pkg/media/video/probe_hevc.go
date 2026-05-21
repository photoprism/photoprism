package video

import (
	"io"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
)

// HevcChunks lists the ISO BMFF sample entry codes that identify an HEVC
// (H.265) video stream, including Dolby Vision wrappers built on top of HEVC.
var HevcChunks = Chunks{
	ChunkHVC1, ChunkHVC2, ChunkHVC3, ChunkDVH1,
	ChunkHEV1, ChunkHEV2, ChunkHEV3, ChunkDVHE,
}

// HevcHeadScanLimit caps how far HEVC chunk scans read into the file. The
// HEVC sample entry sits in the stsd box inside moov; faststart videos place
// moov at the head, and Motion Photos put it just past the JPEG (a few MB
// at most), so 16 MiB comfortably covers both layouts.
const HevcHeadScanLimit = 16 * 1024 * 1024

// IsHEVC reports whether the reader contains an HEVC video stream by scanning
// the head of the file (up to HevcHeadScanLimit) for any HEVC sample entry
// code in a single pass. Returns false on read errors or empty input.
func IsHEVC(file io.ReadSeeker) bool {
	if file == nil {
		return false
	}

	pos, _, err := HevcChunks.DataOffset(file, 0, HevcHeadScanLimit)

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
