package fs

type MediaType string

const (
	MediaImage   MediaType = "image"
	MediaSidecar MediaType = "sidecar"
	MediaRaw     MediaType = "raw"
	MediaVideo   MediaType = "video"
	MediaOther   MediaType = "other"
)

var MediaTypes = map[FileType]MediaType{
	TypeRaw:      MediaRaw,
	TypeJpeg:     MediaImage,
	TypePng:      MediaImage,
	TypeGif:      MediaImage,
	TypeTiff:     MediaImage,
	TypeBitmap:   MediaImage,
	TypeHEIF:     MediaImage,
	TypeAvi:      MediaVideo,
	TypeHEVC:     MediaVideo,
	TypeMp4:      MediaVideo,
	TypeMov:      MediaVideo,
	Type3gp:      MediaVideo,
	Type3g2:      MediaVideo,
	TypeFlv:      MediaVideo,
	TypeMkv:      MediaVideo,
	TypeMpg:      MediaVideo,
	TypeOgv:      MediaVideo,
	TypeWebm:     MediaVideo,
	TypeWMV:      MediaVideo,
	TypeXMP:      MediaSidecar,
	TypeXML:      MediaSidecar,
	TypeAAE:      MediaSidecar,
	TypeYaml:     MediaSidecar,
	TypeText:     MediaSidecar,
	TypeJson:     MediaSidecar,
	TypeToml:     MediaSidecar,
	TypeMarkdown: MediaSidecar,
	TypeOther:    MediaOther,
}

func GetMediaType(fileName string) MediaType {
	if fileName == "" {
		return MediaOther
	}

	result, ok := MediaTypes[GetFileType(fileName)]

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
