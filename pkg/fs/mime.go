package fs

import (
	"os"
	"path/filepath"

	"github.com/h2non/filetype"
)

const (
	MimeTypeJpeg   = "image/jpeg"
	MimeTypePng    = "image/png"
	MimeTypeGif    = "image/gif"
	MimeTypeBitmap = "image/bmp"
	MimeTypeWebP   = "image/webp"
	MimeTypeTiff   = "image/tiff"
	MimeTypeAVIF   = "image/avif"
	MimeTypeHEIF   = "image/heif"
)

// MimeType returns the mime type of a file, an empty string if it is unknown.
func MimeType(filename string) string {
	// Workaround, since "image/avif " cannot be recognized yet.
	if Extensions[filepath.Ext(filename)] == ImageAVIF {
		return MimeTypeAVIF
	}

	handle, err := os.Open(filename)

	if err != nil {
		return ""
	}

	defer handle.Close()

	// Only the first 261 bytes are used to sniff the content type.
	buffer := make([]byte, 261)

	if _, err := handle.Read(buffer); err != nil {
		return ""
	} else if t, err := filetype.Get(buffer); err == nil && t != filetype.Unknown {
		return t.MIME.Value
	} else if t := filetype.GetType(NormalizedExt(filename)); t != filetype.Unknown {
		return t.MIME.Value
	} else {
		return ""
	}
}
