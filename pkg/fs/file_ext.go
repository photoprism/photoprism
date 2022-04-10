package fs

import (
	"path/filepath"
	"strings"
)

// FileExt contains the filename extensions of file formats known to PhotoPrism.
var FileExt = FileExtensions{
	".jpg":  FormatJpeg,
	".jpeg": FormatJpeg,
	".jpe":  FormatJpeg,
	".jif":  FormatJpeg,
	".jfif": FormatJpeg,
	".jfi":  FormatJpeg,
	".thm":  FormatJpeg,
	".3fr":  FormatRaw,
	".ari":  FormatRaw,
	".arw":  FormatRaw,
	".bay":  FormatRaw,
	".cap":  FormatRaw,
	".crw":  FormatRaw,
	".cr2":  FormatRaw,
	".cr3":  FormatRaw,
	".cr4":  FormatRaw,
	".data": FormatRaw,
	".dcs":  FormatRaw,
	".dcr":  FormatRaw,
	".dng":  FormatRaw,
	".drf":  FormatRaw,
	".eip":  FormatRaw,
	".erf":  FormatRaw,
	".fff":  FormatRaw,
	".gpr":  FormatRaw,
	".iiq":  FormatRaw,
	".k25":  FormatRaw,
	".kdc":  FormatRaw,
	".mdc":  FormatRaw,
	".mef":  FormatRaw,
	".mos":  FormatRaw,
	".mrw":  FormatRaw,
	".nef":  FormatRaw,
	".nrw":  FormatRaw,
	".obm":  FormatRaw,
	".orf":  FormatRaw,
	".pef":  FormatRaw,
	".ptx":  FormatRaw,
	".pxn":  FormatRaw,
	".r3d":  FormatRaw,
	".raf":  FormatRaw,
	".raw":  FormatRaw,
	".rwl":  FormatRaw,
	".rwz":  FormatRaw,
	".rw2":  FormatRaw,
	".srf":  FormatRaw,
	".srw":  FormatRaw,
	".sr2":  FormatRaw,
	".x3f":  FormatRaw,
	".png":  FormatPng,
	".pn":   FormatPng,
	".tif":  FormatTiff,
	".tiff": FormatTiff,
	".gif":  FormatGif,
	".bmp":  FormatBitmap,
	".heif": FormatHEIF,
	".heic": FormatHEIF,
	".hevc": FormatHEVC,
	".mov":  FormatMov,
	".avi":  FormatAvi,
	".avc":  FormatAvc,
	".mp":   FormatMp4,
	".mp4":  FormatMp4,
	".m4v":  FormatMp4,
	".mpg":  FormatMpg,
	".mpeg": FormatMpg,
	".3gp":  Format3gp,
	".3g2":  Format3g2,
	".flv":  FormatFlv,
	".mkv":  FormatMkv,
	".mpo":  FormatMpo,
	".mts":  FormatMts,
	".ogv":  FormatOgv,
	".webp": FormatWebP,
	".webm": FormatWebM,
	".wmv":  FormatWMV,
	".aae":  FormatAAE,
	".md":   FormatMarkdown,
	".json": FormatJson,
	".txt":  FormatText,
	".yml":  FormatYaml,
	".yaml": FormatYaml,
	".xmp":  FormatXMP,
	".xml":  FormatXML,
}

// TypeExt contains the default file type extensions.
var TypeExt = FileExt.TypeExt()

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

func (m FileExtensions) TypeExt() TypeExtensions {
	result := make(TypeExtensions)

	if ignoreCase {
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
