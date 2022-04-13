package fs

// MediaType represents a general media type.
type MediaType string

// General categories of media file types.
const (
	MediaImage   MediaType = "image"
	MediaSidecar MediaType = "sidecar"
	MediaRaw     MediaType = "raw"
	MediaVideo   MediaType = "video"
	MediaVector  MediaType = "vector"
	MediaOther   MediaType = "other"
)

// String returns the media name as string.
func (m MediaType) String() string {
	return string(m)
}

// MediaTypes maps file formats to general media types.
var MediaTypes = map[Format]MediaType{
	FormatRaw:      MediaRaw,
	FormatJpeg:     MediaImage,
	FormatPng:      MediaImage,
	FormatGif:      MediaImage,
	FormatTiff:     MediaImage,
	FormatBitmap:   MediaImage,
	FormatMpo:      MediaImage,
	FormatHEIF:     MediaImage,
	FormatHEVC:     MediaVideo,
	FormatWebP:     MediaImage,
	FormatWebM:     MediaVideo,
	FormatAvi:      MediaVideo,
	FormatAVC:      MediaVideo,
	FormatAV1:      MediaVideo,
	FormatMPEG:     MediaVideo,
	FormatMJPEG:    MediaVideo,
	FormatMp2:      MediaVideo,
	FormatMp4:      MediaVideo,
	FormatMkv:      MediaVideo,
	FormatMov:      MediaVideo,
	Format3gp:      MediaVideo,
	Format3g2:      MediaVideo,
	FormatFlv:      MediaVideo,
	FormatMts:      MediaVideo,
	FormatOgv:      MediaVideo,
	FormatWMV:      MediaVideo,
	FormatXMP:      MediaSidecar,
	FormatXML:      MediaSidecar,
	FormatAAE:      MediaSidecar,
	FormatYaml:     MediaSidecar,
	FormatText:     MediaSidecar,
	FormatJson:     MediaSidecar,
	FormatToml:     MediaSidecar,
	FormatMarkdown: MediaSidecar,
	FormatOther:    MediaOther,
}

func GetMediaType(fileName string) MediaType {
	if fileName == "" {
		return MediaOther
	}

	result, ok := MediaTypes[FileFormat(fileName)]

	if !ok {
		result = MediaOther
	}

	return result
}

func IsMedia(fileName string) bool {
	switch GetMediaType(fileName) {
	case MediaRaw, MediaImage, MediaVideo:
		return true
	default:
		return false
	}
}
