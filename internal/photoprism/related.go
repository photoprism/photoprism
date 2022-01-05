package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

// RelatedFiles represents a list of related files to be indexed or imported.
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

// Count returns the number of files without the main file.
func (m RelatedFiles) Count() int {
	if l := m.Len(); l < 1 {
		return l
	} else {
		return l - 1
	}
}

// MainFileType returns the main file type as string.
func (m RelatedFiles) MainFileType() string {
	if m.Main == nil {
		return ""
	}

	return string(m.Main.FileType())
}

// MainLogName returns the main file name for logging.
func (m RelatedFiles) MainLogName() string {
	if m.Main == nil {
		return ""
	}

	return sanitize.Log(m.Main.RelName(Config().OriginalsPath()))
}
