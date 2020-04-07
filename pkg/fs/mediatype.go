package fs

type MediaType string

const (
	MediaRaw     MediaType = "raw"
	MediaImage   MediaType = "image"
	MediaVideo   MediaType = "video"
	MediaSidecar MediaType = "sidecar"
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
	TypeMP4:      MediaVideo,
	TypeMov:      MediaVideo,
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
	result, ok := MediaTypes[GetFileType(fileName)]

	if !ok {
		result = MediaOther
	}

	return result
}
