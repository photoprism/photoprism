package file

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
)

type Type string

const (
	// JPEG image file.
	TypeJpeg Type = "jpg"
	// PNG image file.
	TypePng Type = "png"
	// RAW image file.
	TypeRaw Type = "raw"
	// High Efficiency Image File Format.
	TypeHEIF Type = "heif" // High Efficiency Image File Format
	// Movie file.
	TypeMovie Type = "mov"
	// Adobe XMP sidecar file (XML).
	TypeXMP Type = "xmp"
	// Apple sidecar file (XML).
	TypeAAE Type = "aae"
	// XML metadata / config / sidecar file.
	TypeXML Type = "xml"
	// YAML metadata / config / sidecar file.
	TypeYaml Type = "yml"
	// Text config / sidecar file.
	TypeText Type = "txt"
	// Markdown text sidecar file.
	TypeMarkdown Type = "md"
	// Unknown file format.
	TypeOther Type = "unknown"
)

const (
	// MimeTypeJpeg is jpeg image type
	MimeTypeJpeg = "image/jpeg"
)

// Ext lists all the available and supported image file formats.
var Ext = map[string]Type{
	".crw":  TypeRaw,
	".cr2":  TypeRaw,
	".nef":  TypeRaw,
	".arw":  TypeRaw,
	".dng":  TypeRaw,
	".mov":  TypeMovie,
	".avi":  TypeMovie,
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
	".tif":  TypeRaw,
	".x3f":  TypeRaw,
}
