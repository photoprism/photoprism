package fs

import (
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

type FileFormat string

const (
	FormatJpeg     FileFormat = "jpg"  // JPEG image file.
	FormatPng      FileFormat = "png"  // PNG image file.
	FormatGif      FileFormat = "gif"  // GIF image file.
	FormatTiff     FileFormat = "tiff" // TIFF image file.
	FormatBitmap   FileFormat = "bmp"  // BMP image file.
	FormatRaw      FileFormat = "raw"  // RAW image file.
	FormatHEIF     FileFormat = "heif" // High Efficiency Image File Format
	FormatHEVC     FileFormat = "hevc"
	FormatMov      FileFormat = "mov" // Video files.
	FormatMp4      FileFormat = "mp4"
	FormatMpo      FileFormat = "mpo"
	FormatAvc      FileFormat = "avc"
	FormatAvi      FileFormat = "avi"
	Format3gp      FileFormat = "3gp"
	Format3g2      FileFormat = "3g2"
	FormatFlv      FileFormat = "flv"
	FormatMkv      FileFormat = "mkv"
	FormatMpg      FileFormat = "mpg"
	FormatMts      FileFormat = "mts"
	FormatOgv      FileFormat = "ogv"
	FormatWebm     FileFormat = "webm"
	FormatWMV      FileFormat = "wmv"
	FormatXMP      FileFormat = "xmp"  // Adobe XMP sidecar file (XML).
	FormatAAE      FileFormat = "aae"  // Apple sidecar file (XML).
	FormatXML      FileFormat = "xml"  // XML metadata / config / sidecar file.
	FormatYaml     FileFormat = "yml"  // YAML metadata / config / sidecar file.
	FormatToml     FileFormat = "toml" // Tom's Obvious, Minimal Language sidecar file.
	FormatJson     FileFormat = "json" // JSON metadata / config / sidecar file.
	FormatText     FileFormat = "txt"  // Text config / sidecar file.
	FormatMarkdown FileFormat = "md"   // Markdown text sidecar file.
	FormatOther    FileFormat = ""     // Unknown file format.
)

type FileExtensions map[string]FileFormat
type TypeExtensions map[FileFormat][]string

const (
	YamlExt     = ".yml"
	JpegExt     = ".jpg"
	AvcExt      = ".avc"
	FujiRawExt  = ".raf"
	CanonCr3Ext = ".cr3"
)

// FileExt contains the filename extensions of file formats known to PhotoPrism.
var FileExt = FileExtensions{
	".bmp":  FormatBitmap,
	".gif":  FormatGif,
	".tif":  FormatTiff,
	".tiff": FormatTiff,
	".png":  FormatPng,
	".pn":   FormatPng,
	".crw":  FormatRaw,
	".cr2":  FormatRaw,
	".cr3":  FormatRaw,
	".nef":  FormatRaw,
	".arw":  FormatRaw,
	".dng":  FormatRaw,
	".mov":  FormatMov,
	".avi":  FormatAvi,
	".mp4":  FormatMp4,
	".m4v":  FormatMp4,
	".avc":  FormatAvc,
	".hevc": FormatHEVC,
	".3gp":  Format3gp,
	".3g2":  Format3g2,
	".flv":  FormatFlv,
	".mkv":  FormatMkv,
	".mpg":  FormatMpg,
	".mpeg": FormatMpg,
	".mpo":  FormatMpo,
	".mts":  FormatMts,
	".ogv":  FormatOgv,
	".webm": FormatWebm,
	".wmv":  FormatWMV,
	".yml":  FormatYaml,
	".yaml": FormatYaml,
	".jpg":  FormatJpeg,
	".jpeg": FormatJpeg,
	".jpe":  FormatJpeg,
	".jif":  FormatJpeg,
	".jfif": FormatJpeg,
	".jfi":  FormatJpeg,
	".thm":  FormatJpeg,
	".xmp":  FormatXMP,
	".aae":  FormatAAE,
	".heif": FormatHEIF,
	".heic": FormatHEIF,
	".3fr":  FormatRaw,
	".ari":  FormatRaw,
	".bay":  FormatRaw,
	".cap":  FormatRaw,
	".data": FormatRaw,
	".dcs":  FormatRaw,
	".dcr":  FormatRaw,
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
	".rw2":  FormatRaw,
	".rwz":  FormatRaw,
	".sr2":  FormatRaw,
	".srf":  FormatRaw,
	".srw":  FormatRaw,
	".x3f":  FormatRaw,
	".xml":  FormatXML,
	".txt":  FormatText,
	".md":   FormatMarkdown,
	".json": FormatJson,
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

var TypeExt = FileExt.TypeExt()

// Find returns the first filename with the same base name and a given type.
func (t FileFormat) Find(fileName string, stripSequence bool) string {
	base := BasePrefix(fileName, stripSequence)
	dir := filepath.Dir(fileName)

	prefix := filepath.Join(dir, base)
	prefixLower := filepath.Join(dir, strings.ToLower(base))
	prefixUpper := filepath.Join(dir, strings.ToUpper(base))

	for _, ext := range TypeExt[t] {
		if info, err := os.Stat(prefix + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}

		if ignoreCase {
			continue
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

// GetFileFormat returns the (expected) type for a given file name.
func GetFileFormat(fileName string) FileFormat {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	result, ok := FileExt[fileExt]

	if !ok {
		result = FormatOther
	}

	return result
}

// FindFirst searches a list of directories for the first file with the same base name and a given type.
func (t FileFormat) FindFirst(fileName string, dirs []string, baseDir string, stripSequence bool) string {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

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
			} else if info, err := os.Stat(filepath.Join(dir, fileBasePrefix) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}

			if ignoreCase {
				continue
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseLower) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			} else if info, err := os.Stat(filepath.Join(dir, fileBaseUpper) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}
		}
	}

	return ""
}

// FindAll searches a list of directories for files with the same base name and a given type.
func (t FileFormat) FindAll(fileName string, dirs []string, baseDir string, stripSequence bool) (results []string) {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

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
				results = append(results, filepath.Join(dir, info.Name()))
			}

			if info, err := os.Stat(filepath.Join(dir, fileBasePrefix) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}

			if ignoreCase {
				continue
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseLower) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseUpper) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}
		}
	}

	return results
}

// NormalizedExt returns the file extension without dot and in lowercase.
func NormalizedExt(fileName string) string {
	if dot := strings.LastIndex(fileName, "."); dot != -1 && len(fileName[dot+1:]) >= 1 {
		return strings.ToLower(fileName[dot+1:])
	}

	return ""
}
