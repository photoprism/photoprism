package fs

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
	"os"
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
	TypeMp4      FileType = "mp4"
	TypeHEVC     FileType = "hevc"
	TypeAvi      FileType = "avi"
	Type3gp      FileType = "3gp"
	Type3g2      FileType = "3g2"
	TypeFlv      FileType = "flv"
	TypeMkv      FileType = "mkv"
	TypeMpg      FileType = "mpg"
	TypeOgv      FileType = "ogv"
	TypeWebm     FileType = "webm"
	TypeWMV      FileType = "wmv"
	TypeXMP      FileType = "xmp"  // Adobe XMP sidecar file (XML).
	TypeAAE      FileType = "aae"  // Apple sidecar file (XML).
	TypeXML      FileType = "xml"  // XML metadata / config / sidecar file.
	TypeYaml     FileType = "yml"  // YAML metadata / config / sidecar file.
	TypeToml     FileType = "toml" // Tom's Obvious, Minimal Language sidecar file.
	TypeJson     FileType = "json" // JSON metadata / config / sidecar file.
	TypeText     FileType = "txt"  // Text config / sidecar file.
	TypeMarkdown FileType = "md"   // Markdown text sidecar file.
	TypeOther    FileType = ""     // Unknown file format.
)

type FileExtensions map[string]FileType
type TypeExtensions map[FileType][]string

const (
	YamlExt = ".yml"
	JpegExt = ".jpg"
	AvcExt  = ".mp4"
	HevcExt = ".hevc"
)

// FileExt contains the filename extensions of file formats known to PhotoPrism.
var FileExt = FileExtensions{
	".bmp":  TypeBitmap,
	".gif":  TypeGif,
	".tif":  TypeTiff,
	".tiff": TypeTiff,
	".png":  TypePng,
	".pn":   TypePng,
	".crw":  TypeRaw,
	".cr2":  TypeRaw,
	".cr3":  TypeRaw,
	".nef":  TypeRaw,
	".arw":  TypeRaw,
	".dng":  TypeRaw,
	".mov":  TypeMov,
	".avi":  TypeAvi,
	".mp4":  TypeMp4,
	".hevc": TypeHEVC,
	".3gp":  Type3gp,
	".3g2":  Type3g2,
	".flv":  TypeFlv,
	".mkv":  TypeMkv,
	".mpg":  TypeMpg,
	".mpeg": TypeMpg,
	".ogv":  TypeOgv,
	".webm": TypeWebm,
	".wmv":  TypeWMV,
	".yml":  TypeYaml,
	".yaml": TypeYaml,
	".jpg":  TypeJpeg,
	".jpeg": TypeJpeg,
	".jpe":  TypeJpeg,
	".jif":  TypeJpeg,
	".jfif": TypeJpeg,
	".jfi":  TypeJpeg,
	".thm":  TypeJpeg,
	".xmp":  TypeXMP,
	".aae":  TypeAAE,
	".heif": TypeHEIF,
	".heic": TypeHEIF,
	".3fr":  TypeRaw,
	".ari":  TypeRaw,
	".bay":  TypeRaw,
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

	for ext, t := range m {
		extUpper := strings.ToUpper(ext)
		if _, ok := result[t]; ok {
			result[t] = append(result[t], ext, extUpper)
		} else {
			result[t] = []string{ext, extUpper}
		}
	}

	return result
}

var TypeExt = FileExt.TypeExt()

// Find returns the first filename with the same base name and a given type.
func (t FileType) Find(fileName string, stripSequence bool) string {
	base := BasePrefix(fileName, stripSequence)
	dir := filepath.Dir(fileName)

	prefix := filepath.Join(dir, base)
	prefixLower := filepath.Join(dir, strings.ToLower(base))
	prefixUpper := filepath.Join(dir, strings.ToUpper(base))

	for _, ext := range TypeExt[t] {
		if info, err := os.Stat(prefix + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}

		if info, err := os.Stat(prefixLower + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}

		if info, err := os.Stat(prefixUpper + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}
	}

	return ""
}

// GetFileType returns the (expected) type for a given file name.
func GetFileType(fileName string) FileType {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	result, ok := FileExt[fileExt]

	if !ok {
		result = TypeOther
	}

	return result
}

// FindFirst searches a list of directories for the first file with the same base name and a given type.
func (t FileType) FindFirst(fileName string, dirs []string, baseDir string, stripSequence bool) string {
	fileBase := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBase)
	fileBaseUpper := strings.ToUpper(fileBase)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range TypeExt[t] {
		lastDir := ""

		for _, dir := range search {
			if dir == "" || dir == lastDir {
				continue
			}

			lastDir = dir

			if dir != fileDir {
				if filepath.IsAbs(dir) {
					dir = filepath.Join(dir, RelName(fileDir, baseDir))
				} else {
					dir = filepath.Join(fileDir, dir)
				}
			}

			if info, err := os.Stat(filepath.Join(dir, fileBase) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseLower) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseUpper) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}
		}
	}

	return ""
}

// NormalizedExt returns the file extension without dot and in lowercase.
func NormalizedExt(fileName string) string {
	if dot := strings.LastIndex(fileName, "."); dot != -1 && len(fileName[dot+1:]) >= 1 {
		return strings.ToLower(fileName[dot+1:])
	}

	return ""
}
