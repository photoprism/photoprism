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
	MimeTypeEPS     = "image/eps"
	MimeTypeXML     = "text/xml"
	MimeTypeJSON    = "application/json"
)

// MimeType returns the mime type of a file, or an empty string if it could not be detected.
func MimeType(filename string) (mimeType string) {
	// Workaround, since "image/dng" cannot be recognized yet.
	if ext := Extensions[strings.ToLower(filepath.Ext(filename))]; ext == "" {
		// Continue.
	} else if ext == ImageDNG {
		return MimeTypeDNG
	} else if ext == ImageAVIF {
		return MimeTypeAVIF
	} else if ext == VideoMP4 {
		return MimeTypeMP4
	} else if ext == VideoMOV {
		return MimeTypeMOV
	} else if ext == VectorSVG {
		return MimeTypeSVG
	} else if ext == VectorEPS {
		return MimeTypeEPS
	}

	if t, err := mimetype.DetectFile(filename); err != nil {
		return MimeTypeUnknown
	} else {
		mimeType, _, _ = strings.Cut(t.String(), ";")
	}

	return mimeType
}
