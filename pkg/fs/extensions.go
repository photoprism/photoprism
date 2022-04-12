package fs

import (
	"path/filepath"
	"strings"
)

// FileExtensions maps file extensions to standard formats
type FileExtensions map[string]Format

// Extensions contains the filename extensions of file formats known to PhotoPrism.
var Extensions = FileExtensions{
	".jpg":      FormatJpeg,
	".jpeg":     FormatJpeg,
	".jpe":      FormatJpeg,
	".jif":      FormatJpeg,
	".jfif":     FormatJpeg,
	".jfi":      FormatJpeg,
	".thm":      FormatJpeg,
	".3fr":      FormatRaw,
	".ari":      FormatRaw,
	".arw":      FormatRaw,
	".bay":      FormatRaw,
	".cap":      FormatRaw,
	".crw":      FormatRaw,
	".cr2":      FormatRaw,
	".cr3":      FormatRaw,
	".data":     FormatRaw,
	".dcs":      FormatRaw,
	".dcr":      FormatRaw,
	".dng":      FormatRaw,
	".drf":      FormatRaw,
	".eip":      FormatRaw,
	".erf":      FormatRaw,
	".fff":      FormatRaw,
	".gpr":      FormatRaw,
	".iiq":      FormatRaw,
	".k25":      FormatRaw,
	".kdc":      FormatRaw,
	".mdc":      FormatRaw,
	".mef":      FormatRaw,
	".mos":      FormatRaw,
	".mrw":      FormatRaw,
	".nef":      FormatRaw,
	".nrw":      FormatRaw,
	".obm":      FormatRaw,
	".orf":      FormatRaw,
	".pef":      FormatRaw,
	".ptx":      FormatRaw,
	".pxn":      FormatRaw,
	".r3d":      FormatRaw,
	".raf":      FormatRaw,
	".raw":      FormatRaw,
	".rwl":      FormatRaw,
	".rwz":      FormatRaw,
	".rw2":      FormatRaw,
	".srf":      FormatRaw,
	".srw":      FormatRaw,
	".sr2":      FormatRaw,
	".x3f":      FormatRaw,
	".png":      FormatPng,
	".pn":       FormatPng,
	".tif":      FormatTiff,
	".tiff":     FormatTiff,
	".gif":      FormatGif,
	".bmp":      FormatBitmap,
	".heif":     FormatHEIF,
	".heic":     FormatHEIF,
	".hevc":     FormatHEVC,
	".mov":      FormatMov,
	".qt":       FormatMov,
	".avi":      FormatAvi,
	".av1":      FormatAV1,
	".avc":      FormatAVC,
	".mpg":      FormatMPEG,
	".mpeg":     FormatMPEG,
	".mjpg":     FormatMJPEG,
	".mjpeg":    FormatMJPEG,
	".mp2":      FormatMp2,
	".mpv":      FormatMp2,
	".mp":       FormatMp4,
	".mp4":      FormatMp4,
	".m4v":      FormatMp4,
	".3gp":      Format3gp,
	".3g2":      Format3g2,
	".flv":      FormatFlv,
	".f4v":      FormatFlv,
	".mkv":      FormatMkv,
	".mpo":      FormatMpo,
	".mts":      FormatMts,
	".ogv":      FormatOgv,
	".ogg":      FormatOgv,
	".ogx":      FormatOgv,
	".webp":     FormatWebP,
	".webm":     FormatWebM,
	".wmv":      FormatWMV,
	".aae":      FormatAAE,
	".md":       FormatMarkdown,
	".markdown": FormatMarkdown,
	".json":     FormatJson,
	".toml":     FormatToml,
	".txt":      FormatText,
	".yml":      FormatYaml,
	".yaml":     FormatYaml,
	".xmp":      FormatXMP,
	".xml":      FormatXML,
}

// Known tests if the file extension is known (supported).
func (m FileExtensions) Known(name string) bool {
	if name == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(name))

	if ext == "" {
		return false
	}

	if _, ok := m[ext]; ok {
		return true
	}

	return false
}

// Formats returns all known file extensions by format.
func (m FileExtensions) Formats(noUppercase bool) FileFormats {
	result := make(FileFormats)

	if noUppercase {
		for ext, t := range m {
			if _, ok := result[t]; ok {
				result[t] = append(result[t], ext)
			} else {
				result[t] = []string{ext}
			}
		}
	} else {
		for ext, t := range m {
			extUpper := strings.ToUpper(ext)
			if _, ok := result[t]; ok {
				result[t] = append(result[t], ext, extUpper)
			} else {
				result[t] = []string{ext, extUpper}
			}
		}
	}

	return result
}
