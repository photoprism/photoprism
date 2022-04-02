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
	FormatMpo      FileFormat = "mpo"  // Stereoscopic Image that consists of two JPG images that are combined into one 3D image
	FormatHEIF     FileFormat = "heif" // High Efficiency Image File Format
	FormatWebP     FileFormat = "webp" // Google WebP Image
	FormatWebM     FileFormat = "webm" // Google WebM Video
	FormatHEVC     FileFormat = "hevc" // H.265, High Efficiency Video Coding (HEVC)
	FormatAvc      FileFormat = "avc"  // H.264, Advanced Video Coding (AVC), MPEG-4 Part 10, used internally
	FormatMov      FileFormat = "mov"  // QuickTime File Format, can contain AVC, HEVC,...
	FormatMp4      FileFormat = "mp4"  // Standard MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	FormatAvi      FileFormat = "avi"  // Microsoft Audio Video Interleave (AVI)
	Format3gp      FileFormat = "3gp"  // Mobile Multimedia Container Format, MPEG-4 Part 12
	Format3g2      FileFormat = "3g2"  // Similar to 3GP, consumes less space & bandwidth
	FormatFlv      FileFormat = "flv"  // Flash Video
	FormatMkv      FileFormat = "mkv"  // Matroska Multimedia Container, free and open
	FormatMpg      FileFormat = "mpg"  // Moving Picture Experts Group (MPEG)
	FormatMts      FileFormat = "mts"  // AVCHD (Advanced Video Coding High Definition)
	FormatOgv      FileFormat = "ogv"  // Ogg container format maintained by the Xiph.Org, free and open
	FormatWMV      FileFormat = "wmv"  // Windows Media Video
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
	".mp":   FormatMp4,
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
	".webp": FormatWebP,
	".webm": FormatWebM,
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
