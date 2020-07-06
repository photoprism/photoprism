package photoprism

// List of related files for importing and indexing.
type RelatedFiles struct {
	Files MediaFiles
	Main  *MediaFile
}

// ContainsJpeg returns true if related file list contains a JPEG.
func (rf RelatedFiles) ContainsJpeg() bool {
	for _, f := range rf.Files {
		if f.IsJpeg() {
			return true
		}
	}

	if rf.Main == nil {
		return false
	}

	return rf.Main.IsJpeg()
}
