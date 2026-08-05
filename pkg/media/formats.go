package media

import (
	"sort"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Formats maps file formats to general media types.
var Formats = map[fs.Type]Type{
	fs.ArchiveZip:      Archive,
	fs.DocumentPDF:     Document,
	fs.ImageJpeg:       Image,
	fs.ImageJpegXL:     Image,
	fs.ImageThumb:      Image,
	fs.ImagePng:        Image,
	fs.ImageGif:        Image,
	fs.ImageTiff:       Image,
	fs.ImagePsd:        Image,
	fs.ImageBmp:        Image,
	fs.ImageMPO:        Image,
	fs.ImageAvif:       Image,
	fs.ImageAvifS:      Image,
	fs.ImageHeif:       Image,
	fs.ImageHeic:       Image,
	fs.ImageHeicS:      Image,
	fs.ImageWebp:       Image,
	fs.ImageInsp:       Image,
	fs.ImageRaw:        Raw,
	fs.ImageDng:        Raw,
	fs.SidecarXMP:      Sidecar,
	fs.SidecarXml:      Sidecar,
	fs.SidecarAppleXml: Sidecar,
	fs.SidecarYaml:     Sidecar,
	fs.SidecarJson:     Sidecar,
	fs.SidecarText:     Sidecar,
	fs.SidecarInfo:     Sidecar,
	fs.SidecarMarkdown: Sidecar,
	fs.VectorSVG:       Vector,
	fs.VectorAI:        Vector,
	fs.VectorPS:        Vector,
	fs.VectorEPS:       Vector,
	fs.VideoWebm:       Video,
	fs.VideoAvc:        Video,
	fs.VideoHvc:        Video,
	fs.VideoHev:        Video,
	fs.VideoVvc:        Video,
	fs.VideoEvc:        Video,
	fs.VideoAVI:        Video,
	fs.VideoAv1:        Video,
	fs.VideoVp8:        Video,
	fs.VideoVp9:        Video,
	fs.VideoMpeg:       Video,
	fs.VideoMjpeg:      Video,
	fs.VideoMp2:        Video,
	fs.VideoMp4:        Video,
	fs.VideoM4v:        Video,
	fs.VideoMkv:        Video,
	fs.VideoMov:        Video,
	fs.VideoMXF:        Video,
	fs.Video3GP:        Video,
	fs.Video3G2:        Video,
	fs.VideoFlash:      Video,
	fs.VideoM2TS:       Video,
	fs.VideoAVCHD:      Video,
	fs.VideoTheora:     Video,
	fs.VideoASF:        Video,
	fs.VideoWMV:        Video,
	fs.VideoDV:         Video,
	fs.VideoInsv:       Video,
	fs.TypeUnknown:     Sidecar,
}

// FileTypes returns the file types that belong to the specified media type, sorted by name.
// A media type groups several file types, so Raw covers both fs.ImageRaw and fs.ImageDng.
func FileTypes(t Type) []fs.Type {
	result := make([]fs.Type, 0, len(Formats))

	for fileType, mediaType := range Formats {
		if mediaType == t {
			result = append(result, fileType)
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })

	return result
}

// FileTypeStrings returns the file types that belong to the specified media type as strings.
func FileTypeStrings(t Type) []string {
	fileTypes := FileTypes(t)
	result := make([]string, 0, len(fileTypes))

	for _, fileType := range fileTypes {
		result = append(result, string(fileType))
	}

	return result
}
