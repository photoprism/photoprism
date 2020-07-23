package photoprism

import (
	"strings"
)

// List of related files for importing and indexing.
type RelatedFiles struct {
	Files MediaFiles
	Main  *MediaFile
}

// ContainsJpeg returns true if related file list contains a JPEG.
func (m RelatedFiles) ContainsJpeg() bool {
	for _, f := range m.Files {
		if f.IsJpeg() {
			return true
		}
	}

	if m.Main == nil {
		return false
	}

	return m.Main.IsJpeg()
}

// String returns file names as string.
func (m RelatedFiles) String() string {
	names := make([]string, len(m.Files))

	for i, f := range m.Files {
		names[i] = f.BaseName()
	}

	return strings.Join(names, ", ")
}

// Len returns the number of related files.
func (m RelatedFiles) Len() int {
	return len(m.Files)
}
