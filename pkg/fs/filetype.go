package fs

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"
)

type FileType string

const (
	TypeJpeg     FileType = "jpg"  // JPEG image file.
	TypePng      FileType = "png"  // PNG image file.
	TypeGif      FileType = "gif"  // GIF image file.
	TypeTiff     FileType = "tiff" // TIFF image file.
	TypeBitmap   FileType = "bmp"  // BMP image file.
	TypeRaw      FileType = "raw"  // RAW image file.
	TypeHEIF     FileType = "heif" // High Efficiency Image File Format
	TypeMov      FileType = "mov"  // Video files.
	TypeMP4      FileType = "mp4"
	TypeAvi      FileType = "avi"
	TypeXMP      FileType = "xmp"     // Adobe XMP sidecar file (XML).
	TypeAAE      FileType = "aae"     // Apple sidecar file (XML).
	TypeXML      FileType = "xml"     // XML metadata / config / sidecar file.
	TypeYaml     FileType = "yml"     // YAML metadata / config / sidecar file.
	TypeToml     FileType = "toml"    // Tom's Obvious, Minimal Language sidecar file.
	TypeJson     FileType = "json"    // JSON metadata / config / sidecar file.
	TypeText     FileType = "txt"     // Text config / sidecar file.
	TypeMarkdown FileType = "md"      // Markdown text sidecar file.
	TypeOther    FileType = "unknown" // Unknown file format.
)

// FileExt contains the filename extensions of file formats known to PhotoPrism.
var FileExt = map[string]FileType{
	".bmp":  TypeBitmap,
	".gif":  TypeGif,
	".tif":  TypeTiff,
	".tiff": TypeTiff,
	".png":  TypePng,
	".crw":  TypeRaw,
	".cr2":  TypeRaw,
	".nef":  TypeRaw,
	".arw":  TypeRaw,
	".dng":  TypeRaw,
	".mov":  TypeMov,
	".avi":  TypeAvi,
	".mp4":  TypeMP4,
	".yml":  TypeYaml,
	".jpg":  TypeJpeg,
	".thm":  TypeJpeg,
	".jpeg": TypeJpeg,
	".xmp":  TypeXMP,
	".aae":  TypeAAE,
	".heif": TypeHEIF,
	".heic": TypeHEIF,
	".3fr":  TypeRaw,
	".ari":  TypeRaw,
	".bay":  TypeRaw,
	".cr3":  TypeRaw,
	".cap":  TypeRaw,
	".data": TypeRaw,
	".dcs":  TypeRaw,
	".dcr":  TypeRaw,
	".drf":  TypeRaw,
	".eip":  TypeRaw,
	".erf":  TypeRaw,
	".fff":  TypeRaw,
	".gpr":  TypeRaw,
	".iiq":  TypeRaw,
	".k25":  TypeRaw,
	".kdc":  TypeRaw,
	".mdc":  TypeRaw,
	".mef":  TypeRaw,
	".mos":  TypeRaw,
	".mrw":  TypeRaw,
	".nrw":  TypeRaw,
	".obm":  TypeRaw,
	".orf":  TypeRaw,
	".pef":  TypeRaw,
	".ptx":  TypeRaw,
	".pxn":  TypeRaw,
	".r3d":  TypeRaw,
	".raf":  TypeRaw,
	".raw":  TypeRaw,
	".rwl":  TypeRaw,
	".rw2":  TypeRaw,
	".rwz":  TypeRaw,
	".sr2":  TypeRaw,
	".srf":  TypeRaw,
	".srw":  TypeRaw,
	".x3f":  TypeRaw,
	".xml":  TypeXML,
	".txt":  TypeText,
	".md":   TypeMarkdown,
	".json": TypeJson,
}

func GetFileType(fileName string) FileType {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	result, ok := FileExt[fileExt]

	if !ok {
		result = TypeOther
	}

	return result
}
