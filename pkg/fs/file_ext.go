package fs

import (
	"path/filepath"
	"strings"
)

const (
	ExtNone     = ""
	ExtLocal    = ".local"
	ExtPDF      = ".pdf"
	ExtJpeg     = ".jpg"
	ExtPng      = ".png"
	ExtDng      = ".dng"
	ExtThm      = ".thm"
	ExtH264     = ".h264"
	ExtAvc      = ".avc"
	ExtAvc1     = ".avc1"
	ExtDva      = ".dva"
	ExtDva1     = ".dva1"
	ExtAvc2     = ".avc2"
	ExtAvc3     = ".avc3"
	ExtDvav     = ".dvav"
	ExtAvc10    = ".avc10"
	ExtH265     = ".h265"
	ExtHvc      = ".hvc"
	ExtHvc1     = ".hvc1"
	ExtDvh      = ".dvh"
	ExtDvh1     = ".dvh1"
	ExtHvc2     = ".hvc2"
	ExtHvc3     = ".hvc3"
	ExtHvc10    = ".hvc10"
	ExtHevc     = ".hevc"
	ExtHevc10   = ".hevc10"
	ExtHev      = ".hev"
	ExtDvhe     = ".dvhe"
	ExtHev1     = ".hev1"
	ExtHev2     = ".hev2"
	ExtHev3     = ".hev3"
	ExtHev10    = ".hev10"
	ExtH266     = ".h266"
	ExtVvc      = ".vvc"
	ExtVvc1     = ".vvc1"
	ExtEvc      = ".evc"
	ExtEvc1     = ".evc1"
	ExtMp4      = ".mp4"
	ExtMov      = ".mov"
	ExtQT       = ".qt"
	ExtYml      = ".yml"
	ExtYaml     = ".yaml"
	ExtTml      = ".tml"
	ExtToml     = ".toml"
	ExtJson     = ".json"
	ExtGeoJson  = ".geojson"
	ExtXml      = ".xml"
	ExtXMP      = ".xmp"
	ExtHTM      = ".htm"
	ExtHTML     = ".html"
	ExtXHTML    = ".xhtml"
	ExtTxt      = ".txt"
	ExtMd       = ".md"
	ExtMarkdown = ".markdown"
	ExtPb       = ".pb"
	ExtProto    = ".proto"
	ExtZip      = ".zip"
)

// Ext returns all extension of a file name including the dots.
func Ext(name string) string {
	if name == "" {
		return ""
	}

	ext := filepath.Ext(name)
	name = StripExt(name)

	if Extensions.Known(name) {
		ext = filepath.Ext(name) + ext
	}

	return ext
}

// ArchiveExt returns the normalized archive file extension or an empty string if it is not an archive.
func ArchiveExt(name string) string {
	switch strings.ToLower(Ext(name)) {
	case ExtZip:
		return ExtZip
	default:
		return ""
	}
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
