package fs

import (
	"path/filepath"
	"strings"
)

const (
	ExtJPEG = ".jpg"
	ExtPNG  = ".png"
	ExtDNG  = ".dng"
	ExtTHM  = ".thm"
	ExtAVC  = ".avc"
	ExtHEVC = ".hevc"
	ExtVVC  = ".vvc"
	ExtEVC  = ".evc"
	ExtMP4  = ".mp4"
	ExtMOV  = ".mov"
	ExtYAML = ".yml"
)

// Ext returns all extension of a file name including the dots.
func Ext(name string) string {
	ext := filepath.Ext(name)
	name = StripExt(name)

	if Extensions.Known(name) {
		ext = filepath.Ext(name) + ext
	}

	return ext
}

// NormalizedExt returns the file extension without dot and in lowercase.
func NormalizedExt(fileName string) string {
	if dot := strings.LastIndex(fileName, "."); dot != -1 && len(fileName[dot+1:]) >= 1 {
		return strings.ToLower(fileName[dot+1:])
	}

	return ""
}

// LowerExt returns the file name extension with dot in lower case.
func LowerExt(fileName string) string {
	if fileName == "" {
		return ""
	}

	return strings.ToLower(filepath.Ext(fileName))
}

// TrimExt removes unwanted characters from file extension strings, and makes it lowercase for comparison.
func TrimExt(ext string) string {
	return strings.ToLower(strings.Trim(ext, " .,;:“”'`\""))
}

// StripExt removes the file type extension from a file name (if any).
func StripExt(name string) string {
	if end := strings.LastIndex(name, "."); end != -1 {
		name = name[:end]
	}

	return name
}

// StripKnownExt removes all known file type extension from a file name (if any).
func StripKnownExt(name string) string {
	for Extensions.Known(name) {
		name = StripExt(name)
	}

	return name
}
