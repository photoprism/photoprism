package video

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Standard content types.
const (
	ContentTypeDefault = "application/octet-stream"
	ContentTypeAVC     = fs.MimeTypeMP4 + `; codecs="avc1"`
	ContentTypeMOV     = fs.MimeTypeMOV
)

// ContentType composes the video content type from the given mime type and codec.
func ContentType(mimeType string, codec Codec) string {
	if codec == CodecUnknown {
		return mimeType
	} else if mimeType == "" {
		return ContentTypeDefault
	}

	return fmt.Sprintf("%s; codecs=\"%s\"", mimeType, codec)
}
