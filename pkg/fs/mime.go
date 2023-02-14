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
	MimeTypeMP4     = "video/mp4"
	MimeTypeMOV     = "video/quicktime"
	MimeTypeSVG     = "image/svg+xml"
	MimeTypeAI      = "application/vnd.adobe.illustrator"
	MimeTypePS      = "application/ps"
	MimeTypeEPS     = "image/eps"
	MimeTypeXML     = "text/xml"
	MimeTypeJSON    = "application/json"
)

// MimeType returns the mime type of a file, or an empty string if it could not be detected.
func MimeType(filename string) (mimeType string) {
	if filename == "" {
		return MimeTypeUnknown
	}

	// Workaround for types that cannot be reliably detected.
	switch Extensions[strings.ToLower(filepath.Ext(filename))] {
	case ImageDNG:
		return MimeTypeDNG
	case MimeTypeAVIF:
		return MimeTypeAVIF
	case VideoMP4:
		return MimeTypeMP4
	case MimeTypeMOV:
		return MimeTypeMOV
	case VectorSVG:
		return MimeTypeSVG
	case VectorAI:
		return MimeTypeAI
	case VectorPS:
		return MimeTypePS
	case VectorEPS:
		return MimeTypeEPS
	}

	if t, err := mimetype.DetectFile(filename); err != nil {
		return MimeTypeUnknown
	} else {
		mimeType, _, _ = strings.Cut(t.String(), ";")
	}

	return mimeType
}
