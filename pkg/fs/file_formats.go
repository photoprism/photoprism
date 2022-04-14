package fs

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"sort"
	"strings"
	"unicode"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// FileFormats maps standard formats to file extensions.
type FileFormats map[Format][]string

// Formats contains the default file type extensions.
var Formats = Extensions.Formats(ignoreCase)

// Supported file formats.
const (
	FormatJpeg     Format = "jpg"  // JPEG image file.
	FormatPng      Format = "png"  // PNG image file.
	FormatGif      Format = "gif"  // GIF image file.
	FormatTiff     Format = "tiff" // TIFF image file.
	FormatBitmap   Format = "bmp"  // BMP image file.
	FormatRaw      Format = "raw"  // RAW image file.
	FormatMpo      Format = "mpo"  // Stereoscopic Image that consists of two JPG images that are combined into one 3D image
	FormatHEIF     Format = "heif" // High Efficiency Image File Format
	FormatWebP     Format = "webp" // Google WebP Image
	FormatWebM     Format = "webm" // Google WebM Video
	FormatHEVC     Format = "hevc" // H.265, High Efficiency Video Coding (HEVC)
	FormatAVC      Format = "avc"  // H.264, Advanced Video Coding (AVC), MPEG-4 Part 10, used internally
	FormatAV1      Format = "av1"  // Alliance for Open Media Video
	FormatMPEG     Format = "mpg"  // Moving Picture Experts Group (MPEG)
	FormatMJPEG    Format = "mjpg" // Motion JPEG (M-JPEG)
	FormatMov      Format = "mov"  // QuickTime File Format, can contain AVC, HEVC,...
	FormatMp2      Format = "mp2"  // MPEG-2, H.222/H.262
	FormatMp4      Format = "mp4"  // MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	FormatAvi      Format = "avi"  // Microsoft Audio Video Interleave (AVI)
	Format3gp      Format = "3gp"  // Mobile Multimedia Container, MPEG-4 Part 12
	Format3g2      Format = "3g2"  // Similar to 3GP, consumes less space & bandwidth
	FormatFlv      Format = "flv"  // Flash Video
	FormatMkv      Format = "mkv"  // Matroska Multimedia Container, free and open
	FormatMts      Format = "mts"  // AVCHD (Advanced Video Coding High Definition)
	FormatOgv      Format = "ogv"  // Ogg container format maintained by the Xiph.Org, free and open
	FormatWMV      Format = "wmv"  // Windows Media Video
	FormatXMP      Format = "xmp"  // Adobe XMP sidecar file (XML).
	FormatAAE      Format = "aae"  // Apple sidecar file (XML).
	FormatXML      Format = "xml"  // XML metadata / config / sidecar file.
	FormatYaml     Format = "yml"  // YAML metadata / config / sidecar file.
	FormatToml     Format = "toml" // Tom's Obvious, Minimal Language sidecar file.
	FormatJson     Format = "json" // JSON metadata / config / sidecar file.
	FormatText     Format = "txt"  // Text config / sidecar file.
	FormatMarkdown Format = "md"   // Markdown text sidecar file.
	FormatOther    Format = ""     // Unknown file format.
)

// FormatDesc contains human-readable descriptions for supported file formats
var FormatDesc = map[Format]string{
	FormatJpeg:     "Joint Photographic Experts Group (JPEG)",
	FormatPng:      "Portable Network Graphics",
	FormatGif:      "Graphics Interchange Format",
	FormatTiff:     "Tag Image File Format",
	FormatBitmap:   "Bitmap",
	FormatRaw:      "Unprocessed Sensor Data",
	FormatMpo:      "Stereoscopic (3D JPEG)",
	FormatHEIF:     "High Efficiency Image File Format (HEIF)",
	FormatWebP:     "Google WebP",
	FormatWebM:     "Google WebM",
	FormatHEVC:     "High Efficiency Video Coding (HEVC, HVC1, H.265)",
	FormatAVC:      "Advanced Video Coding (AVC, AVC1, H.264, MPEG-4 Part 10)",
	FormatAV1:      "AOMedia Video 1 (AV1, AV01)",
	FormatMov:      "Apple QuickTime (QT)",
	FormatMp2:      "MPEG 2 (H.262, H.222)",
	FormatMp4:      "Multimedia Container (MPEG-4 Part 14)",
	FormatAvi:      "Microsoft Audio Video Interleave",
	FormatWMV:      "Microsoft Windows Media",
	Format3gp:      "Mobile Multimedia Container (MPEG-4 Part 12)",
	Format3g2:      "Mobile Multimedia Container for CDMA2000 (based on 3GP)",
	FormatFlv:      "Flash Video",
	FormatMkv:      "Matroska Multimedia Container (MKV, MCF, EBML)",
	FormatMPEG:     "Moving Picture Experts Group (MPEG)",
	FormatMJPEG:    "Motion JPEG (M-JPEG)",
	FormatMts:      "Advanced Video Coding High Definition (AVCHD)",
	FormatOgv:      "Ogg Media by Xiph.Org",
	FormatXMP:      "Adobe Extensible Metadata Platform",
	FormatAAE:      "Apple Image Edits",
	FormatXML:      "Extensible Markup Language",
	FormatJson:     "Serialized JSON Data (Exiftool, Google Photos)",
	FormatYaml:     "Serialized YAML Data (Config, Metadata)",
	FormatToml:     "Serialized TOML Data (Tom's Obvious, Minimal Language)",
	FormatText:     "Plain Text",
	FormatMarkdown: "Markdown Formatted Text",
	FormatOther:    "Other",
}

// Report returns a file format documentation table.
func (m FileFormats) Report(withDesc, withType, withExt bool) (rows [][]string, cols []string) {
	cols = make([]string, 0, 4)
	cols = append(cols, "Format")

	t := 0

	if withDesc {
		cols = append(cols, "Description")
	}

	if withType {
		if withDesc {
			t = 2
		} else {
			t = 1
		}

		cols = append(cols, "Type")
	}

	if withExt {
		cols = append(cols, "Extensions")
	}

	rows = make([][]string, 0, len(m))

	ucFirst := func(str string) string {
		for i, v := range str {
			return string(unicode.ToUpper(v)) + str[i+1:]
		}
		return ""
	}

	for f, ext := range m {
		sort.Slice(ext, func(i, j int) bool {
			return ext[i] < ext[j]
		})

		v := make([]string, 0, 4)
		v = append(v, strings.ToUpper(f.String()))

		if withDesc {
			v = append(v, FormatDesc[f])
		}

		if withType {
			v = append(v, ucFirst(string(MediaTypes[f])))
		}

		if withExt {
			v = append(v, strings.Join(ext, ", "))
		}

		rows = append(rows, v)
	}

	sort.Slice(rows, func(i, j int) bool {
		if t > 0 && rows[i][t] == rows[j][t] {
			return rows[i][0] < rows[j][0]
		} else {
			return rows[i][t] < rows[j][t]
		}
	})

	return rows, cols
}
