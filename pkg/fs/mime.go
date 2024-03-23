package fs

import (
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

const (
	MimeTypeUnknown = ""
	MimeTypeBinary  = "application/octet-stream"
	MimeTypeJPEG    = "image/jpeg"
	MimeTypeJPEGXL  = "image/jxl"
	MimeTypePNG     = "image/png"
	MimeTypeAPNG    = "image/vnd.mozilla.apng"
	MimeTypeGIF     = "image/gif"
	MimeTypeBMP     = "image/bmp"
	MimeTypeTIFF    = "image/tiff"
	MimeTypeDNG     = "image/dng"
	MimeTypeAVIF    = "image/avif"
	MimeTypeAVIFS   = "image/avif-sequence"
	MimeTypeHEIC    = "image/heic"
	MimeTypeHEICS   = "image/heic-sequence"
	MimeTypeWebP    = "image/webp"
	MimeTypeMP4     = "video/mp4"
	MimeTypeMOV     = "video/quicktime"
	MimeTypeSVG     = "image/svg+xml"
	MimeTypeAI      = "application/vnd.adobe.illustrator"
	MimeTypePS      = "application/postscript"
	MimeTypeEPS     = "image/eps"
	MimeTypeXML     = "text/xml"
	MimeTypeJSON    = "application/json"
)

// MimeType returns the mime type of a file, or an empty string if it could not be detected.
func MimeType(filename string) (mimeType string) {
	if filename == "" {
		return MimeTypeUnknown
	}

	// Detect file type based on the filename extension.
	fileType := Extensions[strings.ToLower(filepath.Ext(filename))]

	// Determine mime type based on the extension for the following
	// formats, which otherwise cannot be reliably distinguished:
	switch fileType {
	// Apple QuickTime Container
	case VideoMOV:
		return MimeTypeMOV
	// Adobe Digital Negative
	case ImageDNG:
		return MimeTypeDNG
	// Adobe PostScript
	case VectorPS:
		return MimeTypePS
	// Adobe Embedded PostScript
	case VectorEPS:
		return MimeTypeEPS
	// Adobe Illustrator
	case VectorAI:
		return MimeTypeAI
	// Scalable Vector Graphics
	case VectorSVG:
		return MimeTypeSVG
	}

	// Detect mime type based on the file content.
	detectedType, err := mimetype.DetectFile(filename)

	if detectedType != nil && err == nil {
		mimeType, _, _ = strings.Cut(detectedType.String(), ";")
	}

	// Treat "application/octet-stream" as unknown.
	if mimeType == MimeTypeBinary {
		mimeType = MimeTypeUnknown
	}

	// If it could be detected, try to determine mime type from extension:
	if mimeType == MimeTypeUnknown {
		switch fileType {
		// MP4 Multimedia Container
		case VideoMP4:
			return MimeTypeMP4
		// AV1 Image File
		case ImageAVIF:
			return MimeTypeAVIF
		// AV1 Image File Sequence
		case ImageAVIFS:
			return MimeTypeAVIFS
		// High Efficiency Image Container
		case ImageHEIC:
			return MimeTypeHEIC
		// High Efficiency Image Container Sequence
		case ImageHEICS:
			return MimeTypeHEICS
		}
	}

	return mimeType
}
