package photoprism

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
)

const (
	// FileTypeOther is an unkown file format.
	FileTypeOther = "unknown"
	// FileTypeYaml is a yaml file format.
	FileTypeYaml = "yml"
	// FileTypeJpeg is a jpeg file format.
	FileTypeJpeg = "jpg"
	// FileTypePng is a png file format.
	FileTypePng = "png"
	// FileTypeRaw is a raw file format.
	FileTypeRaw = "raw"
	// FileTypeXmp is an xmp file format.
	FileTypeXmp = "xmp"
	// FileTypeAae is an aae file format.
	FileTypeAae = "aae"
	// FileTypeMovie is a movie file format.
	FileTypeMovie = "mov"
	// FileTypeHEIF High Efficiency Image File Format
	FileTypeHEIF = "heif" // High Efficiency Image File Format
)

const (
	// MimeTypeJpeg is jpeg image type
	MimeTypeJpeg = "image/jpeg"
)

// FileExtensions lists all the available and supported image file formats.
var FileExtensions = map[string]string{
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
	".xmp":  FileTypeXmp,
	".aae":  FileTypeAae,
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
