package fs

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/pkg/http/header"
)

const (
	MimeTypeUnknown = ""
)

// DetectMimeType returns the MIME type of the specified file,
// or an error if the type could not be detected.
//
// The IANA and IETF use the term "media type", and consider the term "MIME type" to be obsolete,
// since media types have become used in contexts unrelated to email, such as HTTP:
// https://en.wikipedia.org/wiki/Media_type#Structure
func DetectMimeType(filename string) (mimeType string, err error) {
	// Abort if no filename was specified.
	if filename == "" {
		return MimeTypeUnknown, errors.New("missing filename")
	}

	// Detect file type based on the filename extension.
	fileType := Extensions[strings.ToLower(filepath.Ext(filename))]

	// Determine mime type based on the extension for the following
	// formats, which otherwise cannot be reliably distinguished:
	switch fileType {
	// MPEG-2 Transport Stream
	case VideoM2TS, VideoAVCHD:
		return header.ContentTypeM2TS, nil
	// Apple QuickTime Container
	case VideoMov:
		return header.ContentTypeMov, nil
	// MPEG-4 AVC Video
	case VideoAvc:
		return header.ContentTypeMp4Avc, nil
	// MPEG-4 HEVC Video
	case VideoHvc:
		return header.ContentTypeMp4Hvc, nil
	// MPEG-4 HEVC Bitstream
	case VideoHev:
		return header.ContentTypeMp4Hev, nil
	// Adobe Digital Negative
	case ImageDng:
		return header.ContentTypeDng, nil
	// Adobe Illustrator
	case VectorAI:
		return header.ContentTypeAI, nil
	// Adobe PostScript
	case VectorPS:
		return header.ContentTypePS, nil
	// Adobe Embedded PostScript
	case VectorEPS:
		return header.ContentTypeEPS, nil
	// Scalable Vector Graphics
	case VectorSVG:
		return header.ContentTypeSVG, nil
	}

	// Use "gabriel-vasile/mimetype" to automatically detect the MIME type.
	detectedType, err := mimetype.DetectFile(filename)

	// Check if type could be successfully detected.
	if err == nil {
		if detectedType != nil {
			mimeType = detectedType.String()
		}
	} else if e := err.Error(); strings.HasSuffix(e, ErrPermissionDenied.Error()) {
		return MimeTypeUnknown, ErrPermissionDenied
	} else if strings.Contains(e, EOF.Error()) {
		return MimeTypeUnknown, ErrUnexpectedEOF
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
			return header.ContentTypeMp4, nil
		// AV1 Image File
		case ImageAvif:
			return header.ContentTypeAvif, nil
		// AV1 Image File Sequence
		case ImageAvifS:
			return header.ContentTypeAvifS, nil
		// High Efficiency Image Container
		case ImageHeic, ImageHeif:
			return header.ContentTypeHeic, nil
		// High Efficiency Image Container Sequence
		case ImageHeicS:
			return header.ContentTypeHeicS, nil
		// ZIP Archive File:
		case ArchiveZip:
			return header.ContentTypeZip, nil
		}
	}

	return mimeType, err
}

// MimeType returns the MIME type of the specified file,
// or an empty string if the type could not be detected.
func MimeType(filename string) (mimeType string) {
	mimeType, _ = DetectMimeType(filename)
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
