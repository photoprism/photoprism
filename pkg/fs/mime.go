package fs

import (
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

const (
	MimeTypeUnknown = ""
	MimeTypeJpeg    = "image/jpeg"
	MimeTypePng     = "image/png"
	MimeTypeGif     = "image/gif"
	MimeTypeBitmap  = "image/bmp"
	MimeTypeTiff    = "image/tiff"
	MimeTypeDNG     = "image/dng"
	MimeTypeAVIF    = "image/avif"
	MimeTypeHEIC    = "image/heic"
	MimeTypeWebP    = "image/webp"
	MimeTypeXML     = "text/xml"
	MimeTypeJSON    = "application/json"
)

// Set mime detection read limit.
func init() {
	mimetype.SetLimit(1024)
}

// MimeType returns the mime type of a file, or an empty string if it could not be detected.
func MimeType(filename string) (mimeType string) {
	// Workaround, since "image/dng" cannot be recognized yet.
	if ext := strings.ToLower(filepath.Ext(filename)); ext == "" {
		// Continue.
	} else if Extensions[ext] == ImageDNG {
		return MimeTypeDNG
	} else if Extensions[ext] == ImageAVIF {
		return MimeTypeAVIF
	}

	if t, err := mimetype.DetectFile(filename); err != nil {
		return MimeTypeUnknown
	} else {
		mimeType, _, _ = strings.Cut(t.String(), ";")
	}

	return mimeType
}
