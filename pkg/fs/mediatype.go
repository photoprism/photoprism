package fs

type MediaType string

const (
	MediaImage   MediaType = "image"
	MediaSidecar MediaType = "sidecar"
	MediaRaw     MediaType = "raw"
	MediaVideo   MediaType = "video"
	MediaOther   MediaType = "other"
)

var MediaTypes = map[FileFormat]MediaType{
	FormatRaw:      MediaRaw,
	FormatJpeg:     MediaImage,
	FormatPng:      MediaImage,
	FormatGif:      MediaImage,
	FormatTiff:     MediaImage,
	FormatBitmap:   MediaImage,
	FormatHEIF:     MediaImage,
	FormatMpo:      MediaImage,
	FormatAvi:      MediaVideo,
	FormatHEVC:     MediaVideo,
	FormatAvc:      MediaVideo,
	FormatMp4:      MediaVideo,
	FormatMov:      MediaVideo,
	Format3gp:      MediaVideo,
	Format3g2:      MediaVideo,
	FormatFlv:      MediaVideo,
	FormatMkv:      MediaVideo,
	FormatMpg:      MediaVideo,
	FormatMts:      MediaVideo,
	FormatOgv:      MediaVideo,
	FormatWebm:     MediaVideo,
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

	result, ok := MediaTypes[GetFileFormat(fileName)]

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
