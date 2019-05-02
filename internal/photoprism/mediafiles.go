package photoprism

// MediaFiles provides a Collection for mediafiles.
type MediaFiles []*MediaFile

// Len returns the length of the mediafile collection.
func (f MediaFiles) Len() int {
	return len(f)
}

// Less compares two mediafiles based on filename length.
func (f MediaFiles) Less(i, j int) bool {
	fileName1 := f[i].Filename()
	fileName2 := f[j].Filename()

	if len(fileName1) == len(fileName2) {
		return fileName1 < fileName2
	}

	return len(fileName1) < len(fileName2)
}

// Swap changes two files order around.
func (f MediaFiles) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
