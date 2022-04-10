package fs

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

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

// GetFileFormat returns the (expected) type for a given file name.
func GetFileFormat(fileName string) FileFormat {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	result, ok := FileExt[fileExt]

	if !ok {
		result = FormatOther
	}

	return result
}
