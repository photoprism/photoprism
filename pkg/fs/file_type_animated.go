package fs

import (
	"github.com/photoprism/photoprism/pkg/media/http/header"
)

// TypeAnimated maps animated file types to their mime type.
var TypeAnimated = TypeMap{
	ImageGif:   header.ContentTypeGif,
	ImagePng:   header.ContentTypeAPng,
	ImageWebp:  header.ContentTypeWebp,
	ImageAvif:  header.ContentTypeAvifS,
	ImageAvifS: header.ContentTypeAvifS,
	ImageHeic:  header.ContentTypeHeicS,
	ImageHeicS: header.ContentTypeHeicS,
}
