package fs

import (
	"strings"
)

const (
	YamlExt     = ".yml"
	JpegExt     = ".jpg"
	AvcExt      = ".avc"
	FujiRawExt  = ".raf"
	CanonCr3Ext = ".cr3"
)

// NormalizeExt returns the file extension without dot and in lowercase.
func NormalizeExt(fileName string) string {
	if dot := strings.LastIndex(fileName, "."); dot != -1 && len(fileName[dot+1:]) >= 1 {
		return strings.ToLower(fileName[dot+1:])
	}

	return ""
}

// TrimExt removes unwanted characters from file extension strings, and makes it lowercase for comparison.
func TrimExt(ext string) string {
	return strings.ToLower(strings.Trim(ext, " .,;:“”'`\""))
}
