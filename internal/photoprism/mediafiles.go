package photoprism

// MediaFiles represents a slice of files.
type MediaFiles []*MediaFile

// Len returns the length of the file slice.
func (f MediaFiles) Len() int {
	return len(f)
}

// Less compares two files based on the filename.
func (f MediaFiles) Less(i, j int) bool {
	fileName1 := f[i].FileName()
	fileName2 := f[j].FileName()

	if len(fileName1) == len(fileName2) {
		return fileName1 < fileName2
	}

	return len(fileName1) < len(fileName2)
}

// Swap the position of two files in the slice.
func (f MediaFiles) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
