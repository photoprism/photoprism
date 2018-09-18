package photoprism

type MediaFiles []*MediaFile

func (f MediaFiles) Len() int {
	return len(f)
}

func (f MediaFiles) Less(i, j int) bool {
	fileName1 := f[i].GetFilename()
	fileName2 := f[j].GetFilename()

	if len(fileName1) == len(fileName2) {
		return fileName1 < fileName2
	} else {
		return len(fileName1) < len(fileName2)
	}
}

func (f MediaFiles) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
