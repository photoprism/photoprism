package meta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
)

// ErrExifFileTooLarge is returned when a file exceeds ExifMaxFileBytes.
var ErrExifFileTooLarge = errors.New("file is too large for embedded metadata parsing")

var (
	// ExifMaxFileBytes limits the file size the embedded Exif parsers accept, since they read
	// the whole file into memory before they look for the Exif block.
	ExifMaxFileBytes int64 = 256 << 20

	// ExifMaxTags limits the number of tag values retained from a single file.
	ExifMaxTags = 4096
)

// ExifFileSize returns the size of the named file, or ErrExifFileTooLarge when it exceeds
// ExifMaxFileBytes. ExifTool reads a file as a stream rather than loading it completely, so it
// still reports metadata for one of these; with ExifTool disabled there is no embedded metadata
// for a file this large.
func ExifFileSize(fileName string) (int64, error) {
	s, err := os.Stat(fileName)

	if err != nil {
		return 0, err
	}

	if s.IsDir() {
		return 0, fmt.Errorf("%s is a directory", clean.Log(filepath.Base(fileName)))
	}

	if size := s.Size(); size > ExifMaxFileBytes {
		return size, ErrExifFileTooLarge
	} else {
		return size, nil
	}
}
