package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// Type represents a video format type.
type Type struct {
	File   fs.Type
	Codec  Codec
	Width  int
	Height int
	Public bool
}
