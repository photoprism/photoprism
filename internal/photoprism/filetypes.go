package photoprism

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
)

type FileType string

const (
	// JPEG image file.
	FileTypeJpeg FileType = "jpg"
	// PNG image file.
	FileTypePng FileType = "png"
	// RAW image file.
	FileTypeRaw FileType = "raw"
	// High Efficiency Image File Format.
	FileTypeHEIF FileType = "heif" // High Efficiency Image File Format
	// Movie file.
	FileTypeMovie FileType = "mov"
	// Adobe XMP sidecar file (XML).
	FileTypeXMP FileType = "xmp"
	// Apple sidecar file (XML).
	FileTypeAAE FileType = "aae"
	// XML metadata / config / sidecar file.
	FileTypeXML FileType = "xml"
	// YAML metadata / config / sidecar file.
	FileTypeYaml FileType = "yml"
	// Text config / sidecar file.
	FileTypeText FileType = "txt"
	// Markdown text sidecar file.
	FileTypeMarkdown FileType = "md"
	// Unknown file format.
	FileTypeOther FileType = "unknown"
)

const (
	// MimeTypeJpeg is jpeg image type
	MimeTypeJpeg = "image/jpeg"
)

// FileExtensions lists all the available and supported image file formats.
var FileExtensions = map[string]FileType{
	".crw":  FileTypeRaw,
	".cr2":  FileTypeRaw,
	".nef":  FileTypeRaw,
	".arw":  FileTypeRaw,
	".dng":  FileTypeRaw,
	".mov":  FileTypeMovie,
	".avi":  FileTypeMovie,
	".yml":  FileTypeYaml,
	".jpg":  FileTypeJpeg,
	".thm":  FileTypeJpeg,
	".jpeg": FileTypeJpeg,
	".xmp":  FileTypeXMP,
	".aae":  FileTypeAAE,
	".heif": FileTypeHEIF,
	".heic": FileTypeHEIF,
	".3fr":  FileTypeRaw,
	".ari":  FileTypeRaw,
	".bay":  FileTypeRaw,
	".cr3":  FileTypeRaw,
	".cap":  FileTypeRaw,
	".data": FileTypeRaw,
	".dcs":  FileTypeRaw,
	".dcr":  FileTypeRaw,
	".drf":  FileTypeRaw,
	".eip":  FileTypeRaw,
	".erf":  FileTypeRaw,
	".fff":  FileTypeRaw,
	".gpr":  FileTypeRaw,
	".iiq":  FileTypeRaw,
	".k25":  FileTypeRaw,
	".kdc":  FileTypeRaw,
	".mdc":  FileTypeRaw,
	".mef":  FileTypeRaw,
	".mos":  FileTypeRaw,
	".mrw":  FileTypeRaw,
	".nrw":  FileTypeRaw,
	".obm":  FileTypeRaw,
	".orf":  FileTypeRaw,
	".pef":  FileTypeRaw,
	".ptx":  FileTypeRaw,
	".pxn":  FileTypeRaw,
	".r3d":  FileTypeRaw,
	".raf":  FileTypeRaw,
	".raw":  FileTypeRaw,
	".rwl":  FileTypeRaw,
	".rw2":  FileTypeRaw,
	".rwz":  FileTypeRaw,
	".sr2":  FileTypeRaw,
	".srf":  FileTypeRaw,
	".srw":  FileTypeRaw,
	".tif":  FileTypeRaw,
	".x3f":  FileTypeRaw,
}
