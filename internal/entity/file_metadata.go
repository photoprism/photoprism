package entity

// FileDimensions represents metadata related to the size and orientation of a file.
// see File.UpdateVideoInfos()
type FileDimensions struct {
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float32
}

// FileAppearance represents file metadata related to colors, luminance and perception.
// see File.UpdateVideoInfos()
type FileAppearance struct {
	FileMainColor string
	FileColors    string
	FileLuminance string
	FileDiff      int
	FileChroma    int16
}
