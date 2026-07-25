package photoprism

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// RelatedFiles represents a list of related files to be indexed or imported.
type RelatedFiles struct {
	Files MediaFiles
	Main  *MediaFile
}

// HasPreview checks if the list of files contains a PNG or JPEG image to render a preview in the UI.
func (m RelatedFiles) HasPreview() bool {
	for _, f := range m.Files {
		if f.IsPreviewImage() {
			return true
		}
	}

	if m.Main == nil {
		return false
	}

	return m.Main.IsPreviewImage()
}

// ContainsPreview reports whether the file list itself includes a preview image
// (JPEG/PNG). Unlike HasPreview it ignores Main, so it distinguishes a group
// whose primary preview is being indexed from an incremental sidecar-only update
// where the unchanged preview was filtered out of the list.
func (m RelatedFiles) ContainsPreview() bool {
	for _, f := range m.Files {
		if f.IsPreviewImage() {
			return true
		}
	}

	return false
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

	return clean.Log(m.Main.RootRelName())
}
