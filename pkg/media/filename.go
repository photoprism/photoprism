package media

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// FromName returns the content type matching the file extension.
func FromName(fileName string) Type {
	if fileName == "" {
		return Unknown
	}

	// Find media type based on the file type.
	if result, found := Formats[fs.FileType(fileName)]; found {
		return result
	}

	// Default to sidecar.
	return Sidecar
}

// MainFile checks if the filename belongs to a main content type.
func MainFile(fileName string) bool {
	return FromName(fileName).Main()
}
