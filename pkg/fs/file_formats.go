package fs

import (
	"fmt"
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
	FormatMov      Format = "mov"  // QuickTime File Format, can contain AVC, HEVC,...
	FormatMp2      Format = "mp2"  // MPEG-2, H.222/H.262
	FormatMp4      Format = "mp4"  // MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	FormatAvi      Format = "avi"  // Microsoft Audio Video Interleave (AVI)
	Format3gp      Format = "3gp"  // Mobile Multimedia Container, MPEG-4 Part 12
	Format3g2      Format = "3g2"  // Similar to 3GP, consumes less space & bandwidth
	FormatFlv      Format = "flv"  // Flash Video
	FormatMkv      Format = "mkv"  // Matroska Multimedia Container, free and open
	FormatMpg      Format = "mpg"  // Moving Picture Experts Group (MPEG)
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
	FormatJpeg:     "JPEG (Joint Photographic Experts Group)",
	FormatPng:      "Portable Network Graphics",
	FormatGif:      "Graphics Interchange Format",
	FormatTiff:     "Tag Image File Format",
	FormatBitmap:   "Bitmap",
	FormatRaw:      "Unprocessed Image Data",
	FormatMpo:      "Stereoscopic 3D Format based on JPEG",
	FormatHEIF:     "High Efficiency Image File Format",
	FormatWebP:     "Google WebP",
	FormatWebM:     "Google WebM",
	FormatHEVC:     "H.265, High Efficiency Video Coding",
	FormatAVC:      "H.264, Advanced Video Coding, MPEG-4 Part 10",
	FormatAV1:      "Alliance for Open Media",
	FormatMov:      "QuickTime File Format",
	FormatMp2:      "MPEG-2, H.222, H.262",
	FormatMp4:      "MPEG-4 Part 14 Multimedia Container",
	FormatAvi:      "Microsoft Audio Video Interleave",
	Format3gp:      "MPEG-4 Part 12, Mobile Multimedia Container",
	Format3g2:      "Multimedia Container for 3G CDMA2000, based on 3GP",
	FormatFlv:      "Flash Video",
	FormatMkv:      "Matroska Multimedia Container",
	FormatMpg:      "MPEG (Moving Picture Experts Group)",
	FormatMts:      "AVCHD (Advanced Video Coding High Definition)",
	FormatOgv:      "Ogg Media by Xiph.Org",
	FormatWMV:      "Windows Media",
	FormatXMP:      "Adobe Extensible Metadata Platform",
	FormatAAE:      "Apple Image Edits",
	FormatXML:      "Extensible Markup Language",
	FormatJson:     "Serialized JSON Data (Exiftool, Google Photos)",
	FormatYaml:     "Serialized YAML Data (Metadata, Config Values)",
	FormatToml:     "Serialized TOML Data (Tom's Obvious, Minimal Language)",
	FormatText:     "Plain Text",
	FormatMarkdown: "Markdown Formatted Text",
	FormatOther:    "Other",
}

// Markdown returns a file format table in markdown text format.
func (m FileFormats) Markdown() string {

	results := make([][]string, 0, len(m))

	max := func(x, y int) int {
		if x < y {
			return y
		}

		return x
	}

	ucFirst := func(str string) string {
		for i, v := range str {
			return string(unicode.ToUpper(v)) + str[i+1:]
		}
		return ""
	}

	l0, l1, l2, l3 := 12, 12, 12, 12

	for f, ext := range m {
		sort.Slice(ext, func(i, j int) bool {
			return ext[i] < ext[j]
		})

		v := make([]string, 4)
		v[0] = strings.ToUpper(f.String())
		v[1] = FormatDesc[f]
		v[2] = ucFirst(string(MediaTypes[f]))
		v[3] = strings.Join(ext, ", ")
		l0, l1, l2, l3 = max(l0, len(v[0])), max(l1, len(v[1])), max(l2, len(v[2])), max(l3, len(v[3]))
		results = append(results, v)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i][2] == results[j][2] {
			return results[i][0] < results[j][0]
		} else {
			return results[i][2] < results[j][2]
		}
	})

	rows := make([]string, len(results)+2)

	cols := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds | %%-%ds |\n", l0, l1, l2, l3)

	rows = append(rows, fmt.Sprintf(cols, "Format", "Description", "Type", "Extensions"))
	rows = append(rows, fmt.Sprintf("|:%s-|:%s-|:%s-|:%s-|\n", strings.Repeat("-", l0), strings.Repeat("-", l1), strings.Repeat("-", l2), strings.Repeat("-", l3)))

	for _, r := range results {
		rows = append(rows, fmt.Sprintf(cols, r[0], r[1], r[2], r[3]))
	}

	return strings.Join(rows, "")
}
