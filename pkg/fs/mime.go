package fs

import (
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/pkg/media/http/header"
)

const (
	MimeTypeUnknown = ""
)

// MimeType returns the mimetype of a file, or an empty string if it could not be determined.
//
// The IANA and IETF use the term "media type", and consider the term "MIME type" to be obsolete,
// since media types have become used in contexts unrelated to email, such as HTTP:
// https://en.wikipedia.org/wiki/Media_type#Structure
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
	case VideoMov:
		return header.ContentTypeMov
	// MPEG-4 AVC Video
	case VideoAvc:
		return header.ContentTypeMp4Avc
	// MPEG-4 HEVC Video
	case VideoHvc:
		return header.ContentTypeMp4Hvc
	// MPEG-4 HEVC Bitstream
	case VideoHev:
		return header.ContentTypeMp4Hev
	// Adobe Digital Negative
	case ImageDng:
		return header.ContentTypeDng
	// Adobe Illustrator
	case VectorAI:
		return header.ContentTypeAI
	// Adobe PostScript
	case VectorPS:
		return header.ContentTypePS
	// Adobe Embedded PostScript
	case VectorEPS:
		return header.ContentTypeEPS
	// Adobe PDF
	case DocumentPDF:
		return header.ContentTypePDF
	// Scalable Vector Graphics
	case VectorSVG:
		return header.ContentTypeSVG
	}

	// Detect mime type based on the file content.
	detectedType, err := mimetype.DetectFile(filename)

	if detectedType != nil && err == nil {
		mimeType = detectedType.String()
	}

	// Treat "application/octet-stream" as unknown.
	if mimeType == header.ContentTypeBinary {
		mimeType = MimeTypeUnknown
	}

	// If it could be detected, try to determine mime type from extension:
	if mimeType == MimeTypeUnknown {
		switch fileType {
		// MPEG-4 Multimedia Container
		case VideoMp4:
			return header.ContentTypeMp4
		// AV1 Image File
		case ImageAvif:
			return header.ContentTypeAvif
		// AV1 Image File Sequence
		case ImageAvifS:
			return header.ContentTypeAvifS
		// High Efficiency Image Container
		case ImageHeic, ImageHeif:
			return header.ContentTypeHeic
		// High Efficiency Image Container Sequence
		case ImageHeicS:
			return header.ContentTypeHeicS
		}
	}

	return mimeType
}

// BaseType returns the media type string without any optional parameters.
func BaseType(mimeType string) string {
	if mimeType == "" {
		return ""
	}

	mimeType, _, _ = strings.Cut(mimeType, ";")

	return strings.ToLower(mimeType)
}

// SameType tests if the specified mime types are matching, except for any optional parameters.
func SameType(mime1, mime2 string) bool {
	if mime1 == mime2 {
		return true
	}

	return BaseType(mime1) == BaseType(mime2)
}
